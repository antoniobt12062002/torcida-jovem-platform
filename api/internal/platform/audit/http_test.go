//go:build integration

package audit_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	platformapi "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/api"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

type staticSession struct{ principal authz.Principal }

func (s staticSession) Validate(_ context.Context, token string) (httpx.SessionInfo, error) {
	if token != "ok" {
		return httpx.SessionInfo{}, httpx.ErrUnauthenticated
	}
	return httpx.SessionInfo{Principal: s.principal, SessionID: "s", CSRFToken: "c"}, nil
}

type queryEnv struct {
	env
	engine   *gin.Engine
	contract *testutil.Contract
}

// newQueryEnv mounts the audit query behind Authn, with the contract validator.
func newQueryEnv(t *testing.T, perms ...authz.Permission) queryEnv {
	t.Helper()
	e := newEnv(t)
	authorizer, err := authz.NewAuthorizer([]authz.Definition{{Permission: audit.ReadPermission}}, audit.DeniedHook(e.rec))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := platformapi.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	validate, err := httpx.NewContractValidator(spec)
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(httpx.RequestID())
	r.Use(httpx.Authn(httpx.AuthnConfig{Validator: staticSession{principal(perms...)}}))
	wrapper := platformapi.ServerInterfaceWrapper{
		Handler: struct {
			platformapi.HealthHandler
			audit.Handler
		}{Handler: audit.Handler{Query: &audit.Query{Authz: authorizer, DB: e.app}}},
		ErrorHandler: func(c *gin.Context, _ error, _ int) {
			httpx.WriteProblem(c, http.StatusUnprocessableEntity, "validation_failed", "Parâmetros inválidos.")
		},
	}
	r.GET("/api/v1/audit-logs", validate, wrapper.GetAuditLogs)
	return queryEnv{env: e, engine: r, contract: testutil.LoadContracts(t)}
}

func (q queryEnv) get(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.AddCookie(&http.Cookie{Name: "tj_session", Value: "ok"})
	w := httptest.NewRecorder()
	q.engine.ServeHTTP(w, req)
	q.contract.ValidateResponse(t, req, w)
	return w
}

// seed inserts n entries, one minute apart, oldest first, with the given attributes.
func (q queryEnv) seed(t *testing.T, n int, action, entityType, entityID, outcome string, minuteOffset int) {
	t.Helper()
	err := q.owner.Exec(`INSERT INTO audit_log (occurred_at, actor_type, actor_user_id, action, entity_type, entity_id, outcome, context)
		SELECT '2026-09-26 12:00:00+00'::timestamptz + make_interval(mins => ? + g), 'user', ?::uuid, ?, ?, ?, ?, '{"n":1}'::jsonb
		FROM generate_series(1, ?) AS g`, minuteOffset, actorID, action, entityType, entityID, outcome, n).Error
	if err != nil {
		t.Fatal(err)
	}
}

type page struct {
	Items []struct {
		ID         string         `json:"id"`
		Action     string         `json:"action"`
		EntityType string         `json:"entity_type"`
		EntityID   string         `json:"entity_id"`
		Outcome    string         `json:"outcome"`
		OccurredAt string         `json:"occurred_at"`
		Context    map[string]any `json:"context"`
	} `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

func decodePage(t *testing.T, w *httptest.ResponseRecorder) page {
	t.Helper()
	var p page
	if err := json.Unmarshal(w.Body.Bytes(), &p); err != nil {
		t.Fatalf("corpo: %v: %s", err, w.Body.String())
	}
	return p
}

func problemCode(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var p struct{ Code string }
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	return p.Code
}

// AUD-03.1: mais novo primeiro, 50 por página por padrão, cursor sem repetir nem pular.
func TestAuditQueryReturnsNewestFirstFiftyPerPageAndPagesWithoutGaps(t *testing.T) {
	q := newQueryEnv(t, audit.ReadPermission)
	q.seed(t, 120, "user.create", "user", "u-1", "success", 0)

	first := q.get(t, "/api/v1/audit-logs")
	p1 := decodePage(t, first)
	if first.Code != http.StatusOK || len(p1.Items) != 50 || p1.NextCursor == nil {
		t.Fatalf("primeira página: %d, %d itens, cursor %v", first.Code, len(p1.Items), p1.NextCursor)
	}
	seen := map[string]bool{}
	var times []string
	collect := func(p page) {
		for _, it := range p.Items {
			if seen[it.ID] {
				t.Errorf("entrada repetida: %s", it.ID)
			}
			seen[it.ID] = true
			times = append(times, it.OccurredAt)
		}
	}
	collect(p1)
	next := *p1.NextCursor
	for pages := 0; next != ""; pages++ {
		if pages > 20 {
			t.Fatal("a paginação não termina: o cursor não avança")
		}
		p := decodePage(t, q.get(t, "/api/v1/audit-logs?cursor="+next))
		collect(p)
		next = ""
		if p.NextCursor != nil {
			next = *p.NextCursor
		}
	}
	if len(seen) != 120 {
		t.Errorf("entradas vistas = %d, esperado 120", len(seen))
	}
	if !slices.IsSortedFunc(times, func(a, b string) int { return -1 * compare(a, b) }) {
		t.Error("as páginas deveriam vir do mais novo para o mais antigo")
	}
}

func compare(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// AUD-03.1: entradas com o mesmo instante não se perdem entre páginas (desempate por id).
func TestAuditQueryKeepsEntriesWithTheSameInstantAcrossPages(t *testing.T) {
	q := newQueryEnv(t, audit.ReadPermission)
	// Without the index the planner sorts, so only the (occurred_at, id) order keeps the pages consistent.
	if err := q.owner.Exec(`DROP INDEX audit_log_time_idx`).Error; err != nil {
		t.Fatal(err)
	}
	if err := q.owner.Exec(`INSERT INTO audit_log (occurred_at, actor_type, action, entity_type, entity_id, outcome)
		SELECT '2026-09-26 12:00:00+00', 'system', 'rbac.sync', 'rbac', 'x' || g, 'success' FROM generate_series(1, 25) g`).Error; err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	next := ""
	for pages := 0; ; pages++ {
		if pages > 20 {
			t.Fatal("a paginação não termina: o cursor não avança")
		}
		target := "/api/v1/audit-logs?limit=3"
		if next != "" {
			target += "&cursor=" + next
		}
		p := decodePage(t, q.get(t, target))
		if len(p.Items) > 3 {
			t.Fatalf("a página %d tem %d itens, o limit era 3", pages+1, len(p.Items))
		}
		for _, it := range p.Items {
			seen[it.ID] = true
		}
		if p.NextCursor == nil {
			break
		}
		if *p.NextCursor == "" || *p.NextCursor == next {
			t.Fatalf("o cursor precisa avançar entre as páginas: %q depois de %q", *p.NextCursor, next)
		}
		next = *p.NextCursor
	}
	if len(seen) != 25 {
		t.Errorf("entradas vistas = %d, esperado 25", len(seen))
	}
}

// AUD-03.2: os filtros se combinam por E.
func TestAuditQueryCombinesTheFiltersWithAnd(t *testing.T) {
	q := newQueryEnv(t, audit.ReadPermission)
	q.seed(t, 2, "user.create", "user", "u-1", "success", 0)
	q.seed(t, 3, "user.deactivate", "user", "u-1", "success", 10)
	q.seed(t, 4, "user.deactivate", "user", "u-2", "denied", 20)
	q.seed(t, 5, "admin.promote", "user", "u-2", "denied", 30)
	q.seed(t, 6, "rbac.sync", "rbac", "matrix", "success", 40)

	cases := map[string]int{
		"entity_type=user":                                      14,
		"entity_type=rbac":                                      6,
		"entity_type=user&entity_id=matrix":                     0,
		"entity_id=u-1":                                         5,
		"action=user.deactivate":                                7,
		"outcome=denied":                                        9,
		"action=user.deactivate&outcome=denied":                 4,
		"entity_id=u-2&outcome=denied&action=admin.promote":     5,
		"entity_id=u-1&outcome=denied":                          0,
		"actor_user_id=" + actorID + "&action=user.create":      2,
		"actor_user_id=11111111-1111-4111-8111-111111111111":    0,
		"from=2026-09-26T12:31:00Z&to=2026-09-26T12:33:00Z":     3,
		"from=2026-09-26T12:25:00Z&action=user.deactivate":      0,
		"to=2026-09-26T12:20:00Z&entity_id=u-1&outcome=success": 5,
	}
	for query, want := range cases {
		w := q.get(t, "/api/v1/audit-logs?limit=100&"+query)
		if got := len(decodePage(t, w).Items); w.Code != http.StatusOK || got != want {
			t.Errorf("%s: status = %d, itens = %d, esperado %d", query, w.Code, got, want)
		}
	}
}

// Sem limit, a consulta devolve 50 por página.
func TestAuditQueryServiceDefaultsToFiftyPerPage(t *testing.T) {
	q := newQueryEnv(t, audit.ReadPermission)
	q.seed(t, 60, "user.create", "user", "u-1", "success", 0)
	authorizer, _ := authz.NewAuthorizer([]authz.Definition{{Permission: audit.ReadPermission}}, nil)
	svc := &audit.Query{Authz: authorizer, DB: q.app}

	page, err := svc.Search(context.Background(), principal(audit.ReadPermission), audit.Filter{}, 0, "")

	if err != nil || len(page.Items) != 50 || page.NextCursor == "" {
		t.Errorf("itens = %d, cursor = %q, erro = %v", len(page.Items), page.NextCursor, err)
	}
}

// AUD-03.3: limit acima de 100 é 422 invalid_limit, e cursor que não é cursor também é 422.
func TestAuditQueryRejectsALimitAbove100AndABadCursor(t *testing.T) {
	q := newQueryEnv(t, audit.ReadPermission)

	over := q.get(t, "/api/v1/audit-logs?limit=101")
	bad := q.get(t, "/api/v1/audit-logs?cursor=isto-nao-e-um-cursor")

	if over.Code != http.StatusUnprocessableEntity || problemCode(t, over) != "invalid_limit" {
		t.Errorf("limit=101: %d %s", over.Code, over.Body.String())
	}
	if bad.Code != http.StatusUnprocessableEntity || problemCode(t, bad) != "invalid_cursor" {
		t.Errorf("cursor inválido: %d %s", bad.Code, bad.Body.String())
	}
}

// O serviço também recusa o limite fora da faixa, sem depender do contrato.
func TestAuditQueryServiceRejectsALimitOutOfRange(t *testing.T) {
	q := newQueryEnv(t, audit.ReadPermission)
	authorizer, _ := authz.NewAuthorizer([]authz.Definition{{Permission: audit.ReadPermission}}, nil)
	svc := &audit.Query{Authz: authorizer, DB: q.app}

	for _, limit := range []int{101, -1} {
		if _, err := svc.Search(context.Background(), principal(audit.ReadPermission), audit.Filter{}, limit, ""); err != audit.ErrInvalidLimit {
			t.Errorf("limit %d: err = %v", limit, err)
		}
	}
	if _, err := (&audit.Query{}).Search(context.Background(), principal(audit.ReadPermission), audit.Filter{}, 0, ""); err == nil {
		t.Error("uma consulta mal montada deveria falhar")
	}
}

// AUD-03.4: sem audit:log:read a resposta é 403 forbidden, e a negativa é auditada.
func TestAuditQueryWithoutThePermissionIs403AndTheDenialIsAudited(t *testing.T) {
	q := newQueryEnv(t)
	q.seed(t, 3, "user.create", "user", "u-1", "success", 0)

	w := q.get(t, "/api/v1/audit-logs")

	if w.Code != http.StatusForbidden || problemCode(t, w) != "forbidden" {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	var denied int64
	q.owner.Raw("SELECT count(*) FROM audit_log WHERE action = 'authz.denied'").Scan(&denied)
	if denied != 1 {
		t.Errorf("authz.denied = %d, esperado 1", denied)
	}
}

// A consulta devolve os campos do registro sem inventar nem perder nada.
func TestAuditQueryReturnsTheStoredFields(t *testing.T) {
	q := newQueryEnv(t, audit.ReadPermission)
	if err := q.owner.Exec(`INSERT INTO audit_log (actor_type, actor_user_id, action, entity_type, entity_id, before, after, outcome, reason, context, request_id)
		VALUES ('user', ?::uuid, 'admin.promote', 'user', 'u-9', '{"roles":["ASSOCIADO"]}', '{"roles":["ASSOCIADO","TESOURARIA"]}', 'success', 'motivo da promoção', '{"k":"v"}', 'req-12345678')`, actorID).Error; err != nil {
		t.Fatal(err)
	}

	var body struct {
		Items []map[string]any `json:"items"`
	}
	w := q.get(t, "/api/v1/audit-logs")
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	if len(body.Items) != 1 {
		t.Fatalf("itens = %d: %s", len(body.Items), w.Body.String())
	}
	it := body.Items[0]
	if it["actor_user_id"] != actorID || it["reason"] != "motivo da promoção" || it["request_id"] != "req-12345678" || it["actor_type"] != "user" {
		t.Errorf("registro = %v", it)
	}
	if it["before"].(map[string]any)["roles"].([]any)[0] != "ASSOCIADO" || len(it["after"].(map[string]any)["roles"].([]any)) != 2 || it["context"].(map[string]any)["k"] != "v" {
		t.Errorf("antes, depois e contexto = %v", it)
	}
}

// Edge case: banco indisponível durante a consulta responde 503
// service_unavailable, não 500, e sem vazar o driver ou a string de conexão.
func TestAuditQueryAnswersServiceUnavailableWhenTheDatabaseIsDown(t *testing.T) {
	down := testutil.NewTestDB(t)
	sqlDB, err := down.DB()
	if err != nil {
		t.Fatal(err)
	}
	_ = sqlDB.Close() // simula o banco caindo no meio da requisição

	authorizer, err := authz.NewAuthorizer([]authz.Definition{{Permission: audit.ReadPermission}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := platformapi.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	validate, err := httpx.NewContractValidator(spec)
	if err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(httpx.RequestID())
	r.Use(httpx.Authn(httpx.AuthnConfig{Validator: staticSession{principal(audit.ReadPermission)}}))
	wrapper := platformapi.ServerInterfaceWrapper{
		Handler: struct {
			platformapi.HealthHandler
			audit.Handler
		}{Handler: audit.Handler{Query: &audit.Query{Authz: authorizer, DB: down}}},
	}
	r.GET("/api/v1/audit-logs", validate, wrapper.GetAuditLogs)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/audit-logs", nil)
	req.AddCookie(&http.Cookie{Name: "tj_session", Value: "ok"})
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	var p struct{ Code string }
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	if w.Code != http.StatusServiceUnavailable || p.Code != "service_unavailable" {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	for _, leak := range []string{"tj_app_dev", "connection", "sql:", "gorm", "pgconn"} {
		if strings.Contains(strings.ToLower(w.Body.String()), leak) {
			t.Errorf("a resposta não pode vazar detalhe do driver ou da conexão (%q): %s", leak, w.Body.String())
		}
	}
	testutil.LoadContracts(t).ValidateResponse(t, req, w)
}
