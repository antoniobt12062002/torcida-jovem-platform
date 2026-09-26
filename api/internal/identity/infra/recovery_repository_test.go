//go:build integration

package infra_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
)

type recoveryEnv struct {
	env
	users *infra.UserRepository
	repo  *infra.RecoveryRepository
}

func newRecoveryEnv(t *testing.T) recoveryEnv {
	t.Helper()
	e := newEnv(t)
	return recoveryEnv{env: e, users: infra.NewUserRepository(e.app), repo: infra.NewRecoveryRepository(e.app)}
}

func (r recoveryEnv) user(t *testing.T, email string) domain.User {
	t.Helper()
	u, err := r.users.Create(context.Background(), newDomainUser(email))
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func sha(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// IDN-07.2: só o hash do token é gravado.
func TestCreateTokenStoresOnlyTheSHA256(t *testing.T) {
	r := newRecoveryEnv(t)
	u := r.user(t, "ana@exemplo.com")

	if err := r.repo.CreateToken(context.Background(), u.ID, sha("segredo-do-token"), t0, t0.Add(30*time.Minute), "req-12345678"); err != nil {
		t.Fatal(err)
	}

	var raw string
	if err := r.owner.Raw("SELECT row_to_json(p)::text FROM password_reset_tokens p").Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	if strings.Contains(raw, "segredo-do-token") {
		t.Errorf("o token não pode estar no banco: %s", raw)
	}
}

// IDN-07.5: uso único.
func TestConsumeReturnsTheUserOnceAndThenRefuses(t *testing.T) {
	r := newRecoveryEnv(t)
	u := r.user(t, "ana@exemplo.com")
	_ = r.repo.CreateToken(context.Background(), u.ID, sha("t1"), t0, t0.Add(30*time.Minute), "")

	got, err := r.repo.Consume(context.Background(), sha("t1"), t0.Add(time.Minute))
	if err != nil || got != u.ID {
		t.Fatalf("Consume = %q, %v", got, err)
	}
	if _, err := r.repo.Consume(context.Background(), sha("t1"), t0.Add(2*time.Minute)); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("segundo uso: %v", err)
	}
}

// IDN-07.6: o mesmo erro para desconhecido, expirado e usado.
func TestUnknownExpiredAndUsedTokensReturnTheSameError(t *testing.T) {
	r := newRecoveryEnv(t)
	u := r.user(t, "ana@exemplo.com")
	_ = r.repo.CreateToken(context.Background(), u.ID, sha("expira"), t0, t0.Add(30*time.Minute), "")
	_ = r.repo.CreateToken(context.Background(), u.ID, sha("usado"), t0, t0.Add(30*time.Minute), "")
	_, _ = r.repo.Consume(context.Background(), sha("usado"), t0)

	var msgs []string
	for name, hash := range map[string][]byte{"desconhecido": sha("nao-existe"), "expirado": sha("expira"), "usado": sha("usado")} {
		_, errConsume := r.repo.Consume(context.Background(), hash, t0.Add(31*time.Minute))
		_, errFind := r.repo.FindValid(context.Background(), hash, t0.Add(31*time.Minute))
		if !errors.Is(errConsume, domain.ErrInvalidResetToken) || !errors.Is(errFind, domain.ErrInvalidResetToken) {
			t.Errorf("%s: %v / %v", name, errConsume, errFind)
			continue
		}
		msgs = append(msgs, errConsume.Error())
	}
	for _, m := range msgs {
		if m != msgs[0] {
			t.Errorf("as mensagens deveriam ser idênticas: %q vs %q", m, msgs[0])
		}
	}
}

func TestFindValidDoesNotConsumeTheToken(t *testing.T) {
	r := newRecoveryEnv(t)
	u := r.user(t, "ana@exemplo.com")
	_ = r.repo.CreateToken(context.Background(), u.ID, sha("t1"), t0, t0.Add(30*time.Minute), "")

	for range 2 {
		if id, err := r.repo.FindValid(context.Background(), sha("t1"), t0.Add(time.Minute)); err != nil || id != u.ID {
			t.Fatalf("FindValid = %q, %v", id, err)
		}
	}
	if _, err := r.repo.Consume(context.Background(), sha("t1"), t0.Add(time.Minute)); err != nil {
		t.Errorf("o token ainda deveria valer: %v", err)
	}
}

func TestTokenExpiresExactlyAtItsExpiry(t *testing.T) {
	r := newRecoveryEnv(t)
	u := r.user(t, "ana@exemplo.com")
	_ = r.repo.CreateToken(context.Background(), u.ID, sha("t1"), t0, t0.Add(30*time.Minute), "")

	if _, err := r.repo.FindValid(context.Background(), sha("t1"), t0.Add(30*time.Minute-time.Second)); err != nil {
		t.Errorf("1 segundo antes de expirar: %v", err)
	}
	if _, err := r.repo.FindValid(context.Background(), sha("t1"), t0.Add(30*time.Minute)); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("no instante da expiração: %v", err)
	}
}

// IDN-07.2: criar um token invalida os pendentes do mesmo usuário, e só deles.
func TestCreatingATokenInvalidatesThePendingOnesOfTheSameUserOnly(t *testing.T) {
	r := newRecoveryEnv(t)
	ana, bia := r.user(t, "ana@exemplo.com"), r.user(t, "bia@exemplo.com")
	_ = r.repo.CreateToken(context.Background(), ana.ID, sha("ana-1"), t0, t0.Add(30*time.Minute), "")
	_ = r.repo.CreateToken(context.Background(), bia.ID, sha("bia-1"), t0, t0.Add(30*time.Minute), "")

	if err := r.repo.CreateToken(context.Background(), ana.ID, sha("ana-2"), t0.Add(time.Minute), t0.Add(31*time.Minute), ""); err != nil {
		t.Fatal(err)
	}

	if _, err := r.repo.FindValid(context.Background(), sha("ana-1"), t0.Add(2*time.Minute)); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("o token antigo de Ana deveria ter sido invalidado: %v", err)
	}
	if _, err := r.repo.FindValid(context.Background(), sha("ana-2"), t0.Add(2*time.Minute)); err != nil {
		t.Errorf("o token novo de Ana vale: %v", err)
	}
	if _, err := r.repo.FindValid(context.Background(), sha("bia-1"), t0.Add(2*time.Minute)); err != nil {
		t.Errorf("o token de Bia não pode ser afetado: %v", err)
	}
}

func TestInvalidatePendingClosesEveryPendingTokenOfTheUserAndKeepsTheRows(t *testing.T) {
	r := newRecoveryEnv(t)
	u := r.user(t, "ana@exemplo.com")
	_ = r.repo.CreateToken(context.Background(), u.ID, sha("t1"), t0, t0.Add(30*time.Minute), "")

	if err := r.repo.InvalidatePending(context.Background(), u.ID, t0.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	if _, err := r.repo.FindValid(context.Background(), sha("t1"), t0.Add(2*time.Minute)); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("err = %v", err)
	}
	if n := count(t, r.owner, "password_reset_tokens"); n != 1 {
		t.Errorf("as linhas são preservadas, n = %d", n)
	}
}

// IDN-07.5: duas confirmações simultâneas, só uma vence.
func TestConcurrentConsumesOfTheSameTokenLeaveExactlyOneWinner(t *testing.T) {
	r := newRecoveryEnv(t)
	u := r.user(t, "ana@exemplo.com")
	_ = r.repo.CreateToken(context.Background(), u.ID, sha("t1"), t0, t0.Add(30*time.Minute), "")
	const attempts = 6
	errs := make([]error, attempts)

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Go(func() { _, errs[i] = r.repo.Consume(context.Background(), sha("t1"), t0.Add(time.Minute)) })
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
	if winners != 1 {
		t.Errorf("vencedores = %d", winners)
	}
}

// IDN-07.4: no máximo 3 solicitações por e-mail por hora.
func TestRegisterRequestAllowsThreePerHourPerEmailHash(t *testing.T) {
	r := newRecoveryEnv(t)
	ana, bia := hashOf(t, "ana@exemplo.com"), hashOf(t, "bia@exemplo.com")

	for i := range 3 {
		allowed, err := r.repo.RegisterRequest(context.Background(), ana, t0.Add(time.Duration(i)*time.Minute))
		if err != nil || !allowed {
			t.Fatalf("solicitação %d: allowed = %v, err = %v", i+1, allowed, err)
		}
	}
	if allowed, _ := r.repo.RegisterRequest(context.Background(), ana, t0.Add(5*time.Minute)); allowed {
		t.Error("a quarta solicitação na mesma hora não é permitida")
	}
	if allowed, _ := r.repo.RegisterRequest(context.Background(), bia, t0.Add(5*time.Minute)); !allowed {
		t.Error("o limite é por e-mail")
	}
	if allowed, _ := r.repo.RegisterRequest(context.Background(), ana, t0.Add(61*time.Minute)); !allowed {
		t.Error("depois de uma hora as solicitações voltam a ser permitidas")
	}
}

func TestOnlyTheEmailHMACIsStoredInRequests(t *testing.T) {
	r := newRecoveryEnv(t)
	_, _ = r.repo.RegisterRequest(context.Background(), hashOf(t, "ana@exemplo.com"), t0)

	var raw string
	_ = r.owner.Raw("SELECT row_to_json(q)::text FROM password_reset_requests q").Scan(&raw).Error
	if raw == "" || strings.Contains(raw, "exemplo") {
		t.Errorf("linha = %q", raw)
	}
}

// IDN-04.12: contador próprio da troca de senha, por usuário, separado do login.
func TestPasswordChangeLockoutIsPerUserAndSeparateFromLoginAttempts(t *testing.T) {
	r := newRecoveryEnv(t)
	ana, bia := r.user(t, "ana@exemplo.com"), r.user(t, "bia@exemplo.com")
	attempts := infra.NewAttemptRepository(r.app)
	for i := range 5 {
		_ = attempts.RecordFailure(context.Background(), hashOf(t, "ana@exemplo.com"), t0.Add(time.Duration(i)*time.Second))
	}

	if _, blocked, _ := r.repo.PasswordChangeBlocked(context.Background(), ana.ID, t0.Add(10*time.Second)); blocked {
		t.Fatal("falhas de login não podem bloquear a troca de senha")
	}
	for i := range 4 {
		_ = r.repo.RecordPasswordChangeFailure(context.Background(), ana.ID, t0.Add(time.Duration(i)*time.Minute))
	}
	if _, blocked, _ := r.repo.PasswordChangeBlocked(context.Background(), ana.ID, t0.Add(3*time.Minute)); blocked {
		t.Error("4 erros não bloqueiam")
	}
	_ = r.repo.RecordPasswordChangeFailure(context.Background(), ana.ID, t0.Add(4*time.Minute))

	until, blocked, err := r.repo.PasswordChangeBlocked(context.Background(), ana.ID, t0.Add(4*time.Minute))
	if err != nil || !blocked || !until.Equal(t0.Add(19*time.Minute)) {
		t.Errorf("blocked = %v, until = %v, err = %v", blocked, until, err)
	}
	if _, blocked, _ := r.repo.PasswordChangeBlocked(context.Background(), bia.ID, t0.Add(4*time.Minute)); blocked {
		t.Error("o bloqueio de Ana não atinge Bia")
	}
	if _, blocked, _ := r.repo.PasswordChangeBlocked(context.Background(), ana.ID, t0.Add(19*time.Minute)); blocked {
		t.Error("o bloqueio termina 15 minutos depois do quinto erro")
	}
	if err := r.repo.ClearPasswordChangeFailures(context.Background(), ana.ID); err != nil {
		t.Fatal(err)
	}
	if _, blocked, _ := r.repo.PasswordChangeBlocked(context.Background(), ana.ID, t0.Add(4*time.Minute)); blocked {
		t.Error("o sucesso zera a contagem")
	}
}
