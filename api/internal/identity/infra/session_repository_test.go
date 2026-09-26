//go:build integration

package infra_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

type sessionEnv struct {
	env
	users    *infra.UserRepository
	sessions *infra.SessionRepository
}

func newSessionEnv(t *testing.T) sessionEnv {
	t.Helper()
	e := newEnv(t)
	return sessionEnv{env: e, users: infra.NewUserRepository(e.app), sessions: infra.NewSessionRepository(e.app)}
}

func (s sessionEnv) open(t *testing.T, userID string) (token string, sess domain.Session) {
	t.Helper()
	token, hash, err := domain.NewSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	csrf, _ := domain.NewCSRFToken()
	sess, err = s.sessions.Create(context.Background(), userID, hash, csrf, time.Now().UTC().Add(8*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return token, sess
}

func (s sessionEnv) user(t *testing.T, email string) domain.User {
	t.Helper()
	u, err := s.users.Create(context.Background(), newDomainUser(email))
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// IDN-02.2: só o hash SHA-256 do token é gravado.
func TestCreateStoresOnlyTheSHA256OfTheToken(t *testing.T) {
	s := newSessionEnv(t)
	u := s.user(t, "ana@exemplo.com")

	token, sess := s.open(t, u.ID)

	var raw string
	if err := s.owner.Raw("SELECT row_to_json(s)::text FROM sessions s").Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	if strings.Contains(raw, token) {
		t.Errorf("o token nunca pode estar no banco: %s", raw)
	}
	var stored []byte
	if err := s.owner.Raw("SELECT token_hash FROM sessions WHERE id = ?::uuid", sess.ID).Row().Scan(&stored); err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256([]byte(token))
	if string(stored) != string(want[:]) {
		t.Error("token_hash deveria ser o SHA-256 do token")
	}
	if sess.UserID != u.ID || sess.CSRFToken == "" || sess.RevokedAt != nil || sess.CreatedAt.IsZero() || !sess.LastSeenAt.Equal(sess.CreatedAt) {
		t.Errorf("sessão = %+v", sess)
	}
}

func TestFindByTokenUsesTheHashAndReturnsNotFoundForUnknownTokens(t *testing.T) {
	s := newSessionEnv(t)
	u := s.user(t, "ana@exemplo.com")
	token, sess := s.open(t, u.ID)

	got, err := s.sessions.FindByToken(context.Background(), token)
	if err != nil || got.ID != sess.ID || got.CSRFToken != sess.CSRFToken || got.UserID != u.ID {
		t.Errorf("FindByToken = %+v, %v", got, err)
	}
	for _, bad := range []string{"", "token-inexistente", token + "x"} {
		if _, err := s.sessions.FindByToken(context.Background(), bad); !errors.Is(err, domain.ErrSessionNotFound) {
			t.Errorf("FindByToken(%q): %v", bad, err)
		}
	}
}

func TestRevokeMarksTheSessionAndKeepsFindingIt(t *testing.T) {
	s := newSessionEnv(t)
	u := s.user(t, "ana@exemplo.com")
	token, sess := s.open(t, u.ID)

	if err := s.sessions.Revoke(context.Background(), sess.ID); err != nil {
		t.Fatal(err)
	}
	got, err := s.sessions.FindByToken(context.Background(), token)

	if err != nil || got.RevokedAt == nil {
		t.Errorf("a sessão revogada continua consultável, com RevokedAt: %+v, %v", got, err)
	}
	if err := s.sessions.Revoke(context.Background(), sess.ID); err != nil {
		t.Errorf("revogar de novo é inofensivo: %v", err)
	}
	if err := s.sessions.Revoke(context.Background(), "não-é-uuid"); err != nil {
		t.Errorf("id malformado é inofensivo: %v", err)
	}
}

// IDN-02.6: revogar todas as sessões do usuário, em uma transação.
func TestRevokeAllForUserRevokesOnlyThatUsersActiveSessions(t *testing.T) {
	s := newSessionEnv(t)
	ana, bia := s.user(t, "ana@exemplo.com"), s.user(t, "bia@exemplo.com")
	s.open(t, ana.ID)
	s.open(t, ana.ID)
	tokenBia, _ := s.open(t, bia.ID)

	n, err := s.sessions.RevokeAllForUser(context.Background(), ana.ID)

	if err != nil || n != 2 {
		t.Fatalf("revogadas = %d, %v", n, err)
	}
	var active int64
	_ = s.owner.Raw("SELECT count(*) FROM sessions WHERE user_id = ?::uuid AND revoked_at IS NULL", ana.ID).Scan(&active).Error
	if active != 0 {
		t.Errorf("sessões ativas de Ana = %d", active)
	}
	if got, err := s.sessions.FindByToken(context.Background(), tokenBia); err != nil || got.RevokedAt != nil {
		t.Errorf("a sessão de Bia não pode ser afetada: %+v, %v", got, err)
	}
	if n, _ := s.sessions.RevokeAllForUser(context.Background(), ana.ID); n != 0 {
		t.Errorf("uma segunda revogação não encontra sessões ativas: %d", n)
	}
}

func TestRevokeAllForUserJoinsTheTransactionOfTheContext(t *testing.T) {
	s := newSessionEnv(t)
	u := s.user(t, "ana@exemplo.com")
	s.open(t, u.ID)
	boom := errors.New("boom")

	err := database.WithTx(context.Background(), s.app, func(ctx context.Context) error {
		if _, err := s.sessions.RevokeAllForUser(ctx, u.ID); err != nil {
			return err
		}
		return boom
	})

	var active int64
	_ = s.owner.Raw("SELECT count(*) FROM sessions WHERE revoked_at IS NULL").Scan(&active).Error
	if !errors.Is(err, boom) || active != 1 {
		t.Errorf("err = %v, ativas = %d: a revogação deveria ter sido revertida junto", err, active)
	}
}

func TestRevokeAllExceptKeepsTheGivenSession(t *testing.T) {
	s := newSessionEnv(t)
	u := s.user(t, "ana@exemplo.com")
	_, keep := s.open(t, u.ID)
	other, _ := s.open(t, u.ID)

	n, err := s.sessions.RevokeAllForUserExcept(context.Background(), u.ID, keep.ID)

	if err != nil || n != 1 {
		t.Fatalf("revogadas = %d, %v", n, err)
	}
	if got, _ := s.sessions.FindByToken(context.Background(), other); got.RevokedAt == nil {
		t.Error("a outra sessão deveria ter sido revogada")
	}
	var kept int64
	_ = s.owner.Raw("SELECT count(*) FROM sessions WHERE id = ?::uuid AND revoked_at IS NULL", keep.ID).Scan(&kept).Error
	if kept != 1 {
		t.Error("a sessão preservada não pode ser revogada")
	}
}

// IDN-02.5 e a atualização limitada: no máximo uma vez por minuto.
func TestTouchUpdatesLastSeenAtAtMostOncePerInterval(t *testing.T) {
	s := newSessionEnv(t)
	u := s.user(t, "ana@exemplo.com")
	token, sess := s.open(t, u.ID)
	base := sess.LastSeenAt

	updated, err := s.sessions.Touch(context.Background(), sess.ID, base.Add(30*time.Second), time.Minute)
	if err != nil || updated {
		t.Errorf("30 segundos depois não deveria atualizar: %v, %v", updated, err)
	}
	updated, err = s.sessions.Touch(context.Background(), sess.ID, base.Add(61*time.Second), time.Minute)
	if err != nil || !updated {
		t.Errorf("61 segundos depois deveria atualizar: %v, %v", updated, err)
	}
	got, _ := s.sessions.FindByToken(context.Background(), token)
	if !got.LastSeenAt.Equal(base.Add(61 * time.Second)) {
		t.Errorf("LastSeenAt = %v", got.LastSeenAt)
	}
	if updated, _ := s.sessions.Touch(context.Background(), sess.ID, base.Add(90*time.Second), time.Minute); updated {
		t.Error("29 segundos depois da última atualização não deveria atualizar de novo")
	}
}

func TestTouchNeverRevivesARevokedSession(t *testing.T) {
	s := newSessionEnv(t)
	u := s.user(t, "ana@exemplo.com")
	_, sess := s.open(t, u.ID)
	_ = s.sessions.Revoke(context.Background(), sess.ID)

	if updated, err := s.sessions.Touch(context.Background(), sess.ID, sess.LastSeenAt.Add(time.Hour), time.Minute); err != nil || updated {
		t.Errorf("Touch em sessão revogada: %v, %v", updated, err)
	}
}
