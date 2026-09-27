package httpapi

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

func loadPlatform(t *testing.T) *openapi3.T {
	t.Helper()
	doc, err := httpx.LoadContract(filepath.Join("..", "..", "openapi", "platform.yaml"))
	if err != nil {
		t.Fatalf("o contrato do platform deveria ser válido: %v", err)
	}
	return doc
}

func auditOp(t *testing.T) *openapi3.Operation {
	t.Helper()
	item := loadPlatform(t).Paths.Find("/api/v1/audit-logs")
	if item == nil || item.Get == nil {
		t.Fatal("GET /api/v1/audit-logs deveria existir no contrato do platform")
	}
	return item.Get
}

// AUD-03: a consulta exige sessão; o /healthz continua público.
func TestAuditQueryRequiresASessionAndHealthStaysPublic(t *testing.T) {
	doc := loadPlatform(t)

	if op := doc.Paths.Find("/healthz").Get; !isPublic(op) {
		t.Error("/healthz deveria continuar público")
	}
	op := auditOp(t)
	reqs := doc.Security
	if op.Security != nil {
		reqs = *op.Security
	}
	if len(reqs) != 1 || reqs[0]["sessionCookie"] == nil {
		t.Errorf("a consulta de auditoria deveria exigir o cookie de sessão: %v", reqs)
	}
}

// AUD-03.1 a .3: filtros, cursor e limite de até 100.
func TestAuditQueryHasTheFiltersCursorAndALimitOfAtMost100(t *testing.T) {
	op := auditOp(t)
	got := map[string]*openapi3.Parameter{}
	var names []string
	for _, p := range op.Parameters {
		got[p.Value.Name] = p.Value
		names = append(names, p.Value.Name)
	}
	slices.Sort(names)
	want := []string{"action", "actor_user_id", "cursor", "entity_id", "entity_type", "from", "limit", "outcome", "to"}
	if !slices.Equal(names, want) {
		t.Fatalf("parâmetros = %v, esperado %v", names, want)
	}
	limit := got["limit"].Schema.Value
	if limit.Max == nil || *limit.Max != 100 || limit.Min == nil || *limit.Min != 1 || limit.Default != 50 && limit.Default != float64(50) {
		t.Errorf("limit = min %v, max %v, default %v", limit.Min, limit.Max, limit.Default)
	}
	if got["actor_user_id"].Schema.Value.Format != "uuid" || got["from"].Schema.Value.Format != "date-time" || got["to"].Schema.Value.Format != "date-time" {
		t.Error("actor_user_id é uuid; from e to são date-time")
	}
	outcomes := got["outcome"].Schema.Value.Enum
	if len(outcomes) != 3 {
		t.Errorf("outcome = %v", outcomes)
	}
	for _, p := range got {
		if p.Required {
			t.Errorf("o filtro %s deveria ser opcional", p.Name)
		}
	}
}

// AUD-01.3: o registro tem ator, alvo, resultado, contexto e request_id.
func TestAuditLogSchemaHasActorTargetOutcomeContextAndRequestID(t *testing.T) {
	doc := loadPlatform(t)
	schema := doc.Components.Schemas["AuditLog"].Value

	want := []string{"action", "actor_type", "actor_user_id", "after", "before", "context", "entity_id", "entity_type", "id", "occurred_at", "outcome", "reason", "request_id"}
	if !slices.Equal(props(schema), want) {
		t.Errorf("AuditLog = %v", props(schema))
	}
	if !schema.Properties["actor_user_id"].Value.Nullable || !schema.Properties["reason"].Value.Nullable {
		t.Error("actor_user_id e reason podem ser nulos")
	}
	if len(schema.Properties["actor_type"].Value.Enum) != 3 || len(schema.Properties["outcome"].Value.Enum) != 3 {
		t.Error("actor_type e outcome têm três valores cada")
	}
	page := doc.Components.Schemas["AuditLogPage"].Value
	if !slices.Equal(props(page), []string{"items", "next_cursor"}) || !page.Properties["next_cursor"].Value.Nullable {
		t.Errorf("AuditLogPage = %v", props(page))
	}
}

// AUD-03.3 e .4: 401, 403 e 422 em problem+json; leitura não exige CSRF.
func TestAuditQueryDescribesItsErrorsAndNeedsNoCSRF(t *testing.T) {
	op := auditOp(t)

	for _, code := range []string{"401", "403", "422"} {
		r := op.Responses.Value(code)
		if r == nil || r.Value.Content.Get("application/problem+json") == nil {
			t.Errorf("a resposta %s deveria ser problem+json", code)
		}
	}
	for _, p := range op.Parameters {
		if p.Value.Name == "X-CSRF-Token" {
			t.Error("GET não exige X-CSRF-Token")
		}
	}
}
