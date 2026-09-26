//go:build integration

package infra_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
)

var t0 = time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

func newAttemptRepo(t *testing.T) (env, *infra.AttemptRepository) {
	t.Helper()
	e := newEnv(t)
	return e, infra.NewAttemptRepository(e.app)
}

func hashOf(t *testing.T, email string) []byte {
	t.Helper()
	h, err := domain.HashEmail([]byte("0123456789abcdef0123456789abcdef"), email)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func fail(t *testing.T, repo *infra.AttemptRepository, h []byte, times ...time.Duration) {
	t.Helper()
	for _, d := range times {
		if err := repo.RecordFailure(context.Background(), h, t0.Add(d)); err != nil {
			t.Fatal(err)
		}
	}
}

// IDN-02.8: conta as falhas por hash do e-mail na janela de 15 minutos.
func TestFourFailuresDoNotBlockAndTheFifthDoes(t *testing.T) {
	_, repo := newAttemptRepo(t)
	h := hashOf(t, "ana@exemplo.com")
	fail(t, repo, h, 0, time.Minute, 2*time.Minute, 3*time.Minute)

	if _, blocked, err := repo.Blocked(context.Background(), h, t0.Add(3*time.Minute)); err != nil || blocked {
		t.Fatalf("4 falhas: blocked = %v, err = %v", blocked, err)
	}
	fail(t, repo, h, 4*time.Minute)

	until, blocked, err := repo.Blocked(context.Background(), h, t0.Add(4*time.Minute))
	if err != nil || !blocked || !until.Equal(t0.Add(19*time.Minute)) {
		t.Errorf("5 falhas: blocked = %v, until = %v, err = %v", blocked, until, err)
	}
	if _, blocked, _ := repo.Blocked(context.Background(), h, t0.Add(19*time.Minute)); blocked {
		t.Error("o bloqueio termina 15 minutos depois da quinta falha")
	}
}

func TestFailuresOutsideTheWindowDoNotAddUp(t *testing.T) {
	_, repo := newAttemptRepo(t)
	h := hashOf(t, "ana@exemplo.com")
	fail(t, repo, h, 0, 4*time.Minute, 8*time.Minute, 12*time.Minute, 16*time.Minute)

	if _, blocked, err := repo.Blocked(context.Background(), h, t0.Add(16*time.Minute)); err != nil || blocked {
		t.Errorf("blocked = %v, err = %v", blocked, err)
	}
}

// IDN-02.8: o sucesso zera a contagem.
func TestSuccessResetsTheCount(t *testing.T) {
	e, repo := newAttemptRepo(t)
	h := hashOf(t, "ana@exemplo.com")
	fail(t, repo, h, 0, time.Minute, 2*time.Minute, 3*time.Minute)

	if err := repo.RecordSuccess(context.Background(), h); err != nil {
		t.Fatal(err)
	}
	fail(t, repo, h, 4*time.Minute)

	if _, blocked, _ := repo.Blocked(context.Background(), h, t0.Add(4*time.Minute)); blocked {
		t.Error("depois de um sucesso, a contagem recomeça do zero")
	}
	if n := count(t, e.owner, "login_attempts"); n != 1 {
		t.Errorf("as falhas anteriores ao sucesso deveriam ter sido removidas, linhas = %d", n)
	}
}

func TestDifferentEmailsHaveIndependentCounters(t *testing.T) {
	_, repo := newAttemptRepo(t)
	ana, bia := hashOf(t, "ana@exemplo.com"), hashOf(t, "bia@exemplo.com")
	fail(t, repo, ana, 0, time.Minute, 2*time.Minute, 3*time.Minute, 4*time.Minute)

	if _, blocked, _ := repo.Blocked(context.Background(), bia, t0.Add(4*time.Minute)); blocked {
		t.Error("o bloqueio de Ana não pode atingir Bia")
	}
	if _, blocked, _ := repo.Blocked(context.Background(), ana, t0.Add(4*time.Minute)); !blocked {
		t.Error("Ana deveria estar bloqueada")
	}
	if err := repo.RecordSuccess(context.Background(), bia); err != nil {
		t.Fatal(err)
	}
	if _, blocked, _ := repo.Blocked(context.Background(), ana, t0.Add(4*time.Minute)); !blocked {
		t.Error("o sucesso de Bia não pode zerar a contagem de Ana")
	}
}

func TestNoAttemptsMeansNotBlocked(t *testing.T) {
	_, repo := newAttemptRepo(t)

	if _, blocked, err := repo.Blocked(context.Background(), hashOf(t, "ninguem@exemplo.com"), t0); err != nil || blocked {
		t.Errorf("blocked = %v, err = %v", blocked, err)
	}
}

// Só o hash do e-mail fica no banco.
func TestOnlyTheEmailHashIsStored(t *testing.T) {
	e, repo := newAttemptRepo(t)
	fail(t, repo, hashOf(t, "ana@exemplo.com"), 0)

	var raw string
	if err := e.owner.Raw("SELECT row_to_json(a)::text FROM login_attempts a").Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	if raw == "" || strings.Contains(raw, "ana@exemplo.com") || strings.Contains(raw, "exemplo") {
		t.Errorf("o e-mail não pode estar na linha: %s", raw)
	}
}
