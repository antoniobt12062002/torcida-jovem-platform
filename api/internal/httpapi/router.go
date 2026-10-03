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

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
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
)

type PingFunc func(ctx context.Context) error

// Deps are what the router composes.
type Deps struct {
	Ping     PingFunc
	Identity *identity.Module
	// Financeiro is composed by cmd/api/main.go (06-api-http) but not yet
	// registered here — that wiring (financeirohttp.Register, the upload
	// chain, FIN-D-024) is 06-api-http/T8's own task.
	Financeiro     *financeiro.Module
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
	if d.Ping == nil || d.Identity == nil || d.AuditQuery == nil || d.Log == nil {
		panic("httpapi: NewRouter exige Ping, Identity, AuditQuery e Log")
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
	validate, err := httpx.NewContractValidator(platformSpec, identitySpec)
	if err != nil {
		panic(fmt.Sprintf("validador do contrato: %v", err))
	}

	handler := identityhttp.New(d.Identity, d.Cookie, d.Log)
	origin := httpx.Origin(httpx.OriginConfig{Allowed: d.AllowedOrigins, Cookie: d.Cookie})
	authn := httpx.Authn(httpx.AuthnConfig{
		Validator: handler.SessionValidator(), Cookie: d.Cookie,
		AllowedWhilePasswordChangeIsRequired: identityhttp.AllowedWhilePasswordChangeIsRequired,
		OnAuthenticated:                      identityhttp.WithActor,
	})
	public := []gin.HandlerFunc{origin, httpx.BodyLimit(PublicBodyLimit), validate}
	authenticated := []gin.HandlerFunc{httpx.BodyLimit(AuthenticatedBodyLimit), origin, authn, httpx.CSRF(), validate}

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
	if d.extraRoutes != nil {
		d.extraRoutes(r)
	}

	withoutContract, withoutRoute := parity(routeKeys(r.Routes()), append(operationKeys(platformSpec), operationKeys(identitySpec)...))
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
