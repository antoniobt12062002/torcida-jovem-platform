package httpapi

import (
	"net/http"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

func loadIdentity(t *testing.T) *openapi3.T {
	t.Helper()
	doc, err := httpx.LoadContract(filepath.Join("..", "..", "openapi", "identity.yaml"))
	if err != nil {
		t.Fatalf("o contrato de identidade deveria ser válido: %v", err)
	}
	return doc
}

type operation struct {
	key    string
	method string
	path   string
	op     *openapi3.Operation
}

func operations(doc *openapi3.T) []operation {
	var out []operation
	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			out = append(out, operation{key: method + " " + path, method: method, path: path, op: op})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key < out[j].key })
	return out
}

func isPublic(op *openapi3.Operation) bool { return op.Security != nil && len(*op.Security) == 0 }

// API-01.1: exatamente as operações aprovadas; nada de DELETE no vínculo nem GET /roles.
func TestIdentityContractHasExactlyTheApprovedOperations(t *testing.T) {
	want := []string{
		"GET /api/v1/auth/me", "GET /api/v1/users",
		"POST /api/v1/auth/login", "POST /api/v1/auth/logout", "POST /api/v1/auth/password",
		"POST /api/v1/auth/password-reset/confirm", "POST /api/v1/auth/password-reset/request",
		"POST /api/v1/users", "POST /api/v1/users/{id}/admin-membership", "POST /api/v1/users/{id}/admin-membership/revoke",
		"POST /api/v1/users/{id}/deactivate", "POST /api/v1/users/{id}/password-reset", "POST /api/v1/users/{id}/reactivate",
		"PUT /api/v1/users/{id}/roles",
	}
	var got []string
	for _, o := range operations(loadIdentity(t)) {
		got = append(got, o.key)
	}
	if !slices.Equal(got, want) {
		t.Errorf("operações = %v\nesperado  = %v", got, want)
	}
	for _, o := range got {
		if strings.HasPrefix(o, "DELETE ") || strings.Contains(o, "/roles") && strings.HasPrefix(o, "GET ") {
			t.Errorf("operação que não faz parte da fundação: %s", o)
		}
	}
}

// Só login e as duas de recuperação são públicas; as demais exigem o cookie de sessão.
func TestOnlyLoginAndPasswordRecoveryArePublic(t *testing.T) {
	doc := loadIdentity(t)
	var public []string
	for _, o := range operations(doc) {
		if isPublic(o.op) {
			public = append(public, o.key)
			continue
		}
		reqs := doc.Security
		if o.op.Security != nil {
			reqs = *o.op.Security
		}
		if len(reqs) != 1 || reqs[0]["sessionCookie"] == nil {
			t.Errorf("%s deveria exigir o cookie de sessão: %v", o.key, reqs)
		}
	}
	want := []string{"POST /api/v1/auth/login", "POST /api/v1/auth/password-reset/confirm", "POST /api/v1/auth/password-reset/request"}
	if !slices.Equal(public, want) {
		t.Errorf("rotas públicas = %v, esperado %v", public, want)
	}
}

// IDN-03.1: as operações autenticadas que alteram estado descrevem o X-CSRF-Token.
func TestStateChangingAuthenticatedOperationsDeclareTheCSRFHeader(t *testing.T) {
	for _, o := range operations(loadIdentity(t)) {
		stateChanging := o.method != http.MethodGet
		hasCSRF := false
		for _, p := range o.op.Parameters {
			if p.Value.In == "header" && p.Value.Name == "X-CSRF-Token" && p.Value.Required {
				hasCSRF = true
			}
		}
		if want := stateChanging && !isPublic(o.op); hasCSRF != want {
			t.Errorf("%s: X-CSRF-Token declarado = %v, esperado %v", o.key, hasCSRF, want)
		}
	}
}

// API-02: toda operação descreve erros em problem+json; os 429 dizem quando tentar de novo.
func TestEveryOperationDescribesProblemJSONErrorsAndRateLimitsSayRetryAfter(t *testing.T) {
	for _, o := range operations(loadIdentity(t)) {
		problems := 0
		for code, ref := range o.op.Responses.Map() {
			if ref.Value.Content.Get("application/problem+json") != nil {
				problems++
			}
			if code == "429" && ref.Value.Headers["Retry-After"] == nil {
				t.Errorf("%s: o 429 deveria declarar Retry-After", o.key)
			}
		}
		if problems == 0 {
			t.Errorf("%s não descreve nenhuma resposta de erro problem+json", o.key)
		}
	}
}

// IDN-04.8: a listagem tem cursor, limite e filtros, e devolve next_cursor.
func TestUserListingHasCursorLimitFiltersAndNextCursor(t *testing.T) {
	doc := loadIdentity(t)
	op := doc.Paths.Find("/api/v1/users").Get
	var names []string
	for _, p := range op.Parameters {
		names = append(names, p.Value.Name)
	}
	slices.Sort(names)
	if !slices.Equal(names, []string{"active", "cursor", "limit", "role"}) {
		t.Errorf("parâmetros = %v", names)
	}
	page := op.Responses.Value("200").Value.Content.Get("application/json").Schema.Value
	if page.Properties["next_cursor"] == nil || !page.Properties["next_cursor"].Value.Nullable {
		t.Error("a página deveria ter next_cursor anulável")
	}
}

func props(s *openapi3.Schema) []string {
	var out []string
	for name := range s.Properties {
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}

// IDN-02.1 e .9: login e me devolvem o mesmo contexto, sem dado sensível.
func TestLoginAndMeReturnTheSameAuthContextWithoutSensitiveData(t *testing.T) {
	doc := loadIdentity(t)
	login := doc.Paths.Find("/api/v1/auth/login").Post.Responses.Value("200").Value.Content.Get("application/json").Schema
	me := doc.Paths.Find("/api/v1/auth/me").Get.Responses.Value("200").Value.Content.Get("application/json").Schema

	if login.Ref != me.Ref || !strings.HasSuffix(login.Ref, "AuthContext") {
		t.Errorf("login = %q, me = %q: deveriam usar o mesmo AuthContext", login.Ref, me.Ref)
	}
	ctx := me.Value
	want := []string{"admin_membership", "csrf_token", "must_change_password", "permissions", "roles", "user"}
	if !slices.Equal(props(ctx), want) {
		t.Errorf("AuthContext = %v", props(ctx))
	}
	if !slices.Equal(props(ctx.Properties["user"].Value), []string{"email", "id", "name"}) {
		t.Errorf("user = %v", props(ctx.Properties["user"].Value))
	}
	m := ctx.Properties["admin_membership"].Value
	if !m.Nullable || !slices.Equal(props(m), []string{"granted_at", "reason"}) {
		t.Errorf("admin_membership deveria ser nulo ou só {reason, granted_at}: nullable = %v, %v", m.Nullable, props(m))
	}
	if !slices.Contains(ctx.Required, "admin_membership") {
		t.Error("admin_membership é obrigatório (pode ser null)")
	}
}

func TestNoResponseSchemaCarriesHashesOrInternalTokens(t *testing.T) {
	doc := loadIdentity(t)
	forbidden := []string{"password_hash", "hash", "token_hash", "session_token", "granted_by", "revoked_by", "history"}
	for name, ref := range doc.Components.Schemas {
		for _, prop := range props(ref.Value) {
			if slices.Contains(forbidden, prop) {
				t.Errorf("o esquema %s expõe %s", name, prop)
			}
		}
	}
	tmp := doc.Components.Schemas["TemporaryPassword"].Value
	if !slices.Equal(props(tmp), []string{"temporary_password"}) {
		t.Errorf("TemporaryPassword = %v", props(tmp))
	}
}

// O reset administrativo devolve a senha temporária uma vez, sem cache.
func TestAdministrativeResetResponseIsNoStore(t *testing.T) {
	resp := loadIdentity(t).Paths.Find("/api/v1/users/{id}/password-reset").Post.Responses.Value("200").Value
	h := resp.Headers["Cache-Control"]
	if h == nil || len(h.Value.Schema.Value.Enum) != 1 || h.Value.Schema.Value.Enum[0] != "no-store" {
		t.Error("a resposta 200 deveria declarar Cache-Control: no-store")
	}
}

// O corpo da retirada carrega o motivo; não é uma exclusão.
func TestRevokeAdminMembershipCarriesTheReasonInTheBody(t *testing.T) {
	op := loadIdentity(t).Paths.Find("/api/v1/users/{id}/admin-membership/revoke").Post
	body := op.RequestBody.Value.Content.Get("application/json").Schema.Value
	if !slices.Equal(body.Required, []string{"reason"}) {
		t.Errorf("required = %v", body.Required)
	}
}

// O enum de papéis do contrato é o mesmo dos papéis do domínio.
func TestRoleEnumMatchesTheDomainRoles(t *testing.T) {
	var got []string
	for _, v := range loadIdentity(t).Components.Schemas["Role"].Value.Enum {
		got = append(got, v.(string))
	}
	var want []string
	for _, r := range domain.AllRoles {
		want = append(want, string(r))
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("enum = %v, papéis = %v", got, want)
	}
}

// O login não distingue e-mail inexistente de senha errada: nenhum formato de e-mail no corpo.
func TestLoginAndRecoveryRequestsDoNotValidateTheEmailFormat(t *testing.T) {
	doc := loadIdentity(t)
	for _, name := range []string{"LoginRequest", "PasswordResetRequest"} {
		email := doc.Components.Schemas[name].Value.Properties["email"].Value
		if email.Format != "" {
			t.Errorf("%s.email não pode ter format %q: um 422 revelaria o que é e-mail válido", name, email.Format)
		}
	}
}
