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
	"STORAGE_ENABLED", "S3_ENDPOINT", "S3_REGION", "S3_BUCKET", "S3_ACCESS_KEY", "S3_SECRET_KEY",
	"S3_USE_PATH_STYLE", "DOCUMENT_URL_TTL_SECONDS",
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
	if cfg.StorageEnabled {
		t.Error("StorageEnabled deveria ser false por padrão: nada consome storage ainda")
	}
	if cfg.DocumentURLTTL != 300*time.Second {
		t.Errorf("DocumentURLTTL = %v, esperado 300s", cfg.DocumentURLTTL)
	}
	if cfg.S3UsePathStyle {
		t.Error("S3UsePathStyle deveria ser false por padrão")
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
		{"storage enabled inválido", map[string]string{"STORAGE_ENABLED": "talvez"}, "STORAGE_ENABLED"},
		{"storage habilitado sem nenhuma variável S3", map[string]string{"STORAGE_ENABLED": "true"}, "S3_ENDPOINT"},
		{"storage habilitado sem região", map[string]string{
			"STORAGE_ENABLED": "true", "S3_ENDPOINT": "http://localhost:9000", "S3_BUCKET": "tj-documentos",
			"S3_ACCESS_KEY": "x", "S3_SECRET_KEY": "y",
		}, "S3_REGION"},
		{"storage habilitado sem bucket", map[string]string{
			"STORAGE_ENABLED": "true", "S3_ENDPOINT": "http://localhost:9000", "S3_REGION": "us-east-1",
			"S3_ACCESS_KEY": "x", "S3_SECRET_KEY": "y",
		}, "S3_BUCKET"},
		{"storage habilitado sem chave de acesso", map[string]string{
			"STORAGE_ENABLED": "true", "S3_ENDPOINT": "http://localhost:9000", "S3_REGION": "us-east-1",
			"S3_BUCKET": "tj-documentos", "S3_SECRET_KEY": "y",
		}, "S3_ACCESS_KEY"},
		{"storage habilitado sem chave secreta", map[string]string{
			"STORAGE_ENABLED": "true", "S3_ENDPOINT": "http://localhost:9000", "S3_REGION": "us-east-1",
			"S3_BUCKET": "tj-documentos", "S3_ACCESS_KEY": "x",
		}, "S3_SECRET_KEY"},
		{"path style inválido", map[string]string{"S3_USE_PATH_STYLE": "talvez"}, "S3_USE_PATH_STYLE"},
		{"validade da URL zero", map[string]string{"DOCUMENT_URL_TTL_SECONDS": "0"}, "DOCUMENT_URL_TTL_SECONDS"},
		{"validade da URL negativa", map[string]string{"DOCUMENT_URL_TTL_SECONDS": "-1"}, "DOCUMENT_URL_TTL_SECONDS"},
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

// DOC-03: com STORAGE_ENABLED=true e as cinco variáveis, o storage carrega, e
// nenhum erro de outra variável inválida ecoa a chave secreta de volta.
func TestLoadStorageEnabledReadsAllVariablesAndNeverLeaksTheSecretKey(t *testing.T) {
	setEnv(t, map[string]string{
		"DATABASE_URL": "postgres://x", "STORAGE_ENABLED": "true",
		"S3_ENDPOINT": "http://localhost:9000", "S3_REGION": "us-east-1", "S3_BUCKET": "tj-documentos",
		"S3_ACCESS_KEY": "AKIAEXEMPLO", "S3_SECRET_KEY": "segredo-s3-nao-pode-vazar",
		"S3_USE_PATH_STYLE": "true", "DOCUMENT_URL_TTL_SECONDS": "120",
		"PORT": "abc", // erro proposital de outra variável, para checar o vazamento
	})

	_, err := Load()

	if err == nil || !strings.Contains(err.Error(), "PORT") {
		t.Fatalf("esperava erro nomeando PORT, veio %v", err)
	}
	if strings.Contains(err.Error(), "segredo-s3-nao-pode-vazar") {
		t.Errorf("erro vazou a chave secreta do S3: %v", err)
	}

	setEnv(t, map[string]string{
		"DATABASE_URL": "postgres://x", "STORAGE_ENABLED": "true",
		"S3_ENDPOINT": "http://localhost:9000", "S3_REGION": "us-east-1", "S3_BUCKET": "tj-documentos",
		"S3_ACCESS_KEY": "AKIAEXEMPLO", "S3_SECRET_KEY": "segredo-s3-nao-pode-vazar",
		"S3_USE_PATH_STYLE": "true", "DOCUMENT_URL_TTL_SECONDS": "120",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !cfg.StorageEnabled || cfg.S3Endpoint != "http://localhost:9000" || cfg.S3Region != "us-east-1" ||
		cfg.S3Bucket != "tj-documentos" || cfg.S3AccessKey != "AKIAEXEMPLO" || cfg.S3SecretKey != "segredo-s3-nao-pode-vazar" {
		t.Errorf("configuração de storage = %+v", cfg)
	}
	if !cfg.S3UsePathStyle {
		t.Error("S3UsePathStyle deveria ser true")
	}
	if cfg.DocumentURLTTL != 120*time.Second {
		t.Errorf("DocumentURLTTL = %v, esperado 120s", cfg.DocumentURLTTL)
	}
}

// Com storage desabilitado, as variáveis S3 continuam opcionais.
func TestLoadStorageDisabledDoesNotRequireS3Variables(t *testing.T) {
	setEnv(t, map[string]string{"DATABASE_URL": "postgres://x"})

	cfg, err := Load()

	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if cfg.S3Endpoint != "" || cfg.S3Region != "" || cfg.S3Bucket != "" || cfg.S3AccessKey != "" || cfg.S3SecretKey != "" {
		t.Errorf("nenhum campo S3 deveria ser lido com storage desabilitado: %+v", cfg)
	}
}
