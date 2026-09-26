package httpapi

import (
	"context"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

var ginParam = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

// routeKeys turns the routes registered in Gin into "METHOD /path/{param}".
func routeKeys(routes []gin.RouteInfo) []string {
	keys := make([]string, 0, len(routes))
	for _, r := range routes {
		keys = append(keys, r.Method+" "+ginParam.ReplaceAllString(r.Path, "{$1}"))
	}
	sort.Strings(keys)
	return keys
}

// contractKeys lists "METHOD /path" for every operation of every module contract.
func contractKeys(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("../../openapi/*.yaml")
	if err != nil || len(files) == 0 {
		t.Fatalf("contratos não encontrados: %v", err)
	}
	var keys []string
	for _, f := range files {
		if filepath.Base(f) == "common.yaml" {
			continue
		}
		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true
		doc, err := loader.LoadFromFile(f)
		if err != nil {
			t.Fatalf("carregar %s: %v", f, err)
		}
		if err := doc.Validate(context.Background()); err != nil {
			t.Fatalf("contrato inválido %s: %v", f, err)
		}
		for path, item := range doc.Paths.Map() {
			for method := range item.Operations() {
				keys = append(keys, method+" "+path)
			}
		}
	}
	sort.Strings(keys)
	return keys
}

// parity returns the routes with no operation in any contract and the
// contract operations with no route.
func parity(routes, operations []string) (routesWithoutContract, operationsWithoutRoute []string) {
	for _, r := range routes {
		if !slices.Contains(operations, r) {
			routesWithoutContract = append(routesWithoutContract, r)
		}
	}
	for _, o := range operations {
		if !slices.Contains(routes, o) {
			operationsWithoutRoute = append(operationsWithoutRoute, o)
		}
	}
	return routesWithoutContract, operationsWithoutRoute
}

// contractPendingRoutes (contract_pending_routes) lists contract operations
// whose handlers do not exist yet. It is temporary: phase 10 creates the identity
// contract before the handlers of phase 11, which must empty this list.
var contractPendingRoutes = []string{
	"GET /api/v1/auth/me", "GET /api/v1/users",
	"POST /api/v1/auth/login", "POST /api/v1/auth/logout", "POST /api/v1/auth/password",
	"POST /api/v1/auth/password-reset/confirm", "POST /api/v1/auth/password-reset/request",
	"POST /api/v1/users", "POST /api/v1/users/{id}/admin-membership", "POST /api/v1/users/{id}/admin-membership/revoke",
	"POST /api/v1/users/{id}/deactivate", "POST /api/v1/users/{id}/password-reset", "POST /api/v1/users/{id}/reactivate",
	"PUT /api/v1/users/{id}/roles",
}

// API-01.6: toda rota registrada no Gin existe em algum contrato de módulo, e
// toda operação de contrato tem rota, salvo as pendentes explicitamente listadas.
func TestEveryRegisteredRouteHasAContractOperation(t *testing.T) {
	router := NewRouter(func(context.Context) error { return nil })

	withoutContract, withoutRoute := parity(routeKeys(router.Routes()), contractKeys(t))
	withoutRoute = slices.DeleteFunc(withoutRoute, func(k string) bool { return slices.Contains(contractPendingRoutes, k) })

	if len(withoutContract) > 0 {
		t.Errorf("rotas registradas sem operação em nenhum contrato: %v", withoutContract)
	}
	if len(withoutRoute) > 0 {
		t.Errorf("operações de contrato sem rota registrada: %v", withoutRoute)
	}
}

func TestParityDetectsARouteWithoutOperation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/healthz", func(*gin.Context) {})
	r.POST("/api/v1/users/:id/roles", func(*gin.Context) {})

	withoutContract, withoutRoute := parity(routeKeys(r.Routes()), []string{"GET /healthz"})

	if !slices.Equal(withoutContract, []string{"POST /api/v1/users/{id}/roles"}) || len(withoutRoute) != 0 {
		t.Errorf("sem contrato = %v, sem rota = %v", withoutContract, withoutRoute)
	}
}

func TestParityDetectsAnOperationWithoutRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/healthz", func(*gin.Context) {})

	withoutContract, withoutRoute := parity(routeKeys(r.Routes()), []string{"GET /healthz", "GET /api/v1/users"})

	if len(withoutContract) != 0 || !slices.Equal(withoutRoute, []string{"GET /api/v1/users"}) {
		t.Errorf("sem contrato = %v, sem rota = %v", withoutContract, withoutRoute)
	}
}

func TestRouteKeysConvertGinParametersToOpenAPIStyle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/api/v1/users/:id/sessions/:session_id", func(*gin.Context) {})

	got := routeKeys(r.Routes())
	if !slices.Equal(got, []string{"DELETE /api/v1/users/{id}/sessions/{session_id}"}) {
		t.Errorf("routeKeys = %v", got)
	}
}

// A lista de pendências só pode guardar operações reais que ainda não têm rota:
// registrar o handler obriga a tirar a operação da lista.
func TestPendingRoutesAreRealContractOperationsWithoutARoute(t *testing.T) {
	router := NewRouter(func(context.Context) error { return nil })
	contract, registered := contractKeys(t), routeKeys(router.Routes())

	for _, pending := range contractPendingRoutes {
		if !slices.Contains(contract, pending) {
			t.Errorf("%s está pendente, mas não existe em nenhum contrato", pending)
		}
		if slices.Contains(registered, pending) {
			t.Errorf("%s já tem rota: remova-a de contractPendingRoutes", pending)
		}
	}
	if !slices.IsSorted(contractPendingRoutes) {
		t.Error("mantenha a lista ordenada")
	}
}
