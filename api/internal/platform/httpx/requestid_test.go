package httpx

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

var uuidV4 = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func newEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func doRequest(t *testing.T, r *gin.Engine, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRequestIDReusesValidHeader(t *testing.T) {
	r := newEngine()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	for _, id := range []string{"abc-12345", "REQ-2026-09-27-a1b2c3d4", strings.Repeat("a", 8), strings.Repeat("a", 64)} {
		w := doRequest(t, r, map[string]string{"X-Request-Id": id})
		if got := w.Header().Get("X-Request-Id"); got != id {
			t.Errorf("id válido %q não foi reaproveitado: %q", id, got)
		}
	}
}

func TestRequestIDGeneratesWhenMissingOrInvalid(t *testing.T) {
	r := newEngine()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	invalid := map[string]string{
		"ausente":           "",
		"curto demais":      "abc-123",
		"longo demais":      strings.Repeat("a", 65),
		"caracter inválido": "abc def!!! 123",
		"quebra de linha":   "abc-12345\r\nX-Evil: 1",
	}
	for name, id := range invalid {
		headers := map[string]string{}
		if id != "" {
			headers["X-Request-Id"] = id
		}
		w := doRequest(t, r, headers)
		got := w.Header().Get("X-Request-Id")
		if !uuidV4.MatchString(got) {
			t.Errorf("%s: esperava um UUID v4 gerado, veio %q", name, got)
		}
	}
}

func TestRequestIDGeneratedValuesAreUnique(t *testing.T) {
	r := newEngine()
	r.Use(RequestID())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	seen := map[string]bool{}
	for range 50 {
		id := doRequest(t, r, nil).Header().Get("X-Request-Id")
		if seen[id] {
			t.Fatalf("id repetido: %s", id)
		}
		seen[id] = true
	}
}

func TestRequestIDIsAvailableInContexts(t *testing.T) {
	r := newEngine()
	r.Use(RequestID())
	var fromRequestCtx, fromGin string
	r.GET("/ping", func(c *gin.Context) {
		fromRequestCtx = RequestIDFrom(c.Request.Context())
		fromGin = c.GetString("request_id")
		c.String(http.StatusOK, "ok")
	})

	w := doRequest(t, r, map[string]string{"X-Request-Id": "abc-12345"})
	if fromRequestCtx != "abc-12345" || fromGin != "abc-12345" {
		t.Errorf("contexto da requisição = %q, gin = %q, esperado abc-12345", fromRequestCtx, fromGin)
	}
	if w.Header().Get("X-Request-Id") != "abc-12345" {
		t.Errorf("cabeçalho de resposta = %q", w.Header().Get("X-Request-Id"))
	}
}

func TestRequestIDFromWithoutMiddlewareIsEmpty(t *testing.T) {
	if got := RequestIDFrom(httptest.NewRequest(http.MethodGet, "/", nil).Context()); got != "" {
		t.Errorf("esperava vazio, veio %q", got)
	}
}
