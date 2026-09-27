package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

const thingContract = `openapi: 3.0.3
info: {title: t, version: "1"}
servers:
  - url: /
components:
  securitySchemes:
    sessionCookie: {type: apiKey, in: cookie, name: tj_session}
paths:
  /api/v1/protected:
    get:
      operationId: getProtected
      security:
        - sessionCookie: []
      responses:
        '200': {description: ok}
  /api/v1/things:
    post:
      operationId: createThing
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name, quantity]
              properties:
                name: {type: string, minLength: 1}
                quantity: {type: integer}
                kind: {type: string, enum: [a, b]}
                owner_id: {type: string, format: uuid}
      responses:
        '201': {description: criado}
    get:
      operationId: listThings
      parameters:
        - {name: limit, in: query, schema: {type: integer, minimum: 1, maximum: 100}}
      responses:
        '200': {description: ok}
`

func loadThingContract(t *testing.T) *openapi3.T {
	t.Helper()
	path := filepath.Join(t.TempDir(), "thing.yaml")
	if err := os.WriteFile(path, []byte(thingContract), 0o600); err != nil {
		t.Fatal(err)
	}
	doc, err := LoadContract(path)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

// engine wires RequestID + validator; reached records whether a handler ran.
func engine(t *testing.T, reached *bool, domain func(*gin.Context)) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	validator, err := NewContractValidator(loadThingContract(t))
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(RequestID(), validator)
	h := func(c *gin.Context) {
		*reached = true
		if domain != nil {
			domain(c)
			return
		}
		c.Status(http.StatusCreated)
	}
	r.POST("/api/v1/things", h)
	r.GET("/api/v1/things", h)
	r.GET("/livre", h)
	return r
}

func doContract(r http.Handler, method, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

type contractProblem struct {
	Status    int          `json:"status"`
	Code      string       `json:"code"`
	RequestID string       `json:"request_id"`
	Errors    []FieldError `json:"errors"`
}

func decodeContractProblem(t *testing.T, rec *httptest.ResponseRecorder) contractProblem {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q", ct)
	}
	var p contractProblem
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("corpo não é JSON: %v: %s", err, rec.Body.String())
	}
	return p
}

func hasError(errs []FieldError, field, code string) bool {
	for _, e := range errs {
		if e.Field == field && e.Code == code {
			return true
		}
	}
	return false
}

func TestValidatorRejectsMissingRequiredField(t *testing.T) {
	var reached bool
	rec := doContract(engine(t, &reached, nil), http.MethodPost, "/api/v1/things", `{"quantity": 2}`)

	if rec.Code != http.StatusUnprocessableEntity || reached {
		t.Fatalf("status = %d, handler executado = %v", rec.Code, reached)
	}
	p := decodeContractProblem(t, rec)
	if p.Code != "validation_failed" || p.Status != 422 || p.RequestID == "" {
		t.Errorf("problem = %+v", p)
	}
	if !hasError(p.Errors, "name", "required") {
		t.Errorf("errors = %+v", p.Errors)
	}
}

func TestValidatorRejectsWrongTypeEnumAndFormat(t *testing.T) {
	var reached bool
	body := `{"name":"x","quantity":"muitos","kind":"z","owner_id":"não-uuid"}`
	rec := doContract(engine(t, &reached, nil), http.MethodPost, "/api/v1/things", body)

	if rec.Code != http.StatusUnprocessableEntity || reached {
		t.Fatalf("status = %d, handler executado = %v", rec.Code, reached)
	}
	p := decodeContractProblem(t, rec)
	for _, want := range []FieldError{
		{Field: "quantity", Code: "invalid_type"},
		{Field: "kind", Code: "invalid_enum"},
		{Field: "owner_id", Code: "invalid_format"},
	} {
		if !hasError(p.Errors, want.Field, want.Code) {
			t.Errorf("faltou %+v em %+v", want, p.Errors)
		}
	}
}

func TestValidatorReportsInvalidQueryParameter(t *testing.T) {
	var reached bool
	rec := doContract(engine(t, &reached, nil), http.MethodGet, "/api/v1/things?limit=abc", "")

	if rec.Code != http.StatusUnprocessableEntity || reached {
		t.Fatalf("status = %d", rec.Code)
	}
	if p := decodeContractProblem(t, rec); !hasError(p.Errors, "limit", "invalid_type") {
		t.Errorf("errors = %+v", p.Errors)
	}
}

// API-02.5: o limite fora de 1 a 100 tem código próprio, sem perder o erro por campo.
func TestValidatorNamesTheCodeInvalidLimitWhenTheLimitIsOutOfRange(t *testing.T) {
	for target, field := range map[string]string{
		"/api/v1/things?limit=101": "too_large",
		"/api/v1/things?limit=0":   "too_small",
	} {
		var reached bool
		rec := doContract(engine(t, &reached, nil), http.MethodGet, target, "")

		p := decodeContractProblem(t, rec)
		if rec.Code != http.StatusUnprocessableEntity || reached || p.Code != "invalid_limit" || !hasError(p.Errors, "limit", field) {
			t.Errorf("%s: status = %d, code = %q, errors = %+v", target, rec.Code, p.Code, p.Errors)
		}
	}
}

func TestValidatorKeepsValidationFailedWhenTheLimitIsNotANumber(t *testing.T) {
	var reached bool
	rec := doContract(engine(t, &reached, nil), http.MethodGet, "/api/v1/things?limit=abc", "")

	if p := decodeContractProblem(t, rec); p.Code != "validation_failed" {
		t.Errorf("code = %q", p.Code)
	}
}

func TestValidatorRejectsMalformedJSONWith400(t *testing.T) {
	var reached bool
	rec := doContract(engine(t, &reached, nil), http.MethodPost, "/api/v1/things", `{"name": `)

	if rec.Code != http.StatusBadRequest || reached {
		t.Fatalf("status = %d, handler executado = %v", rec.Code, reached)
	}
	p := decodeContractProblem(t, rec)
	if p.Code != "invalid_json" || p.RequestID == "" || len(p.Errors) != 0 {
		t.Errorf("problem = %+v", p)
	}
}

func TestValidatorRejectsMissingRequiredBody(t *testing.T) {
	var reached bool
	rec := doContract(engine(t, &reached, nil), http.MethodPost, "/api/v1/things", "")

	if reached || (rec.Code != http.StatusUnprocessableEntity && rec.Code != http.StatusBadRequest) {
		t.Fatalf("status = %d, handler executado = %v", rec.Code, reached)
	}
	decodeContractProblem(t, rec)
}

func TestValidRequestReachesTheHandlerWithBodyIntact(t *testing.T) {
	var reached bool
	var seen string
	r := engine(t, &reached, func(c *gin.Context) {
		var in struct{ Name string }
		if !BindJSON(c, &in) {
			return
		}
		seen = in.Name
		c.Status(http.StatusCreated)
	})
	rec := doContract(r, http.MethodPost, "/api/v1/things", `{"name":"copo","quantity":3,"kind":"a"}`)

	if rec.Code != http.StatusCreated || !reached || seen != "copo" {
		t.Errorf("status = %d, handler = %v, name = %q", rec.Code, reached, seen)
	}
}

func TestValidatorDoesNotAlterDomainErrors(t *testing.T) {
	var reached bool
	r := engine(t, &reached, func(c *gin.Context) {
		WriteProblem(c, http.StatusConflict, "stock_negative", "Estoque insuficiente.")
	})
	rec := doContract(r, http.MethodPost, "/api/v1/things", `{"name":"copo","quantity":3}`)

	if rec.Code != http.StatusConflict || !reached {
		t.Fatalf("status = %d, handler = %v", rec.Code, reached)
	}
	if p := decodeContractProblem(t, rec); p.Code != "stock_negative" || len(p.Errors) != 0 {
		t.Errorf("problem = %+v", p)
	}
}

// Rotas sem operação no contrato seguem para o handler: recusá-las em produção
// é decisão futura (registrada no STATE.md), não regra deste middleware.
func TestValidatorPassesThroughRoutesWithoutContract(t *testing.T) {
	var reached bool
	rec := doContract(engine(t, &reached, nil), http.MethodGet, "/livre", "")

	if rec.Code != http.StatusCreated || !reached {
		t.Errorf("status = %d, handler = %v", rec.Code, reached)
	}
}

func TestNewContractValidatorRequiresAContract(t *testing.T) {
	if _, err := NewContractValidator(); err == nil {
		t.Error("esperava erro sem contratos")
	}
}

// Autenticação e autorização são dos módulos de domínio: o validador não pode
// recusar uma rota protegida por falta de cookie.
func TestValidatorDoesNotEnforceSecuritySchemes(t *testing.T) {
	var reached bool
	r := engine(t, &reached, nil)
	r.GET("/api/v1/protected", func(c *gin.Context) { reached = true; c.Status(http.StatusOK) })
	rec := doContract(r, http.MethodGet, "/api/v1/protected", "")

	if rec.Code != http.StatusOK || !reached {
		t.Errorf("status = %d, handler = %v", rec.Code, reached)
	}
}
