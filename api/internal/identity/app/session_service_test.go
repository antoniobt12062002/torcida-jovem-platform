//go:build integration

package app_test

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

type sessionEnv struct {
	db, owner *gorm.DB
	users     *infra.UserRepository
	sessions  *infra.SessionRepository
	roles     *infra.RoleRepository
	svc       *app.SessionService
	now       time.Time
}

func newSessionEnv(t *testing.T) *sessionEnv {
	t.Helper()
	db := testutil.NewTestDB(t)
	rec := audit.NewRecorder(db, logx.New("error", &bytes.Buffer{}))
	e := &sessionEnv{
		db: db, owner: testutil.OwnerDBFor(t, db), now: t0,
		users: infra.NewUserRepository(db), sessions: infra.NewSessionRepository(db), roles: infra.NewRoleRepository(db, rec),
	}
	e.svc = &app.SessionService{
		Sessions: e.sessions, Users: e.users, Roles: e.roles, Audit: rec,
		Idle: 60 * time.Minute, TouchInterval: time.Minute, Now: func() time.Time { return e.now },
	}
	m, err := app.BuildMatrix(append(app.FoundationContributions(), financeiro.Contribution())...)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.roles.Sync(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	return e
}

// open creates an active user with the given roles and a session that lives 8 hours from now.
func (e *sessionEnv) open(t *testing.T, email string, roles ...domain.Role) (domain.User, string, domain.Session) {
	t.Helper()
	u, err := e.users.Create(context.Background(), domain.User{Email: email, Name: "Fulano", PasswordHash: "h", Active: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) == 0 {
		roles = []domain.Role{domain.RoleAssociado}
	}
	if err := e.users.SetRoles(context.Background(), u.ID, roles); err != nil {
		t.Fatal(err)
	}
	token, hash, _ := domain.NewSessionToken()
	csrf, _ := domain.NewCSRFToken()
	s, err := e.sessions.Create(context.Background(), u.ID, hash, csrf, e.now, e.now.Add(8*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return u, token, s
}

func (e *sessionEnv) lastSeen(t *testing.T, token string) time.Time {
	t.Helper()
	s, err := e.sessions.FindByToken(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	return s.LastSeenAt
}

func TestValidateReturnsThePrincipalTheCSRFTokenAndTheUser(t *testing.T) {
	e := newSessionEnv(t)
	u, token, sess := e.open(t, "ana@exemplo.com", domain.RoleConselhoFiscal, domain.RoleDiretoria)

	got, err := e.svc.Validate(context.Background(), token)

	if err != nil {
		t.Fatal(err)
	}
	if got.Principal.UserID != u.ID || got.CSRFToken != sess.CSRFToken || got.SessionID != sess.ID || got.User.Email != "ana@exemplo.com" || got.MustChangePassword {
		t.Errorf("resultado = %+v", got)
	}
	if !slices.Equal(got.Principal.Roles, []string{"CONSELHO_FISCAL", "DIRETORIA"}) {
		t.Errorf("papéis = %v", got.Principal.Roles)
	}
	for _, p := range []authz.Permission{"audit:log:read", "financeiro:parecer:opine", "identity:user:read"} {
		if !got.Principal.Has(p) {
			t.Errorf("o principal deveria ter %s", p)
		}
	}
	if got.Principal.Has("identity:user:create") {
		t.Error("o principal não pode ter permissões de papéis que não possui")
	}
}

func TestValidateReportsMustChangePassword(t *testing.T) {
	e := newSessionEnv(t)
	u, token, _ := e.open(t, "ana@exemplo.com")
	_ = e.users.SetPassword(context.Background(), u.ID, "h", true)

	got, err := e.svc.Validate(context.Background(), token)

	if err != nil || !got.MustChangePassword {
		t.Errorf("resultado = %+v, err = %v", got, err)
	}
}

func TestValidateRejectsMissingUnknownAndMalformedTokens(t *testing.T) {
	e := newSessionEnv(t)
	e.open(t, "ana@exemplo.com")

	for _, token := range []string{"", "token-inexistente", "%%%", string(make([]byte, 5000))} {
		if _, err := e.svc.Validate(context.Background(), token); !errors.Is(err, app.ErrUnauthenticated) {
			t.Errorf("token %.10q: err = %v, esperava ErrUnauthenticated", token, err)
		}
	}
}

// IDN-02.5: mais de 60 minutos sem uso, ou mais de 8 horas, expira com session_expired.
func TestValidateExpiresAfterTheIdleLimit(t *testing.T) {
	e := newSessionEnv(t)
	_, token, _ := e.open(t, "ana@exemplo.com")

	e.now = t0.Add(60 * time.Minute)
	if _, err := e.svc.Validate(context.Background(), token); err != nil {
		t.Fatalf("exatamente 60 minutos ainda vale: %v", err)
	}
	e.now = t0.Add(60*time.Minute + 61*time.Minute)
	if _, err := e.svc.Validate(context.Background(), token); !errors.Is(err, app.ErrSessionExpired) {
		t.Errorf("mais de 60 minutos sem uso: %v", err)
	}
}

func TestValidateExpiresAfterTheAbsoluteLimitEvenWithConstantUse(t *testing.T) {
	e := newSessionEnv(t)
	_, token, _ := e.open(t, "ana@exemplo.com")

	for i := 1; i <= 8*2; i++ { // uso a cada 30 minutos durante 8 horas
		e.now = t0.Add(time.Duration(i) * 30 * time.Minute)
		if _, err := e.svc.Validate(context.Background(), token); err != nil {
			t.Fatalf("aos %d minutos: %v", i*30, err)
		}
	}
	e.now = t0.Add(8*time.Hour + time.Second)

	if _, err := e.svc.Validate(context.Background(), token); !errors.Is(err, app.ErrSessionExpired) {
		t.Errorf("depois de 8 horas: %v", err)
	}
}

// last_seen_at é atualizado no máximo uma vez por minuto, e o uso renova a ociosidade.
func TestValidateTouchesLastSeenAtAtMostOncePerMinute(t *testing.T) {
	e := newSessionEnv(t)
	_, token, _ := e.open(t, "ana@exemplo.com")
	created := e.lastSeen(t, token)

	e.now = created.Add(30 * time.Second)
	_, _ = e.svc.Validate(context.Background(), token)
	if got := e.lastSeen(t, token); !got.Equal(created) {
		t.Errorf("30 segundos depois não deveria escrever: %v -> %v", created, got)
	}
	e.now = created.Add(61 * time.Second)
	_, _ = e.svc.Validate(context.Background(), token)
	if got := e.lastSeen(t, token); !got.Equal(e.now) {
		t.Errorf("61 segundos depois deveria atualizar: %v", got)
	}
}

func TestUseRenewsTheIdleLimit(t *testing.T) {
	e := newSessionEnv(t)
	_, token, _ := e.open(t, "ana@exemplo.com")
	e.now = t0.Add(50 * time.Minute)
	_, _ = e.svc.Validate(context.Background(), token)

	e.now = t0.Add(50*time.Minute + 55*time.Minute) // 105 min desde a criação, 55 desde o último uso

	if _, err := e.svc.Validate(context.Background(), token); err != nil {
		t.Errorf("o uso renova a ociosidade: %v", err)
	}
}

// RBAC-02.5: usuário inativo invalida todas as suas sessões.
func TestDeactivatedUserInvalidatesEverySession(t *testing.T) {
	e := newSessionEnv(t)
	u, token1, _ := e.open(t, "ana@exemplo.com")
	token2, hash2, _ := domain.NewSessionToken()
	_, _ = e.sessions.Create(context.Background(), u.ID, hash2, "csrf", e.now, e.now.Add(time.Hour))
	if err := e.users.SetActive(context.Background(), u.ID, false); err != nil {
		t.Fatal(err)
	}

	for _, token := range []string{token1, token2} {
		if _, err := e.svc.Validate(context.Background(), token); !errors.Is(err, app.ErrUnauthenticated) {
			t.Errorf("sessão de usuário inativo: %v", err)
		}
	}
}

// A permissão vale na requisição seguinte à troca de papéis (sem cache).
func TestRoleChangesApplyOnTheNextRequest(t *testing.T) {
	e := newSessionEnv(t)
	u, token, _ := e.open(t, "ana@exemplo.com")
	before, _ := e.svc.Validate(context.Background(), token)
	if before.Principal.Has("audit:log:read") {
		t.Fatal("ASSOCIADO não tem audit:log:read")
	}

	_ = e.users.SetRoles(context.Background(), u.ID, []domain.Role{domain.RoleConselhoFiscal})

	after, err := e.svc.Validate(context.Background(), token)
	if err != nil || !after.Principal.Has("audit:log:read") {
		t.Errorf("a mudança deveria valer de imediato: %+v, %v", after.Principal.Permissions, err)
	}
}

// IDN-02.6 e AUD-04.4
func TestLogoutRevokesTheSessionAndAuditsIt(t *testing.T) {
	e := newSessionEnv(t)
	u, token, _ := e.open(t, "ana@exemplo.com")

	if err := e.svc.Logout(context.Background(), token); err != nil {
		t.Fatal(err)
	}

	if _, err := e.svc.Validate(context.Background(), token); !errors.Is(err, app.ErrUnauthenticated) {
		t.Errorf("depois do logout: %v", err)
	}
	var rows []struct {
		Action, ActorType, EntityID string
		ActorUserID                 *string
	}
	_ = e.owner.Raw(`SELECT action, actor_type, entity_id, actor_user_id::text AS actor_user_id FROM audit_log WHERE action = 'auth.logout'`).Scan(&rows).Error
	if len(rows) != 1 || rows[0].ActorType != "user" || rows[0].ActorUserID == nil || *rows[0].ActorUserID != u.ID {
		t.Errorf("eventos auth.logout = %+v", rows)
	}
}

func TestLogoutIsIdempotentAndIgnoresUnknownTokens(t *testing.T) {
	e := newSessionEnv(t)
	_, token, _ := e.open(t, "ana@exemplo.com")

	for _, tk := range []string{token, token, "token-inexistente", ""} {
		if err := e.svc.Logout(context.Background(), tk); err != nil {
			t.Errorf("Logout(%.10q): %v", tk, err)
		}
	}
	var n int64
	_ = e.owner.Raw("SELECT count(*) FROM audit_log WHERE action = 'auth.logout'").Scan(&n).Error
	if n != 1 {
		t.Errorf("só o primeiro logout válido gera evento: %d", n)
	}
}

func TestValidateAfterAnExpiredSessionDoesNotTouchIt(t *testing.T) {
	e := newSessionEnv(t)
	_, token, _ := e.open(t, "ana@exemplo.com")
	created := e.lastSeen(t, token)

	e.now = t0.Add(3 * time.Hour)
	_, _ = e.svc.Validate(context.Background(), token)

	if got := e.lastSeen(t, token); !got.Equal(created) {
		t.Errorf("uma sessão expirada não pode ser renovada pelo uso: %v -> %v", created, got)
	}
}

func TestSessionServiceRequiresAllDependencies(t *testing.T) {
	if _, err := (&app.SessionService{}).Validate(context.Background(), "x"); err == nil || errors.Is(err, app.ErrUnauthenticated) {
		t.Errorf("um serviço mal montado deveria falhar com erro de configuração, veio %v", err)
	}
	if err := (&app.SessionService{}).Logout(context.Background(), "x"); err == nil {
		t.Error("Logout de um serviço mal montado deveria falhar")
	}
}
