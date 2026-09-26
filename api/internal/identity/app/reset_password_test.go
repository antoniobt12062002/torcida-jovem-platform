//go:build integration

package app_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
)

func (e *ucEnv) resetWithToken() *app.ResetPasswordWithToken {
	return &app.ResetPasswordWithToken{
		Recovery: e.recovery, Users: e.users, Roles: e.roles, Sessions: e.sessions, Passwords: e.hasher, Denylist: e.deny,
		Audit: e.rec, Tx: e.tx, Now: e.clock,
	}
}

// issueToken creates a pending token for the user and returns the plain value.
func (e *ucEnv) issueToken(t *testing.T, userID string) string {
	t.Helper()
	token, hash, _ := domain.NewSessionToken()
	if err := e.recovery.CreateToken(context.Background(), userID, hash, e.now, e.now.Add(30*time.Minute), ""); err != nil {
		t.Fatal(err)
	}
	return token
}

const newPassword = "uma-senha-nova-e-forte-9"

// IDN-07.5: redefinição completa numa transação.
func TestResetWithATokenChangesThePasswordConsumesTheTokenRevokesSessionsAndAudits(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleTesouraria) // must_change_password = true
	e.sessionFor(t, u.ID)
	e.sessionFor(t, u.ID)
	token := e.issueToken(t, u.ID)

	err := e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: token, NewPassword: newPassword})

	if err != nil {
		t.Fatal(err)
	}
	if !e.passwordIs(t, u.ID, newPassword) || e.passwordIs(t, u.ID, currentPassword) {
		t.Error("a senha deveria ter sido trocada")
	}
	if stored, _ := e.users.FindByID(context.Background(), u.ID); stored.MustChangePassword {
		t.Error("must_change_password deveria ter sido limpo")
	}
	if e.activeSessions(t, u.ID) != 0 {
		t.Error("todas as sessões deveriam ter sido revogadas")
	}
	if err := e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: token, NewPassword: newPassword + "x"}); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("o token é de uso único: %v", err)
	}
	evs := e.events(t, "auth.password_reset_completed")
	if len(evs) != 1 || evs[0].Outcome != "success" || evs[0].ActorType != "user" || evs[0].ActorID == nil || *evs[0].ActorID != u.ID || evs[0].EntityID != u.ID {
		t.Fatalf("eventos = %+v", evs)
	}
	blob := evs[0].Context
	if evs[0].Before != nil {
		blob += *evs[0].Before
	}
	if evs[0].After != nil {
		blob += *evs[0].After
	}
	for _, secret := range []string{token, newPassword, "argon2"} {
		if strings.Contains(blob, secret) {
			t.Errorf("a auditoria não pode conter %q", secret)
		}
	}
}

func TestResetInvalidatesTheOtherPendingTokensOfTheUser(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)
	first := e.issueToken(t, u.ID)
	// dois tokens pendentes ao mesmo tempo: o segundo gravado direto, sem invalidar o primeiro
	other, otherHash, _ := domain.NewSessionToken()
	_ = e.owner.Exec(`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES (?::uuid, ?, ?)`, u.ID, otherHash, e.now.Add(time.Hour)).Error

	if err := e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: first, NewPassword: newPassword}); err != nil {
		t.Fatal(err)
	}

	if err := e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: other, NewPassword: newPassword + "y"}); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("o outro token pendente deveria ter sido invalidado: %v", err)
	}
}

// IDN-07.6 e .8: o mesmo erro para desconhecido, expirado e usado, com evento de segurança.
func TestUnknownExpiredAndUsedTokensGetTheSameErrorAndAFailureEvent(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)
	expired := e.issueToken(t, u.ID)
	used := e.issueToken(t, u.ID)
	_ = e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: used, NewPassword: newPassword})
	e.now = t0.Add(time.Hour)
	before := len(e.events(t, "auth.password_reset_failed"))

	var msgs []string
	for name, token := range map[string]string{"desconhecido": "token-que-nao-existe", "expirado": expired, "usado": used, "vazio": "", "enorme": strings.Repeat("x", 5000)} {
		err := e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: token, NewPassword: newPassword + "z"})
		if !errors.Is(err, domain.ErrInvalidResetToken) {
			t.Errorf("%s: %v", name, err)
			continue
		}
		msgs = append(msgs, err.Error())
	}

	for _, m := range msgs {
		if m != msgs[0] {
			t.Errorf("as mensagens deveriam ser idênticas: %q vs %q", m, msgs[0])
		}
	}
	evs := e.events(t, "auth.password_reset_failed")
	if len(evs)-before != 5 {
		t.Fatalf("eventos novos = %d, esperado 5", len(evs)-before)
	}
	for _, ev := range evs[before:] {
		if ev.Outcome != "failure" || ev.ActorType != "anonymous" || strings.Contains(ev.Context, "token-que-nao") || strings.Contains(ev.Context, expired) {
			t.Errorf("evento = %+v", ev)
		}
	}
	if !e.passwordIs(t, u.ID, newPassword) {
		t.Error("as tentativas inválidas não podem alterar a senha já redefinida")
	}
}

// IDN-07.7: violação de política mantém o token válido.
func TestPolicyViolationsKeepTheTokenValidAndChangeNothing(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleTesouraria)
	e.sessionFor(t, u.ID)
	token := e.issueToken(t, u.ID)
	uc := e.resetWithToken()

	for pw, want := range map[string]domain.PasswordViolation{
		"curta-123":              domain.PasswordTooShort,
		strings.Repeat("a", 129): domain.PasswordTooLong,
		"senhacomum1234":         domain.PasswordCompromised,
	} {
		err := uc.Execute(e.ctx, app.ResetInput{Token: token, NewPassword: pw})
		if pe, ok := errors.AsType[*domain.PasswordError](err); !ok || pe.Violation != want {
			t.Errorf("%.12q: err = %v, esperado %s", pw, err, want)
		}
	}

	if e.activeSessions(t, u.ID) != 1 || !e.passwordIs(t, u.ID, currentPassword) {
		t.Error("nada pode ter mudado")
	}
	if len(e.events(t, "auth.password_reset_failed")) != 0 {
		t.Error("uma senha fraca não é falha de token")
	}
	if err := uc.Execute(e.ctx, app.ResetInput{Token: token, NewPassword: newPassword}); err != nil {
		t.Errorf("o token ainda deveria valer: %v", err)
	}
}

// IDN-07.9
func TestAnInactiveUsersTokenIsTreatedAsInvalid(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)
	token := e.issueToken(t, u.ID)
	_ = e.users.SetActive(context.Background(), u.ID, false)

	err := e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: token, NewPassword: newPassword})

	if !errors.Is(err, domain.ErrInvalidResetToken) || !e.passwordIs(t, u.ID, currentPassword) {
		t.Errorf("err = %v", err)
	}
}

// IDN-07.5: confirmações simultâneas, só uma vence.
func TestSimultaneousConfirmationsOfTheSameTokenLeaveExactlyOneWinner(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)
	token := e.issueToken(t, u.ID)
	const attempts = 5
	errs := make([]error, attempts)

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Go(func() {
			errs[i] = e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: token, NewPassword: newPassword + itoa(i)})
		})
	}
	wg.Wait()

	winners := 0
	for _, err := range errs {
		switch {
		case err == nil:
			winners++
		case !errors.Is(err, domain.ErrInvalidResetToken):
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if winners != 1 || len(e.events(t, "auth.password_reset_completed")) != 1 {
		t.Errorf("vencedores = %d, eventos = %d", winners, len(e.events(t, "auth.password_reset_completed")))
	}
	if len(e.events(t, "auth.password_reset_failed")) != attempts-1 {
		t.Errorf("as perdedoras geram auth.password_reset_failed: %d", len(e.events(t, "auth.password_reset_failed")))
	}
}

// AUD-01.2: sem auditoria, nada muda, e o token continua valendo.
func TestResetIsRolledBackWhenTheAuditFails(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)
	e.sessionFor(t, u.ID)
	token := e.issueToken(t, u.ID)
	e.breakAudit(t)

	err := e.resetWithToken().Execute(e.ctx, app.ResetInput{Token: token, NewPassword: newPassword})

	if !errors.Is(err, audit.ErrWrite) {
		t.Fatalf("err = %v", err)
	}
	if !e.passwordIs(t, u.ID, currentPassword) || e.activeSessions(t, u.ID) != 1 {
		t.Error("a senha e as sessões deveriam continuar como estavam")
	}
	if _, err := e.recovery.FindValid(context.Background(), sha256of(token), e.now); err != nil {
		t.Errorf("o token deveria continuar válido: %v", err)
	}
}

func TestResetRequiresAllDependencies(t *testing.T) {
	if err := (&app.ResetPasswordWithToken{}).Execute(context.Background(), app.ResetInput{}); err == nil || errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("err = %v", err)
	}
}

// barrierTokens makes every caller pass FindValid before any of them can go on,
// so the single-use guarantee is decided by Consume and not by lucky timing.
type barrierTokens struct {
	app.ResetTokens
	arrive sync.WaitGroup
}

func (b *barrierTokens) FindValid(ctx context.Context, hash []byte, now time.Time) (string, error) {
	id, err := b.ResetTokens.FindValid(ctx, hash, now)
	b.arrive.Done()
	b.arrive.Wait()
	return id, err
}

func TestSingleUseHoldsEvenWhenEveryConfirmationSawTheTokenAsValid(t *testing.T) {
	e := newUCEnv(t)
	u := e.changer(t, domain.RoleAssociado)
	token := e.issueToken(t, u.ID)
	const attempts = 4
	barrier := &barrierTokens{ResetTokens: e.recovery}
	barrier.arrive.Add(attempts)
	errs := make([]error, attempts)

	var wg sync.WaitGroup
	for i := range attempts {
		uc := e.resetWithToken()
		uc.Recovery = barrier
		wg.Go(func() { errs[i] = uc.Execute(e.ctx, app.ResetInput{Token: token, NewPassword: newPassword + itoa(i)}) })
	}
	wg.Wait()

	winners := 0
	for _, err := range errs {
		if err == nil {
			winners++
		} else if !errors.Is(err, domain.ErrInvalidResetToken) {
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if winners != 1 || len(e.events(t, "auth.password_reset_completed")) != 1 || len(e.events(t, "auth.password_reset_failed")) != attempts-1 {
		t.Errorf("vencedores = %d, completed = %d, failed = %d", winners,
			len(e.events(t, "auth.password_reset_completed")), len(e.events(t, "auth.password_reset_failed")))
	}
}
