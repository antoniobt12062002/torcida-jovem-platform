package logx

import (
	"io"
	"log/slog"
	"strings"
)

var sensitiveKeys = map[string]struct{}{
	"password":      {},
	"password_hash": {},
	"token":         {},
	"session_token": {},
	"csrf_token":    {},
	"cookie":        {},
	"authorization": {},
}

func New(level string, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: parseLevel(level),
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if _, ok := sensitiveKeys[strings.ToLower(a.Key)]; ok {
				return slog.String(a.Key, "[redacted]")
			}
			return a
		},
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
