//go:build integration

package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func (e *ucEnv) revoke() *app.RevokeAdmin {
	return &app.RevokeAdmin{
		Authz: e.authorizer, Users: e.users, Members: e.members, Perms: e.roles, Sessions: e.sessions, Audit: e.rec, Tx: e.tx, Now: e.clock,
	}
}

func revokeIn(actor authz.Principal, target, reason string) app.RevokeInput {
	return app.RevokeInput{Actor: actor, TargetID: target, Reason: reason}
}

// IDN-06.6: retirada completa numa transação.
func TestRevokeClosesMembershipKeepsOnlyAssociadoRevokesSessionsAndAudits(t *testing.T) {
	e := newUCEnv(t)
	presidenteU, presidente := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria, domain.RoleEventos)
	e.sessionFor(t, target.ID)

	err := e.revoke().Execute(e.ctx, revokeIn(presidente, target.ID, "fim do mandato"))

	if err != nil {
		t.Fatal(err)
	}
	if got := e.rolesOf(t, target.ID); len(got) != 1 || got[0] != domain.RoleAssociado {
		t.Errorf("papéis = %v: só ASSOCIADO deve ficar", got)
	}
	if _, ok, _ := e.members.Active(context.Background(), target.ID); ok {
		t.Error("o vínculo deveria estar encerrado")
	}
	history, _ := e.members.History(context.Background(), target.ID)
	if len(history) != 1 || history[0].RevokeReason == nil || *history[0].RevokeReason != "fim do mandato" || history[0].RevokedBy == nil || *history[0].RevokedBy != presidenteU.ID {
		t.Errorf("o histórico deveria registrar quem, quando e por quê: %+v", history)
	}
	if e.activeSessions(t, target.ID) != 0 {
		t.Error("as sessões deveriam ter sido revogadas")
	}
	stored, _ := e.users.FindByID(context.Background(), target.ID)
	if !stored.Active {
		t.Error("retirar o acesso não desativa a conta")
	}
	evs := e.events(t, "admin.revoke")
	if len(evs) != 1 || evs[0].Outcome != "success" || evs[0].Reason == nil || *evs[0].Reason != "fim do mandato" || evs[0].EntityID != target.ID {
		t.Errorf("eventos = %+v", evs)
	}
}

func TestRevokeChecksThePermissionFirst(t *testing.T) {
	e := newUCEnv(t)
	_, diretoria := e.actor(t, domain.RoleDiretoria)
	target, _ := e.actor(t, domain.RoleTesouraria)

	err := e.revoke().Execute(e.ctx, revokeIn(diretoria, target.ID, "fim do mandato"))

	if !errors.Is(err, authz.ErrForbidden) || len(e.events(t, "authz.denied")) != 1 {
		t.Errorf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), target.ID); !ok {
		t.Error("nada pode ter mudado")
	}
}

// RBAC-03.3
func TestNobodyRevokesTheirOwnAccess(t *testing.T) {
	e := newUCEnv(t)
	u, presidente := e.actor(t, domain.RolePresidente)

	err := e.revoke().Execute(e.ctx, revokeIn(presidente, u.ID, "fim do mandato"))

	if !errors.Is(err, domain.ErrSelfChangeForbidden) {
		t.Fatalf("err = %v", err)
	}
	if evs := e.events(t, "role.change_denied"); len(evs) != 1 || evs[0].Outcome != "denied" {
		t.Errorf("eventos = %+v", evs)
	}
	if _, ok, _ := e.members.Active(context.Background(), u.ID); !ok {
		t.Error("nada pode ter mudado")
	}
}

// RBAC-03.2: não se retira o acesso de quem tem permissões que o ator não tem.
func TestRevokeRefusesATargetWithPermissionsTheActorLacks(t *testing.T) {
	e := newUCEnv(t)
	_, adminSistema := e.actor(t, domain.RoleAdminSistema)
	conselheiro, _ := e.actor(t, domain.RoleConselhoFiscal)

	err := e.revoke().Execute(e.ctx, revokeIn(adminSistema, conselheiro.ID, "fim do mandato"))

	if !errors.Is(err, domain.ErrPrivilegeEscalation) {
		t.Fatalf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), conselheiro.ID); !ok {
		t.Error("nada pode ter mudado")
	}
	if evs := e.events(t, "role.change_denied"); len(evs) != 1 || evs[0].Outcome != "denied" {
		t.Errorf("eventos = %+v", evs)
	}
}

func TestRevokeRequiresAnActiveMembershipAndAReason(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	plain, _ := e.actor(t) // ASSOCIADO, sem vínculo
	admin, _ := e.actor(t, domain.RoleTesouraria)

	if err := e.revoke().Execute(e.ctx, revokeIn(presidente, plain.ID, "fim do mandato")); !errors.Is(err, domain.ErrNotAdmin) {
		t.Errorf("sem vínculo: %v", err)
	}
	if err := e.revoke().Execute(e.ctx, revokeIn(presidente, admin.ID, "   ")); !errors.Is(err, domain.ErrReasonRequired) {
		t.Errorf("sem motivo: %v", err)
	}
	if err := e.revoke().Execute(e.ctx, revokeIn(presidente, "0f8fad5b-d9cb-469f-a165-70867728950e", "fim do mandato")); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("alvo inexistente: %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), admin.ID); !ok {
		t.Error("um pedido recusado não pode encerrar o vínculo")
	}
}

// IDN-06.8
func TestRevokeRefusesTheLastActiveHolderOfAdminGrant(t *testing.T) {
	e := newUCEnv(t)
	only, _ := e.actor(t, domain.RoleAdminSistema)
	actor := e.withPermsOf(t, domain.RolePresidente)

	err := e.revoke().Execute(e.ctx, revokeIn(actor, only.ID, "fim do mandato"))

	if !errors.Is(err, domain.ErrLastAdmin) {
		t.Fatalf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), only.ID); !ok || len(e.events(t, "admin.revoke")) != 0 {
		t.Error("nada pode ter mudado nem sido auditado")
	}
}

func TestRevokingAnAdministratorWhoDoesNotHoldAdminGrantIsAlwaysAllowed(t *testing.T) {
	e := newUCEnv(t)
	only, _ := e.actor(t, domain.RoleAdminSistema)
	tesoureiro, _ := e.actor(t, domain.RoleTesouraria)
	actor := e.withPermsOf(t, domain.RolePresidente)

	if err := e.revoke().Execute(e.ctx, revokeIn(actor, tesoureiro.ID, "fim do mandato")); err != nil {
		t.Errorf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), only.ID); !ok {
		t.Error("o único administrador continua")
	}
}

func TestSimultaneousRevocationsNeverLeaveTheSystemWithoutAnAdministrator(t *testing.T) {
	e := newUCEnv(t)
	a, _ := e.actor(t, domain.RoleAdminSistema)
	b, _ := e.actor(t, domain.RoleAdminSistema)
	actor := e.withPermsOf(t, domain.RolePresidente)
	errs := make([]error, 2)

	var wg sync.WaitGroup
	for i, id := range []string{a.ID, b.ID} {
		wg.Go(func() { errs[i] = e.revoke().Execute(e.ctx, revokeIn(actor, id, "fim do mandato")) })
	}
	wg.Wait()

	ok, last := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, domain.ErrLastAdmin):
			last++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if ok != 1 || last != 1 {
		t.Errorf("sucessos = %d, last_admin = %d", ok, last)
	}
	if n, _ := e.users.ActiveHoldersOf(context.Background(), "identity:admin:grant", ""); n != 1 {
		t.Errorf("titulares = %d", n)
	}
}

// AUD-01.2
func TestRevokeIsRolledBackWhenTheAuditFails(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)
	e.sessionFor(t, target.ID)
	e.breakAudit(t)

	err := e.revoke().Execute(e.ctx, revokeIn(presidente, target.ID, "fim do mandato"))

	if !errors.Is(err, audit.ErrWrite) {
		t.Errorf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), target.ID); !ok || e.activeSessions(t, target.ID) != 1 {
		t.Error("vínculo e sessões deveriam continuar como estavam")
	}
	if got := e.rolesOf(t, target.ID); len(got) != 2 {
		t.Errorf("papéis = %v", got)
	}
}

func TestRevokeRequiresAllDependencies(t *testing.T) {
	if err := (&app.RevokeAdmin{}).Execute(context.Background(), app.RevokeInput{}); err == nil || errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v", err)
	}
}
