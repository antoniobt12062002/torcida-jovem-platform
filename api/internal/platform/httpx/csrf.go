package httpx

import (
	"crypto/subtle"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func changesState(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

// OriginConfig wires the Origin middleware.
type OriginConfig struct {
	// Allowed are the exact origins (scheme://host[:port]) that may call the API.
	// Empty means no origin is allowed (fail closed).
	Allowed []string
	// Cookie tells whether the request carries a session (authenticated request).
	Cookie SessionCookie
}

// Origin defends state-changing requests (POST, PUT, PATCH, DELETE), public
// routes included: an Origin header outside the allowed list answers 403
// `origin_not_allowed`. A missing Origin does not block by itself (non-browser
// clients omit it), except when the request carries a session cookie and the
// browser says it is cross-site (`Sec-Fetch-Site: cross-site`).
func Origin(cfg OriginConfig) gin.HandlerFunc {
	allowed := slices.Clone(cfg.Allowed)
	return func(c *gin.Context) {
		if !changesState(c.Request.Method) {
			c.Next()
			return
		}
		origin := c.GetHeader("Origin")
		if origin != "" {
			if !slices.Contains(allowed, origin) {
				deny(c)
				return
			}
			c.Next()
			return
		}
		if c.GetHeader("Sec-Fetch-Site") == "cross-site" {
			if token, ok := cfg.Cookie.Read(c); ok && token != "" {
				deny(c)
				return
			}
		}
		c.Next()
	}
}

func deny(c *gin.Context) {
	WriteProblem(c, http.StatusForbidden, "origin_not_allowed", "Origem não permitida.")
}

// CSRF requires, on authenticated state-changing requests, an X-CSRF-Token equal
// to the session's token (constant-time compare), else 403 `csrf_invalid`. It
// runs after Authn: routes with no session (public ones, unknown routes) are left
// alone, except that a state-changing request on a known route with no session in
// the context is denied, since that is a wiring mistake.
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !changesState(c.Request.Method) || c.FullPath() == "" {
			c.Next()
			return
		}
		info, ok := SessionFrom(c.Request.Context())
		if !ok {
			if publicRoute(c) {
				c.Next()
				return
			}
			invalidCSRF(c)
			return
		}
		got := c.GetHeader("X-CSRF-Token")
		if info.CSRFToken == "" || subtle.ConstantTimeCompare([]byte(got), []byte(info.CSRFToken)) != 1 {
			invalidCSRF(c)
			return
		}
		c.Next()
	}
}

func invalidCSRF(c *gin.Context) {
	WriteProblem(c, http.StatusForbidden, "csrf_invalid", "Token CSRF ausente ou inválido.")
}
