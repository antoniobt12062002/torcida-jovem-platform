//go:build integration

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email/emailtest"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/storage"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

const (
	firstPassword  = "senha-inicial-do-admin-1"
	secondPassword = "segunda-senha-do-admin-2"
)

// e2e is the whole API, over a real database, driven only through HTTP. Every
// response is checked against the OpenAPI contracts.
type e2e struct {
	t        *testing.T
	router   http.Handler
	mod      *identity.Module
	mail     *emailtest.Recorder
	logs     *bytes.Buffer
	contract *testutil.Contract
	now      time.Time
	secrets  []string // values that must never show up in logs or audit
}

type sess struct {
	cookie *http.Cookie
	csrf   string
}

func newE2E(t *testing.T) *e2e {
	t.Helper()
	db := testutil.NewTestDB(t)
	logs := &bytes.Buffer{}
	log := logx.New("info", logs)
	rec := audit.NewRecorder(db, log)
	matrix, err := app.BuildMatrix(append(app.FoundationContributions(), financeiro.Contribution(), estoque.Contribution())...)
	if err != nil {
		t.Fatal(err)
	}
	hasher, _ := password.NewHasher(password.Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1})
	authorizer, _ := authz.NewAuthorizer(matrix.Definitions, audit.DeniedHook(rec))
	e := &e2e{t: t, mail: &emailtest.Recorder{}, logs: logs, contract: testutil.LoadContracts(t), now: time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)}
	e.mod = identity.New(identity.Deps{
		DB: db, Recorder: rec, Authorizer: authorizer, Matrix: matrix, Hasher: hasher, Denylist: password.NewDenylist("12345678\n"),
		Sender: e.mail, Log: log, HashKey: []byte("0123456789abcdef0123456789abcdef"),
		SessionIdle: time.Hour, SessionAbsolute: 8 * time.Hour, ResetTTL: 30 * time.Minute, BaseURL: appOrigin,
		Now: func() time.Time { return e.now },
	})
	e.mod.RequestReset.Async = func(fn func()) { fn() }
	if _, err := e.mod.Roles.Sync(context.Background(), matrix); err != nil {
		t.Fatal(err)
	}
	s3cfg := testutil.SharedS3(t)
	s3, err := storage.NewS3(context.Background(), storage.Config{
		Endpoint: s3cfg.Endpoint, Region: s3cfg.Region, Bucket: s3cfg.Bucket,
		AccessKey: s3cfg.AccessKey, SecretKey: s3cfg.SecretKey, UsePathStyle: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	docs := &documents.Service{DB: db, Storage: s3, Authz: authorizer, Audit: rec, URLTTL: 5 * time.Minute}
	fin := financeiro.New(financeiro.Deps{DB: db, Recorder: rec, Authorizer: authorizer, Documents: docs})
	est := estoque.New(estoque.Deps{DB: db, Recorder: rec, Authorizer: authorizer})
	e.router = NewRouter(Deps{
		Ping: func(context.Context) error { return nil }, Identity: e.mod, Financeiro: fin, Estoque: est,
		AuditQuery: &audit.Query{Authz: authorizer, DB: db}, Log: log,
		AllowedOrigins: []string{appOrigin}, Cookie: httpx.SessionCookie{Secure: true},
	})
	return e
}

type opt func(*http.Request)

func withOrigin(o string) opt { return func(r *http.Request) { r.Header.Set("Origin", o) } }
func noOrigin() opt          { return func(r *http.Request) { r.Header.Del("Origin") } }
func header(k, v string) opt { return func(r *http.Request) { r.Header.Set(k, v) } }

// do sends the request and validates the response against the contracts.
func (e *e2e) do(s *sess, method, path string, body any, opts ...opt) *httptest.ResponseRecorder {
	e.t.Helper()
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Origin", appOrigin)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if s != nil {
		if s.cookie != nil {
			req.AddCookie(s.cookie)
		}
		if s.csrf != "" {
			req.Header.Set("X-CSRF-Token", s.csrf)
		}
	}
	for _, o := range opts {
		o(req)
	}
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	e.contract.ValidateResponse(e.t, req, w)
	return w
}

func body(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("corpo não é JSON: %v: %s", err, w.Body.String())
	}
	return m
}

func code(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	c, _ := body(t, w)["code"].(string)
	return c
}

func (e *e2e) expect(w *httptest.ResponseRecorder, status int, problem string) {
	e.t.Helper()
	if w.Code != status || (problem != "" && code(e.t, w) != problem) {
		e.t.Fatalf("status = %d (esperado %d), código = %q (esperado %q), corpo = %s", w.Code, status, code(e.t, w), problem, w.Body.String())
	}
}

// bootstrap creates the first administrator, as the command does.
func (e *e2e) bootstrap(role domain.Role) domain.User {
	e.t.Helper()
	u, err := e.mod.Bootstrap.Execute(context.Background(), app.BootstrapInput{Email: "admin@exemplo.com", Name: "Administrador", Password: firstPassword, Role: role})
	if err != nil {
		e.t.Fatal(err)
	}
	e.secrets = append(e.secrets, firstPassword)
	return u
}

func (e *e2e) login(email, pw string) (*httptest.ResponseRecorder, *sess) {
	e.t.Helper()
	w := e.do(nil, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": email, "password": pw})
	if w.Code != http.StatusOK {
		return w, nil
	}
	return w, &sess{cookie: w.Result().Cookies()[0], csrf: body(e.t, w)["csrf_token"].(string)}
}

func (e *e2e) mustLogin(email, pw string) *sess {
	e.t.Helper()
	w, s := e.login(email, pw)
	if s == nil {
		e.t.Fatalf("login de %s: %d %s", email, w.Code, w.Body.String())
	}
	return s
}

// adminSession is the bootstrapped PRESIDENTE, signed in with the password already changed.
func (e *e2e) adminSession() *sess {
	e.t.Helper()
	e.bootstrap(domain.RolePresidente)
	s := e.mustLogin("admin@exemplo.com", firstPassword)
	e.expect(e.do(s, http.MethodPost, "/api/v1/auth/password", map[string]string{"current_password": firstPassword, "new_password": secondPassword}), http.StatusNoContent, "")
	e.secrets = append(e.secrets, secondPassword)
	return e.mustLogin("admin@exemplo.com", secondPassword)
}

// createUser creates a user through the API and returns its id.
func (e *e2e) createUser(s *sess, email string) string {
	e.t.Helper()
	w := e.do(s, http.MethodPost, "/api/v1/users", map[string]string{"email": email, "name": "Pessoa " + email, "password": "senha-inicial-1234"})
	e.expect(w, http.StatusCreated, "")
	return body(e.t, w)["id"].(string)
}

// events lists the audit entries of an action through the API.
func (e *e2e) events(s *sess, action string) []map[string]any {
	e.t.Helper()
	w := e.do(s, http.MethodGet, "/api/v1/audit-logs?limit=100&action="+action, nil)
	e.expect(w, http.StatusOK, "")
	var out []map[string]any
	for _, it := range body(e.t, w)["items"].([]any) {
		out = append(out, it.(map[string]any))
	}
	return out
}

// promoted creates a user, promotes them and returns their session (password already changed).
func (e *e2e) promoted(admin *sess, email string, roles ...string) (string, *sess) {
	e.t.Helper()
	id := e.createUser(admin, email)
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/admin-membership", map[string]any{"reason": "assumiu a função na diretoria", "roles": roles}), http.StatusNoContent, "")
	s := e.mustLogin(email, "senha-inicial-1234")
	e.expect(e.do(s, http.MethodPost, "/api/v1/auth/password", map[string]string{"current_password": "senha-inicial-1234", "new_password": "senha-definitiva-5678"}), http.StatusNoContent, "")
	return id, e.mustLogin(email, "senha-definitiva-5678")
}

// IDN-01, IDN-02.1 e .9: o administrador criado pelo bootstrap entra e recebe o contexto completo.
func TestE2EBootstrapAdministratorSignsInWithTheFullContext(t *testing.T) {
	e := newE2E(t)
	e.bootstrap(domain.RoleAdminSistema)

	w, s := e.login("admin@exemplo.com", firstPassword)

	if s == nil {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	b := body(t, w)
	if b["must_change_password"] != true || len(b["permissions"].([]any)) == 0 {
		t.Errorf("contexto = %v", b)
	}
	m := b["admin_membership"].(map[string]any)
	if m["reason"] != "bootstrap-admin" || len(m) != 2 {
		t.Errorf("admin_membership = %v", m)
	}
	me := e.do(s, http.MethodGet, "/api/v1/auth/me", nil)
	if me.Code != http.StatusOK || strings.TrimSpace(me.Body.String()) != strings.TrimSpace(w.Body.String()) {
		t.Errorf("o me deveria devolver o mesmo corpo do login")
	}
}

// RBAC-02.6: o administrador criado pelo bootstrap só troca a senha antes de usar o resto.
func TestE2EFirstLoginForcesThePasswordChange(t *testing.T) {
	e := newE2E(t)
	e.bootstrap(domain.RolePresidente)
	s := e.mustLogin("admin@exemplo.com", firstPassword)

	e.expect(e.do(s, http.MethodGet, "/api/v1/users", nil), http.StatusForbidden, "password_change_required")
	e.expect(e.do(s, http.MethodPost, "/api/v1/auth/password", map[string]string{"current_password": firstPassword, "new_password": secondPassword}), http.StatusNoContent, "")

	e.expect(e.do(s, http.MethodGet, "/api/v1/users", nil), http.StatusOK, "")
	if body(t, e.do(s, http.MethodGet, "/api/v1/auth/me", nil))["must_change_password"] != false {
		t.Error("a flag deveria cair depois da troca")
	}
}

// O fluxo completo do enunciado: bootstrap, login, criar usuário, promover, trocar a senha obrigatória,
// recuperar o acesso por e-mail, redefinir por administrador, atribuir papel, retirar o acesso, desativar e consultar a auditoria.
func TestE2EFullIdentityFlow(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()

	// criar e promover a TESOURARIA (a promoção derruba as sessões e exige a troca)
	id := e.createUser(admin, "tesoureira@exemplo.com")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/admin-membership", map[string]any{"reason": "eleita tesoureira da diretoria", "roles": []string{"TESOURARIA"}}), http.StatusNoContent, "")
	first := e.mustLogin("tesoureira@exemplo.com", "senha-inicial-1234")
	e.expect(e.do(first, http.MethodGet, "/api/v1/auth/me", nil), http.StatusOK, "")
	e.expect(e.do(first, http.MethodGet, "/api/v1/users", nil), http.StatusForbidden, "password_change_required")
	e.expect(e.do(first, http.MethodPost, "/api/v1/auth/password", map[string]string{"current_password": "senha-inicial-1234", "new_password": "senha-definitiva-5678"}), http.StatusNoContent, "")
	tes := e.mustLogin("tesoureira@exemplo.com", "senha-definitiva-5678")

	// recuperação por e-mail: solicitação, token, redefinição, sessões antigas revogadas
	e.expect(e.do(nil, http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": "tesoureira@exemplo.com"}), http.StatusAccepted, "")
	msg, _ := e.mail.Last()
	_, after, _ := strings.Cut(msg.TextBody, appOrigin+"/redefinir-senha#token=")
	token := strings.Fields(after)[0]
	e.secrets = append(e.secrets, token)
	e.expect(e.do(nil, http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": token, "new_password": "senha-recuperada-9012"}), http.StatusNoContent, "")
	e.expect(e.do(tes, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "unauthenticated")
	tes = e.mustLogin("tesoureira@exemplo.com", "senha-recuperada-9012")
	e.expect(e.do(tes, http.MethodGet, "/api/v1/auth/me", nil), http.StatusOK, "")

	// redefinição administrativa: senha temporária uma vez, sessões caem
	reset := e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/password-reset", map[string]string{"reason": "perdeu o acesso ao e-mail cadastrado"})
	e.expect(reset, http.StatusOK, "")
	temp := body(t, reset)["temporary_password"].(string)
	e.secrets = append(e.secrets, temp)
	e.expect(e.do(tes, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "")
	_, s := e.login("tesoureira@exemplo.com", temp)
	if s == nil {
		t.Fatal("a senha temporária deveria entrar")
	}

	// atribuir papel, retirar o acesso e desativar
	e.expect(e.do(admin, http.MethodPut, "/api/v1/users/"+id+"/roles", map[string][]string{"roles": {"EVENTOS"}}), http.StatusNoContent, "")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/admin-membership/revoke", map[string]string{"reason": "deixou a diretoria em assembleia"}), http.StatusNoContent, "")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/deactivate", nil), http.StatusNoContent, "")
	if w, s := e.login("tesoureira@exemplo.com", temp); s != nil || w.Code != http.StatusUnauthorized {
		t.Errorf("o desativado não entra: %d", w.Code)
	}

	// consultar a auditoria: cada passo está lá
	for _, action := range []string{
		"user.bootstrap", "auth.login", "user.create", "admin.promote", "user.password_change", "auth.password_reset_requested",
		"auth.password_reset_completed", "user.password_reset", "user.roles_set", "admin.revoke", "user.deactivate",
	} {
		if len(e.events(admin, action)) == 0 {
			t.Errorf("a auditoria deveria ter %s", action)
		}
	}
	e.noSecretsLeaked()
}

// noSecretsLeaked: nenhuma senha, senha temporária ou token aparece nos logs nem em qualquer resposta de auditoria.
func (e *e2e) noSecretsLeaked() {
	e.t.Helper()
	admin := e.mustLogin("admin@exemplo.com", secondPassword)
	all := e.do(admin, http.MethodGet, "/api/v1/audit-logs?limit=100", nil).Body.String()
	for _, secret := range e.secrets {
		if strings.Contains(e.logs.String(), secret) || strings.Contains(all, secret) {
			e.t.Errorf("um segredo apareceu nos logs ou na auditoria: %.6s...", secret)
		}
	}
}

// IDN-03.2: sem o token CSRF da sessão, o que altera estado é 403 csrf_invalid.
func TestE2EStateChangesWithoutCSRFAreRefused(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	noCSRF := &sess{cookie: admin.cookie}
	wrong := &sess{cookie: admin.cookie, csrf: "token-errado"}
	newUser := map[string]string{"email": "x@exemplo.com", "name": "X", "password": "senha-inicial-1234"}

	e.expect(e.do(noCSRF, http.MethodPost, "/api/v1/users", newUser), http.StatusForbidden, "csrf_invalid")
	e.expect(e.do(wrong, http.MethodPost, "/api/v1/users", newUser), http.StatusForbidden, "csrf_invalid")
	e.expect(e.do(noCSRF, http.MethodGet, "/api/v1/users", nil), http.StatusOK, "")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users", newUser), http.StatusCreated, "")
}

// IDN-03.3, .4 e .6: origem estranha é 403 (autenticada ou no login); sem Origin só bloqueia cross-site com cookie.
func TestE2EOriginIsCheckedOnStateChanges(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	newUser := map[string]string{"email": "x@exemplo.com", "name": "X", "password": "senha-inicial-1234"}

	e.expect(e.do(admin, http.MethodPost, "/api/v1/users", newUser, withOrigin("https://evil.example")), http.StatusForbidden, "origin_not_allowed")
	e.expect(e.do(nil, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "admin@exemplo.com", "password": secondPassword}, withOrigin("https://evil.example")), http.StatusForbidden, "origin_not_allowed")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users", newUser, noOrigin(), header("Sec-Fetch-Site", "cross-site")), http.StatusForbidden, "origin_not_allowed")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users", newUser, noOrigin()), http.StatusCreated, "")
}

// RBAC-02.1: toda rota fora da lista pública, sem sessão, é 401.
func TestE2ERoutesWithoutASessionAreUnauthenticated(t *testing.T) {
	e := newE2E(t)

	for _, r := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/auth/me"}, {http.MethodGet, "/api/v1/users"}, {http.MethodGet, "/api/v1/audit-logs"},
		{http.MethodPost, "/api/v1/users"}, {http.MethodPost, "/api/v1/auth/logout"},
	} {
		e.expect(e.do(nil, r.method, r.path, nil, header("X-CSRF-Token", "x")), http.StatusUnauthorized, "unauthenticated")
	}
}

// RBAC-02.2 e .7: sem a permissão é 403 forbidden, sem revelar se o recurso existe, e a negativa é auditada.
func TestE2EWithoutThePermissionIs403WithoutRevealingTheResource(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	_, plain := e.promoted(admin, "diretor@exemplo.com", "DIRETORIA")
	existing := e.createUser(admin, "alvo@exemplo.com")

	onExisting := e.do(plain, http.MethodPost, "/api/v1/users/"+existing+"/deactivate", nil)
	onMissing := e.do(plain, http.MethodPost, "/api/v1/users/00000000-0000-4000-8000-000000000000/deactivate", nil)

	e.expect(onExisting, http.StatusForbidden, "forbidden")
	e.expect(onMissing, http.StatusForbidden, "forbidden")
	a, b := body(t, onExisting), body(t, onMissing)
	delete(a, "request_id")
	delete(b, "request_id")
	if a["detail"] != b["detail"] || a["code"] != b["code"] || a["status"] != b["status"] {
		t.Errorf("a resposta não pode revelar se o recurso existe: %v x %v", a, b)
	}
	if len(e.events(admin, "authz.denied")) < 2 {
		t.Error("as negativas deveriam gerar authz.denied")
	}
}

// IDN-04.1 e .3: criar usuário e desativar (as sessões do desativado caem).
func TestE2ECreateAndDeactivateAUser(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	id := e.createUser(admin, "maria@exemplo.com")
	maria := e.mustLogin("maria@exemplo.com", "senha-inicial-1234")

	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/deactivate", nil), http.StatusNoContent, "")

	e.expect(e.do(maria, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "")
	if w, s := e.login("maria@exemplo.com", "senha-inicial-1234"); s != nil {
		t.Errorf("o desativado não entra: %d", w.Code)
	}
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/reactivate", nil), http.StatusNoContent, "")
	e.mustLogin("maria@exemplo.com", "senha-inicial-1234")
}

// IDN-04.7: os papéis novos valem na próxima requisição do alvo.
func TestE2ERolesChangeApplyOnTheNextRequest(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	id, target := e.promoted(admin, "eventos@exemplo.com", "EVENTOS")

	e.expect(e.do(admin, http.MethodPut, "/api/v1/users/"+id+"/roles", map[string][]string{"roles": {"DIRETORIA"}}), http.StatusNoContent, "")

	roles := body(t, e.do(target, http.MethodGet, "/api/v1/auth/me", nil))["roles"].([]any)
	if len(roles) != 2 || roles[0] != "ASSOCIADO" || roles[1] != "DIRETORIA" {
		t.Errorf("roles = %v", roles)
	}
	e.expect(e.do(target, http.MethodGet, "/api/v1/users", nil), http.StatusOK, "")
}

// IDN-06.1 e .6: promover e retirar; o histórico fica e as sessões caem.
func TestE2EPromoteAndRevokeAdministrativeAccess(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	id, target := e.promoted(admin, "tesouro@exemplo.com", "TESOURARIA")

	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/admin-membership/revoke", map[string]string{"reason": "deixou a diretoria em assembleia"}), http.StatusNoContent, "")

	e.expect(e.do(target, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "")
	again := e.mustLogin("tesouro@exemplo.com", "senha-definitiva-5678")
	me := body(t, e.do(again, http.MethodGet, "/api/v1/auth/me", nil))
	if me["admin_membership"] != nil || len(me["roles"].([]any)) != 1 {
		t.Errorf("depois da retirada só resta ASSOCIADO: %v", me)
	}
	history, _ := e.mod.Members.History(context.Background(), id)
	if len(history) != 1 || history[0].RevokedAt == nil {
		t.Errorf("o histórico do vínculo deveria ficar, encerrado: %+v", history)
	}
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/admin-membership/revoke", map[string]string{"reason": "retirar de novo quem já saiu"}), http.StatusConflict, "not_admin")
}

// RBAC-03.1: quem não cobre as permissões do papel não o concede, e a negativa é auditada.
func TestE2EPrivilegeEscalationIsRefusedAndAudited(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	_, sysadmin := e.promoted(admin, "sistema@exemplo.com", "ADMIN_SISTEMA")
	target := e.createUser(admin, "alvo@exemplo.com")

	w := e.do(sysadmin, http.MethodPost, "/api/v1/users/"+target+"/admin-membership", map[string]any{"reason": "promoção acima do que o ator pode", "roles": []string{"CONSELHO_FISCAL"}})

	e.expect(w, http.StatusForbidden, "privilege_escalation")
	if _, ok, _ := e.mod.Members.Active(context.Background(), target); ok {
		t.Error("nada deveria mudar")
	}
	if len(e.events(admin, "role.change_denied")) != 1 {
		t.Error("a negativa deveria gerar role.change_denied")
	}
}

// RBAC-03.3: ninguém altera o próprio acesso.
func TestE2ESelfChangeIsRefused(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	me := body(t, e.do(admin, http.MethodGet, "/api/v1/auth/me", nil))["user"].(map[string]any)["id"].(string)

	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+me+"/deactivate", nil), http.StatusForbidden, "self_change_forbidden")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+me+"/admin-membership/revoke", map[string]string{"reason": "retirando o meu próprio acesso"}), http.StatusForbidden, "self_change_forbidden")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+me+"/password-reset", map[string]string{"reason": "redefinindo a minha própria senha"}), http.StatusForbidden, "self_change_forbidden")
}

// IDN-07.1, .5 e .6: recuperação idêntica para conta existente e inexistente, token de uso único, sessões antigas revogadas.
func TestE2EPasswordRecoveryByEmail(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	e.createUser(admin, "maria@exemplo.com")
	old := e.mustLogin("maria@exemplo.com", "senha-inicial-1234")

	known := e.do(nil, http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": "maria@exemplo.com"})
	unknown := e.do(nil, http.MethodPost, "/api/v1/auth/password-reset/request", map[string]string{"email": "ninguem@exemplo.com"})
	if known.Code != http.StatusAccepted || known.Body.String() != unknown.Body.String() {
		t.Fatalf("respostas diferentes: %d %s x %d %s", known.Code, known.Body.String(), unknown.Code, unknown.Body.String())
	}
	msg, _ := e.mail.Last()
	_, after, _ := strings.Cut(msg.TextBody, "#token=")
	token := strings.Fields(after)[0]

	e.expect(e.do(nil, http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": token, "new_password": "senha-recuperada-9012"}), http.StatusNoContent, "")

	e.expect(e.do(old, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "")
	e.mustLogin("maria@exemplo.com", "senha-recuperada-9012")
	again := e.do(nil, http.MethodPost, "/api/v1/auth/password-reset/confirm", map[string]string{"token": token, "new_password": "outra-senha-recuperada-3"})
	e.expect(again, http.StatusBadRequest, "invalid_reset_token")
	if len(e.events(admin, "auth.password_reset_failed")) != 1 {
		t.Error("o token usado deveria gerar auth.password_reset_failed")
	}
}

// IDN-08.1: a redefinição administrativa devolve a senha temporária uma vez, sem cache, e derruba as sessões.
func TestE2EAdministrativePasswordReset(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	id := e.createUser(admin, "maria@exemplo.com")
	maria := e.mustLogin("maria@exemplo.com", "senha-inicial-1234")

	w := e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/password-reset", map[string]string{"reason": "perdeu o acesso ao e-mail cadastrado"})

	e.expect(w, http.StatusOK, "")
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q", w.Header().Get("Cache-Control"))
	}
	temp := body(t, w)["temporary_password"].(string)
	e.secrets = append(e.secrets, temp)
	e.expect(e.do(maria, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "")
	next := e.mustLogin("maria@exemplo.com", temp)
	if body(t, e.do(next, http.MethodGet, "/api/v1/auth/me", nil))["must_change_password"] != true {
		t.Error("a senha temporária exige a troca no acesso seguinte")
	}
	e.noSecretsLeaked()
}

// AUD-04.1 a .5 e AUD-03.1: os eventos de segurança são consultáveis pela API.
func TestE2ESecurityEventsAreQueryableThroughTheAPI(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	id, tes := e.promoted(admin, "tesouro@exemplo.com", "TESOURARIA")
	e.mustLogin("tesouro@exemplo.com", "senha-definitiva-5678") // segundo login
	for i := 0; i < 6; i++ {
		e.login("tesouro@exemplo.com", "senha-errada-123") // 5 falhas e 1 bloqueio
	}
	e.expect(e.do(tes, http.MethodPost, "/api/v1/auth/logout", nil), http.StatusNoContent, "")
	e.expect(e.do(admin, http.MethodPut, "/api/v1/users/"+id+"/roles", map[string][]string{"roles": {"EVENTOS"}}), http.StatusNoContent, "")
	_, sysadmin := e.promoted(admin, "sistema@exemplo.com", "ADMIN_SISTEMA")
	other := e.createUser(admin, "alvo@exemplo.com")
	e.expect(e.do(sysadmin, http.MethodPost, "/api/v1/users/"+other+"/admin-membership", map[string]any{"reason": "promoção acima do que o ator pode", "roles": []string{"CONSELHO_FISCAL"}}), http.StatusForbidden, "privilege_escalation")
	e.expect(e.do(admin, http.MethodPost, "/api/v1/users/"+id+"/admin-membership/revoke", map[string]string{"reason": "deixou a diretoria em assembleia"}), http.StatusNoContent, "")

	for _, action := range []string{"auth.login", "auth.login_failed", "auth.login_blocked", "auth.logout", "admin.promote", "admin.revoke", "user.roles_set", "role.change_denied"} {
		if len(e.events(admin, action)) == 0 {
			t.Errorf("o evento %s deveria ser consultável", action)
		}
	}
}

// AUD-04.2: a falha de login não guarda a senha nem o e-mail em claro, e o ator é anônimo.
func TestE2EFailedLoginIsAuditedWithoutThePasswordOrThePlainEmail(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()

	e.login("desconhecido@exemplo.com", "senha-errada-que-nao-pode-vazar")

	evs := e.events(admin, "auth.login_failed")
	if len(evs) != 1 || evs[0]["actor_type"] != "anonymous" || evs[0]["outcome"] != "failure" {
		t.Fatalf("eventos = %v", evs)
	}
	raw, _ := json.Marshal(evs[0])
	for _, leak := range []string{"senha-errada-que-nao-pode-vazar", "desconhecido@exemplo.com"} {
		if strings.Contains(string(raw), leak) || strings.Contains(e.logs.String(), leak) {
			t.Errorf("o evento ou o log não pode conter %q", leak)
		}
	}
}

// IDN-02.5: a sessão ociosa por mais de 60 minutos expira e a resposta é session_expired, limpando o cookie.
func TestE2EAnIdleSessionExpires(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()

	e.now = e.now.Add(61 * time.Minute)
	w := e.do(admin, http.MethodGet, "/api/v1/auth/me", nil)

	e.expect(w, http.StatusUnauthorized, "session_expired")
	if c := w.Result().Cookies(); len(c) != 1 || c[0].MaxAge >= 0 {
		t.Errorf("o cookie deveria ser limpo: %v", c)
	}
}

// AUD-03.1: a consulta exige audit:log:read; quem não a tem recebe 403.
func TestE2EAuditQueryNeedsItsPermission(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	_, diretoria := e.promoted(admin, "diretor@exemplo.com", "DIRETORIA")

	e.expect(e.do(diretoria, http.MethodGet, "/api/v1/audit-logs", nil), http.StatusForbidden, "forbidden")
	e.expect(e.do(admin, http.MethodGet, "/api/v1/audit-logs", nil), http.StatusOK, "")
}

// IDN-04.8: a listagem de usuários pagina por cursor e o limite acima de 100 é 422 invalid_limit.
func TestE2EUserListingPaginatesAndRejectsAnExcessiveLimit(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	for _, email := range []string{"a@exemplo.com", "b@exemplo.com"} {
		e.createUser(admin, email)
	}

	first := e.do(admin, http.MethodGet, "/api/v1/users?limit=2", nil)
	e.expect(first, http.StatusOK, "")
	next, _ := body(t, first)["next_cursor"].(string)
	second := e.do(admin, http.MethodGet, "/api/v1/users?limit=2&cursor="+next, nil)

	e.expect(second, http.StatusOK, "")
	if len(body(t, first)["items"].([]any)) != 2 || len(body(t, second)["items"].([]any)) != 1 {
		t.Errorf("páginas = %v e %v", body(t, first), body(t, second))
	}
	e.expect(e.do(admin, http.MethodGet, "/api/v1/users?limit=101", nil), http.StatusUnprocessableEntity, "invalid_limit")
}

// Nenhuma resposta de usuário traz hash de senha nem token.
func TestE2ENoUserResponseCarriesAHashOrAToken(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	e.createUser(admin, "maria@exemplo.com")

	for _, path := range []string{"/api/v1/users", "/api/v1/auth/me"} {
		out := strings.ToLower(e.do(admin, http.MethodGet, path, nil).Body.String())
		for _, leak := range []string{"argon2", "password_hash", "token_hash", "session_token"} {
			if strings.Contains(out, leak) {
				t.Errorf("%s contém %q", path, leak)
			}
		}
	}
}

// Uma rota que não existe é 404 em problem+json, sem passar pela autenticação como se existisse.
func TestE2EUnknownRouteIs404(t *testing.T) {
	e := newE2E(t)

	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/nao-existe", nil))

	if w.Code != http.StatusNotFound || w.Header().Get("Content-Type") != "application/problem+json" {
		t.Errorf("status = %d, tipo = %s", w.Code, w.Header().Get("Content-Type"))
	}
}

// IDN-03.5: GET não altera estado (nem usuários, nem vínculos).
func TestE2EReadsNeverChangeState(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()
	count := func() int {
		page := body(t, e.do(admin, http.MethodGet, "/api/v1/users?limit=100", nil))
		return len(page["items"].([]any))
	}
	before := count()

	for _, path := range []string{"/api/v1/users", "/api/v1/auth/me", "/api/v1/audit-logs", "/healthz"} {
		e.expect(e.do(admin, http.MethodGet, path, nil, noOrigin()), http.StatusOK, "")
	}

	if after := count(); after != before {
		t.Errorf("usuários: %d antes, %d depois", before, after)
	}
}

// IDN-02.6: o logout revoga a sessão e limpa o cookie.
func TestE2ELogoutRevokesTheSessionAndClearsTheCookie(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()

	w := e.do(admin, http.MethodPost, "/api/v1/auth/logout", nil)

	e.expect(w, http.StatusNoContent, "")
	if c := w.Result().Cookies(); len(c) != 1 || c[0].MaxAge >= 0 {
		t.Errorf("o cookie deveria ser limpo: %v", c)
	}
	e.expect(e.do(admin, http.MethodGet, "/api/v1/auth/me", nil), http.StatusUnauthorized, "")
}

// API-01: o contrato valida a estrutura da requisição, nas rotas autenticadas e nas públicas,
// com 422 validation_failed e o erro por campo (antes de qualquer caso de uso).
func TestE2EContractValidationRunsOnAuthenticatedAndPublicRoutes(t *testing.T) {
	e := newE2E(t)
	admin := e.adminSession()

	auth := e.do(admin, http.MethodPost, "/api/v1/users", map[string]string{"email": "x@exemplo.com", "password": "senha-inicial-1234"})
	pub := e.do(nil, http.MethodPost, "/api/v1/auth/login", map[string]string{"email": "x@exemplo.com"})

	for name, w := range map[string]*httptest.ResponseRecorder{"autenticada": auth, "pública": pub} {
		e.expect(w, http.StatusUnprocessableEntity, "validation_failed")
		if errs, _ := body(t, w)["errors"].([]any); len(errs) == 0 {
			t.Errorf("%s: deveria listar o campo faltante: %s", name, w.Body.String())
		}
	}
}
