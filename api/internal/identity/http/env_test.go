//go:build integration

package identityhttp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	identityhttp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/http"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email/emailtest"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

const (
	appOrigin    = "https://app.tj.example"
	goodPassword = "uma-senha-correta-123"
)

var t0 = time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

// env is the identity module behind the real middleware chains, over a private
// database: what the handler tests call through HTTP.
type env struct {
	t        *testing.T
	db       *gorm.DB
	owner    *gorm.DB
	mod      *identity.Module
	engine   *gin.Engine
	contract *testutil.Contract
	mail     *emailtest.Recorder
	hasher   *password.Hasher
	logs     *bytes.Buffer
	now      time.Time
	seq      int
}

func newEnv(t *testing.T) *env {
	t.Helper()
	db := testutil.NewTestDB(t)
	logs := &bytes.Buffer{}
	log := logx.New("info", logs)
	rec := audit.NewRecorder(db, log)
	matrix, err := app.BuildMatrix(append(app.FoundationContributions(), financeiro.Contribution())...)
	if err != nil {
		t.Fatal(err)
	}
	hasher, err := password.NewHasher(password.Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1})
	if err != nil {
		t.Fatal(err)
	}
	authorizer, err := authz.NewAuthorizer(matrix.Definitions, audit.DeniedHook(rec))
	if err != nil {
		t.Fatal(err)
	}
	e := &env{t: t, db: db, hasher: hasher, owner: testutil.OwnerDBFor(t, db), mail: &emailtest.Recorder{}, logs: logs, now: t0, contract: testutil.LoadContracts(t)}
	e.mod = identity.New(identity.Deps{
		DB: db, Recorder: rec, Authorizer: authorizer, Matrix: matrix, Hasher: hasher,
		Denylist: password.NewDenylist("12345678\nsenhacomum1234\n"), Sender: e.mail, Log: log,
		HashKey: []byte("0123456789abcdef0123456789abcdef"), SessionIdle: 60 * time.Minute, SessionAbsolute: 8 * time.Hour,
		ResetTTL: 30 * time.Minute, BaseURL: appOrigin, Now: func() time.Time { return e.now },
	})
	e.mod.RequestReset.Async = func(fn func()) { fn() }
	if _, err := e.mod.Roles.Sync(context.Background(), matrix); err != nil {
		t.Fatal(err)
	}
	e.engine = e.newEngine()
	return e
}

func (e *env) newEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	spec, err := identityhttp.GetSpec()
	if err != nil {
		e.t.Fatal(err)
	}
	validate, err := httpx.NewContractValidator(spec)
	if err != nil {
		e.t.Fatal(err)
	}
	h := identityhttp.New(e.mod, httpx.SessionCookie{Secure: true}, logx.New("info", e.logs))
	origin := httpx.Origin(httpx.OriginConfig{Allowed: []string{appOrigin}, Cookie: httpx.SessionCookie{Secure: true}})
	authn := httpx.Authn(httpx.AuthnConfig{
		Validator: h.SessionValidator(), Cookie: httpx.SessionCookie{Secure: true},
		AllowedWhilePasswordChangeIsRequired: identityhttp.AllowedWhilePasswordChangeIsRequired,
		OnAuthenticated:                      identityhttp.WithActor,
	})
	r := gin.New()
	r.Use(httpx.RequestID())
	identityhttp.Register(r, h, h, identityhttp.Chains{
		Public:        []gin.HandlerFunc{origin, httpx.BodyLimit(64 << 10), validate},
		Authenticated: []gin.HandlerFunc{httpx.BodyLimit(1 << 20), origin, authn, httpx.CSRF(), validate},
	})
	return r
}

type opt func(*http.Request)

func cookie(token string) opt {
	return func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "tj_session", Value: token}) }
}
func csrf(token string) opt { return func(r *http.Request) { r.Header.Set("X-CSRF-Token", token) } }
func origin(o string) opt   { return func(r *http.Request) { r.Header.Set("Origin", o) } }

// session is an authenticated client: the cookie and the CSRF token.
type session struct{ token, csrf string }

func (s session) opts() []opt { return []opt{cookie(s.token), csrf(s.csrf)} }

// do sends the request through the real chain, with the allowed Origin unless an
// opt changes it, and validates the response against the contract.
func (e *env) do(method, path string, body any, opts ...opt) *httptest.ResponseRecorder {
	e.t.Helper()
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			e.t.Fatal(err)
		}
		rd = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Origin", appOrigin)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, o := range opts {
		o(req)
	}
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)
	e.contract.ValidateResponse(e.t, req, w)
	return w
}

func (e *env) as(s session, method, path string, body any, more ...opt) *httptest.ResponseRecorder {
	e.t.Helper()
	return e.do(method, path, body, append(s.opts(), more...)...)
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("corpo não é JSON: %v: %s", err, w.Body.String())
	}
	return m
}

func codeOf(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	c, _ := decode(t, w)["code"].(string)
	return c
}

// newUser creates an active user; administrative roles come with a membership.
func (e *env) newUser(roles ...domain.Role) domain.User {
	e.t.Helper()
	e.seq++
	hash, err := e.hasher.Hash(goodPassword)
	if err != nil {
		e.t.Fatal(err)
	}
	u, err := e.mod.Users.Create(context.Background(), domain.User{
		Email: "user" + itoa(e.seq) + "@exemplo.com", Name: "Usuário " + itoa(e.seq), PasswordHash: hash, Active: true,
	})
	if err != nil {
		e.t.Fatal(err)
	}
	if len(roles) == 0 {
		roles = []domain.Role{domain.RoleAssociado}
	}
	if !slices.Contains(roles, domain.RoleAssociado) {
		roles = append(roles, domain.RoleAssociado)
	}
	if err := e.mod.Users.SetRoles(context.Background(), u.ID, roles); err != nil {
		e.t.Fatal(err)
	}
	for _, r := range roles {
		if r != domain.RoleAssociado {
			m, _ := domain.NewAdminMembership(u.ID, "membro da diretoria eleita", nil, e.now)
			if _, err := e.mod.Members.Grant(context.Background(), m); err != nil {
				e.t.Fatal(err)
			}
			break
		}
	}
	return u
}

// login signs the user in with the good password and returns the session.
func (e *env) login(u domain.User) session {
	e.t.Helper()
	w := e.do(http.MethodPost, "/api/v1/auth/login", map[string]string{"email": u.Email, "password": goodPassword})
	if w.Code != http.StatusOK {
		e.t.Fatalf("login de %s: %d %s", u.Email, w.Code, w.Body.String())
	}
	var token string
	for _, c := range w.Result().Cookies() {
		if c.Name == "tj_session" {
			token = c.Value
		}
	}
	return session{token: token, csrf: decode(e.t, w)["csrf_token"].(string)}
}

// signedIn creates a user with the roles and signs them in.
func (e *env) signedIn(roles ...domain.Role) (domain.User, session) {
	e.t.Helper()
	u := e.newUser(roles...)
	return u, e.login(u)
}

func (e *env) events(action string) []map[string]any {
	e.t.Helper()
	var rows []struct {
		Outcome, ActorType, EntityID, Context string
	}
	err := e.owner.Raw(`SELECT outcome, actor_type, entity_id, context::text AS context FROM audit_log WHERE action = ? ORDER BY occurred_at, id`, action).Scan(&rows).Error
	if err != nil {
		e.t.Fatal(err)
	}
	out := make([]map[string]any, len(rows))
	for i, r := range rows {
		out[i] = map[string]any{"outcome": r.Outcome, "actor_type": r.ActorType, "entity_id": r.EntityID, "context": r.Context}
	}
	return out
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// tokenFromMail extracts the recovery token from the last e-mail.
func (e *env) tokenFromMail() string {
	e.t.Helper()
	msg, ok := e.mail.Last()
	if !ok {
		e.t.Fatal("nenhum e-mail enviado")
	}
	_, after, found := strings.Cut(msg.TextBody, appOrigin+"/redefinir-senha#token=")
	if !found {
		e.t.Fatalf("o corpo deveria ter o link com o token no fragmento: %q", msg.TextBody)
	}
	return strings.Fields(after)[0]
}
