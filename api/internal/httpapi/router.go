package httpapi

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	platformapi "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/api"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

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
	platformapi.RegisterHandlersWithOptions(r, platformapi.HealthHandler{Ping: ping}, platformapi.GinServerOptions{
		Middlewares: []platformapi.MiddlewareFunc{platformapi.MiddlewareFunc(validate)},
	})
	return r
}
