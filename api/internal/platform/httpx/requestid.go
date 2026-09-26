package httpx

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-Id"

const requestIDKey = "request_id"

type ctxKey struct{}

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if !validRequestID.MatchString(id) {
			id = newUUIDv4()
		}
		c.Set(requestIDKey, id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKey{}, id))
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

// WithRequestID returns a context that carries the request id, for code that
// runs outside a request (tests, background work started by a request).
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

func newUUIDv4() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
