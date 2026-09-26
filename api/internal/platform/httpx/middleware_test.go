package httpx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
)

func logLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for l := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("linha de log não é JSON: %q", l)
		}
		out = append(out, m)
	}
	return out
}

func serve(r *gin.Engine, method, target string, body io.Reader, mutate func(*http.Request)) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("X-Request-Id", "req-abc-12345")
	if mutate != nil {
		mutate(req)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRecoverTurnsPanicInto500WithoutInternalDetails(t *testing.T) {
	var logs bytes.Buffer
	r := newEngine()
	r.Use(RequestID(), Recover(logx.New("info", &logs)))
	r.GET("/boom", func(c *gin.Context) { panic("boom-segredo-interno") })
	r.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := serve(r, http.MethodGet, "/boom", nil, nil)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, esperado 500", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q", ct)
	}
	p := decodeProblem(t, w)
	if p["code"] != "internal_error" || p["request_id"] != "req-abc-12345" {
		t.Errorf("corpo = %v", p)
	}
	body := w.Body.String()
	for _, leak := range []string{"boom-segredo-interno", "goroutine", ".go:"} {
		if strings.Contains(body, leak) {
			t.Errorf("a resposta vazou %q: %s", leak, body)
		}
	}
	if !strings.Contains(logs.String(), "req-abc-12345") {
		t.Errorf("o log do panic deveria conter o request_id: %s", logs.String())
	}

	if again := serve(r, http.MethodGet, "/ok", nil, nil); again.Code != http.StatusOK {
		t.Errorf("o servidor deveria continuar atendendo após o panic, status = %d", again.Code)
	}
}

func TestAccessLogWritesOneStructuredLinePerRequest(t *testing.T) {
	var logs bytes.Buffer
	r := newEngine()
	r.Use(RequestID(), AccessLog(logx.New("info", &logs)))
	r.GET("/items", func(c *gin.Context) { c.String(http.StatusCreated, "ok") })

	serve(r, http.MethodGet, "/items?token=super-segredo&a=b", nil, nil)

	lines := logLines(t, &logs)
	if len(lines) != 1 {
		t.Fatalf("esperava 1 linha, veio %d: %s", len(lines), logs.String())
	}
	l := lines[0]
	if l["method"] != "GET" || l["path"] != "/items" || l["status"] != float64(201) || l["request_id"] != "req-abc-12345" {
		t.Errorf("linha = %v", l)
	}
	if d, ok := l["duration_ms"].(float64); !ok || d < 0 {
		t.Errorf("duration_ms = %v, esperado número maior ou igual a zero", l["duration_ms"])
	}
	if _, has := l["user_id"]; has {
		t.Errorf("user_id não deveria aparecer sem usuário: %v", l)
	}
	if strings.Contains(logs.String(), "super-segredo") || strings.Contains(logs.String(), "a=b") {
		t.Errorf("a query string vazou no log: %s", logs.String())
	}
}

func TestAccessLogIncludesUserIDWhenPresent(t *testing.T) {
	var logs bytes.Buffer
	r := newEngine()
	r.Use(RequestID(), AccessLog(logx.New("info", &logs)))
	r.GET("/me", func(c *gin.Context) {
		SetUserID(c, "user-42")
		c.String(http.StatusOK, "ok")
	})

	serve(r, http.MethodGet, "/me", nil, nil)
	if got := logLines(t, &logs)[0]["user_id"]; got != "user-42" {
		t.Errorf("user_id = %v, esperado user-42", got)
	}
}

func TestAccessLogRecordsStatus500AfterRecoveredPanic(t *testing.T) {
	var logs bytes.Buffer
	logger := logx.New("info", &logs)
	r := newEngine()
	r.Use(RequestID(), AccessLog(logger), Recover(logger))
	r.GET("/boom", func(c *gin.Context) { panic("x") })

	serve(r, http.MethodGet, "/boom", nil, nil)

	var access map[string]any
	for _, l := range logLines(t, &logs) {
		if l["msg"] == "request" {
			access = l
		}
	}
	if access == nil || access["status"] != float64(500) {
		t.Errorf("a linha de acesso deveria registrar status 500: %v", access)
	}
}

func bodyLimitEngine(limit int64, reached *bool, read *int) *gin.Engine {
	r := newEngine()
	r.Use(RequestID(), BodyLimit(limit))
	r.POST("/upload", func(c *gin.Context) {
		*reached = true
		b, _ := io.ReadAll(c.Request.Body)
		*read = len(b)
		c.String(http.StatusOK, "ok")
	})
	return r
}

func TestBodyLimitRejectsKnownLengthAboveLimit(t *testing.T) {
	var reached bool
	var read int
	r := bodyLimitEngine(10, &reached, &read)

	w := serve(r, http.MethodPost, "/upload", strings.NewReader(strings.Repeat("a", 11)), nil)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, esperado 413", w.Code)
	}
	p := decodeProblem(t, w)
	if p["code"] != "payload_too_large" || p["request_id"] != "req-abc-12345" {
		t.Errorf("corpo = %v", p)
	}
	if reached {
		t.Error("o handler não deveria executar")
	}
}

func TestBodyLimitRejectsUnknownLengthAboveLimit(t *testing.T) {
	var reached bool
	var read int
	r := bodyLimitEngine(10, &reached, &read)

	w := serve(r, http.MethodPost, "/upload", strings.NewReader(strings.Repeat("a", 11)), func(req *http.Request) {
		req.ContentLength = -1
	})

	if w.Code != http.StatusRequestEntityTooLarge || decodeProblem(t, w)["code"] != "payload_too_large" {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if reached {
		t.Error("o handler não deveria executar")
	}
}

func TestBodyLimitAcceptsBodyExactlyAtLimit(t *testing.T) {
	for name, mutate := range map[string]func(*http.Request){
		"tamanho conhecido":    nil,
		"tamanho desconhecido": func(req *http.Request) { req.ContentLength = -1 },
	} {
		var reached bool
		var read int
		r := bodyLimitEngine(10, &reached, &read)
		w := serve(r, http.MethodPost, "/upload", strings.NewReader(strings.Repeat("a", 10)), mutate)
		if w.Code != http.StatusOK || !reached || read != 10 {
			t.Errorf("%s: status = %d, handler = %v, bytes lidos = %d", name, w.Code, reached, read)
		}
	}
}

func TestDefaultBodyLimitIsOneMiB(t *testing.T) {
	var reached bool
	var read int
	r := bodyLimitEngine(DefaultBodyLimit, &reached, &read)

	over := serve(r, http.MethodPost, "/upload", bytes.NewReader(make([]byte, 1<<20+1)), nil)
	if over.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("1 MiB + 1 byte: status = %d, esperado 413", over.Code)
	}

	reached = false
	exact := serve(r, http.MethodPost, "/upload", bytes.NewReader(make([]byte, 1<<20)), nil)
	if exact.Code != http.StatusOK || read != 1<<20 {
		t.Errorf("1 MiB exato: status = %d, bytes lidos = %d", exact.Code, read)
	}
}
