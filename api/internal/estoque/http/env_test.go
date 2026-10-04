//go:build integration

// env_test.go is the shared integration harness for this package's handler
// tests (T4 of 05-api-http). Unlike financeiro/http/env_test.go (which must
// hand-wire routes because financeiro's Register only returns the wrapper,
// FIN-D-026/027/029), this harness calls the real estoquehttp.Register —
// *Handler satisfies StrictServerInterface in full as soon as this single
// task adds its 6 methods, same simpler shape as identity/http/env_test.go.
package estoquehttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	identityapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
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

const appOrigin = "https://app.tj.example"
const goodPassword = "uma-senha-correta-123"

// env is identity (for real login/session) and estoque (under test) behind
// the real middleware chains, over a private database.
type env struct {
	t        *testing.T
	db       *gorm.DB
	idMod    *identity.Module
	estMod   *estoque.Module
	engine   *gin.Engine
	contract *testutil.Contract
	hasher   *password.Hasher
	seq      int
}

func newEnv(t *testing.T) *env {
	t.Helper()
	db := testutil.NewTestDB(t)
	log := logx.New("error", io.Discard)
	rec := audit.NewRecorder(db, log)
	matrix, err := identityapp.BuildMatrix(append(identityapp.FoundationContributions(), estoque.Contribution())...)
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
	idMod := identity.New(identity.Deps{
		DB: db, Recorder: rec, Authorizer: authorizer, Matrix: matrix, Hasher: hasher,
		Denylist: password.NewDenylist("12345678\nsenhacomum1234\n"), Sender: &emailtest.Recorder{}, Log: log,
		HashKey: []byte("0123456789abcdef0123456789abcdef"), SessionIdle: 60 * time.Minute, SessionAbsolute: 8 * time.Hour,
		ResetTTL: 30 * time.Minute, BaseURL: appOrigin, Now: time.Now,
	})
	if _, err := idMod.Roles.Sync(context.Background(), matrix); err != nil {
		t.Fatal(err)
	}
	estMod := estoque.New(estoque.Deps{DB: db, Recorder: rec, Authorizer: authorizer})
	e := &env{t: t, db: db, idMod: idMod, estMod: estMod, hasher: hasher, contract: testutil.LoadContracts(t)}
	e.engine = e.newEngine(log)
	return e
}

// newEngine wires the real chains (same shape as httpapi/router.go's
// "authenticated") over identity's real routes (for login) plus estoque's
// real routes under test, mounted through the real estoquehttp.Register.
func (e *env) newEngine(log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	idSpec, err := identityhttp.GetSpec()
	if err != nil {
		e.t.Fatal(err)
	}
	estSpec, err := GetSpec()
	if err != nil {
		e.t.Fatal(err)
	}
	validate, err := httpx.NewContractValidator(idSpec, estSpec)
	if err != nil {
		e.t.Fatal(err)
	}
	cookie := httpx.SessionCookie{Secure: true}
	idH := identityhttp.New(e.idMod, cookie, log)
	origin := httpx.Origin(httpx.OriginConfig{Allowed: []string{appOrigin}, Cookie: cookie})
	authn := httpx.Authn(httpx.AuthnConfig{
		Validator: idH.SessionValidator(), Cookie: cookie,
		AllowedWhilePasswordChangeIsRequired: identityhttp.AllowedWhilePasswordChangeIsRequired,
		OnAuthenticated:                      identityhttp.WithActor,
	})
	authenticated := []gin.HandlerFunc{httpx.BodyLimit(1 << 20), origin, authn, httpx.CSRF(), validate}

	r := gin.New()
	r.Use(httpx.RequestID())
	identityhttp.Register(r, idH, idH, identityhttp.Chains{
		Public:        []gin.HandlerFunc{origin, httpx.BodyLimit(64 << 10), validate},
		Authenticated: authenticated,
	})
	h := New(e.estMod, log)
	Register(r, h, h, Chains{Authenticated: authenticated})
	return r
}

type opt func(*http.Request)

func cookieOpt(token string) opt {
	return func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "tj_session", Value: token}) }
}
func csrfOpt(token string) opt { return func(r *http.Request) { r.Header.Set("X-CSRF-Token", token) } }

type session struct{ token, csrf string }

func (s session) opts() []opt { return []opt{cookieOpt(s.token), csrfOpt(s.csrf)} }

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

// newUser creates an active user with the given roles (ASSOCIADO if none).
func (e *env) newUser(roles ...domain.Role) domain.User {
	e.t.Helper()
	e.seq++
	hash, err := e.hasher.Hash(goodPassword)
	if err != nil {
		e.t.Fatal(err)
	}
	email := "user" + uuid.NewString() + "@exemplo.com"
	u, err := e.idMod.Users.Create(context.Background(), domain.User{Email: email, Name: "Usuário de Teste", PasswordHash: hash, Active: true})
	if err != nil {
		e.t.Fatal(err)
	}
	if len(roles) == 0 {
		roles = []domain.Role{domain.RoleAssociado}
	}
	if err := e.idMod.Users.SetRoles(context.Background(), u.ID, roles); err != nil {
		e.t.Fatal(err)
	}
	return u
}

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
