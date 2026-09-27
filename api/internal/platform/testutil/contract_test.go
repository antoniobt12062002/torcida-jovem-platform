package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	platformapi "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/api"
)

// healthOnly serves the health check only; the audit query is not under test here.
type healthOnly struct{ platformapi.HealthHandler }

func (healthOnly) GetAuditLogs(*gin.Context, platformapi.GetAuditLogsParams) {}

func healthEngine(ping func(context.Context) error) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	platformapi.RegisterHandlers(r, healthOnly{platformapi.HealthHandler{Ping: ping}})
	return r
}

func get(r http.Handler, target string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return req, rec
}

func TestValidateResponseAcceptsTheRealHealthResponses(t *testing.T) {
	contracts := LoadContracts(t)

	req, rec := get(healthEngine(func(context.Context) error { return nil }), "/healthz")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	contracts.ValidateResponse(t, req, rec)

	req, rec = get(healthEngine(func(context.Context) error { return context.DeadlineExceeded }), "/healthz")
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
	contracts.ValidateResponse(t, req, rec)
}

func TestCheckResponseRejectsWhatTheContractDoesNotDescribe(t *testing.T) {
	contracts := LoadContracts(t)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	cases := []struct {
		name        string
		status      int
		contentType string
		body        string
	}{
		{"status não documentado", 500, "application/json", `{"status":"ok"}`},
		{"valor fora do enum", 200, "application/json", `{"status":"desconhecido"}`},
		{"campo obrigatório ausente", 200, "application/json", `{}`},
		{"tipo de campo errado", 200, "application/json", `{"status":1}`},
		{"tipo de conteúdo não documentado", 200, "text/plain", `ok`},
	}
	for _, tc := range cases {
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", tc.contentType)
		rec.WriteHeader(tc.status)
		_, _ = rec.WriteString(tc.body)
		if err := contracts.CheckResponse(req, rec); err == nil {
			t.Errorf("%s: a resposta deveria ser recusada", tc.name)
		}
	}
}

func TestCheckResponseRejectsARouteWithoutContract(t *testing.T) {
	contracts := LoadContracts(t)
	req := httptest.NewRequest(http.MethodGet, "/rota-sem-contrato", nil)
	rec := httptest.NewRecorder()
	rec.WriteHeader(http.StatusOK)

	err := contracts.CheckResponse(req, rec)
	if err == nil || !strings.Contains(err.Error(), "/rota-sem-contrato") {
		t.Errorf("esperava erro citando a rota sem contrato, veio %v", err)
	}
}

const problemContract = `openapi: 3.0.3
info: {title: t, version: "1"}
servers:
  - url: /
paths:
  /api/v1/thing:
    get:
      operationId: getThing
      responses:
        '403':
          description: proibido
          content:
            application/problem+json:
              schema:
                type: object
                required: [type, title, status, detail, code, request_id]
                properties:
                  type: {type: string}
                  title: {type: string}
                  status: {type: integer}
                  detail: {type: string}
                  code: {type: string}
                  request_id: {type: string}
`

func TestCheckResponseSupportsProblemJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "thing.yaml")
	if err := os.WriteFile(path, []byte(problemContract), 0o600); err != nil {
		t.Fatal(err)
	}
	contracts := LoadContractFiles(t, path)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/thing", nil)

	write := func(body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Type", "application/problem+json")
		rec.WriteHeader(http.StatusForbidden)
		_, _ = rec.WriteString(body)
		return rec
	}
	good := `{"type":"about:blank","title":"Forbidden","status":403,"detail":"x","code":"forbidden","request_id":"r-1"}`
	if err := contracts.CheckResponse(req, write(good)); err != nil {
		t.Errorf("problem+json válido foi recusado: %v", err)
	}
	if err := contracts.CheckResponse(req, write(`{"type":"about:blank","title":"Forbidden","status":403,"detail":"x","request_id":"r-1"}`)); err == nil {
		t.Error("problem+json sem `code` deveria ser recusado")
	}
}
