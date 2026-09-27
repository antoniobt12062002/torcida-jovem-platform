package httpapi

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	platformapi "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/api"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

// platformServer is the platform ServerInterface: the health check and the audit
// query. The audit query is not registered yet (see contractPendingRoutes).
type platformServer struct {
	platformapi.HealthHandler
	audit.Handler
}

type PingFunc func(ctx context.Context) error

func NewRouter(ping PingFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(httpx.RequestID(), gin.Recovery(), gin.Logger())

	// Each module validates its own requests against its own contract.
	spec, err := platformapi.GetSpec()
	if err != nil {
		panic(fmt.Sprintf("contrato platform embutido inválido: %v", err))
	}
	validate, err := httpx.NewContractValidator(spec)
	if err != nil {
		panic(fmt.Sprintf("validador do contrato platform: %v", err))
	}
	// Only /healthz is registered for now: the audit query is in the contract but
	// its handler comes in phase 11 (see contractPendingRoutes in routes_test.go).
	platform := platformapi.ServerInterfaceWrapper{
		Handler:            platformServer{HealthHandler: platformapi.HealthHandler{Ping: ping}},
		HandlerMiddlewares: []platformapi.MiddlewareFunc{platformapi.MiddlewareFunc(validate)},
		ErrorHandler: func(c *gin.Context, err error, status int) {
			httpx.WriteProblem(c, status, "validation_failed", "Parâmetros inválidos.")
		},
	}
	r.GET("/healthz", platform.GetHealth)
	return r
}
