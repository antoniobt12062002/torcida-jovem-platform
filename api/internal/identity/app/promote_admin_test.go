//go:build integration

package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

const promoReason = "eleito tesoureiro na assembleia de setembro"

func (e *ucEnv) promote() *app.PromoteToAdmin {
	return &app.PromoteToAdmin{
		Authz: e.authorizer, Users: e.users, Members: e.members, Roles: e.roles, Sessions: e.sessions, Matrix: e.matrix,
		Audit: e.rec, Tx: e.tx, Now: e.clock,
	}
}

func promoteIn(actor authz.Principal, target string, reason string, roles ...domain.Role) app.PromoteInput {
	return app.PromoteInput{Actor: actor, TargetID: target, Roles: roles, Reason: reason}
}

// IDN-06.1, IDN-05.5: promoção completa numa transação.
func TestPromoteGrantsMembershipAndRolesForcesPasswordChangeRevokesSessionsAndAudits(t *testing.T) {
	e := newUCEnv(t)
	presidenteU, presidente := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t) // ASSOCIADO
	e.sessionFor(t, target.ID)
	e.sessionFor(t, target.ID)

	err := e.promote().Execute(e.ctx, promoteIn(presidente, target.ID, "  "+promoReason+"  ", domain.RoleTesouraria, domain.RoleTesouraria))

	if err != nil {
		t.Fatal(err)
	}
	m, ok, _ := e.members.Active(context.Background(), target.ID)
	if !ok || m.Reason != promoReason || m.GrantedBy == nil || *m.GrantedBy != presidenteU.ID {
		t.Errorf("vínculo = %+v, %v", m, ok)
	}
	got := e.rolesOf(t, target.ID)
	slices.Sort(got)
	if !slices.Equal(got, []domain.Role{domain.RoleAssociado, domain.RoleTesouraria}) {
		t.Errorf("papéis = %v (ASSOCIADO é mantido e repetições são ignoradas)", got)
	}
	stored, _ := e.users.FindByID(context.Background(), target.ID)
	if !stored.MustChangePassword {
		t.Error("must_change_password deveria ser verdadeiro")
	}
	if e.activeSessions(t, target.ID) != 0 {
		t.Error("as sessões do alvo deveriam ter sido revogadas")
	}
	evs := e.events(t, "admin.promote")
	if len(evs) != 1 || evs[0].Outcome != "success" || evs[0].EntityID != target.ID || evs[0].Reason == nil || *evs[0].Reason != promoReason ||
		evs[0].ActorID == nil || *evs[0].ActorID != presidenteU.ID {
		t.Fatalf("eventos = %+v", evs)
	}
	assertPromoteBeforeAfter(t, evs[0], []string{"ASSOCIADO"}, []string{"ASSOCIADO", "TESOURARIA"})
}

// assertPromoteBeforeAfter decodes before/after and checks the exact keys and
// values IDN-06.1 requires: the previous and new role names, nothing more and
// nothing less (a Contains check would miss a swapped or padded role list).
func assertPromoteBeforeAfter(t *testing.T, ev auditEvent, beforeRoles, afterRoles []string) {
	t.Helper()
	if ev.Before == nil || ev.After == nil {
		t.Fatalf("before/after = %v / %v", ev.Before, ev.After)
	}
	var before, after map[string]any
	if err := json.Unmarshal([]byte(*ev.Before), &before); err != nil {
		t.Fatalf("before não é JSON: %v: %s", err, *ev.Before)
	}
	if err := json.Unmarshal([]byte(*ev.After), &after); err != nil {
		t.Fatalf("after não é JSON: %v: %s", err, *ev.After)
	}
	wantKeys := []string{"admin_membership", "roles"}
	if !slices.Equal(sortedKeys(before), wantKeys) {
		t.Errorf("before tem chaves %v, esperado exatamente %v", sortedKeys(before), wantKeys)
	}
	if !slices.Equal(sortedKeys(after), []string{"admin_membership", "must_change_password", "roles"}) {
		t.Errorf("after tem chaves %v", sortedKeys(after))
	}
	if !slices.Equal(anySliceToStrings(before["roles"]), beforeRoles) {
		t.Errorf("before.roles = %v, esperado %v", before["roles"], beforeRoles)
	}
	if !slices.Equal(anySliceToStrings(after["roles"]), afterRoles) {
		t.Errorf("after.roles = %v, esperado %v", after["roles"], afterRoles)
	}
	if before["admin_membership"] != false {
		t.Errorf("before.admin_membership = %v, esperado false", before["admin_membership"])
	}
	if after["admin_membership"] != true || after["must_change_password"] != true {
		t.Errorf("after.admin_membership = %v, after.must_change_password = %v, esperado true e true", after["admin_membership"], after["must_change_password"])
	}
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func anySliceToStrings(v any) []string {
	list, _ := v.([]any)
	out := make([]string, len(list))
	for i, x := range list {
		out[i], _ = x.(string)
	}
	return out
}

func TestPromoteChecksThePermissionFirst(t *testing.T) {
	e := newUCEnv(t)
	_, diretoria := e.actor(t, domain.RoleDiretoria)
	target, _ := e.actor(t)

	err := e.promote().Execute(e.ctx, promoteIn(diretoria, target.ID, "sem-motivo", domain.RoleTesouraria))

	if !errors.Is(err, authz.ErrForbidden) || len(e.events(t, "authz.denied")) != 1 {
		t.Errorf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), target.ID); ok {
		t.Error("nada pode ter sido escrito")
	}
}

// RBAC-03.3
func TestNobodyPromotesThemselves(t *testing.T) {
	e := newUCEnv(t)
	u, adminSistema := e.actor(t, domain.RoleAdminSistema)

	err := e.promote().Execute(e.ctx, promoteIn(adminSistema, u.ID, promoReason, domain.RoleEventos))

	if !errors.Is(err, domain.ErrSelfChangeForbidden) {
		t.Fatalf("err = %v", err)
	}
	if evs := e.events(t, "role.change_denied"); len(evs) != 1 || evs[0].Outcome != "denied" {
		t.Errorf("eventos = %+v", evs)
	}
}

// RBAC-03.1 e .4: só concede o que possui, pelas permissões efetivas.
func TestPromoteRefusesRolesWithPermissionsTheActorLacks(t *testing.T) {
	e := newUCEnv(t)
	_, adminSistema := e.actor(t, domain.RoleAdminSistema)
	_, presidente := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t)

	err := e.promote().Execute(e.ctx, promoteIn(adminSistema, target.ID, promoReason, domain.RoleConselhoFiscal))

	if !errors.Is(err, domain.ErrPrivilegeEscalation) {
		t.Fatalf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), target.ID); ok {
		t.Error("nada pode ter mudado")
	}
	evs := e.events(t, "role.change_denied")
	if len(evs) != 1 || evs[0].Outcome != "denied" || !strings.Contains(evs[0].Context, "CONSELHO_FISCAL") || !strings.Contains(evs[0].Context, "financeiro:parecer:opine") {
		t.Errorf("a negação deveria registrar os papéis pedidos e as permissões que faltam: %+v", evs)
	}
	if err := e.promote().Execute(e.ctx, promoteIn(presidente, target.ID, promoReason, domain.RoleConselhoFiscal)); err != nil {
		t.Errorf("o PRESIDENTE cobre as permissões do CONSELHO_FISCAL: %v", err)
	}
	if err := e.promote().Execute(e.ctx, promoteIn(adminSistema, e.newTarget(t), promoReason, domain.RoleTesouraria)); err != nil {
		t.Errorf("o ADMIN_SISTEMA pode conceder papéis sem permissões que ele não tenha: %v", err)
	}
}

func (e *ucEnv) newTarget(t *testing.T) string {
	t.Helper()
	u, _ := e.actor(t)
	return u.ID
}

// IDN-06.10 e a validação dos papéis.
func TestPromoteRequiresAtLeastOneAdministrativeRole(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	target := e.newTarget(t)

	for _, roles := range [][]domain.Role{nil, {}, {domain.RoleAssociado}} {
		if err := e.promote().Execute(e.ctx, promoteIn(presidente, target, promoReason, roles...)); !errors.Is(err, domain.ErrAdminRoleRequired) {
			t.Errorf("%v: err = %v", roles, err)
		}
	}
	if err := e.promote().Execute(e.ctx, promoteIn(presidente, target, promoReason, domain.RoleEventos, "CHEFE")); !errors.Is(err, domain.ErrUnknownRole) {
		t.Errorf("papel desconhecido: %v", err)
	}
}

// IDN-06.2, .3, .4
func TestPromoteValidatesReasonTargetStateAndExistingMembership(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	uc := e.promote()
	target := e.newTarget(t)

	if err := uc.Execute(e.ctx, promoteIn(presidente, target, "curto", domain.RoleEventos)); !errors.Is(err, domain.ErrReasonRequired) {
		t.Errorf("motivo curto: %v", err)
	}
	inactive := e.newTarget(t)
	_ = e.users.SetActive(context.Background(), inactive, false)
	if err := uc.Execute(e.ctx, promoteIn(presidente, inactive, promoReason, domain.RoleEventos)); !errors.Is(err, domain.ErrUserInactive) {
		t.Errorf("alvo inativo: %v", err)
	}
	if err := uc.Execute(e.ctx, promoteIn(presidente, "0f8fad5b-d9cb-469f-a165-70867728950e", promoReason, domain.RoleEventos)); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("alvo inexistente: %v", err)
	}
	if err := uc.Execute(e.ctx, promoteIn(presidente, target, promoReason, domain.RoleEventos)); err != nil {
		t.Fatal(err)
	}
	before := len(e.events(t, "admin.promote"))
	if err := uc.Execute(e.ctx, promoteIn(presidente, target, promoReason, domain.RoleTesouraria)); !errors.Is(err, domain.ErrAlreadyAdmin) {
		t.Errorf("já é administrador: %v", err)
	}
	if len(e.events(t, "admin.promote")) != before {
		t.Error("uma promoção recusada não audita")
	}
	if got := e.rolesOf(t, target); slices.Contains(got, domain.RoleTesouraria) {
		t.Errorf("os papéis não podem ter mudado: %v", got)
	}
}

// Edge case: dois administradores promovendo o mesmo usuário ao mesmo tempo.
func TestConcurrentPromotionsOfTheSameUserLeaveExactlyOneWinner(t *testing.T) {
	e := newUCEnv(t)
	_, a := e.actor(t, domain.RolePresidente)
	_, b := e.actor(t, domain.RolePresidente)
	target := e.newTarget(t)
	errs := make([]error, 2)

	var wg sync.WaitGroup
	for i, actor := range []authz.Principal{a, b} {
		wg.Go(func() {
			errs[i] = e.promote().Execute(e.ctx, promoteIn(actor, target, promoReason, domain.RoleEventos))
		})
	}
	wg.Wait()

	ok, already := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, domain.ErrAlreadyAdmin):
			already++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if ok != 1 || already != 1 || len(e.events(t, "admin.promote")) != 1 {
		t.Errorf("sucessos = %d, already_admin = %d, eventos = %d", ok, already, len(e.events(t, "admin.promote")))
	}
}

// AUD-01.2
func TestPromoteIsRolledBackWhenTheAuditFails(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	target := e.newTarget(t)
	e.sessionFor(t, target)
	e.breakAudit(t)

	err := e.promote().Execute(e.ctx, promoteIn(presidente, target, promoReason, domain.RoleEventos))

	if !errors.Is(err, audit.ErrWrite) {
		t.Errorf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), target); ok || e.activeSessions(t, target) != 1 {
		t.Error("vínculo, papéis, senha e sessões deveriam continuar como estavam")
	}
	if got := e.rolesOf(t, target); len(got) != 1 {
		t.Errorf("papéis = %v", got)
	}
	if stored, _ := e.users.FindByID(context.Background(), target); stored.MustChangePassword {
		t.Error("must_change_password não pode ter mudado")
	}
}

func TestPromoteRequiresAllDependencies(t *testing.T) {
	if err := (&app.PromoteToAdmin{}).Execute(context.Background(), app.PromoteInput{}); err == nil || errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v", err)
	}
}
