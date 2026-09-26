package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

type fakeValidator struct {
	calls  int
	tokens []string
	info   SessionInfo
	err    error
}

func (f *fakeValidator) Validate(_ context.Context, token string) (SessionInfo, error) {
	f.calls++
	f.tokens = append(f.tokens, token)
	return f.info, f.err
}

func validInfo() SessionInfo {
	return SessionInfo{
		Principal: authz.Principal{UserID: "u-1", Roles: []string{"X"}, Permissions: map[authz.Permission]struct{}{"identity:user:read": {}}},
		SessionID: "s-1", CSRFToken: "csrf-abc",
	}
}

func authnConfig(v SessionValidator) AuthnConfig {
	return AuthnConfig{
		Validator: v,
		Cookie:    SessionCookie{Secure: true},
		Public:    map[string]bool{"POST /login": true},
		AllowedWhilePasswordChangeIsRequired: map[string]bool{
			"POST /logout": true, "GET /me": true, "POST /password": true,
		},
	}
}

func authnEngine(cfg AuthnConfig, seen *SessionInfo) *gin.Engine {
	r := newEngine()
	r.Use(RequestID(), Authn(cfg))
	ok := func(c *gin.Context) {
		if seen != nil {
			if info, found := SessionFrom(c.Request.Context()); found {
				*seen = info
			}
		}
		c.String(http.StatusOK, "ok")
	}
	r.POST("/login", ok)
	r.POST("/logout", ok)
	r.GET("/me", ok)
	r.POST("/password", ok)
	r.GET("/users", ok)
	r.POST("/users", ok)
	return r
}

func withCookie(token string) func(*http.Request) {
	return func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "tj_session", Value: token}) }
}

// Rota pública: passa sem cookie e sem consultar a sessão.
func TestAuthnLetsPublicRoutesThroughWithoutCallingTheValidator(t *testing.T) {
	v := &fakeValidator{err: ErrUnauthenticated}
	r := authnEngine(authnConfig(v), nil)

	w := serve(r, http.MethodPost, "/login", nil, nil)

	if w.Code != http.StatusOK || v.calls != 0 {
		t.Errorf("status = %d, chamadas = %d", w.Code, v.calls)
	}
}

// RBAC-02.1: negado por padrão. Só a lista pública dispensa sessão.
func TestAuthnRequiresASessionOnEveryRouteOutsideThePublicList(t *testing.T) {
	v := &fakeValidator{info: validInfo()}
	r := authnEngine(authnConfig(v), nil)

	for _, route := range []struct{ method, path string }{{"GET", "/users"}, {"POST", "/users"}, {"GET", "/me"}, {"POST", "/logout"}} {
		w := serve(r, route.method, route.path, nil, nil)
		p := decodeProblem(t, w)
		if w.Code != http.StatusUnauthorized || p["code"] != "unauthenticated" || p["request_id"] != "req-abc-12345" {
			t.Errorf("%s %s sem cookie: status = %d, corpo = %v", route.method, route.path, w.Code, p)
		}
	}
	if v.calls != 0 {
		t.Errorf("sem cookie não há o que validar: %d chamadas", v.calls)
	}
}

func TestAuthnRejectsMalformedAndUnknownCookies(t *testing.T) {
	v := &fakeValidator{err: ErrUnauthenticated}
	r := authnEngine(authnConfig(v), nil)

	for name, token := range map[string]string{"desconhecido": "abc", "enorme": strings.Repeat("x", 600)} {
		w := serve(r, http.MethodGet, "/users", nil, withCookie(token))
		if w.Code != http.StatusUnauthorized || decodeProblem(t, w)["code"] != "unauthenticated" {
			t.Errorf("%s: status = %d, corpo = %s", name, w.Code, w.Body.String())
		}
	}
	empty := serve(r, http.MethodGet, "/users", nil, withCookie(""))
	if empty.Code != http.StatusUnauthorized {
		t.Errorf("cookie vazio: %d", empty.Code)
	}
	if v.calls != 1 {
		t.Errorf("o cookie enorme e o vazio nem chegam ao validador: %d chamadas", v.calls)
	}
}

// IDN-02.5: sessão expirada devolve session_expired e limpa o cookie.
func TestAuthnAnswersSessionExpiredAndClearsTheCookie(t *testing.T) {
	r := authnEngine(authnConfig(&fakeValidator{err: ErrSessionExpired}), nil)

	w := serve(r, http.MethodGet, "/users", nil, withCookie("velho"))

	p := decodeProblem(t, w)
	if w.Code != http.StatusUnauthorized || p["code"] != "session_expired" {
		t.Fatalf("status = %d, corpo = %v", w.Code, p)
	}
	set := w.Header().Get("Set-Cookie")
	if !strings.Contains(set, "tj_session=;") || !strings.Contains(set, "Max-Age=0") || !strings.Contains(set, "HttpOnly") || !strings.Contains(set, "Path=/") {
		t.Errorf("Set-Cookie = %q, deveria limpar o cookie", set)
	}
}

func TestAuthnDoesNotLeakInternalErrors(t *testing.T) {
	r := authnEngine(authnConfig(&fakeValidator{err: errors.New("banco caiu: senha=segredo-interno")}), nil)

	w := serve(r, http.MethodGet, "/users", nil, withCookie("abc"))

	p := decodeProblem(t, w)
	if w.Code != http.StatusInternalServerError || p["code"] != "internal_error" || p["request_id"] != "req-abc-12345" {
		t.Errorf("status = %d, corpo = %v", w.Code, p)
	}
	if strings.Contains(w.Body.String(), "segredo-interno") || strings.Contains(w.Body.String(), "banco") {
		t.Errorf("vazou detalhe interno: %s", w.Body.String())
	}
}

// Sessão válida: Principal, token CSRF e usuário chegam ao handler.
func TestAuthnPutsThePrincipalAndTheSessionInTheContext(t *testing.T) {
	v := &fakeValidator{info: validInfo()}
	var seen SessionInfo
	var actor string
	cfg := authnConfig(v)
	cfg.OnAuthenticated = func(ctx context.Context, info SessionInfo) context.Context {
		actor = info.Principal.UserID
		return ctx
	}
	r := authnEngine(cfg, &seen)

	w := serve(r, http.MethodGet, "/users", nil, withCookie("meu-token"))

	if w.Code != http.StatusOK || v.tokens[0] != "meu-token" {
		t.Fatalf("status = %d, token = %v", w.Code, v.tokens)
	}
	if seen.Principal.UserID != "u-1" || seen.CSRFToken != "csrf-abc" || seen.SessionID != "s-1" || !seen.Principal.Has("identity:user:read") {
		t.Errorf("sessão no contexto = %+v", seen)
	}
	if actor != "u-1" {
		t.Errorf("OnAuthenticated deveria receber a sessão: %q", actor)
	}
}

func TestAuthnLetsTheAccessLogKnowTheUser(t *testing.T) {
	r := newEngine()
	r.Use(RequestID(), Authn(authnConfig(&fakeValidator{info: validInfo()})))
	var uid string
	r.GET("/users", func(c *gin.Context) { uid = c.GetString(userIDKey) })

	serve(r, http.MethodGet, "/users", nil, withCookie("abc"))

	if uid != "u-1" {
		t.Errorf("user_id = %q", uid)
	}
}

func TestOnAuthenticatedContextReachesTheHandler(t *testing.T) {
	type key struct{}
	cfg := authnConfig(&fakeValidator{info: validInfo()})
	cfg.OnAuthenticated = func(ctx context.Context, _ SessionInfo) context.Context { return context.WithValue(ctx, key{}, "enriquecido") }
	r := newEngine()
	r.Use(RequestID(), Authn(cfg))
	var got any
	r.GET("/users", func(c *gin.Context) { got = c.Request.Context().Value(key{}) })

	serve(r, http.MethodGet, "/users", nil, withCookie("abc"))

	if got != "enriquecido" {
		t.Errorf("valor = %v", got)
	}
}

// RBAC-02.6: com must_change_password, só logout, me e a troca de senha.
func TestAuthnBlocksEverythingButTheAllowedRoutesWhileAPasswordChangeIsRequired(t *testing.T) {
	info := validInfo()
	info.MustChangePassword = true
	v := &fakeValidator{info: info}
	r := authnEngine(authnConfig(v), nil)

	for _, ok := range []struct{ method, path string }{{"POST", "/logout"}, {"GET", "/me"}, {"POST", "/password"}} {
		if w := serve(r, ok.method, ok.path, nil, withCookie("abc")); w.Code != http.StatusOK {
			t.Errorf("%s %s deveria passar: %d", ok.method, ok.path, w.Code)
		}
	}
	for _, blocked := range []struct{ method, path string }{{"GET", "/users"}, {"POST", "/users"}} {
		w := serve(r, blocked.method, blocked.path, nil, withCookie("abc"))
		if w.Code != http.StatusForbidden || decodeProblem(t, w)["code"] != "password_change_required" {
			t.Errorf("%s %s: status = %d, corpo = %s", blocked.method, blocked.path, w.Code, w.Body.String())
		}
	}
}

func TestAuthnDoesNotBlockRoutesWhenNoPasswordChangeIsRequired(t *testing.T) {
	r := authnEngine(authnConfig(&fakeValidator{info: validInfo()}), nil)

	if w := serve(r, http.MethodGet, "/users", nil, withCookie("abc")); w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
}

// Rotas que não existem seguem para o 404 do Gin, sem virar 401.
func TestAuthnLeavesUnknownRoutesTo404(t *testing.T) {
	r := authnEngine(authnConfig(&fakeValidator{}), nil)

	if w := serve(r, http.MethodGet, "/nao-existe", nil, nil); w.Code != http.StatusNotFound {
		t.Errorf("status = %d", w.Code)
	}
}

func TestAuthnPanicsAtSetupWithoutAValidator(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("um Authn sem validador é erro de montagem: deveria falhar na partida")
		}
	}()
	Authn(AuthnConfig{})
}

// ---- cookie da sessão

func cookieFor(t *testing.T, set func(*gin.Context)) *http.Cookie {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	set(c)
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %v", cookies)
	}
	return cookies[0]
}

func TestSessionCookieSetHasTheSecurityAttributes(t *testing.T) {
	sc := SessionCookie{Secure: true, Domain: "tj.example"}

	ck := cookieFor(t, func(c *gin.Context) { sc.Set(c, "token-123", 8*time.Hour) })

	if ck.Name != "tj_session" || ck.Value != "token-123" || !ck.HttpOnly || !ck.Secure || ck.SameSite != http.SameSiteLaxMode ||
		ck.Path != "/" || ck.Domain != "tj.example" || ck.MaxAge != 8*3600 {
		t.Errorf("cookie = %+v", ck)
	}
}

func TestSessionCookieSecureAndDomainFollowTheConfiguration(t *testing.T) {
	sc := SessionCookie{Secure: false}

	ck := cookieFor(t, func(c *gin.Context) { sc.Set(c, "t", time.Hour) })

	if ck.Secure || ck.Domain != "" || !ck.HttpOnly {
		t.Errorf("cookie = %+v", ck)
	}
}

func TestSessionCookieClearExpiresIt(t *testing.T) {
	sc := SessionCookie{Secure: true, Domain: "tj.example"}

	ck := cookieFor(t, func(c *gin.Context) { sc.Clear(c) })

	if ck.Name != "tj_session" || ck.Value != "" || ck.MaxAge >= 0 || !ck.HttpOnly || !ck.Secure || ck.Path != "/" || ck.Domain != "tj.example" || ck.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie = %+v", ck)
	}
}

func TestSessionCookieReadReturnsTheValueOnlyWhenPresent(t *testing.T) {
	sc := SessionCookie{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req

	if v, ok := sc.Read(c); ok || v != "" {
		t.Errorf("sem cookie: %q, %v", v, ok)
	}
	req.AddCookie(&http.Cookie{Name: "tj_session", Value: "abc"})
	if v, ok := sc.Read(c); !ok || v != "abc" {
		t.Errorf("com cookie: %q, %v", v, ok)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "outro", Value: "x"})
	c.Request = req2
	if _, ok := sc.Read(c); ok {
		t.Error("outro cookie não conta")
	}
}
