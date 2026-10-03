//go:build integration

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email/emailtest"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

const goodPassword = "uma-senha-correta-123"

// realRouter is the router over a real database: the chains with a session.
func realRouter(t *testing.T) (http.Handler, *identity.Module) {
	t.Helper()
	db := testutil.NewTestDB(t)
	log := logx.New("error", &bytes.Buffer{})
	rec := audit.NewRecorder(db, log)
	matrix, err := app.BuildMatrix(append(app.FoundationContributions(), financeiro.Contribution())...)
	if err != nil {
		t.Fatal(err)
	}
	hasher, _ := password.NewHasher(password.Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1})
	authorizer, _ := authz.NewAuthorizer(matrix.Definitions, audit.DeniedHook(rec))
	mod := identity.New(identity.Deps{
		DB: db, Recorder: rec, Authorizer: authorizer, Matrix: matrix, Hasher: hasher, Denylist: password.NewDenylist("12345678\n"),
		Sender: &emailtest.Recorder{}, Log: log, HashKey: []byte("0123456789abcdef0123456789abcdef"),
		SessionIdle: time.Hour, SessionAbsolute: 8 * time.Hour, ResetTTL: 30 * time.Minute, BaseURL: appOrigin,
	})
	if _, err := mod.Roles.Sync(context.Background(), matrix); err != nil {
		t.Fatal(err)
	}
	return NewRouter(Deps{
		Ping: func(context.Context) error { return nil }, Identity: mod, AuditQuery: &audit.Query{Authz: authorizer, DB: db}, Log: log,
		AllowedOrigins: []string{appOrigin}, Cookie: httpx.SessionCookie{Secure: true},
	}), mod
}

func seedUser(t *testing.T, mod *identity.Module, mustChange bool, roles ...domain.Role) domain.User {
	t.Helper()
	hasher, _ := password.NewHasher(password.Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1})
	hash, _ := hasher.Hash(goodPassword)
	u, err := mod.Users.Create(context.Background(), domain.User{Email: "ana@exemplo.com", Name: "Ana", PasswordHash: hash, Active: true, MustChangePassword: mustChange})
	if err != nil {
		t.Fatal(err)
	}
	if err := mod.Users.SetRoles(context.Background(), u.ID, append([]domain.Role{domain.RoleAssociado}, roles...)); err != nil {
		t.Fatal(err)
	}
	if len(roles) > 0 {
		m, _ := domain.NewAdminMembership(u.ID, "membro da diretoria eleita", nil, time.Now())
		if _, err := mod.Members.Grant(context.Background(), m); err != nil {
			t.Fatal(err)
		}
	}
	return u
}

// signIn logs in and returns the session cookie and the CSRF token.
func signIn(t *testing.T, r http.Handler, email string) (*http.Cookie, string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": goodPassword})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Origin", appOrigin)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login: %d %s", w.Code, w.Body.String())
	}
	var ctx struct {
		Csrf string `json:"csrf_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &ctx)
	return w.Result().Cookies()[0], ctx.Csrf
}

func call(r http.Handler, method, path, body string, c *http.Cookie, csrf string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Origin", appOrigin)
	req.Header.Set("Content-Type", "application/json")
	if c != nil {
		req.AddCookie(c)
	}
	if csrf != "" {
		req.Header.Set("X-CSRF-Token", csrf)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// IDN-03: com sessão válida, o que altera estado exige o token CSRF da sessão; a leitura não.
func TestAuthenticatedStateChangesNeedTheSessionsCSRFToken(t *testing.T) {
	r, mod := realRouter(t)
	seedUser(t, mod, false, domain.RolePresidente)
	cookie, csrf := signIn(t, r, "ana@exemplo.com")
	newUser := `{"email":"novo@exemplo.com","name":"Novo","password":"senha-inicial-123"}`

	without := call(r, http.MethodPost, "/api/v1/users", newUser, cookie, "")
	wrong := call(r, http.MethodPost, "/api/v1/users", newUser, cookie, "token-errado")
	ok := call(r, http.MethodPost, "/api/v1/users", newUser, cookie, csrf)
	read := call(r, http.MethodGet, "/api/v1/users", "", cookie, "")

	if without.Code != http.StatusForbidden || !bytes.Contains(without.Body.Bytes(), []byte("csrf_invalid")) || wrong.Code != http.StatusForbidden {
		t.Errorf("sem token: %d %s; errado: %d", without.Code, without.Body.String(), wrong.Code)
	}
	if ok.Code != http.StatusCreated || read.Code != http.StatusOK {
		t.Errorf("com token: %d %s; leitura: %d", ok.Code, ok.Body.String(), read.Code)
	}
}

// IDN-02: com troca de senha obrigatória, as demais rotas respondem 403 password_change_required.
func TestMustChangePasswordBlocksTheOtherRoutesThroughTheRouter(t *testing.T) {
	r, mod := realRouter(t)
	seedUser(t, mod, true, domain.RolePresidente)
	cookie, csrf := signIn(t, r, "ana@exemplo.com")

	blocked := call(r, http.MethodGet, "/api/v1/users", "", cookie, "")
	auditResp := call(r, http.MethodGet, "/api/v1/audit-logs", "", cookie, "")
	me := call(r, http.MethodGet, "/api/v1/auth/me", "", cookie, "")

	for name, w := range map[string]*httptest.ResponseRecorder{"users": blocked, "audit-logs": auditResp} {
		if w.Code != http.StatusForbidden || !bytes.Contains(w.Body.Bytes(), []byte("password_change_required")) {
			t.Errorf("%s: %d %s", name, w.Code, w.Body.String())
		}
	}
	if me.Code != http.StatusOK || csrf == "" {
		t.Errorf("me: %d", me.Code)
	}
}

// AUD-03: a consulta de auditoria está montada atrás da sessão e da permissão.
func TestAuditQueryIsMountedBehindSessionAndPermission(t *testing.T) {
	r, mod := realRouter(t)
	seedUser(t, mod, false, domain.RolePresidente)
	cookie, _ := signIn(t, r, "ana@exemplo.com")

	w := call(r, http.MethodGet, "/api/v1/audit-logs?action=auth.login", "", cookie, "")

	var page struct{ Items []map[string]any }
	_ = json.Unmarshal(w.Body.Bytes(), &page)
	if w.Code != http.StatusOK || len(page.Items) != 1 {
		t.Errorf("status = %d, itens = %d, corpo = %s", w.Code, len(page.Items), w.Body.String())
	}
}
