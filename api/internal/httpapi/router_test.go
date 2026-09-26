package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	cases := []struct {
		name string
		ping PingFunc
		code int
		body string
	}{
		{"banco ok", func(context.Context) error { return nil }, http.StatusOK, `{"status":"ok"}`},
		{"banco fora", func(context.Context) error { return errors.New("down") }, http.StatusServiceUnavailable, `{"status":"degraded"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			NewRouter(tc.ping).ServeHTTP(w, req)
			if w.Code != tc.code || w.Body.String() != tc.body {
				t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}
