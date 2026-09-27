//go:build integration

package identityhttp_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func createBody(email string) map[string]string {
	return map[string]string{"email": email, "name": "Maria da Silva", "password": "senha-inicial-123"}
}

// IDN-04.1: quem tem identity:user:create cria um usuário só com ASSOCIADO e troca de senha obrigatória.
func TestCreateUserAnswers201WithOnlyTheAssociadoRoleAndMustChangePassword(t *testing.T) {
	e := newEnv(t)
	actor, s := e.signedIn(domain.RolePresidente)

	w := e.as(s, http.MethodPost, "/api/v1/users", createBody("maria@exemplo.com"))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["email"] != "maria@exemplo.com" || body["name"] != "Maria da Silva" || body["active"] != true || body["must_change_password"] != true {
		t.Errorf("corpo = %v", body)
	}
	if roles := body["roles"].([]any); len(roles) != 1 || roles[0] != "ASSOCIADO" {
		t.Errorf("roles = %v", roles)
	}
	if v, present := body["admin_membership"]; !present || v != nil {
		t.Errorf("admin_membership deveria ser null: %v", v)
	}
	evs := e.events("user.create")
	if len(evs) != 1 || evs[0]["entity_id"] != body["id"] {
		t.Errorf("eventos user.create = %v", evs)
	}
	_ = actor
}

// IDN-04.2: o e-mail repetido, sem distinguir maiúsculas, é 409 email_taken.
func TestCreateUserWithATakenEmailIs409IgnoringCase(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RolePresidente)
	e.as(s, http.MethodPost, "/api/v1/users", createBody("maria@exemplo.com"))

	w := e.as(s, http.MethodPost, "/api/v1/users", createBody("MARIA@Exemplo.COM"))

	if w.Code != http.StatusConflict || codeOf(t, w) != "email_taken" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// Sem a permissão a resposta é 403 forbidden, e a negativa fica na auditoria.
func TestCreateUserWithoutThePermissionIs403AndTheDenialIsAudited(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleDiretoria)

	w := e.as(s, http.MethodPost, "/api/v1/users", createBody("maria@exemplo.com"))

	if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if len(e.events("authz.denied")) != 1 || len(e.events("user.create")) != 0 {
		t.Errorf("authz.denied = %d, user.create = %d", len(e.events("authz.denied")), len(e.events("user.create")))
	}
}

// IDN-04.9: nenhuma resposta traz hash de senha nem token de sessão.
func TestUserResponsesNeverCarryPasswordHashesOrTokens(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RolePresidente)
	created := e.as(s, http.MethodPost, "/api/v1/users", createBody("maria@exemplo.com"))
	listed := e.as(s, http.MethodGet, "/api/v1/users", nil)

	for name, w := range map[string]string{"criação": created.Body.String(), "listagem": listed.Body.String()} {
		for _, leak := range []string{"argon2", "password", "token", "hash", "senha-inicial"} {
			if strings.Contains(strings.ToLower(strings.ReplaceAll(w, "must_change_password", "")), leak) {
				t.Errorf("%s contém %q: %s", name, leak, w)
			}
		}
	}
}

// IDN-04.3: a desativação derruba as sessões do usuário e ele deixa de entrar.
func TestDeactivateUserRevokesTheSessionsAndBlocksTheLogin(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target, targetSession := e.signedIn()

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/deactivate", nil)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if e.as(targetSession, http.MethodGet, "/api/v1/auth/me", nil).Code != http.StatusUnauthorized {
		t.Error("a sessão do usuário desativado deveria cair")
	}
	if e.do(http.MethodPost, "/api/v1/auth/login", loginBody(target, goodPassword)).Code != http.StatusUnauthorized {
		t.Error("o usuário desativado não entra")
	}
	if len(e.events("user.deactivate")) != 1 {
		t.Error("a desativação deveria gerar user.deactivate")
	}
}

// IDN-04.10: reativar devolve o acesso, sem restaurar papéis administrativos.
func TestReactivateUserDoesNotRestoreAdministrativeRoles(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser(domain.RoleTesouraria)
	e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/deactivate", nil)

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/reactivate", nil)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	roles, err := e.mod.Roles.RolesOf(t.Context(), target.ID)
	if err != nil || !slices.Contains(roles, domain.RoleTesouraria) {
		t.Errorf("os papéis não são restaurados nem removidos pela reativação: %v %v", roles, err)
	}
	if len(e.events("user.reactivate")) != 1 {
		t.Error("a reativação deveria gerar user.reactivate")
	}
	if e.do(http.MethodPost, "/api/v1/auth/login", loginBody(target, goodPassword)).Code != http.StatusOK {
		t.Error("o usuário reativado entra")
	}
}

// Autodesativação é proibida; um usuário inexistente é 404.
func TestDeactivateRefusesSelfAndUnknownUsers(t *testing.T) {
	e := newEnv(t)
	me, s := e.signedIn(domain.RolePresidente)

	self := e.as(s, http.MethodPost, "/api/v1/users/"+me.ID+"/deactivate", nil)
	unknown := e.as(s, http.MethodPost, "/api/v1/users/00000000-0000-4000-8000-000000000000/deactivate", nil)

	if self.Code != http.StatusForbidden || codeOf(t, self) != "self_change_forbidden" {
		t.Errorf("autodesativação: %d %s", self.Code, self.Body.String())
	}
	if unknown.Code != http.StatusNotFound || codeOf(t, unknown) != "not_found" {
		t.Errorf("inexistente: %d %s", unknown.Code, unknown.Body.String())
	}
}

// IDN-04.7: a troca de papéis vale na próxima requisição do alvo e registra antes e depois.
func TestSetUserRolesAppliesFromTheNextRequestAndAuditsBeforeAndAfter(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target, targetSession := e.signedIn(domain.RoleTesouraria)

	w := e.as(admin, http.MethodPut, "/api/v1/users/"+target.ID+"/roles", map[string][]string{"roles": {"EVENTOS"}})

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	me := decode(t, e.as(targetSession, http.MethodGet, "/api/v1/auth/me", nil))
	roles := me["roles"].([]any)
	if len(roles) != 2 || roles[0] != "ASSOCIADO" || roles[1] != "EVENTOS" {
		t.Errorf("roles depois da troca = %v", roles)
	}
	evs := e.events("user.roles_set")
	if len(evs) != 1 || !strings.Contains(evs[0]["entity_id"].(string), target.ID) {
		t.Errorf("eventos user.roles_set = %v", evs)
	}
}

// Quem não tem vínculo administrativo não recebe papéis: 409 admin_membership_required.
func TestSetUserRolesWithoutAMembershipIs409(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser()

	w := e.as(admin, http.MethodPut, "/api/v1/users/"+target.ID+"/roles", map[string][]string{"roles": {"EVENTOS"}})

	if w.Code != http.StatusConflict || codeOf(t, w) != "admin_membership_required" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// IDN-04.8: a listagem tem cursor, filtros e devolve next_cursor.
func TestListUsersPaginatesFiltersAndReturnsTheNextCursor(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RolePresidente)
	inactive := e.newUser()
	e.newUser()
	e.newUser()
	e.as(s, http.MethodPost, "/api/v1/users/"+inactive.ID+"/deactivate", nil)

	first := decode(t, e.as(s, http.MethodGet, "/api/v1/users?limit=2", nil))
	next, _ := first["next_cursor"].(string)
	if len(first["items"].([]any)) != 2 || next == "" {
		t.Fatalf("primeira página = %v", first)
	}
	second := decode(t, e.as(s, http.MethodGet, "/api/v1/users?limit=2&cursor="+next, nil))
	if len(second["items"].([]any)) != 2 || second["next_cursor"] != nil {
		t.Errorf("segunda página = %v", second)
	}
	seen := map[any]bool{}
	for _, page := range []map[string]any{first, second} {
		for _, it := range page["items"].([]any) {
			seen[it.(map[string]any)["id"]] = true
		}
	}
	if len(seen) != 4 {
		t.Errorf("as páginas não deveriam repetir usuários: %d distintos", len(seen))
	}
	inactives := decode(t, e.as(s, http.MethodGet, "/api/v1/users?active=false", nil))["items"].([]any)
	if len(inactives) != 1 || inactives[0].(map[string]any)["id"] != inactive.ID {
		t.Errorf("filtro active=false = %v", inactives)
	}
	presidentes := decode(t, e.as(s, http.MethodGet, "/api/v1/users?role=PRESIDENTE", nil))["items"].([]any)
	if len(presidentes) != 1 {
		t.Fatalf("filtro role=PRESIDENTE = %v", presidentes)
	}
	m, _ := presidentes[0].(map[string]any)["admin_membership"].(map[string]any)
	if !slices.Equal(keys(m), []string{"granted_at", "reason"}) || m["reason"] != "membro da diretoria eleita" {
		t.Errorf("a listagem mostra o vínculo só com motivo e data: %v", m)
	}
}

// IDN-04.8: limit acima de 100 é 422.
func TestListUsersWithALimitAbove100Is422(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RolePresidente)

	w := e.as(s, http.MethodGet, "/api/v1/users?limit=101", nil)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if c := codeOf(t, w); c != "invalid_limit" {
		t.Errorf("code = %q, esperado invalid_limit", c)
	}
}

// A leitura é de quem tem identity:user:read (DIRETORIA lê); sem ela, 403.
func TestListUsersNeedsTheReadPermission(t *testing.T) {
	e := newEnv(t)
	_, reader := e.signedIn(domain.RoleDiretoria)
	_, plain := e.signedIn()

	if w := e.as(reader, http.MethodGet, "/api/v1/users", nil); w.Code != http.StatusOK {
		t.Errorf("DIRETORIA lê a listagem: %d %s", w.Code, w.Body.String())
	}
	if w := e.as(plain, http.MethodGet, "/api/v1/users", nil); w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
		t.Errorf("ASSOCIADO não lê: %d %s", w.Code, w.Body.String())
	}
}
