package httpapi

import (
	"context"

	"github.com/gin-gonic/gin"

	platformapi "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/api"
)

type PingFunc func(ctx context.Context) error

func NewRouter(ping PingFunc) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	platformapi.RegisterHandlers(r, platformapi.HealthHandler{Ping: ping})
	return r
}
