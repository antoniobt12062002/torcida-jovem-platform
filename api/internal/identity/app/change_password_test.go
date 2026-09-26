//go:build integration

package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
)

const currentPassword = "senha-atual-correta-1"

func (e *ucEnv) changePassword() *app.ChangePassword {
	return &app.ChangePassword{
		Users: e.users, Roles: e.roles, Throttle: e.recovery, Sessions: e.sessions, Passwords: e.hasher, Denylist: e.deny,
		Audit: e.rec, Tx: e.tx, Now: e.clock,
	}
}

// changer returns a user with the given roles whose password is currentPassword.
func (e *ucEnv) changer(t *testing.T, roles ...domain.Role) domain.User {
	t.Helper()
	u, _ := e.actor(t, roles...)
	hash, _ := e.hasher.Hash(currentPassword)
	if err := e.users.SetPassword(context.Background(), u.ID, hash, true); err != nil {
		t.Fatal(err)
	}
	return u
}

func (e *ucEnv) passwordIs(t *testing.T, userID, plain string) bool {
	t.Helper()
	u, err := e.users.FindByID(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	ok, _ := e.hasher.Verify(plain, u.PasswordHash)
	return ok
}

// IDN-04.5: troca de senha bem-sucedida.
func TestChangePasswordUpdatesTheHashClearsMustChangeRevokesOtherSessionsAndAudits(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleTesouraria)
	_, current := e.sessionFor(t, u.ID)
	e.sessionFor(t, u.ID)
	e.sessionFor(t, u.ID)

	err := e.changePassword().Execute(e.ctx, app.ChangePasswordInput{UserID: u.ID, CurrentSessionID: current.ID, Current: currentPassword, New: "uma-senha-nova-e-forte-2"})

	if err != nil {
		t.Fatal(err)
	}
	if !e.passwordIs(t, u.ID, "uma-senha-nova-e-forte-2") || e.passwordIs(t, u.ID, currentPassword) {
		t.Error("o hash deveria ter sido trocado")
	}
	stored, _ := e.users.FindByID(context.Background(), u.ID)
	if stored.MustChangePassword {
		t.Error("must_change_password deveria ter sido limpo (IDN-05.6)")
	}
	if e.activeSessions(t, u.ID) != 1 {
		t.Errorf("só a sessão atual deveria continuar: %d", e.activeSessions(t, u.ID))
	}
	evs := e.events(t, "user.password_change")
	if len(evs) != 1 || evs[0].Outcome != "success" || evs[0].ActorID == nil || *evs[0].ActorID != u.ID {
		t.Fatalf("eventos = %+v", evs)
	}
	for _, s := range []string{"uma-senha-nova", currentPassword, "argon2"} {
		if (evs[0].After != nil && strings.Contains(*evs[0].After, s)) || strings.Contains(evs[0].Context, s) {
			t.Errorf("a auditoria não pode conter %q", s)
		}
	}
}

// IDN-04.6: senha atual errada.
func TestWrongCurrentPasswordIsRefusedCountedAndAudited(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleTesouraria)
	_, current := e.sessionFor(t, u.ID)
	e.sessionFor(t, u.ID)

	err := e.changePassword().Execute(e.ctx, app.ChangePasswordInput{UserID: u.ID, CurrentSessionID: current.ID, Current: "senha-errada-123", New: "uma-senha-nova-e-forte-2"})

	if !errors.Is(err, domain.ErrInvalidCurrentPassword) {
		t.Fatalf("err = %v", err)
	}
	if !e.passwordIs(t, u.ID, currentPassword) || e.activeSessions(t, u.ID) != 2 {
		t.Error("nada pode mudar")
	}
	if e.count(t, "password_change_attempts") != 1 || e.count(t, "login_attempts") != 0 {
		t.Error("o erro conta no contador próprio, não no do login")
	}
	if evs := e.events(t, "user.password_change"); len(evs) != 1 || evs[0].Outcome != "failure" {
		t.Errorf("eventos = %+v", evs)
	}
}

// IDN-04.12: cinco erros em 15 minutos bloqueiam a troca por 15 minutos.
func TestFiveWrongCurrentPasswordsBlockThePasswordChangeTemporarily(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)
	uc := e.changePassword()
	in := func(cur string) app.ChangePasswordInput {
		return app.ChangePasswordInput{UserID: u.ID, Current: cur, New: "uma-senha-nova-e-forte-2"}
	}
	for i := range 5 {
		e.now = t0.Add(time.Duration(i) * time.Minute)
		if err := uc.Execute(e.ctx, in("errada-"+itoa(i)+"-xx")); !errors.Is(err, domain.ErrInvalidCurrentPassword) {
			t.Fatalf("erro %d: %v", i+1, err)
		}
	}
	before := e.count(t, "password_change_attempts")

	e.now = t0.Add(5 * time.Minute)
	err := uc.Execute(e.ctx, in(currentPassword)) // até a senha certa é recusada durante o bloqueio
	blocked, ok := errors.AsType[*app.PasswordChangeBlockedError](err)
	if !ok || blocked.RetryAfter != 14*time.Minute {
		t.Fatalf("err = %v", err)
	}
	if e.count(t, "password_change_attempts") != before {
		t.Error("tentativas bloqueadas não contam nem estendem o bloqueio")
	}
	if !e.passwordIs(t, u.ID, currentPassword) {
		t.Error("nada pode mudar durante o bloqueio")
	}

	e.now = t0.Add(19 * time.Minute)
	if err := uc.Execute(e.ctx, in(currentPassword)); err != nil {
		t.Errorf("passados 15 minutos a troca volta a funcionar: %v", err)
	}
	if e.count(t, "password_change_attempts") != 0 {
		t.Error("o sucesso zera a contagem")
	}
}

// Nova senha igual à atual.
func TestNewPasswordEqualToTheCurrentOneIsRefusedWithoutCounting(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)

	err := e.changePassword().Execute(e.ctx, app.ChangePasswordInput{UserID: u.ID, Current: currentPassword, New: currentPassword})

	if !errors.Is(err, domain.ErrPasswordUnchanged) {
		t.Errorf("err = %v", err)
	}
	if e.count(t, "password_change_attempts") != 0 {
		t.Error("recusar a mesma senha não é erro de senha atual")
	}
}

// IDN-05.1 a .4: política por papel; violação não conta como erro e não revoga sessões.
func TestChangePasswordAppliesThePolicyByRoleWithoutSideEffects(t *testing.T) {
	e := newUCEnv(t)
	admin := e.changer(t, domain.RoleTesouraria)
	assoc := e.changer(t, domain.RoleAssociado)
	e.sessionFor(t, admin.ID)
	uc := e.changePassword()

	violation := func(id, newPassword string) domain.PasswordViolation {
		err := uc.Execute(e.ctx, app.ChangePasswordInput{UserID: id, Current: currentPassword, New: newPassword})
		if pe, ok := errors.AsType[*domain.PasswordError](err); ok {
			return pe.Violation
		}
		return ""
	}
	if v := violation(admin.ID, "curta-123"); v != domain.PasswordTooShort { // 9 caracteres
		t.Errorf("9 caracteres para papel administrativo: %q", v)
	}
	if v := violation(assoc.ID, "1234567"); v != domain.PasswordTooShort {
		t.Errorf("7 caracteres para ASSOCIADO: %q", v)
	}
	if v := violation(admin.ID, strings.Repeat("a", 129)); v != domain.PasswordTooLong {
		t.Errorf("129 caracteres: %q", v)
	}
	if v := violation(admin.ID, "senhacomum1234"); v != domain.PasswordCompromised {
		t.Errorf("senha comprometida: %q", v)
	}
	if e.count(t, "password_change_attempts") != 0 || e.activeSessions(t, admin.ID) != 1 || !e.passwordIs(t, admin.ID, currentPassword) {
		t.Error("violações de política não contam, não revogam sessões e não trocam a senha")
	}
	if err := uc.Execute(e.ctx, app.ChangePasswordInput{UserID: assoc.ID, Current: currentPassword, New: "8chars!!"}); err != nil {
		t.Errorf("8 caracteres bastam para ASSOCIADO: %v", err)
	}
}

// AUD-01.2
func TestChangePasswordIsRolledBackWhenTheAuditFails(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)
	_, current := e.sessionFor(t, u.ID)
	e.sessionFor(t, u.ID)
	e.breakAudit(t)

	err := e.changePassword().Execute(e.ctx, app.ChangePasswordInput{UserID: u.ID, CurrentSessionID: current.ID, Current: currentPassword, New: "uma-senha-nova-e-forte-2"})

	if !errors.Is(err, audit.ErrWrite) {
		t.Errorf("err = %v", err)
	}
	if !e.passwordIs(t, u.ID, currentPassword) || e.activeSessions(t, u.ID) != 2 {
		t.Error("a senha e as sessões deveriam continuar como estavam")
	}
}

func TestChangePasswordForAnUnknownUserAndMissingDependencies(t *testing.T) {
	e := newUCEnv(t)

	err := e.changePassword().Execute(e.ctx, app.ChangePasswordInput{UserID: "0f8fad5b-d9cb-469f-a165-70867728950e", Current: "x", New: "y"})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("err = %v", err)
	}
	if err := (&app.ChangePassword{}).Execute(e.ctx, app.ChangePasswordInput{}); err == nil {
		t.Error("um caso de uso mal montado deveria falhar")
	}
}
