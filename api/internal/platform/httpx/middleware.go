package httpx

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

const DefaultBodyLimit int64 = 1 << 20

const userIDKey = "user_id"

func SetUserID(c *gin.Context, id string) {
	c.Set(userIDKey, id)
}

// Recover turns a panic into a 500 problem response. Place it inside AccessLog
// so the recovered request is logged with status 500.
func Recover(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			logger.Error("panic recovered",
				"request_id", RequestIDFrom(c.Request.Context()),
				"panic", fmt.Sprint(rec),
				"stack", string(debug.Stack()))
			WriteProblem(c, http.StatusInternalServerError, "internal_error", "Erro interno. Informe o request_id ao suporte.")
		}()
		c.Next()
	}
}

func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", RequestIDFrom(c.Request.Context()),
		}
		if uid := c.GetString(userIDKey); uid != "" {
			attrs = append(attrs, "user_id", uid)
		}
		logger.Info("request", attrs...)
	}
}

func BodyLimit(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			tooLarge(c)
			return
		}
		if c.Request.ContentLength == -1 && c.Request.Body != nil && c.Request.Body != http.NoBody {
			buf, err := io.ReadAll(io.LimitReader(c.Request.Body, limit+1))
			if err != nil {
				WriteProblem(c, http.StatusBadRequest, "invalid_body", "Não foi possível ler o corpo da requisição.")
				return
			}
			if int64(len(buf)) > limit {
				tooLarge(c)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(buf))
		}
		c.Next()
	}
}

func tooLarge(c *gin.Context) {
	WriteProblem(c, http.StatusRequestEntityTooLarge, "payload_too_large", "O corpo da requisição excede o limite permitido.")
}
