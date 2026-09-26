package logx

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func lines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for l := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("linha não é JSON: %q: %v", l, err)
		}
		out = append(out, m)
	}
	return out
}

func TestSensitiveKeysAreRedacted(t *testing.T) {
	keys := []string{"password", "password_hash", "token", "session_token", "csrf_token", "cookie", "authorization"}
	for _, key := range keys {
		var buf bytes.Buffer
		New("info", &buf).Info("evento", key, "valor-secreto-123", "seguro", "visivel")
		got := lines(t, &buf)
		if len(got) != 1 {
			t.Fatalf("%s: esperava 1 linha, veio %d", key, len(got))
		}
		if got[0][key] != "[redacted]" {
			t.Errorf("%s = %v, esperado [redacted]", key, got[0][key])
		}
		if got[0]["seguro"] != "visivel" {
			t.Errorf("campo não sensível foi alterado: %v", got[0]["seguro"])
		}
		if strings.Contains(buf.String(), "valor-secreto-123") {
			t.Errorf("%s: o valor secreto apareceu no log", key)
		}
	}
}

func TestRedactionIgnoresKeyCase(t *testing.T) {
	var buf bytes.Buffer
	New("info", &buf).Info("evento", "Authorization", "Bearer abc", "Cookie", "tj_session=abc")
	got := lines(t, &buf)[0]
	if got["Authorization"] != "[redacted]" || got["Cookie"] != "[redacted]" {
		t.Errorf("chaves capitalizadas não foram redigidas: %v", got)
	}
}

func TestRedactionInNestedGroups(t *testing.T) {
	var buf bytes.Buffer
	New("info", &buf).Info("evento",
		slog.Group("req",
			slog.String("token", "abc-segredo"),
			slog.String("path", "/x"),
			slog.Group("headers", slog.String("authorization", "Bearer xyz-segredo")),
		))
	if strings.Contains(buf.String(), "abc-segredo") || strings.Contains(buf.String(), "xyz-segredo") {
		t.Fatalf("valor sensível vazou em grupo aninhado: %s", buf.String())
	}
	req := lines(t, &buf)[0]["req"].(map[string]any)
	if req["token"] != "[redacted]" {
		t.Errorf("req.token = %v, esperado [redacted]", req["token"])
	}
	if req["path"] != "/x" {
		t.Errorf("req.path = %v, esperado /x", req["path"])
	}
	headers := req["headers"].(map[string]any)
	if headers["authorization"] != "[redacted]" {
		t.Errorf("req.headers.authorization = %v, esperado [redacted]", headers["authorization"])
	}
}

func TestLogFormatIsStructuredJSON(t *testing.T) {
	var buf bytes.Buffer
	New("info", &buf).Info("olá", "chave", "valor")
	got := lines(t, &buf)[0]
	for _, k := range []string{"time", "level", "msg"} {
		if _, ok := got[k]; !ok {
			t.Errorf("campo %q ausente: %v", k, got)
		}
	}
	if got["msg"] != "olá" || got["chave"] != "valor" || got["level"] != "INFO" {
		t.Errorf("conteúdo inesperado: %v", got)
	}
}

func TestLevelIsConfigurable(t *testing.T) {
	cases := []struct {
		level     string
		debugSeen bool
		infoSeen  bool
		warnSeen  bool
	}{
		{"debug", true, true, true},
		{"info", false, true, true},
		{"warn", false, false, true},
		{"error", false, false, false},
		{"desconhecido", false, true, true},
	}
	for _, tc := range cases {
		var buf bytes.Buffer
		l := New(tc.level, &buf)
		l.Debug("d")
		l.Info("i")
		l.Warn("w")
		out := buf.String()
		if strings.Contains(out, `"msg":"d"`) != tc.debugSeen ||
			strings.Contains(out, `"msg":"i"`) != tc.infoSeen ||
			strings.Contains(out, `"msg":"w"`) != tc.warnSeen {
			t.Errorf("nível %q: saída inesperada:\n%s", tc.level, out)
		}
	}
}
