package config

import (
	"strings"
	"testing"
	"time"
)

func load(t *testing.T, env map[string]string) (Config, error) {
	t.Helper()
	all := map[string]string{"DATABASE_URL": "postgres://x"}
	for k, v := range env {
		all[k] = v
	}
	setEnv(t, all)
	return Load()
}

// IDN-07.2 e PLT-01.1: padrões de autenticação e de e-mail.
func TestAuthAndEmailDefaultsOutsideDevelopment(t *testing.T) {
	cfg, err := load(t, nil)
	if err != nil {
		t.Fatal(err)
	}

	if string(cfg.AuthHashKey) != testHashKey {
		t.Error("AuthHashKey não foi carregada")
	}
	if cfg.Argon2MemoryKiB != 19456 || cfg.Argon2Iterations != 2 || cfg.Argon2Parallelism != 1 {
		t.Errorf("argon2 = %d/%d/%d, esperado 19456/2/1", cfg.Argon2MemoryKiB, cfg.Argon2Iterations, cfg.Argon2Parallelism)
	}
	if cfg.PasswordResetTTL != 30*time.Minute {
		t.Errorf("PasswordResetTTL = %v", cfg.PasswordResetTTL)
	}
	if cfg.EmailProvider != "disabled" || cfg.EmailFrom != "" || cfg.AppBaseURL != "" {
		t.Errorf("e-mail = %q/%q/%q", cfg.EmailProvider, cfg.EmailFrom, cfg.AppBaseURL)
	}
}

func TestDevelopmentDefaultsToTheLogSenderAndLocalBaseURL(t *testing.T) {
	cfg, err := load(t, map[string]string{"APP_ENV": "development", "AUTH_HASH_KEY": ""})
	if err != nil {
		t.Fatal(err)
	}

	if cfg.EmailProvider != "log" || cfg.AppBaseURL != "http://localhost:3000" || cfg.EmailFrom == "" {
		t.Errorf("e-mail = %q/%q/%q", cfg.EmailProvider, cfg.EmailFrom, cfg.AppBaseURL)
	}
	if len(cfg.AuthHashKey) < 32 {
		t.Errorf("em development a chave tem um valor próprio de desenvolvimento, len = %d", len(cfg.AuthHashKey))
	}
}

func TestAuthHashKeyIsRequiredOutsideDevelopmentAndMustHave32Bytes(t *testing.T) {
	for _, env := range []string{"production", "staging"} {
		if _, err := load(t, map[string]string{"APP_ENV": env, "AUTH_HASH_KEY": ""}); err == nil || !strings.Contains(err.Error(), "AUTH_HASH_KEY") {
			t.Errorf("%s sem chave: esperava erro nomeando AUTH_HASH_KEY, veio %v", env, err)
		}
	}
	for _, env := range []string{"development", "production"} {
		short := "curta-demais-mas-secreta"
		_, err := load(t, map[string]string{"APP_ENV": env, "AUTH_HASH_KEY": short})
		if err == nil || !strings.Contains(err.Error(), "AUTH_HASH_KEY") {
			t.Errorf("%s com chave curta: esperava erro, veio %v", env, err)
		} else if strings.Contains(err.Error(), short) {
			t.Errorf("o erro vazou a chave: %v", err)
		}
	}
	if _, err := load(t, map[string]string{"AUTH_HASH_KEY": strings.Repeat("k", 32)}); err != nil {
		t.Errorf("chave de 32 bytes deveria ser aceita: %v", err)
	}
}

func TestArgon2AndResetTTLAreConfigurable(t *testing.T) {
	cfg, err := load(t, map[string]string{
		"ARGON2_MEMORY_KIB": "65536", "ARGON2_ITERATIONS": "3", "ARGON2_PARALLELISM": "4", "PASSWORD_RESET_TTL_MINUTES": "15",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Argon2MemoryKiB != 65536 || cfg.Argon2Iterations != 3 || cfg.Argon2Parallelism != 4 || cfg.PasswordResetTTL != 15*time.Minute {
		t.Errorf("cfg = %+v", cfg)
	}
}

func TestInvalidAuthAndEmailVariablesNameTheVariable(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"memória zero", map[string]string{"ARGON2_MEMORY_KIB": "0"}, "ARGON2_MEMORY_KIB"},
		{"memória não numérica", map[string]string{"ARGON2_MEMORY_KIB": "muita"}, "ARGON2_MEMORY_KIB"},
		{"memória absurda", map[string]string{"ARGON2_MEMORY_KIB": "99999999"}, "ARGON2_MEMORY_KIB"},
		{"iterações zero", map[string]string{"ARGON2_ITERATIONS": "0"}, "ARGON2_ITERATIONS"},
		{"iterações absurdas", map[string]string{"ARGON2_ITERATIONS": "1000"}, "ARGON2_ITERATIONS"},
		{"paralelismo zero", map[string]string{"ARGON2_PARALLELISM": "0"}, "ARGON2_PARALLELISM"},
		{"paralelismo acima de 255", map[string]string{"ARGON2_PARALLELISM": "256"}, "ARGON2_PARALLELISM"},
		{"validade zero", map[string]string{"PASSWORD_RESET_TTL_MINUTES": "0"}, "PASSWORD_RESET_TTL_MINUTES"},
		{"validade acima de 24h", map[string]string{"PASSWORD_RESET_TTL_MINUTES": "1441"}, "PASSWORD_RESET_TTL_MINUTES"},
		{"provedor desconhecido", map[string]string{"EMAIL_PROVIDER": "smtp"}, "EMAIL_PROVIDER"},
		{"log sem remetente", map[string]string{"EMAIL_PROVIDER": "log", "APP_BASE_URL": "https://app.tj.example"}, "EMAIL_FROM"},
		{"remetente inválido", map[string]string{"EMAIL_PROVIDER": "log", "EMAIL_FROM": "não é um endereço", "APP_BASE_URL": "https://app.tj.example"}, "EMAIL_FROM"},
		{"log sem URL base", map[string]string{"EMAIL_PROVIDER": "log", "EMAIL_FROM": "no-reply@tj.example"}, "APP_BASE_URL"},
		{"URL base relativa", map[string]string{"EMAIL_PROVIDER": "log", "EMAIL_FROM": "no-reply@tj.example", "APP_BASE_URL": "/painel"}, "APP_BASE_URL"},
		{"URL base http em produção", map[string]string{"EMAIL_PROVIDER": "log", "EMAIL_FROM": "no-reply@tj.example", "APP_BASE_URL": "http://app.tj.example"}, "APP_BASE_URL"},
		{"URL base com consulta", map[string]string{"EMAIL_PROVIDER": "log", "EMAIL_FROM": "no-reply@tj.example", "APP_BASE_URL": "https://app.tj.example/?x=1"}, "APP_BASE_URL"},
		{"URL base com credenciais", map[string]string{"EMAIL_PROVIDER": "log", "EMAIL_FROM": "no-reply@tj.example", "APP_BASE_URL": "https://u:p@app.tj.example"}, "APP_BASE_URL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := load(t, tc.env)
			if err == nil {
				t.Fatal("esperava erro")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("erro %q deveria nomear %s", err, tc.want)
			}
		})
	}
}

func TestLogProviderOutsideDevelopmentWithFromAndHTTPSBaseURL(t *testing.T) {
	cfg, err := load(t, map[string]string{
		"EMAIL_PROVIDER": "log", "EMAIL_FROM": "TJ Platform <no-reply@tj.example>", "APP_BASE_URL": "https://app.tj.example/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EmailProvider != "log" || cfg.AppBaseURL != "https://app.tj.example" || !strings.Contains(cfg.EmailFrom, "no-reply@tj.example") {
		t.Errorf("e-mail = %q/%q/%q", cfg.EmailProvider, cfg.EmailFrom, cfg.AppBaseURL)
	}
}

func TestHTTPBaseURLIsAllowedInDevelopment(t *testing.T) {
	cfg, err := load(t, map[string]string{"APP_ENV": "development", "APP_BASE_URL": "http://localhost:4000"})
	if err != nil || cfg.AppBaseURL != "http://localhost:4000" {
		t.Errorf("cfg = %+v, err = %v", cfg, err)
	}
}

func TestDisabledProviderNeedsNoFromOrBaseURL(t *testing.T) {
	if _, err := load(t, map[string]string{"EMAIL_PROVIDER": "disabled"}); err != nil {
		t.Errorf("provedor disabled não exige remetente nem URL base: %v", err)
	}
}
