//go:build integration

package app_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

var (
	hashKey = []byte("0123456789abcdef0123456789abcdef")
	t0      = time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	params  = password.Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1}
)

const goodPassword = "uma-senha-correta-123"

// countingVerifier wraps a real Hasher and counts the work done for unknown users.
type countingVerifier struct {
	*password.Hasher
	burned int
}

func (c *countingVerifier) BurnCycles(plain string) { c.burned++; c.Hasher.BurnCycles(plain) }

type loginEnv struct {
	db       *gorm.DB
	owner    *gorm.DB
	users    *infra.UserRepository
	sessions *infra.SessionRepository
	attempts *infra.AttemptRepository
	auth     *app.Authenticator
	verifier *countingVerifier
	logs     *bytes.Buffer
	now      time.Time
	user     domain.User
}

func newLoginEnv(t *testing.T) *loginEnv {
	t.Helper()
	db := testutil.NewTestDB(t)
	logs := &bytes.Buffer{}
	hasher, err := password.NewHasher(params)
	if err != nil {
		t.Fatal(err)
	}
	e := &loginEnv{
		db: db, owner: testutil.OwnerDBFor(t, db), logs: logs, now: t0,
		users: infra.NewUserRepository(db), sessions: infra.NewSessionRepository(db), attempts: infra.NewAttemptRepository(db),
		verifier: &countingVerifier{Hasher: hasher},
	}
	e.auth = &app.Authenticator{
		Users: e.users, Sessions: e.sessions, Attempts: e.attempts, Audit: audit.NewRecorder(db, logx.New("info", logs)),
		Passwords: e.verifier, HashKey: hashKey, SessionAbsolute: 8 * time.Hour,
		Tx:  func(ctx context.Context, fn func(context.Context) error) error { return database.WithTx(ctx, db, fn) },
		Now: func() time.Time { return e.now },
	}
	hash, _ := hasher.Hash(goodPassword)
	e.user, err = e.users.Create(context.Background(), domain.User{Email: "ana@exemplo.com", Name: "Ana", PasswordHash: hash, Active: true})
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func (e *loginEnv) login(email, pw, presented string) (app.LoginResult, error) {
	ctx := httpx.WithRequestID(context.Background(), "req-12345678")
	return e.auth.Login(ctx, app.LoginInput{Email: email, Password: pw, PresentedToken: presented})
}

type auditRow struct {
	ActorType  string
	ActorID    *string
	Action     string
	EntityType string
	EntityID   string
	Outcome    string
	Context    string
}

func (e *loginEnv) events(t *testing.T) []auditRow {
	t.Helper()
	var rows []auditRow
	err := e.owner.Raw(`SELECT actor_type, actor_user_id::text AS actor_id, action, entity_type, entity_id, outcome, context::text AS context
		FROM audit_log ORDER BY occurred_at, id`).Scan(&rows).Error
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func (e *loginEnv) count(t *testing.T, table string) int64 {
	t.Helper()
	var n int64
	if err := e.owner.Raw("SELECT count(*) FROM " + table).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

// IDN-02.1, .2 e AUD-04.1: sessão nova, só o hash no banco, evento auth.login.
func TestSuccessfulLoginCreatesASessionAndAuditsIt(t *testing.T) {
	e := newLoginEnv(t)

	res, err := e.login("  ANA@Exemplo.com ", goodPassword, "")

	if err != nil {
		t.Fatal(err)
	}
	if res.Token == "" || res.UserID != e.user.ID || res.Session.CSRFToken == "" || res.MustChangePassword {
		t.Errorf("resultado = %+v", res)
	}
	if got, err := e.sessions.FindByToken(context.Background(), res.Token); err != nil || got.ID != res.Session.ID || got.RevokedAt != nil {
		t.Errorf("a sessão deveria ser encontrada pelo token: %+v, %v", got, err)
	}
	if !res.Session.ExpiresAt.Equal(t0.Add(8 * time.Hour)) {
		t.Errorf("ExpiresAt = %v", res.Session.ExpiresAt)
	}
	events := e.events(t)
	if len(events) != 1 || events[0].Action != "auth.login" || events[0].Outcome != "success" || events[0].ActorType != "user" ||
		events[0].ActorID == nil || *events[0].ActorID != e.user.ID || events[0].EntityID != e.user.ID {
		t.Errorf("eventos = %+v", events)
	}
}

func TestLoginReportsMustChangePassword(t *testing.T) {
	e := newLoginEnv(t)
	if err := e.users.SetPassword(context.Background(), e.user.ID, e.user.PasswordHash, true); err != nil {
		t.Fatal(err)
	}

	res, err := e.login("ana@exemplo.com", goodPassword, "")

	if err != nil || !res.MustChangePassword {
		t.Errorf("res = %+v, err = %v", res, err)
	}
}

// IDN-02.3 e .4: o mesmo erro para e-mail inexistente, senha errada e usuário inativo.
func TestUnknownWrongPasswordAndInactiveUsersGetTheSameError(t *testing.T) {
	e := newLoginEnv(t)
	inactive, _ := e.users.Create(context.Background(), domain.User{Email: "inativa@exemplo.com", Name: "I", PasswordHash: e.user.PasswordHash, Active: true})
	_ = e.users.SetActive(context.Background(), inactive.ID, false)

	cases := map[string][2]string{
		"e-mail inexistente": {"ninguem@exemplo.com", goodPassword},
		"senha errada":       {"ana@exemplo.com", "senha-errada-123"},
		"usuário inativo":    {"inativa@exemplo.com", goodPassword},
		"e-mail inválido":    {"isto não é um e-mail", goodPassword},
	}
	var messages []string
	for name, c := range cases {
		res, err := e.login(c[0], c[1], "")
		if !errors.Is(err, app.ErrInvalidCredentials) || res.Token != "" {
			t.Errorf("%s: res = %+v, err = %v", name, res, err)
			continue
		}
		messages = append(messages, err.Error())
	}
	for _, m := range messages {
		if m != messages[0] {
			t.Errorf("as mensagens de erro deveriam ser idênticas: %q vs %q", m, messages[0])
		}
	}
	if e.count(t, "sessions") != 0 {
		t.Error("nenhuma falha pode criar sessão")
	}
}

// AUD-04.2: falha grava auth.login_failed com categoria e hash, sem senha nem e-mail em claro.
func TestFailedLoginsAreAuditedWithCategoryAndEmailHashOnly(t *testing.T) {
	e := newLoginEnv(t)
	inactive, _ := e.users.Create(context.Background(), domain.User{Email: "inativa@exemplo.com", Name: "I", PasswordHash: e.user.PasswordHash, Active: true})
	_ = e.users.SetActive(context.Background(), inactive.ID, false)

	_, _ = e.login("ninguem@exemplo.com", "senha-secreta-123", "")
	_, _ = e.login("ana@exemplo.com", "senha-secreta-456", "")
	_, _ = e.login("inativa@exemplo.com", goodPassword, "")

	events := e.events(t)
	if len(events) != 3 {
		t.Fatalf("eventos = %+v", events)
	}
	categories := map[string]bool{}
	for _, ev := range events {
		if ev.Action != "auth.login_failed" || ev.Outcome != "failure" || ev.ActorType != "anonymous" || ev.ActorID != nil {
			t.Errorf("evento = %+v", ev)
		}
		for _, secret := range []string{"senha-secreta", goodPassword, "ninguem@", "ana@", "inativa@", "exemplo.com"} {
			if strings.Contains(ev.Context, secret) || strings.Contains(ev.EntityID, secret) {
				t.Errorf("o evento vazou %q: %+v", secret, ev)
			}
		}
		if !strings.Contains(ev.Context, "email_hash") {
			t.Errorf("o contexto deveria ter o hash do e-mail: %s", ev.Context)
		}
		switch {
		case strings.Contains(ev.Context, "unknown_user"):
			categories["unknown_user"] = true
		case strings.Contains(ev.Context, "wrong_password"):
			categories["wrong_password"] = true
			if ev.EntityType != "user" || ev.EntityID != e.user.ID {
				t.Errorf("com a conta existente, o alvo é o usuário: %+v", ev)
			}
		case strings.Contains(ev.Context, "inactive"):
			categories["inactive"] = true
		}
	}
	if len(categories) != 3 {
		t.Errorf("categorias = %v", categories)
	}
}

// Tempo de resposta: o usuário inexistente gasta o mesmo esforço de uma verificação.
func TestUnknownUserBurnsCyclesToEqualizeTheResponseTime(t *testing.T) {
	e := newLoginEnv(t)

	_, _ = e.login("ninguem@exemplo.com", goodPassword, "")
	_, _ = e.login("isto não é um e-mail", goodPassword, "")
	_, _ = e.login("ana@exemplo.com", "senha-errada-123", "")

	if e.verifier.burned != 2 {
		t.Errorf("BurnCycles chamado %d vezes, esperado 2 (só para quem não existe)", e.verifier.burned)
	}
}

// IDN-02.8 e .10: a quinta falha bloqueia por 15 minutos; nunca de forma permanente.
func TestFifthFailureBlocksForFifteenMinutesAndTheLockIsTemporary(t *testing.T) {
	e := newLoginEnv(t)
	for i := range 5 {
		e.now = t0.Add(time.Duration(i) * time.Minute)
		if _, err := e.login("ana@exemplo.com", "senha-errada-123", ""); !errors.Is(err, app.ErrInvalidCredentials) {
			t.Fatalf("falha %d: %v", i+1, err)
		}
	}

	e.now = t0.Add(5 * time.Minute)
	_, err := e.login("ana@exemplo.com", goodPassword, "") // até a senha certa é recusada
	var locked *app.LockedError
	if !errors.As(err, &locked) {
		t.Fatalf("esperava LockedError, veio %v", err)
	}
	if locked.RetryAfter != 14*time.Minute {
		t.Errorf("RetryAfter = %v, esperado 14m (o bloqueio vai até 15 minutos depois da quinta falha)", locked.RetryAfter)
	}
	if e.count(t, "sessions") != 0 {
		t.Error("durante o bloqueio não se cria sessão")
	}

	e.now = t0.Add(4*time.Minute + 15*time.Minute)
	if _, err := e.login("ana@exemplo.com", goodPassword, ""); err != nil {
		t.Errorf("depois de 15 minutos o login volta a funcionar: %v", err)
	}
}

func TestAttemptsDuringTheLockDoNotCountOrExtendIt(t *testing.T) {
	e := newLoginEnv(t)
	for range 5 {
		_, _ = e.login("ana@exemplo.com", "senha-errada-123", "")
	}
	before := e.count(t, "login_attempts")

	for i := range 10 {
		e.now = t0.Add(time.Duration(i) * time.Minute)
		var locked *app.LockedError
		if _, err := e.login("ana@exemplo.com", "senha-errada-123", ""); !errors.As(err, &locked) {
			t.Fatalf("tentativa %d durante o bloqueio: %v", i, err)
		}
	}

	if after := e.count(t, "login_attempts"); after != before {
		t.Errorf("as tentativas bloqueadas não devem ser contadas: %d -> %d", before, after)
	}
	e.now = t0.Add(15 * time.Minute)
	if _, err := e.login("ana@exemplo.com", goodPassword, ""); err != nil {
		t.Errorf("o bloqueio não pode ter sido estendido: %v", err)
	}
}

// AUD-04.3
func TestBlockedAttemptIsAuditedAsLoginBlocked(t *testing.T) {
	e := newLoginEnv(t)
	for range 5 {
		_, _ = e.login("ana@exemplo.com", "senha-errada-123", "")
	}

	_, _ = e.login("ana@exemplo.com", goodPassword, "")

	events := e.events(t)
	last := events[len(events)-1]
	if last.Action != "auth.login_blocked" || last.Outcome != "denied" || last.ActorType != "anonymous" || strings.Contains(last.Context, "ana@") {
		t.Errorf("último evento = %+v", last)
	}
}

func TestSuccessClearsTheFailureCount(t *testing.T) {
	e := newLoginEnv(t)
	for range 4 {
		_, _ = e.login("ana@exemplo.com", "senha-errada-123", "")
	}
	if _, err := e.login("ana@exemplo.com", goodPassword, ""); err != nil {
		t.Fatal(err)
	}
	for range 4 {
		_, _ = e.login("ana@exemplo.com", "senha-errada-123", "")
	}

	if _, err := e.login("ana@exemplo.com", goodPassword, ""); err != nil {
		t.Errorf("4 + sucesso + 4 falhas não bloqueiam: %v", err)
	}
}

// IDN-02.7: login emite token novo e revoga a sessão apresentada; várias sessões são permitidas.
func TestLoginIssuesANewTokenAndRevokesThePresentedSession(t *testing.T) {
	e := newLoginEnv(t)
	first, _ := e.login("ana@exemplo.com", goodPassword, "")

	second, err := e.login("ana@exemplo.com", goodPassword, first.Token)

	if err != nil {
		t.Fatal(err)
	}
	if second.Token == first.Token {
		t.Error("o token deveria ser novo")
	}
	if old, _ := e.sessions.FindByToken(context.Background(), first.Token); old.RevokedAt == nil {
		t.Error("a sessão apresentada deveria ter sido revogada")
	}
	if cur, _ := e.sessions.FindByToken(context.Background(), second.Token); cur.RevokedAt != nil {
		t.Error("a sessão nova deveria estar ativa")
	}
}

func TestLoginWithoutAPresentedTokenKeepsOtherSessionsAndUnknownTokensAreIgnored(t *testing.T) {
	e := newLoginEnv(t)
	first, _ := e.login("ana@exemplo.com", goodPassword, "")

	second, err := e.login("ana@exemplo.com", goodPassword, "token-que-nao-existe")

	if err != nil {
		t.Fatalf("um cookie desconhecido não pode impedir o login: %v", err)
	}
	if a, _ := e.sessions.FindByToken(context.Background(), first.Token); a.RevokedAt != nil {
		t.Error("várias sessões simultâneas são permitidas: a primeira continua ativa")
	}
	if b, _ := e.sessions.FindByToken(context.Background(), second.Token); b.RevokedAt != nil {
		t.Error("a segunda deveria estar ativa")
	}
}

func TestFailedLoginDoesNotRevokeThePresentedSession(t *testing.T) {
	e := newLoginEnv(t)
	first, _ := e.login("ana@exemplo.com", goodPassword, "")

	_, _ = e.login("ana@exemplo.com", "senha-errada-123", first.Token)

	if a, _ := e.sessions.FindByToken(context.Background(), first.Token); a.RevokedAt != nil {
		t.Error("um login que falha não pode derrubar a sessão em uso")
	}
}

// AUD-04.6: falha ao gravar o evento não impede a resposta do login.
func TestLoginSucceedsEvenWhenTheAuditWriteFails(t *testing.T) {
	e := newLoginEnv(t)
	if err := e.owner.Exec("ALTER TABLE audit_log RENAME TO audit_log_off").Error; err != nil {
		t.Fatal(err)
	}

	res, err := e.login("ana@exemplo.com", goodPassword, "")

	if err != nil || res.Token == "" {
		t.Errorf("res = %+v, err = %v", res, err)
	}
	if !strings.Contains(e.logs.String(), "audit_security_write_failed") {
		t.Errorf("o incidente deveria estar no log: %s", e.logs.String())
	}
	if _, err := e.login("ana@exemplo.com", "senha-errada-123", ""); !errors.Is(err, app.ErrInvalidCredentials) {
		t.Errorf("a falha de auditoria não muda a resposta de uma falha de login: %v", err)
	}
}

// Só o HMAC do e-mail fica nas tentativas.
func TestLoginAttemptsStoreTheHMACOfTheEmailNotTheEmail(t *testing.T) {
	e := newLoginEnv(t)
	_, _ = e.login("Ana@Exemplo.com", "senha-errada-123", "")

	var got []byte
	if err := e.owner.Raw("SELECT email_hash FROM login_attempts").Row().Scan(&got); err != nil {
		t.Fatal(err)
	}
	want, _ := domain.HashEmail(hashKey, "ana@exemplo.com")
	if !bytes.Equal(got, want) {
		t.Errorf("email_hash = %x, esperado o HMAC %x", got, want)
	}
}

func TestLoginRequiresAllDependencies(t *testing.T) {
	_, err := (&app.Authenticator{}).Login(context.Background(), app.LoginInput{Email: "a@b.com", Password: "x"})

	if err == nil || errors.Is(err, app.ErrInvalidCredentials) {
		t.Errorf("um Authenticator mal montado deveria falhar com erro de configuração, veio %v", err)
	}
}
