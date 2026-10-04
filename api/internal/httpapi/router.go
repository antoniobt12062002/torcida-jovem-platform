package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque"
	estoquehttp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/http"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	financeirohttp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/http"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	identityhttp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/http"
	platformapi "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/api"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

const (
	// AuthenticatedBodyLimit and PublicBodyLimit are the request body limits
	// (API-02.7): public routes take small bodies and are the ones anyone can reach.
	AuthenticatedBodyLimit = 1 << 20
	PublicBodyLimit        = 64 << 10
	// FinanceiroUploadBodyLimit is the body limit for financeiro's one
	// multipart route (anexar comprovante): larger than
	// AuthenticatedBodyLimit to accommodate platform/documents's own 10 MiB
	// file limit plus multipart/protocol overhead (FIN-D-024, 06-api-http
	// T7). 06-api-http/T8 builds the dedicated chain with this constant, the
	// same way it builds `authenticated` below — the chain itself cannot
	// exist here yet because financeirohttp.Register, the only place that
	// would use it, requires *financeiro/http.Handler to satisfy the full
	// StrictServerInterface, only true once T4-T7 are all done (FIN-D-026).
	FinanceiroUploadBodyLimit = 11 << 20
)

type PingFunc func(ctx context.Context) error

// Deps are what the router composes.
type Deps struct {
	Ping           PingFunc
	Identity       *identity.Module
	Financeiro     *financeiro.Module
	Estoque        *estoque.Module
	AuditQuery     *audit.Query
	Log            *slog.Logger
	AllowedOrigins []string
	Cookie         httpx.SessionCookie

	// extraRoutes lets a test register a route the contract does not describe.
	extraRoutes func(gin.IRoutes)
}

// platformServer is the platform ServerInterface: the health check and the audit query.
type platformServer struct {
	platformapi.HealthHandler
	audit.Handler
}

// NewRouter composes the HTTP API. It panics when the wiring is incomplete or
// when the routes and the contracts disagree: the OpenAPI contract is the source
// of truth, so a route with no operation, or an operation with no route, keeps
// the API from starting.
//
// Chains: the public routes (login and the two recovery routes) go through
// origin, body limit (64 KiB) and contract validation, with no session and no
// CSRF. Every other route goes through body limit (1 MiB), origin, session
// (deny by default), CSRF and contract validation, in that order, so a caller
// who is not signed in learns nothing of the structure of the API.
func NewRouter(d Deps) *gin.Engine {
	if d.Ping == nil || d.Identity == nil || d.Financeiro == nil || d.Estoque == nil || d.AuditQuery == nil || d.Log == nil {
		panic("httpapi: NewRouter exige Ping, Identity, Financeiro, Estoque, AuditQuery e Log")
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(httpx.RequestID(), httpx.Recover(d.Log), httpx.AccessLog(d.Log))
	r.NoRoute(func(c *gin.Context) {
		httpx.WriteProblem(c, http.StatusNotFound, "not_found", "Rota não encontrada.")
	})

	platformSpec, err := platformapi.GetSpec()
	if err != nil {
		panic(fmt.Sprintf("contrato platform embutido inválido: %v", err))
	}
	identitySpec, err := identityhttp.GetSpec()
	if err != nil {
		panic(fmt.Sprintf("contrato identity embutido inválido: %v", err))
	}
	financeiroSpec, err := financeirohttp.GetSpec()
	if err != nil {
		panic(fmt.Sprintf("contrato financeiro embutido inválido: %v", err))
	}
	estoqueSpec, err := estoquehttp.GetSpec()
	if err != nil {
		panic(fmt.Sprintf("contrato estoque embutido inválido: %v", err))
	}
	validate, err := httpx.NewContractValidator(platformSpec, identitySpec, financeiroSpec, estoqueSpec)
	if err != nil {
		panic(fmt.Sprintf("validador do contrato: %v", err))
	}

	handler := identityhttp.New(d.Identity, d.Cookie, d.Log)
	finHandler := financeirohttp.New(d.Financeiro, d.Log)
	origin := httpx.Origin(httpx.OriginConfig{Allowed: d.AllowedOrigins, Cookie: d.Cookie})
	authn := httpx.Authn(httpx.AuthnConfig{
		Validator: handler.SessionValidator(), Cookie: d.Cookie,
		AllowedWhilePasswordChangeIsRequired: identityhttp.AllowedWhilePasswordChangeIsRequired,
		OnAuthenticated:                      identityhttp.WithActor,
	})
	public := []gin.HandlerFunc{origin, httpx.BodyLimit(PublicBodyLimit), validate}
	authenticated := []gin.HandlerFunc{httpx.BodyLimit(AuthenticatedBodyLimit), origin, authn, httpx.CSRF(), validate}
	// authenticatedUpload is financeiro's one multipart route (anexar
	// comprovante): same chain as authenticated, but with the larger body
	// limit FIN-D-024/FIN-D-031 already settled on.
	authenticatedUpload := []gin.HandlerFunc{httpx.BodyLimit(FinanceiroUploadBodyLimit), origin, authn, httpx.CSRF(), validate}

	platform := platformapi.ServerInterfaceWrapper{
		Handler: platformServer{
			HealthHandler: platformapi.HealthHandler{Ping: d.Ping},
			Handler:       audit.Handler{Query: d.AuditQuery},
		},
		ErrorHandler: func(c *gin.Context, _ error, _ int) {
			httpx.WriteProblem(c, http.StatusUnprocessableEntity, "validation_failed", "Parâmetros inválidos.")
		},
	}
	r.GET("/healthz", validate, platform.GetHealth)
	r.GET("/api/v1/audit-logs", append(append([]gin.HandlerFunc{}, authenticated...), platform.GetAuditLogs)...)
	identityhttp.Register(r, handler, handler, identityhttp.Chains{Public: public, Authenticated: authenticated})
	finWrapper := financeirohttp.Register(finHandler, finHandler)
	auth := func(op gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, authenticated...), op)
	}
	upload := func(op gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, authenticatedUpload...), op)
	}
	r.GET("/api/v1/financeiro/contas", auth(finWrapper.ListContas)...)
	r.POST("/api/v1/financeiro/contas", auth(finWrapper.CreateConta)...)
	r.PATCH("/api/v1/financeiro/contas/:id", auth(finWrapper.RenameConta)...)
	r.POST("/api/v1/financeiro/contas/:id/deactivate", auth(finWrapper.DeactivateConta)...)
	r.GET("/api/v1/financeiro/lancamentos", auth(finWrapper.ListLancamentos)...)
	r.POST("/api/v1/financeiro/lancamentos", auth(finWrapper.CreateLancamento)...)
	r.POST("/api/v1/financeiro/lancamentos/devolucoes", auth(finWrapper.CreateDevolucao)...)
	r.PUT("/api/v1/financeiro/lancamentos/:id", auth(finWrapper.UpdateLancamento)...)
	r.POST("/api/v1/financeiro/lancamentos/:id/receive", auth(finWrapper.ReceiveLancamento)...)
	r.POST("/api/v1/financeiro/lancamentos/:id/pay", auth(finWrapper.PayLancamento)...)
	r.POST("/api/v1/financeiro/lancamentos/:id/cancel", auth(finWrapper.CancelLancamento)...)
	r.GET("/api/v1/financeiro/saldo", auth(finWrapper.GetSaldo)...)
	r.POST("/api/v1/financeiro/lancamentos/:id/comprovantes", upload(finWrapper.CreateComprovante)...)
	r.GET("/api/v1/financeiro/lancamentos/:id/comprovantes", auth(finWrapper.ListComprovantes)...)
	r.GET("/api/v1/financeiro/comprovantes/:documentId/url", auth(finWrapper.GetComprovanteUrl)...)
	estHandler := estoquehttp.New(d.Estoque, d.Log)
	estoquehttp.Register(r, estHandler, estHandler, estoquehttp.Chains{Authenticated: authenticated})
	if d.extraRoutes != nil {
		d.extraRoutes(r)
	}

	operations := append(append(append(operationKeys(platformSpec), operationKeys(identitySpec)...), operationKeys(financeiroSpec)...), operationKeys(estoqueSpec)...)
	withoutContract, withoutRoute := parity(routeKeys(r.Routes()), operations)
	if len(withoutContract) > 0 || len(withoutRoute) > 0 {
		panic(fmt.Sprintf("rotas e contratos divergem: rotas sem operação %v; operações sem rota %v", withoutContract, withoutRoute))
	}
	return r
}

var ginParam = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

// routeKeys turns the routes registered in Gin into sorted "METHOD /path/{param}".
func routeKeys(routes []gin.RouteInfo) []string {
	keys := make([]string, 0, len(routes))
	for _, r := range routes {
		keys = append(keys, r.Method+" "+ginParam.ReplaceAllString(r.Path, "{$1}"))
	}
	sort.Strings(keys)
	return keys
}

// operationKeys lists "METHOD /path" for every operation of a contract, sorted.
func operationKeys(doc *openapi3.T) []string {
	var keys []string
	for path, item := range doc.Paths.Map() {
		for method := range item.Operations() {
			keys = append(keys, method+" "+path)
		}
	}
	sort.Strings(keys)
	return keys
}

// parity returns the routes with no contract operation and the operations with no route.
func parity(routes, operations []string) (routesWithoutContract, operationsWithoutRoute []string) {
	has := func(list []string, key string) bool {
		i := sort.SearchStrings(list, key)
		return i < len(list) && list[i] == key
	}
	sortedRoutes, sortedOps := append([]string{}, routes...), append([]string{}, operations...)
	sort.Strings(sortedRoutes)
	sort.Strings(sortedOps)
	for _, r := range sortedRoutes {
		if !has(sortedOps, r) {
			routesWithoutContract = append(routesWithoutContract, r)
		}
	}
	for _, o := range sortedOps {
		if !has(sortedRoutes, o) {
			operationsWithoutRoute = append(operationsWithoutRoute, o)
		}
	}
	return routesWithoutContract, operationsWithoutRoute
}
