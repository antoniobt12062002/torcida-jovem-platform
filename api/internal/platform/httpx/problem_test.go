package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type signup struct {
	Email   string  `json:"email" binding:"required,email"`
	Name    string  `json:"name" binding:"required,min=2"`
	Age     int     `json:"age"`
	Address address `json:"address" binding:"required"`
}

type address struct {
	Zip string `json:"zip" binding:"required"`
}

func problemEngine() *gin.Engine {
	r := newEngine()
	r.Use(RequestID())
	return r
}

func decodeProblem(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("corpo não é JSON: %q: %v", w.Body.String(), err)
	}
	return m
}

func post(r *gin.Engine, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Request-Id", "req-abc-12345")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func signupEngine() *gin.Engine {
	r := problemEngine()
	r.POST("/signup", func(c *gin.Context) {
		var in signup
		if !BindJSON(c, &in) {
			return
		}
		c.JSON(http.StatusOK, in)
	})
	return r
}

func errorsOf(t *testing.T, p map[string]any) map[string]string {
	t.Helper()
	list, ok := p["errors"].([]any)
	if !ok {
		t.Fatalf("campo errors ausente ou inválido: %v", p)
	}
	out := map[string]string{}
	for _, item := range list {
		m := item.(map[string]any)
		out[m["field"].(string)] = m["code"].(string)
	}
	return out
}

func TestWriteProblemBody(t *testing.T) {
	r := problemEngine()
	r.GET("/x", func(c *gin.Context) {
		WriteProblem(c, http.StatusConflict, "email_taken", "E-mail já cadastrado.")
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Request-Id", "req-abc-12345")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, esperado 409", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, esperado application/problem+json", ct)
	}
	p := decodeProblem(t, w)
	want := map[string]any{
		"type": "about:blank", "title": "Conflict", "status": float64(409),
		"detail": "E-mail já cadastrado.", "code": "email_taken", "request_id": "req-abc-12345",
	}
	for k, v := range want {
		if p[k] != v {
			t.Errorf("%s = %v, esperado %v", k, p[k], v)
		}
	}
	if _, has := p["errors"]; has {
		t.Errorf("errors não deveria existir sem erros de campo: %v", p)
	}
}

func TestWriteProblemWithFieldErrors(t *testing.T) {
	r := problemEngine()
	r.GET("/x", func(c *gin.Context) {
		WriteProblem(c, http.StatusUnprocessableEntity, "validation_failed", "Campos inválidos.",
			FieldError{Field: "email", Code: "required"}, FieldError{Field: "name", Code: "min"})
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	got := errorsOf(t, decodeProblem(t, w))
	if got["email"] != "required" || got["name"] != "min" || len(got) != 2 {
		t.Errorf("errors = %v", got)
	}
}

func TestWriteProblemStopsTheChain(t *testing.T) {
	r := problemEngine()
	reached := false
	r.GET("/x", func(c *gin.Context) {
		WriteProblem(c, http.StatusForbidden, "forbidden", "Acesso negado.")
	}, func(c *gin.Context) { reached = true })
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	if reached {
		t.Error("o handler seguinte não deveria executar após WriteProblem")
	}
}

func TestBindJSONMalformedReturns400(t *testing.T) {
	r := signupEngine()
	for name, body := range map[string]string{
		"sintaxe inválida": `{"email":`,
		"corpo vazio":      ``,
		"não é JSON":       `isto nao e json`,
	} {
		w := post(r, body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, esperado 400", name, w.Code)
		}
		p := decodeProblem(t, w)
		if p["code"] != "invalid_json" || p["request_id"] != "req-abc-12345" {
			t.Errorf("%s: corpo = %v", name, p)
		}
	}
}

func TestBindJSONValidationReturns422WithAllFields(t *testing.T) {
	r := signupEngine()
	w := post(r, `{"email":"nao-e-email","name":"a"}`)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, esperado 422", w.Code)
	}
	p := decodeProblem(t, w)
	if p["code"] != "validation_failed" {
		t.Errorf("code = %v, esperado validation_failed", p["code"])
	}
	got := errorsOf(t, p)
	want := map[string]string{"email": "email", "name": "min", "address.zip": "required"}
	for f, c := range want {
		if got[f] != c {
			t.Errorf("errors[%s] = %q, esperado %q (todos: %v)", f, got[f], c, got)
		}
	}
}

func TestBindJSONMissingFieldsAreRequired(t *testing.T) {
	w := post(signupEngine(), `{}`)
	got := errorsOf(t, decodeProblem(t, w))
	for _, f := range []string{"email", "name", "address.zip"} {
		if got[f] != "required" {
			t.Errorf("errors[%s] = %q, esperado required (todos: %v)", f, got[f], got)
		}
	}
}

func TestBindJSONTypeMismatchReturns422(t *testing.T) {
	w := post(signupEngine(), `{"email":"a@b.co","name":"Ana","age":"vinte","address":{"zip":"86300"}}`)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, esperado 422", w.Code)
	}
	p := decodeProblem(t, w)
	if p["code"] != "validation_failed" {
		t.Errorf("code = %v", p["code"])
	}
	if got := errorsOf(t, p); got["age"] != "invalid_type" {
		t.Errorf("errors = %v, esperado age=invalid_type", got)
	}
}

func TestBindJSONValidBodyPopulatesAndContinues(t *testing.T) {
	w := post(signupEngine(), `{"email":"a@b.co","name":"Ana","age":30,"address":{"zip":"86300"}}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado 200: %s", w.Code, w.Body.String())
	}
	var got signup
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Email != "a@b.co" || got.Name != "Ana" || got.Age != 30 || got.Address.Zip != "86300" {
		t.Errorf("corpo = %+v", got)
	}
}
