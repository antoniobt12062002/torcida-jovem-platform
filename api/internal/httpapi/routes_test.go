package httpapi

import (
	"context"
	"path/filepath"
	"slices"
	"sort"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// contractKeys lists "METHOD /path" for every operation of every module contract
// found in api/openapi, read from the files (independent of the embedded specs).
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
		keys = append(keys, operationKeys(doc)...)
	}
	sort.Strings(keys)
	return keys
}

// API-01.6: toda rota registrada no Gin existe em algum contrato de módulo e toda
// operação de contrato tem rota. Não há mais lista de pendências.
func TestEveryRegisteredRouteHasAContractOperationAndViceVersa(t *testing.T) {
	router := NewRouter(testDeps(nil))

	withoutContract, withoutRoute := parity(routeKeys(router.Routes()), contractKeys(t))

	if len(withoutContract) > 0 {
		t.Errorf("rotas registradas sem operação em nenhum contrato: %v", withoutContract)
	}
	if len(withoutRoute) > 0 {
		t.Errorf("operações de contrato sem rota registrada: %v", withoutRoute)
	}
	if len(routeKeys(router.Routes())) != 16 {
		t.Errorf("rotas = %d, esperado 16 (healthz, auditoria e 14 de identidade)", len(routeKeys(router.Routes())))
	}
}

// O contrato lido dos arquivos e as rotas do router são exatamente a mesma lista.
func TestRoutesAndContractFilesAreTheSameList(t *testing.T) {
	router := NewRouter(testDeps(nil))

	if !slices.Equal(routeKeys(router.Routes()), contractKeys(t)) {
		t.Errorf("rotas = %v; contrato = %v", routeKeys(router.Routes()), contractKeys(t))
	}
}

// Decisão 5: uma rota registrada sem operação no contrato é recusada.
func TestNewRouterRefusesARouteWithoutAContractOperation(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("um roteador com rota fora do contrato deveria falhar na montagem")
		}
	}()
	d := testDeps(nil)
	d.extraRoutes = func(r gin.IRoutes) { r.GET("/api/v1/fora-do-contrato", func(*gin.Context) {}) }

	NewRouter(d)
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
