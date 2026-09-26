package config

import (
	"strings"
	"testing"
	"time"
)

var allVars = []string{
	"APP_ENV", "LOG_LEVEL", "PORT", "DATABASE_URL", "ALLOWED_ORIGINS",
	"COOKIE_SECURE", "COOKIE_DOMAIN", "SESSION_IDLE_MINUTES", "SESSION_ABSOLUTE_HOURS",
	"AUTH_HASH_KEY", "ARGON2_MEMORY_KIB", "ARGON2_ITERATIONS", "ARGON2_PARALLELISM",
	"PASSWORD_RESET_TTL_MINUTES", "EMAIL_PROVIDER", "EMAIL_FROM", "APP_BASE_URL",
}

// testHashKey satisfies the 32-byte minimum; setEnv uses it unless a test sets
// AUTH_HASH_KEY itself.
const testHashKey = "0123456789abcdef0123456789abcdef"

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for _, k := range allVars {
		t.Setenv(k, "")
	}
	if _, ok := kv["AUTH_HASH_KEY"]; !ok {
		t.Setenv("AUTH_HASH_KEY", testHashKey)
	}
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func TestLoadDefaults(t *testing.T) {
	setEnv(t, map[string]string{"DATABASE_URL": "postgres://x"})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.AppEnv != "production" {
		t.Errorf("AppEnv = %q, esperado production", cfg.AppEnv)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, esperado info", cfg.LogLevel)
	}
	if !cfg.CookieSecure {
		t.Error("CookieSecure deveria ser true por padrão")
	}
	if cfg.SessionIdle != 60*time.Minute {
		t.Errorf("SessionIdle = %v, esperado 60m", cfg.SessionIdle)
	}
	if cfg.SessionAbsolute != 8*time.Hour {
		t.Errorf("SessionAbsolute = %v, esperado 8h", cfg.SessionAbsolute)
	}
	if len(cfg.AllowedOrigins) != 0 {
		t.Errorf("AllowedOrigins = %v, esperado vazio", cfg.AllowedOrigins)
	}
}

func TestLoadCustomValues(t *testing.T) {
	setEnv(t, map[string]string{
		"DATABASE_URL":           "postgres://x",
		"APP_ENV":                "staging",
		"LOG_LEVEL":              "debug",
		"COOKIE_DOMAIN":          ".tj.example",
		"SESSION_IDLE_MINUTES":   "30",
		"SESSION_ABSOLUTE_HOURS": "4",
		"ALLOWED_ORIGINS":        "https://app.tj.example, https://admin.tj.example",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.AppEnv != "staging" || cfg.LogLevel != "debug" || cfg.CookieDomain != ".tj.example" {
		t.Errorf("valores não lidos: %+v", cfg)
	}
	if cfg.SessionIdle != 30*time.Minute || cfg.SessionAbsolute != 4*time.Hour {
		t.Errorf("sessão = %v/%v", cfg.SessionIdle, cfg.SessionAbsolute)
	}
	want := []string{"https://app.tj.example", "https://admin.tj.example"}
	if strings.Join(cfg.AllowedOrigins, "|") != strings.Join(want, "|") {
		t.Errorf("AllowedOrigins = %v, esperado %v", cfg.AllowedOrigins, want)
	}
}

func TestLoadDevelopmentDefaults(t *testing.T) {
	setEnv(t, map[string]string{"DATABASE_URL": "postgres://x", "APP_ENV": "development", "COOKIE_SECURE": "false"})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.CookieSecure {
		t.Error("CookieSecure deveria ser false em development quando configurado")
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("AllowedOrigins = %v, esperado [http://localhost:3000]", cfg.AllowedOrigins)
	}
}

func TestLoadInsecureCookieOutsideDevelopment(t *testing.T) {
	for _, env := range []string{"production", "staging", ""} {
		setEnv(t, map[string]string{"DATABASE_URL": "postgres://x", "APP_ENV": env, "COOKIE_SECURE": "false"})
		_, err := Load()
		if err == nil {
			t.Fatalf("APP_ENV=%q com COOKIE_SECURE=false deveria ser recusado", env)
		}
		if !strings.Contains(err.Error(), "COOKIE_SECURE") {
			t.Errorf("erro deveria nomear COOKIE_SECURE: %v", err)
		}
	}
}

func TestLoadInvalidVariablesNameTheVariable(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"database ausente", map[string]string{}, "DATABASE_URL"},
		{"app env inválido", map[string]string{"APP_ENV": "prod"}, "APP_ENV"},
		{"log level inválido", map[string]string{"LOG_LEVEL": "verbose"}, "LOG_LEVEL"},
		{"porta não numérica", map[string]string{"PORT": "abc"}, "PORT"},
		{"porta fora da faixa", map[string]string{"PORT": "70000"}, "PORT"},
		{"cookie secure inválido", map[string]string{"COOKIE_SECURE": "talvez"}, "COOKIE_SECURE"},
		{"ociosidade zero", map[string]string{"SESSION_IDLE_MINUTES": "0"}, "SESSION_IDLE_MINUTES"},
		{"ociosidade não numérica", map[string]string{"SESSION_IDLE_MINUTES": "abc"}, "SESSION_IDLE_MINUTES"},
		{"teto negativo", map[string]string{"SESSION_ABSOLUTE_HOURS": "-1"}, "SESSION_ABSOLUTE_HOURS"},
		{"origem sem esquema", map[string]string{"ALLOWED_ORIGINS": "app.tj.example"}, "ALLOWED_ORIGINS"},
		{"origem com caminho", map[string]string{"ALLOWED_ORIGINS": "https://app.tj.example/painel"}, "ALLOWED_ORIGINS"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := map[string]string{"DATABASE_URL": "postgres://x"}
			for k, v := range tc.env {
				env[k] = v
			}
			if tc.name == "database ausente" {
				delete(env, "DATABASE_URL")
			}
			setEnv(t, env)
			_, err := Load()
			if err == nil {
				t.Fatal("esperava erro")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("erro %q deveria nomear %s", err, tc.want)
			}
		})
	}
}

func TestLoadErrorsNeverContainSecrets(t *testing.T) {
	setEnv(t, map[string]string{
		"DATABASE_URL": "postgres://tj:s3cr3t-password@db:5432/tj",
		"PORT":         "abc",
	})
	_, err := Load()
	if err == nil {
		t.Fatal("esperava erro")
	}
	if strings.Contains(err.Error(), "s3cr3t-password") {
		t.Errorf("erro vazou o segredo: %v", err)
	}
}
