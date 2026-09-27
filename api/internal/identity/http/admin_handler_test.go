//go:build integration

package identityhttp_test

import (
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func promoteBody(reason string, roles ...string) map[string]any {
	return map[string]any{"reason": reason, "roles": roles}
}

func reasonBody(reason string) map[string]string { return map[string]string{"reason": reason} }

// auditMentions reports whether any audit entry carries the text anywhere.
func (e *env) auditMentions(text string) bool {
	e.t.Helper()
	var n int64
	err := e.owner.Raw(`SELECT count(*) FROM audit_log
		WHERE position(? in concat_ws('|', entity_id, reason, before::text, after::text, context::text)) > 0`, text).Scan(&n).Error
	if err != nil {
		e.t.Fatal(err)
	}
	return n > 0
}

func (e *env) rolesOf(id string) []domain.Role {
	e.t.Helper()
	roles, err := e.mod.Roles.RolesOf(e.t.Context(), id)
	if err != nil {
		e.t.Fatal(err)
	}
	return roles
}

// IDN-06.1: a promoção cria o vínculo, atribui os papéis, derruba as sessões e audita.
func TestPromoteCreatesTheMembershipAssignsRolesAndRevokesSessions(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target, targetSession := e.signedIn()

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership", promoteBody("assumiu a tesouraria da diretoria", "TESOURARIA"))

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if e.as(targetSession, http.MethodGet, "/api/v1/auth/me", nil).Code != http.StatusUnauthorized {
		t.Error("as sessões do promovido deveriam cair")
	}
	m, ok, err := e.mod.Members.Active(t.Context(), target.ID)
	if err != nil || !ok || m.Reason != "assumiu a tesouraria da diretoria" || m.GrantedBy == nil {
		t.Errorf("vínculo = %+v, ativo = %v, erro = %v", m, ok, err)
	}
	if roles := e.rolesOf(target.ID); !slices.Contains(roles, domain.RoleTesouraria) || !slices.Contains(roles, domain.RoleAssociado) {
		t.Errorf("papéis = %v", roles)
	}
	if len(e.events("admin.promote")) != 1 {
		t.Error("a promoção deveria gerar admin.promote")
	}
	next := decode(t, e.do(http.MethodPost, "/api/v1/auth/login", loginBody(target, goodPassword)))
	if next["must_change_password"] != true {
		t.Errorf("a promoção exige a troca de senha no próximo login: %v", next["must_change_password"])
	}
}

// IDN-06.2: motivo curto é 422 reason_required e nada muda.
func TestPromoteWithAShortReasonIs422AndChangesNothing(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser()

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership", promoteBody("curto", "TESOURARIA"))

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "reason_required" {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if _, ok, _ := e.mod.Members.Active(t.Context(), target.ID); ok || len(e.rolesOf(target.ID)) != 1 {
		t.Error("nada deveria mudar")
	}
}

// IDN-06.3: quem já tem vínculo ativo não é promovido de novo.
func TestPromoteAnExistingAdministratorIs409(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser(domain.RoleTesouraria)

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership", promoteBody("outra promoção qualquer", "EVENTOS"))

	if w.Code != http.StatusConflict || codeOf(t, w) != "already_admin" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// IDN-06.4: usuário inativo não é promovido.
func TestPromoteAnInactiveUserIs409(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser()
	if err := e.mod.Users.SetActive(t.Context(), target.ID, false); err != nil {
		t.Fatal(err)
	}

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership", promoteBody("promoção de usuário inativo", "EVENTOS"))

	if w.Code != http.StatusConflict || codeOf(t, w) != "user_inactive" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// IDN-06.10: a promoção precisa de ao menos um papel além de ASSOCIADO.
func TestPromoteWithOnlyAssociadoIs422(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser()

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership", promoteBody("promoção sem papel administrativo", "ASSOCIADO"))

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "admin_role_required" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// IDN-06.6: a retirada fecha o vínculo com quem, quando e por quê, deixa só ASSOCIADO,
// derruba as sessões e preserva o histórico.
func TestRevokeClosesTheMembershipKeepsTheHistoryAndLeavesOnlyAssociado(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target, targetSession := e.signedIn(domain.RoleTesouraria)

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership/revoke", reasonBody("saiu da diretoria em assembleia"))

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if e.as(targetSession, http.MethodGet, "/api/v1/auth/me", nil).Code != http.StatusUnauthorized {
		t.Error("as sessões de quem perdeu o acesso deveriam cair")
	}
	if roles := e.rolesOf(target.ID); len(roles) != 1 || roles[0] != domain.RoleAssociado {
		t.Errorf("papéis = %v", roles)
	}
	history, err := e.mod.Members.History(t.Context(), target.ID)
	if err != nil || len(history) != 1 || history[0].RevokedAt == nil || history[0].RevokedBy == nil || history[0].RevokeReason == nil ||
		*history[0].RevokeReason != "saiu da diretoria em assembleia" {
		t.Errorf("o histórico do vínculo deveria ficar, fechado: %+v %v", history, err)
	}
	if len(e.events("admin.revoke")) != 1 {
		t.Error("a retirada deveria gerar admin.revoke")
	}
}

// IDN-06.7: quem não tem vínculo ativo não perde o que não tem.
func TestRevokeAUserWithoutMembershipIs409(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser()

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership/revoke", reasonBody("retirada de quem não é administrador"))

	if w.Code != http.StatusConflict || codeOf(t, w) != "not_admin" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// RBAC-03.1: ninguém concede o que não tem, e a negativa fica na auditoria.
func TestPromoteToARoleTheActorDoesNotCoverIs403AndAuditsTheDenial(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RoleAdminSistema)
	target := e.newUser()

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership", promoteBody("promoção acima do que o ator pode", "CONSELHO_FISCAL"))

	if w.Code != http.StatusForbidden || codeOf(t, w) != "privilege_escalation" {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if _, ok, _ := e.mod.Members.Active(t.Context(), target.ID); ok {
		t.Error("nada deveria mudar")
	}
	if len(e.events("role.change_denied")) != 1 {
		t.Error("a negativa deveria gerar role.change_denied")
	}
}

// RBAC-03.2: ninguém retira o acesso de quem tem mais poder do que ele.
func TestRevokeSomeoneMorePowerfulIs403(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RoleAdminSistema)
	target := e.newUser(domain.RolePresidente)

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/admin-membership/revoke", reasonBody("tentativa de retirar o presidente"))

	if w.Code != http.StatusForbidden || codeOf(t, w) != "privilege_escalation" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if _, ok, _ := e.mod.Members.Active(t.Context(), target.ID); !ok {
		t.Error("o vínculo do presidente deveria continuar")
	}
}

// RBAC-03.3: ninguém altera o próprio vínculo nem redefine a própria senha.
func TestActorsCannotChangeTheirOwnMembershipOrResetTheirOwnPassword(t *testing.T) {
	e := newEnv(t)
	me, s := e.signedIn(domain.RolePresidente)

	revoke := e.as(s, http.MethodPost, "/api/v1/users/"+me.ID+"/admin-membership/revoke", reasonBody("retirando o meu próprio acesso"))
	reset := e.as(s, http.MethodPost, "/api/v1/users/"+me.ID+"/password-reset", reasonBody("redefinindo a minha própria senha"))

	for name, w := range map[string]string{"revoke": codeOf(t, revoke), "reset": codeOf(t, reset)} {
		if w != "self_change_forbidden" {
			t.Errorf("%s: code = %q", name, w)
		}
	}
	if revoke.Code != http.StatusForbidden || reset.Code != http.StatusForbidden {
		t.Errorf("status = %d e %d", revoke.Code, reset.Code)
	}
	if len(e.events("user.password_reset")) != 1 || e.events("user.password_reset")[0]["outcome"] != "denied" {
		t.Errorf("a negativa do reset é auditada com outcome denied: %v", e.events("user.password_reset"))
	}
}

// IDN-08.1 e .7: a senha temporária vem uma vez, sem cache, e não vai para log, auditoria nem e-mail.
func TestAdminResetReturnsTheTemporaryPasswordOnceWithoutCachingOrLeaking(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target, targetSession := e.signedIn()

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/password-reset", reasonBody("perdeu o acesso ao e-mail cadastrado"))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q", cc)
	}
	temp, _ := decode(t, w)["temporary_password"].(string)
	if len(temp) < 10 {
		t.Fatalf("senha temporária = %q", temp)
	}
	if e.as(targetSession, http.MethodGet, "/api/v1/auth/me", nil).Code != http.StatusUnauthorized {
		t.Error("as sessões do alvo deveriam cair")
	}
	relogin := e.do(http.MethodPost, "/api/v1/auth/login", loginBody(target, temp))
	if relogin.Code != http.StatusOK || decode(t, relogin)["must_change_password"] != true {
		t.Errorf("a senha temporária entra e exige a troca: %d %s", relogin.Code, relogin.Body.String())
	}
	if e.auditMentions(temp) || strings.Contains(e.logs.String(), temp) {
		t.Error("a senha temporária não pode aparecer na auditoria nem nos logs")
	}
	if len(e.mail.Messages()) != 0 {
		t.Error("a senha temporária não vai por e-mail")
	}
	if len(e.events("user.password_reset")) != 1 || !e.auditMentions("perdeu o acesso ao e-mail cadastrado") {
		t.Error("o reset deveria gerar user.password_reset com o motivo")
	}
}

// IDN-08.2: não se redefine a senha de quem tem mais poder do que o ator.
func TestAdminResetOfSomeoneMorePowerfulIs403(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RoleAdminSistema)
	target := e.newUser(domain.RolePresidente)

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/password-reset", reasonBody("tentativa de redefinir o presidente"))

	if w.Code != http.StatusForbidden || codeOf(t, w) != "privilege_escalation" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") == "no-store" && strings.Contains(w.Body.String(), "temporary_password") {
		t.Error("uma negativa não devolve senha")
	}
}

// IDN-08.4: usuário inativo não tem a senha redefinida.
func TestAdminResetOfAnInactiveUserIs409(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser()
	if err := e.mod.Users.SetActive(t.Context(), target.ID, false); err != nil {
		t.Fatal(err)
	}

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/password-reset", reasonBody("reset de usuário desativado"))

	if w.Code != http.StatusConflict || codeOf(t, w) != "user_inactive" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// IDN-08: o motivo curto é 422 reason_required e nenhuma senha sai.
func TestAdminResetWithAShortReasonIs422(t *testing.T) {
	e := newEnv(t)
	_, admin := e.signedIn(domain.RolePresidente)
	target := e.newUser()

	w := e.as(admin, http.MethodPost, "/api/v1/users/"+target.ID+"/password-reset", reasonBody("curto"))

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "reason_required" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}
