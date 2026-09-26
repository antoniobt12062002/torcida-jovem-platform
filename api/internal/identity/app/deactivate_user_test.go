//go:build integration

package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func (e *ucEnv) activation() *app.UserActivation {
	return &app.UserActivation{
		Authz: e.authorizer, Users: e.users, Perms: e.roles, Sessions: e.sessions, Recovery: e.recovery, Audit: e.rec,
		Tx: e.tx, Now: e.clock,
	}
}

// withPermsOf is a principal for an unrelated user that carries the permissions
// of the given roles, as a session would: used to act on the only administrator.
func (e *ucEnv) withPermsOf(t *testing.T, roles ...domain.Role) authz.Principal {
	t.Helper()
	u, _ := e.actor(t) // ASSOCIADO only: holds no permission in the database
	p := authz.Principal{UserID: u.ID, Permissions: map[authz.Permission]struct{}{}}
	for _, perm := range e.matrix.PermissionsOf(roles) {
		p.Permissions[perm] = struct{}{}
	}
	return p
}

func isActive(t *testing.T, e *ucEnv, id string) bool {
	t.Helper()
	u, err := e.users.FindByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	return u.Active
}

// IDN-04.3: desativar revoga sessões e tokens de recuperação, e audita.
func TestDeactivateRevokesSessionsAndPendingRecoveryTokensAndAudits(t *testing.T) {
	e := newUCEnv(t)
	adminU, admin := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)
	token, _ := e.sessionFor(t, target.ID)
	_ = e.recovery.CreateToken(context.Background(), target.ID, sha256of("tok"), e.now, e.now.Add(30*time.Minute), "")

	err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: admin, TargetID: target.ID})

	if err != nil {
		t.Fatal(err)
	}
	if isActive(t, e, target.ID) || e.activeSessions(t, target.ID) != 0 {
		t.Error("o usuário deveria estar inativo e sem sessões ativas")
	}
	if _, err := e.recovery.FindValid(context.Background(), sha256of("tok"), e.now); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("o token de recuperação pendente deveria ter sido invalidado: %v", err)
	}
	_ = token
	evs := e.events(t, "user.deactivate")
	if len(evs) != 1 || evs[0].Outcome != "success" || evs[0].EntityID != target.ID || evs[0].ActorID == nil || *evs[0].ActorID != adminU.ID {
		t.Errorf("eventos = %+v", evs)
	}
}

func TestDeactivateChecksThePermissionFirst(t *testing.T) {
	e := newUCEnv(t)
	_, noPerm := e.actor(t, domain.RoleDiretoria)
	target, _ := e.actor(t, domain.RoleTesouraria)

	err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: noPerm, TargetID: target.ID})

	if !errors.Is(err, authz.ErrForbidden) || !isActive(t, e, target.ID) || len(e.events(t, "authz.denied")) != 1 {
		t.Errorf("err = %v", err)
	}
}

// RBAC-03.3: ninguém desativa a própria conta.
func TestNobodyDeactivatesTheirOwnAccount(t *testing.T) {
	e := newUCEnv(t)
	u, admin := e.actor(t, domain.RolePresidente)

	err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: admin, TargetID: u.ID})

	if !errors.Is(err, domain.ErrSelfChangeForbidden) || !isActive(t, e, u.ID) {
		t.Fatalf("err = %v", err)
	}
	evs := e.events(t, "user.deactivate")
	if len(evs) != 1 || evs[0].Outcome != "denied" {
		t.Errorf("a negação deveria ser um evento de segurança: %+v", evs)
	}
}

// RBAC-03.6: a regra sem escalada vale para desativar.
func TestDeactivateRefusesATargetWithPermissionsTheActorLacks(t *testing.T) {
	e := newUCEnv(t)
	_, adminSistema := e.actor(t, domain.RoleAdminSistema) // tem identity:user:update, mas não as institucionais
	presidente, _ := e.actor(t, domain.RolePresidente)

	err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: adminSistema, TargetID: presidente.ID})

	if !errors.Is(err, domain.ErrPrivilegeEscalation) || !isActive(t, e, presidente.ID) {
		t.Fatalf("err = %v", err)
	}
	evs := e.events(t, "user.deactivate")
	if len(evs) != 1 || evs[0].Outcome != "denied" || evs[0].EntityID != presidente.ID {
		t.Errorf("eventos = %+v", evs)
	}
}

func TestPresidenteMayDeactivateAnAdminSistema(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	admin, _ := e.actor(t, domain.RoleAdminSistema)

	if err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: presidente, TargetID: admin.ID}); err != nil {
		t.Errorf("o PRESIDENTE cobre as permissões do ADMIN_SISTEMA: %v", err)
	}
}

// IDN-04.4: nunca fica sem quem administre acesso.
func TestDeactivateRefusesTheLastActiveHolderOfAdminGrant(t *testing.T) {
	e := newUCEnv(t)
	only, _ := e.actor(t, domain.RoleAdminSistema)
	actor := e.withPermsOf(t, domain.RolePresidente)

	err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: actor, TargetID: only.ID})

	if !errors.Is(err, domain.ErrLastAdmin) || !isActive(t, e, only.ID) {
		t.Errorf("err = %v", err)
	}
	if len(e.events(t, "user.deactivate")) != 0 {
		t.Error("uma operação recusada por last_admin não gera auditoria")
	}
}

// O lock do conjunto de administradores impede duas desativações simultâneas de zerarem o conjunto.
func TestSimultaneousDeactivationsNeverLeaveTheSystemWithoutAnAdministrator(t *testing.T) {
	e := newUCEnv(t)
	a, _ := e.actor(t, domain.RoleAdminSistema)
	b, _ := e.actor(t, domain.RoleAdminSistema)
	actor := e.withPermsOf(t, domain.RolePresidente)
	errs := make([]error, 2)

	var wg sync.WaitGroup
	for i, id := range []string{a.ID, b.ID} {
		wg.Go(func() { errs[i] = e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: actor, TargetID: id}) })
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
		t.Errorf("titulares ativos = %d, deveria sobrar exatamente 1", n)
	}
}

func TestDeactivatingAnAlreadyInactiveUserIsANoOp(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)
	_ = e.users.SetActive(context.Background(), target.ID, false)

	if err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: admin, TargetID: target.ID}); err != nil {
		t.Errorf("err = %v", err)
	}
	if len(e.events(t, "user.deactivate")) != 0 {
		t.Error("um no-op não audita")
	}
}

func TestDeactivateUnknownTargetReturnsNotFound(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RolePresidente)

	for _, id := range []string{"0f8fad5b-d9cb-469f-a165-70867728950e", "não-é-uuid"} {
		if err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: admin, TargetID: id}); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("%q: %v", id, err)
		}
	}
}

// AUD-01.2: sem auditoria, nada muda.
func TestDeactivateIsRolledBackWhenTheAuditFails(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)
	e.sessionFor(t, target.ID)
	e.breakAudit(t)

	err := e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: admin, TargetID: target.ID})

	if !errors.Is(err, audit.ErrWrite) || !isActive(t, e, target.ID) || e.activeSessions(t, target.ID) != 1 {
		t.Errorf("err = %v: o usuário e as sessões deveriam continuar como estavam", err)
	}
}

// IDN-04.10: reativar não restaura privilégios.
func TestReactivateAuditsAndDoesNotRestoreAClosedMembership(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)
	_, _ = e.members.Revoke(context.Background(), target.ID, admin.UserID, "fim do mandato", e.now)
	_ = e.users.SetRoles(context.Background(), target.ID, []domain.Role{domain.RoleAssociado})
	_ = e.activation().Deactivate(e.ctx, app.ActivationInput{Actor: admin, TargetID: target.ID})

	err := e.activation().Reactivate(e.ctx, app.ActivationInput{Actor: admin, TargetID: target.ID})

	if err != nil || !isActive(t, e, target.ID) {
		t.Fatalf("err = %v", err)
	}
	if _, ok, _ := e.members.Active(context.Background(), target.ID); ok {
		t.Error("o vínculo encerrado não pode voltar")
	}
	if got := e.rolesOf(t, target.ID); len(got) != 1 || got[0] != domain.RoleAssociado {
		t.Errorf("papéis = %v", got)
	}
	if evs := e.events(t, "user.reactivate"); len(evs) != 1 || evs[0].Outcome != "success" {
		t.Errorf("eventos = %+v", evs)
	}
}

func TestReactivateFollowsTheSameEscalationAndSelfRules(t *testing.T) {
	e := newUCEnv(t)
	_, adminSistema := e.actor(t, domain.RoleAdminSistema)
	presidente, _ := e.actor(t, domain.RolePresidente)
	_ = e.users.SetActive(context.Background(), presidente.ID, false)

	err := e.activation().Reactivate(e.ctx, app.ActivationInput{Actor: adminSistema, TargetID: presidente.ID})

	if !errors.Is(err, domain.ErrPrivilegeEscalation) || isActive(t, e, presidente.ID) {
		t.Errorf("err = %v", err)
	}
	evs := e.events(t, "user.reactivate")
	if len(evs) != 1 || evs[0].Outcome != "denied" {
		t.Errorf("eventos = %+v", evs)
	}
	if err := e.activation().Reactivate(e.ctx, app.ActivationInput{Actor: adminSistema, TargetID: adminSistema.UserID}); !errors.Is(err, domain.ErrSelfChangeForbidden) {
		t.Errorf("autorreativação: %v", err)
	}
}

func TestUserActivationRequiresAllDependencies(t *testing.T) {
	if err := (&app.UserActivation{}).Deactivate(context.Background(), app.ActivationInput{}); err == nil || errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v", err)
	}
}
