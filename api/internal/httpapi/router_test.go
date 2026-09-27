package httpapi

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
)

const appOrigin = "https://app.tj.example"

// testDeps builds a router with no database: nothing here reaches it, because
// every request is refused before a use case runs.
func testDeps(ping PingFunc) Deps {
	if ping == nil {
		ping = func(context.Context) error { return nil }
	}
	return Deps{
		Ping:           ping,
		Identity:       identity.New(identity.Deps{}),
		AuditQuery:     &audit.Query{},
		Log:            logx.New("error", io.Discard),
		AllowedOrigins: []string{appOrigin},
		Cookie:         httpx.SessionCookie{Secure: true},
	}
}

func send(r http.Handler, method, target string, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthz(t *testing.T) {
	cases := []struct {
		name string
		ping PingFunc
		code int
		body string
	}{
		{"banco ok", func(context.Context) error { return nil }, http.StatusOK, `{"status":"ok"}`},
		{"banco fora", func(context.Context) error { return errors.New("down") }, http.StatusServiceUnavailable, `{"status":"degraded"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := send(NewRouter(testDeps(tc.ping)), http.MethodGet, "/healthz", "", nil)
			if w.Code != tc.code || w.Body.String() != tc.body {
				t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

// RBAC-02.1: toda rota fora da lista pública exige sessão. O teste percorre as rotas
// registradas e falha se uma delas responder de outro jeito sem cookie.
func TestEveryRouteOutsideThePublicListRequiresASession(t *testing.T) {
	router := NewRouter(testDeps(nil))
	public := []string{
		"GET /healthz", "POST /api/v1/auth/login",
		"POST /api/v1/auth/password-reset/request", "POST /api/v1/auth/password-reset/confirm",
	}

	for _, key := range routeKeys(router.Routes()) {
		if slices.Contains(public, key) {
			continue
		}
		method, path, _ := strings.Cut(key, " ")
		path = strings.NewReplacer("{id}", "0f8fad5b-d9cb-469f-a165-70867728950e").Replace(path)
		// Origin permitido e CSRF presente: só a falta de sessão pode barrar.
		w := send(router, method, path, "", map[string]string{"Origin": appOrigin, "X-CSRF-Token": "x"})

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, esperado 401 sem sessão (corpo %s)", key, w.Code, w.Body.String())
		}
	}
}

// RBAC-02.4: só o /healthz e as três rotas de acesso são públicas.
func TestThePublicRoutesAreExactlyHealthzLoginAndRecovery(t *testing.T) {
	router := NewRouter(testDeps(nil))
	var open []string
	for _, key := range routeKeys(router.Routes()) {
		method, path, _ := strings.Cut(key, " ")
		w := send(router, method, strings.ReplaceAll(path, "{id}", "0f8fad5b-d9cb-469f-a165-70867728950e"), "", map[string]string{"Origin": appOrigin, "X-CSRF-Token": "x"})
		if w.Code != http.StatusUnauthorized {
			open = append(open, key)
		}
	}
	want := []string{"GET /healthz", "POST /api/v1/auth/login", "POST /api/v1/auth/password-reset/confirm", "POST /api/v1/auth/password-reset/request"}
	if !slices.Equal(open, want) {
		t.Errorf("rotas públicas = %v, esperado %v", open, want)
	}
}

// IDN-07.10 e IDN-03: as rotas públicas checam a origem; o /healthz (GET) não.
func TestPublicRoutesCheckTheOriginAndHealthzDoesNot(t *testing.T) {
	router := NewRouter(testDeps(nil))

	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/password-reset/request", "/api/v1/auth/password-reset/confirm"} {
		w := send(router, http.MethodPost, path, `{}`, map[string]string{"Origin": "https://evil.example", "Content-Type": "application/json"})
		if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "origin_not_allowed") {
			t.Errorf("%s com origem estranha: %d %s", path, w.Code, w.Body.String())
		}
	}
	if w := send(router, http.MethodGet, "/healthz", "", map[string]string{"Origin": "https://evil.example"}); w.Code != http.StatusOK {
		t.Errorf("/healthz fora da checagem de origem: %d", w.Code)
	}
}

// Rotas autenticadas: a origem é checada antes da sessão; o CSRF vem depois da sessão.
func TestAuthenticatedRoutesCheckTheOriginBeforeTheSession(t *testing.T) {
	router := NewRouter(testDeps(nil))

	w := send(router, http.MethodPost, "/api/v1/users", `{}`, map[string]string{"Origin": "https://evil.example", "Content-Type": "application/json"})

	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "origin_not_allowed") {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

// API-02.7: 1 MiB nas rotas autenticadas e 64 KiB nas públicas, com 413 payload_too_large.
func TestBodyLimitsAre1MiBAuthenticatedAnd64KiBPublic(t *testing.T) {
	router := NewRouter(testDeps(nil))
	big := func(n int) string { return `{"email":"` + strings.Repeat("a", n) + `"}` }
	headers := map[string]string{"Origin": appOrigin, "Content-Type": "application/json", "X-CSRF-Token": "x"}

	pub := send(router, http.MethodPost, "/api/v1/auth/login", big(65*1024), headers)
	auth := send(router, http.MethodPost, "/api/v1/users", big(1024*1024+1), headers)
	small := send(router, http.MethodPost, "/api/v1/auth/login", big(100), headers)

	for name, w := range map[string]*httptest.ResponseRecorder{"pública": pub, "autenticada": auth} {
		if w.Code != http.StatusRequestEntityTooLarge || !strings.Contains(w.Body.String(), "payload_too_large") {
			t.Errorf("%s: status = %d, corpo = %.120s", name, w.Code, w.Body.String())
		}
	}
	if small.Code == http.StatusRequestEntityTooLarge {
		t.Error("um corpo pequeno não passa do limite")
	}
	// Na autenticada, 64 KiB ainda cabe (chega até a checagem de sessão).
	if w := send(router, http.MethodPost, "/api/v1/users", big(100*1024), headers); w.Code != http.StatusUnauthorized {
		t.Errorf("100 KiB numa rota autenticada: %d", w.Code)
	}
}

// PLT-01.3: o log de acesso é o próprio, sem a query string (que pode ter dados).
func TestAccessLogIsOursAndKeepsTheQueryStringOut(t *testing.T) {
	var logs bytes.Buffer
	d := testDeps(nil)
	d.Log = logx.New("info", &logs)
	router := NewRouter(d)

	send(router, http.MethodGet, "/healthz?token=segredo-na-url", "", nil)

	out := logs.String()
	if !strings.Contains(out, `"path":"/healthz"`) || strings.Contains(out, "segredo-na-url") {
		t.Errorf("log de acesso = %s", out)
	}
}

// Uma rota que não existe responde 404 em problem+json.
func TestUnknownRoutesAnswerProblemJSON404(t *testing.T) {
	w := send(NewRouter(testDeps(nil)), http.MethodGet, "/api/v1/nao-existe", "", nil)

	if w.Code != http.StatusNotFound || w.Header().Get("Content-Type") != "application/problem+json" || !strings.Contains(w.Body.String(), `"code":"not_found"`) {
		t.Errorf("status = %d, tipo = %s, corpo = %s", w.Code, w.Header().Get("Content-Type"), w.Body.String())
	}
}

func TestNewRouterRefusesToBeBuiltWithoutItsDependencies(t *testing.T) {
	for name, mutate := range map[string]func(*Deps){
		"sem ping":      func(d *Deps) { d.Ping = nil },
		"sem identity":  func(d *Deps) { d.Identity = nil },
		"sem auditoria": func(d *Deps) { d.AuditQuery = nil },
		"sem log":       func(d *Deps) { d.Log = nil },
	} {
		d := testDeps(nil)
		mutate(&d)
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: deveria falhar na montagem", name)
				}
			}()
			NewRouter(d)
		}()
	}
}

var _ = gin.Version
