package httpx

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func originEngine(allowed []string) *gin.Engine {
	r := newEngine()
	r.Use(RequestID(), Origin(OriginConfig{Allowed: allowed, Cookie: SessionCookie{}}))
	ok := func(c *gin.Context) { c.String(http.StatusOK, "ok") }
	r.GET("/x", ok)
	r.HEAD("/x", ok)
	r.OPTIONS("/x", ok)
	r.POST("/x", ok)
	r.PUT("/x", ok)
	r.PATCH("/x", ok)
	r.DELETE("/x", ok)
	return r
}

func headers(kv ...string) func(*http.Request) {
	return func(r *http.Request) {
		for i := 0; i+1 < len(kv); i += 2 {
			r.Header.Set(kv[i], kv[i+1])
		}
	}
}

var stateChanging = []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

// IDN-03: Origin fora da lista bloqueia métodos que alteram estado.
func TestOriginBlocksStateChangingRequestsFromOutsideTheAllowedList(t *testing.T) {
	r := originEngine([]string{"https://app.tj.example"})

	for _, m := range stateChanging {
		w := serve(r, m, "/x", nil, headers("Origin", "https://evil.example"))
		p := decodeProblem(t, w)
		if w.Code != http.StatusForbidden || p["code"] != "origin_not_allowed" || p["request_id"] != "req-abc-12345" {
			t.Errorf("%s: status = %d, corpo = %v", m, w.Code, p)
		}
	}
}

func TestOriginAllowsTheConfiguredOriginsExactlyAndNothingSimilar(t *testing.T) {
	r := originEngine([]string{"https://app.tj.example", "https://admin.tj.example:8443"})

	for _, ok := range []string{"https://app.tj.example", "https://admin.tj.example:8443"} {
		if w := serve(r, http.MethodPost, "/x", nil, headers("Origin", ok)); w.Code != http.StatusOK {
			t.Errorf("%s deveria passar: %d", ok, w.Code)
		}
	}
	for _, bad := range []string{"http://app.tj.example", "https://app.tj.example:444", "https://app.tj.example.evil.example", "https://APP.tj.example", "null", "https://app.tj.example/"} {
		if w := serve(r, http.MethodPost, "/x", nil, headers("Origin", bad)); w.Code != http.StatusForbidden {
			t.Errorf("%q deveria ser bloqueada: %d", bad, w.Code)
		}
	}
}

// Lista vazia bloqueia toda origem informada (falha fechada).
func TestOriginWithAnEmptyListBlocksEveryOrigin(t *testing.T) {
	r := originEngine(nil)

	if w := serve(r, http.MethodPost, "/x", nil, headers("Origin", "https://app.tj.example")); w.Code != http.StatusForbidden {
		t.Errorf("status = %d", w.Code)
	}
}

func TestOriginDoesNotCheckSafeMethods(t *testing.T) {
	r := originEngine([]string{"https://app.tj.example"})

	for _, m := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		if w := serve(r, m, "/x", nil, headers("Origin", "https://evil.example")); w.Code != http.StatusOK {
			t.Errorf("%s: status = %d", m, w.Code)
		}
	}
}

// IDN-03.6: a ausência de Origin, sozinha, não bloqueia (clientes não-navegador).
func TestOriginAbsenceAloneDoesNotBlock(t *testing.T) {
	r := originEngine([]string{"https://app.tj.example"})

	for name, h := range map[string]func(*http.Request){
		"sem cabeçalhos":                       nil,
		"cookie sem Origin":                    withCookie("abc"),
		"same-origin":                          headers("Sec-Fetch-Site", "same-origin"),
		"cross-site sem cookie (rota pública)": headers("Sec-Fetch-Site", "cross-site"),
	} {
		if w := serve(r, http.MethodPost, "/x", nil, h); w.Code != http.StatusOK {
			t.Errorf("%s: status = %d", name, w.Code)
		}
	}
}

// IDN-03.6: requisição autenticada, sem Origin e cross-site é bloqueada.
func TestOriginBlocksAuthenticatedCrossSiteRequestsWithoutOrigin(t *testing.T) {
	r := originEngine([]string{"https://app.tj.example"})

	for _, m := range stateChanging {
		w := serve(r, m, "/x", nil, func(req *http.Request) {
			withCookie("abc")(req)
			req.Header.Set("Sec-Fetch-Site", "cross-site")
		})
		if w.Code != http.StatusForbidden || decodeProblem(t, w)["code"] != "origin_not_allowed" {
			t.Errorf("%s: status = %d, corpo = %s", m, w.Code, w.Body.String())
		}
	}
	// GET cross-site autenticado não é verificado.
	w := serve(r, http.MethodGet, "/x", nil, func(req *http.Request) {
		withCookie("abc")(req)
		req.Header.Set("Sec-Fetch-Site", "cross-site")
	})
	if w.Code != http.StatusOK {
		t.Errorf("GET: status = %d", w.Code)
	}
}

func TestOriginAllowedOriginPassesEvenWithCrossSiteFetchMetadata(t *testing.T) {
	r := originEngine([]string{"https://app.tj.example"})

	w := serve(r, http.MethodPost, "/x", nil, func(req *http.Request) {
		withCookie("abc")(req)
		req.Header.Set("Origin", "https://app.tj.example")
		req.Header.Set("Sec-Fetch-Site", "cross-site")
	})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
}

// ---- CSRF

func csrfEngine(info SessionInfo, authed bool) *gin.Engine {
	r := newEngine()
	r.Use(RequestID())
	if authed {
		r.Use(Authn(AuthnConfig{Validator: &fakeValidator{info: info}, Public: map[string]bool{"POST /pub": true}}))
	}
	r.Use(CSRF())
	ok := func(c *gin.Context) { c.String(http.StatusOK, "ok") }
	r.GET("/x", ok)
	r.POST("/x", ok)
	r.PUT("/x", ok)
	r.PATCH("/x", ok)
	r.DELETE("/x", ok)
	r.POST("/pub", ok)
	return r
}

// IDN-03.1: sem token, ou com token diferente do da sessão, é 403 csrf_invalid.
func TestCSRFRejectsStateChangingRequestsWithoutAMatchingToken(t *testing.T) {
	r := csrfEngine(validInfo(), true)

	for _, m := range stateChanging {
		for name, h := range map[string]func(*http.Request){
			"sem token":    withCookie("abc"),
			"token errado": func(req *http.Request) { withCookie("abc")(req); req.Header.Set("X-CSRF-Token", "csrf-xyz") },
			"token vazio":  func(req *http.Request) { withCookie("abc")(req); req.Header.Set("X-CSRF-Token", "") },
			"prefixo":      func(req *http.Request) { withCookie("abc")(req); req.Header.Set("X-CSRF-Token", "csrf-ab") },
			"token maior":  func(req *http.Request) { withCookie("abc")(req); req.Header.Set("X-CSRF-Token", "csrf-abcd") },
		} {
			w := serve(r, m, "/x", nil, h)
			p := decodeProblem(t, w)
			if w.Code != http.StatusForbidden || p["code"] != "csrf_invalid" || p["request_id"] != "req-abc-12345" {
				t.Errorf("%s %s: status = %d, corpo = %v", m, name, w.Code, p)
			}
		}
	}
}

func TestCSRFAcceptsTheSessionToken(t *testing.T) {
	r := csrfEngine(validInfo(), true)

	for _, m := range stateChanging {
		w := serve(r, m, "/x", nil, func(req *http.Request) { withCookie("abc")(req); req.Header.Set("X-CSRF-Token", "csrf-abc") })
		if w.Code != http.StatusOK {
			t.Errorf("%s: status = %d", m, w.Code)
		}
	}
}

func TestCSRFDoesNotCheckReads(t *testing.T) {
	r := csrfEngine(validInfo(), true)

	if w := serve(r, http.MethodGet, "/x", nil, withCookie("abc")); w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
}

// Rotas públicas não têm sessão: o CSRF não se aplica (a defesa delas é a origem).
func TestCSRFDoesNotApplyToPublicRoutes(t *testing.T) {
	r := csrfEngine(validInfo(), true)

	if w := serve(r, http.MethodPost, "/pub", nil, nil); w.Code != http.StatusOK {
		t.Errorf("status = %d", w.Code)
	}
}

// Sem sessão no contexto numa rota que muda estado, a montagem está errada: falha fechada.
func TestCSRFWithoutAnAuthenticatedSessionFailsClosed(t *testing.T) {
	r := csrfEngine(SessionInfo{}, false)

	w := serve(r, http.MethodPost, "/x", nil, headers("X-CSRF-Token", "qualquer"))

	if w.Code != http.StatusForbidden || decodeProblem(t, w)["code"] != "csrf_invalid" {
		t.Errorf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
}

func TestCSRFWithAnEmptySessionTokenNeverMatches(t *testing.T) {
	info := validInfo()
	info.CSRFToken = ""
	r := csrfEngine(info, true)

	w := serve(r, http.MethodPost, "/x", nil, func(req *http.Request) { withCookie("abc")(req); req.Header.Set("X-CSRF-Token", "") })

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d", w.Code)
	}
}

// Rota inexistente segue para o 404 do roteador, sem virar csrf_invalid.
func TestCSRFLeavesUnknownRoutesTo404(t *testing.T) {
	r := csrfEngine(validInfo(), true)

	if w := serve(r, http.MethodPost, "/nao-existe", nil, nil); w.Code != http.StatusNotFound {
		t.Errorf("status = %d", w.Code)
	}
}
