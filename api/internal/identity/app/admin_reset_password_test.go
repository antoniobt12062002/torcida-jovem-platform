//go:build integration

package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

const resetReason = "usuária perdeu o acesso ao e-mail"

func (e *ucEnv) adminReset() *app.AdminResetPassword {
	return &app.AdminResetPassword{
		Authz: e.authorizer, Users: e.users, Perms: e.roles, Roles: e.roles, Sessions: e.sessions, Recovery: e.recovery,
		Passwords: e.hasher, Denylist: e.deny, Audit: e.rec, Tx: e.tx, Now: e.clock,
	}
}

func resetIn(actor authz.Principal, target, reason string) app.AdminResetInput {
	return app.AdminResetInput{Actor: actor, TargetID: target, Reason: reason}
}

// IDN-08.1 e IDN-05.8: senha temporária mostrada uma vez, troca obrigatória, sessões e tokens revogados.
func TestAdminResetSetsATemporaryPasswordForcesChangeRevokesEverythingAndAudits(t *testing.T) {
	e := newUCEnv(t)
	adminU, admin := e.actor(t, domain.RoleAdminSistema)
	target := e.changer(t, domain.RoleTesouraria)
	_ = e.users.SetMustChangePassword(context.Background(), target.ID, false)
	e.sessionFor(t, target.ID)
	e.sessionFor(t, target.ID)
	pending := e.issueToken(t, target.ID)

	res, err := e.adminReset().Execute(e.ctx, resetIn(admin, target.ID, "  "+resetReason+"  "))

	if err != nil {
		t.Fatal(err)
	}
	if res.TemporaryPassword == "" || !e.passwordIs(t, target.ID, res.TemporaryPassword) || e.passwordIs(t, target.ID, currentPassword) {
		t.Error("a senha temporária deveria estar em vigor e a antiga não")
	}
	if stored, _ := e.users.FindByID(context.Background(), target.ID); !stored.MustChangePassword {
		t.Error("must_change_password deveria ser verdadeiro")
	}
	if e.activeSessions(t, target.ID) != 0 {
		t.Error("as sessões deveriam ter sido revogadas")
	}
	if _, err := e.recovery.FindValid(context.Background(), sha256of(pending), e.now); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("o token de recuperação pendente deveria ter sido invalidado: %v", err)
	}
	evs := e.events(t, "user.password_reset")
	if len(evs) != 1 || evs[0].Outcome != "success" || evs[0].EntityID != target.ID || evs[0].ActorID == nil || *evs[0].ActorID != adminU.ID ||
		evs[0].Reason == nil || *evs[0].Reason != resetReason {
		t.Fatalf("eventos = %+v", evs)
	}
	blob := evs[0].Context + e.logs.String()
	if evs[0].Before != nil {
		blob += *evs[0].Before
	}
	if evs[0].After != nil {
		blob += *evs[0].After
	}
	for _, secret := range []string{res.TemporaryPassword, currentPassword, "argon2"} {
		if strings.Contains(blob, secret) {
			t.Errorf("nada disso pode estar na auditoria ou no log: %q", secret)
		}
	}
}

func TestEachAdminResetGeneratesADifferentPassword(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RoleAdminSistema)
	target := e.changer(t, domain.RoleAssociado)

	first, _ := e.adminReset().Execute(e.ctx, resetIn(admin, target.ID, resetReason))
	second, _ := e.adminReset().Execute(e.ctx, resetIn(admin, target.ID, resetReason))

	if first.TemporaryPassword == "" || first.TemporaryPassword == second.TemporaryPassword {
		t.Errorf("senhas = %q / %q", first.TemporaryPassword, second.TemporaryPassword)
	}
	if e.passwordIs(t, target.ID, first.TemporaryPassword) {
		t.Error("só a última senha temporária vale")
	}
}

func TestAdminResetChecksThePermissionFirst(t *testing.T) {
	e := newUCEnv(t)
	_, diretoria := e.actor(t, domain.RoleDiretoria)
	target := e.changer(t, domain.RoleAssociado)

	res, err := e.adminReset().Execute(e.ctx, resetIn(diretoria, target.ID, "x"))

	if !errors.Is(err, authz.ErrForbidden) || res.TemporaryPassword != "" || len(e.events(t, "authz.denied")) != 1 {
		t.Errorf("res = %+v, err = %v", res, err)
	}
	if !e.passwordIs(t, target.ID, currentPassword) {
		t.Error("nada pode ter mudado")
	}
}

// RBAC-03.3
func TestNobodyResetsTheirOwnPasswordThroughTheAdministrativePath(t *testing.T) {
	e := newUCEnv(t)
	u, admin := e.actor(t, domain.RoleAdminSistema)

	res, err := e.adminReset().Execute(e.ctx, resetIn(admin, u.ID, resetReason))

	if !errors.Is(err, domain.ErrSelfChangeForbidden) || res.TemporaryPassword != "" {
		t.Fatalf("res = %+v, err = %v", res, err)
	}
	if evs := e.events(t, "user.password_reset"); len(evs) != 1 || evs[0].Outcome != "denied" {
		t.Errorf("eventos = %+v", evs)
	}
}

// RBAC-03.6
func TestAdminResetRefusesATargetWithPermissionsTheActorLacks(t *testing.T) {
	e := newUCEnv(t)
	_, adminSistema := e.actor(t, domain.RoleAdminSistema)
	presidente := e.changer(t, domain.RolePresidente)

	res, err := e.adminReset().Execute(e.ctx, resetIn(adminSistema, presidente.ID, resetReason))

	if !errors.Is(err, domain.ErrPrivilegeEscalation) || res.TemporaryPassword != "" {
		t.Fatalf("res = %+v, err = %v", res, err)
	}
	if !e.passwordIs(t, presidente.ID, currentPassword) {
		t.Error("nada pode ter mudado")
	}
	evs := e.events(t, "user.password_reset")
	if len(evs) != 1 || evs[0].Outcome != "denied" || evs[0].EntityID != presidente.ID {
		t.Errorf("eventos = %+v", evs)
	}
	_, presidenteActor := e.actor(t, domain.RolePresidente)
	admin := e.changer(t, domain.RoleAdminSistema)
	if _, err := e.adminReset().Execute(e.ctx, resetIn(presidenteActor, admin.ID, resetReason)); err != nil {
		t.Errorf("o PRESIDENTE cobre as permissões do ADMIN_SISTEMA: %v", err)
	}
}

// IDN-08.6
func TestAdminResetRequiresAReasonOfAtLeastTenCharacters(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RoleAdminSistema)
	target := e.changer(t, domain.RoleAssociado)

	for _, reason := range []string{"", "   ", "curto", "123456789"} {
		if _, err := e.adminReset().Execute(e.ctx, resetIn(admin, target.ID, reason)); !errors.Is(err, domain.ErrReasonRequired) {
			t.Errorf("motivo %q: %v", reason, err)
		}
	}
	if !e.passwordIs(t, target.ID, currentPassword) {
		t.Error("nada pode ter mudado")
	}
	if _, err := e.adminReset().Execute(e.ctx, resetIn(admin, target.ID, "1234567890")); err != nil {
		t.Errorf("10 caracteres bastam: %v", err)
	}
}

// IDN-08.4
func TestAdminResetRefusesInactiveAndUnknownTargets(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RoleAdminSistema)
	off := e.changer(t, domain.RoleAssociado)
	_ = e.users.SetActive(context.Background(), off.ID, false)

	if _, err := e.adminReset().Execute(e.ctx, resetIn(admin, off.ID, resetReason)); !errors.Is(err, domain.ErrUserInactive) {
		t.Errorf("inativo: %v", err)
	}
	if _, err := e.adminReset().Execute(e.ctx, resetIn(admin, "0f8fad5b-d9cb-469f-a165-70867728950e", resetReason)); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("inexistente: %v", err)
	}
}

// AUD-01.2: sem auditoria, nada muda e nenhuma senha é entregue.
func TestAdminResetIsRolledBackWhenTheAuditFailsAndDeliversNoPassword(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RoleAdminSistema)
	target := e.changer(t, domain.RoleAssociado)
	e.sessionFor(t, target.ID)
	e.breakAudit(t)

	res, err := e.adminReset().Execute(e.ctx, resetIn(admin, target.ID, resetReason))

	if !errors.Is(err, audit.ErrWrite) || res.TemporaryPassword != "" {
		t.Errorf("res = %+v, err = %v", res, err)
	}
	if !e.passwordIs(t, target.ID, currentPassword) || e.activeSessions(t, target.ID) != 1 {
		t.Error("a senha e as sessões deveriam continuar como estavam")
	}
}

func TestAdminResetRequiresAllDependencies(t *testing.T) {
	if _, err := (&app.AdminResetPassword{}).Execute(context.Background(), app.AdminResetInput{}); err == nil || errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v", err)
	}
}
