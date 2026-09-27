package platformapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler implements the platform operations of ServerInterface.
type HealthHandler struct {
	Ping func(ctx context.Context) error
}

func (h HealthHandler) GetHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := h.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, Health{Status: Degraded})
		return
	}
	c.JSON(http.StatusOK, Health{Status: Ok})
}
