package httpx

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// Errors a SessionValidator returns; Authn maps them to 401 responses. Any
// other error is an internal failure (500, with no detail).
var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrSessionExpired  = errors.New("session_expired")
)

// SessionInfo is what a valid session gives the request.
type SessionInfo struct {
	Principal          authz.Principal
	SessionID          string
	CSRFToken          string
	MustChangePassword bool
}

// SessionValidator is defined here, not in identity, so the HTTP layer does not
// import the module: identity adapts its session service to it.
type SessionValidator interface {
	Validate(ctx context.Context, token string) (SessionInfo, error)
}

const (
	sessionCookieName = "tj_session"
	maxCookieTokenLen = 512
)

// SessionCookie sets, clears and reads the session cookie: HttpOnly,
// SameSite=Lax, Path=/, Secure and Domain from the configuration.
type SessionCookie struct {
	Domain string
	Secure bool
}

func (sc SessionCookie) cookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name: sessionCookieName, Value: value, Path: "/", Domain: sc.Domain, MaxAge: maxAge,
		HttpOnly: true, Secure: sc.Secure, SameSite: http.SameSiteLaxMode,
	}
}

// Set issues the cookie, living as long as the session's absolute limit.
func (sc SessionCookie) Set(c *gin.Context, token string, lifetime time.Duration) {
	http.SetCookie(c.Writer, sc.cookie(token, int(lifetime.Seconds())))
}

// Clear expires the cookie in the browser.
func (sc SessionCookie) Clear(c *gin.Context) {
	http.SetCookie(c.Writer, sc.cookie("", -1))
}

// Read returns the cookie value, if the request has the session cookie.
func (sc SessionCookie) Read(c *gin.Context) (string, bool) {
	ck, err := c.Request.Cookie(sessionCookieName)
	if err != nil {
		return "", false
	}
	return ck.Value, true
}

// AuthnConfig wires the Authn middleware. Routes are identified by
// "METHOD /path/{pattern}" as the router registered them (c.FullPath()).
type AuthnConfig struct {
	Validator SessionValidator
	Cookie    SessionCookie
	// Public lists the routes that need no session. Every other route needs one.
	Public map[string]bool
	// AllowedWhilePasswordChangeIsRequired lists the routes a user with
	// must_change_password may still call; all others answer 403.
	AllowedWhilePasswordChangeIsRequired map[string]bool
	// OnAuthenticated may enrich the request context (for example with the
	// audit actor) once the session is valid.
	OnAuthenticated func(ctx context.Context, info SessionInfo) context.Context
}

const publicRouteKey = "httpx.public_route"

// publicRoute reports whether Authn let the route through as public.
func publicRoute(c *gin.Context) bool { return c.GetBool(publicRouteKey) }

type sessionKey struct{}

// SessionFrom returns the session Authn stored in the request context.
func SessionFrom(ctx context.Context) (SessionInfo, bool) {
	info, ok := ctx.Value(sessionKey{}).(SessionInfo)
	return info, ok
}

// PrincipalFrom returns the authenticated principal, if any.
func PrincipalFrom(ctx context.Context) (authz.Principal, bool) {
	info, ok := SessionFrom(ctx)
	return info.Principal, ok
}

// Authn authenticates the request from the session cookie: everything outside
// the public list needs a valid session (deny by default). It answers 401
// `unauthenticated` (no, malformed or unknown session) or `session_expired`
// (which also clears the cookie), 500 with no detail for other failures, and 403
// `password_change_required` while must_change_password is true, except on the
// allowed routes. A route that does not exist is left to the 404 of the router.
func Authn(cfg AuthnConfig) gin.HandlerFunc {
	if cfg.Validator == nil {
		panic("httpx: Authn exige um SessionValidator")
	}
	return func(c *gin.Context) {
		if c.FullPath() == "" { // no such route: the router answers 404
			c.Next()
			return
		}
		route := c.Request.Method + " " + c.FullPath()
		if cfg.Public[route] {
			c.Set(publicRouteKey, true)
			c.Next()
			return
		}
		token, ok := cfg.Cookie.Read(c)
		if !ok || token == "" || len(token) > maxCookieTokenLen {
			WriteProblem(c, http.StatusUnauthorized, "unauthenticated", "É preciso entrar para continuar.")
			return
		}
		info, err := cfg.Validator.Validate(c.Request.Context(), token)
		switch {
		case errors.Is(err, ErrSessionExpired):
			cfg.Cookie.Clear(c)
			WriteProblem(c, http.StatusUnauthorized, "session_expired", "A sessão expirou. Entre novamente.")
			return
		case errors.Is(err, ErrUnauthenticated):
			WriteProblem(c, http.StatusUnauthorized, "unauthenticated", "É preciso entrar para continuar.")
			return
		case database.Unavailable(err):
			WriteProblem(c, http.StatusServiceUnavailable, "service_unavailable", "Serviço temporariamente indisponível. Tente novamente em instantes.")
			return
		case err != nil:
			WriteProblem(c, http.StatusInternalServerError, "internal_error", "Erro interno. Informe o request_id ao suporte.")
			return
		}
		if info.MustChangePassword && !cfg.AllowedWhilePasswordChangeIsRequired[route] {
			WriteProblem(c, http.StatusForbidden, "password_change_required", "Troque a senha para continuar.")
			return
		}
		ctx := context.WithValue(c.Request.Context(), sessionKey{}, info)
		if cfg.OnAuthenticated != nil {
			ctx = cfg.OnAuthenticated(ctx, info)
		}
		c.Request = c.Request.WithContext(ctx)
		SetUserID(c, info.Principal.UserID)
		c.Next()
	}
}
