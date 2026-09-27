package config

import (
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv          string
	LogLevel        string
	Port            string
	DatabaseURL     string
	AllowedOrigins  []string
	CookieSecure    bool
	CookieDomain    string
	SessionIdle     time.Duration
	SessionAbsolute time.Duration

	// AuthHashKey keys the HMAC-SHA256 of e-mails (login attempts, audit). It is
	// a secret: it never appears in an error message or a log.
	AuthHashKey       []byte
	Argon2MemoryKiB   uint32
	Argon2Iterations  uint32
	Argon2Parallelism uint8
	PasswordResetTTL  time.Duration

	EmailProvider string // "log" or "disabled"; real providers arrive later
	EmailFrom     string
	AppBaseURL    string // base of the links sent by e-mail, without trailing slash

	// StorageEnabled gates the S3 variables below: nothing in the API requires
	// object storage yet, so it stays off unless a consuming module needs it.
	StorageEnabled bool
	S3Endpoint     string
	S3Region       string
	S3Bucket       string
	// S3AccessKey and S3SecretKey are secrets: they never appear in an error
	// message or a log.
	S3AccessKey    string
	S3SecretKey    string
	S3UsePathStyle bool // required by local S3 emulators
	DocumentURLTTL time.Duration
}

// devAuthHashKey is used only when APP_ENV=development and AUTH_HASH_KEY is unset.
const devAuthHashKey = "dev-only-auth-hash-key-do-not-use-in-production"

const (
	envDevelopment = "development"
	envStaging     = "staging"
	envProduction  = "production"
)

func Load() (Config, error) {
	cfg := Config{
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL e obrigatoria")
	}

	port, err := strconv.Atoi(cfg.Port)
	if err != nil || port < 1 || port > 65535 {
		return Config{}, invalid("PORT", "deve ser um numero entre 1 e 65535")
	}

	cfg.AppEnv = getenv("APP_ENV", envProduction)
	switch cfg.AppEnv {
	case envDevelopment, envStaging, envProduction:
	default:
		return Config{}, invalid("APP_ENV", "deve ser development, staging ou production")
	}

	cfg.LogLevel = getenv("LOG_LEVEL", "info")
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return Config{}, invalid("LOG_LEVEL", "deve ser debug, info, warn ou error")
	}

	cfg.CookieSecure = true
	if raw := os.Getenv("COOKIE_SECURE"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, invalid("COOKIE_SECURE", "deve ser true ou false")
		}
		cfg.CookieSecure = v
	}
	if !cfg.CookieSecure && cfg.AppEnv != envDevelopment {
		return Config{}, invalid("COOKIE_SECURE", "so pode ser false quando APP_ENV=development")
	}
	cfg.CookieDomain = os.Getenv("COOKIE_DOMAIN")

	idle, err := positiveInt("SESSION_IDLE_MINUTES", 60)
	if err != nil {
		return Config{}, err
	}
	absolute, err := positiveInt("SESSION_ABSOLUTE_HOURS", 8)
	if err != nil {
		return Config{}, err
	}
	cfg.SessionIdle = time.Duration(idle) * time.Minute
	cfg.SessionAbsolute = time.Duration(absolute) * time.Hour

	cfg.AllowedOrigins, err = origins(cfg.AppEnv)
	if err != nil {
		return Config{}, err
	}
	if err := loadAuth(&cfg); err != nil {
		return Config{}, err
	}
	if err := loadEmail(&cfg); err != nil {
		return Config{}, err
	}
	if err := loadStorage(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func loadAuth(cfg *Config) error {
	key := os.Getenv("AUTH_HASH_KEY")
	switch {
	case key == "" && cfg.AppEnv == envDevelopment:
		key = devAuthHashKey
	case key == "":
		return invalid("AUTH_HASH_KEY", "e obrigatoria fora de development")
	case len(key) < 32:
		return invalid("AUTH_HASH_KEY", "deve ter no minimo 32 bytes")
	}
	cfg.AuthHashKey = []byte(key)

	memory, err := boundedInt("ARGON2_MEMORY_KIB", 19456, 1<<20)
	if err != nil {
		return err
	}
	iterations, err := boundedInt("ARGON2_ITERATIONS", 2, 20)
	if err != nil {
		return err
	}
	parallelism, err := boundedInt("ARGON2_PARALLELISM", 1, 255)
	if err != nil {
		return err
	}
	ttl, err := boundedInt("PASSWORD_RESET_TTL_MINUTES", 30, 24*60)
	if err != nil {
		return err
	}
	cfg.Argon2MemoryKiB = uint32(memory)
	cfg.Argon2Iterations = uint32(iterations)
	cfg.Argon2Parallelism = uint8(parallelism)
	cfg.PasswordResetTTL = time.Duration(ttl) * time.Minute
	return nil
}

func loadEmail(cfg *Config) error {
	dev := cfg.AppEnv == envDevelopment
	defProvider := "disabled"
	if dev {
		defProvider = "log"
	}
	cfg.EmailProvider = getenv("EMAIL_PROVIDER", defProvider)
	switch cfg.EmailProvider {
	case "log", "disabled":
	default:
		return invalid("EMAIL_PROVIDER", "deve ser log ou disabled")
	}

	defBase, defFrom := "", ""
	if dev {
		defBase, defFrom = "http://localhost:3000", "TJ Platform <no-reply@localhost>"
	}
	base, from := getenv("APP_BASE_URL", defBase), getenv("EMAIL_FROM", defFrom)
	if cfg.EmailProvider == "disabled" && os.Getenv("APP_BASE_URL") == "" && os.Getenv("EMAIL_FROM") == "" {
		return nil
	}
	if cfg.EmailProvider != "disabled" {
		if from == "" {
			return invalid("EMAIL_FROM", "e obrigatoria quando EMAIL_PROVIDER nao e disabled")
		}
		if base == "" {
			return invalid("APP_BASE_URL", "e obrigatoria quando EMAIL_PROVIDER nao e disabled")
		}
	}
	if from != "" {
		if _, err := mail.ParseAddress(from); err != nil {
			return invalid("EMAIL_FROM", "deve ser um endereco de e-mail valido")
		}
		cfg.EmailFrom = from
	}
	if base != "" {
		u, err := url.Parse(base)
		switch {
		case err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https"):
			return invalid("APP_BASE_URL", "deve ser uma URL absoluta http ou https")
		case u.Scheme != "https" && !dev:
			return invalid("APP_BASE_URL", "deve usar https fora de development")
		case u.User != nil || u.RawQuery != "" || u.Fragment != "":
			return invalid("APP_BASE_URL", "nao pode ter credenciais, consulta nem fragmento")
		}
		cfg.AppBaseURL = strings.TrimSuffix(base, "/")
	}
	return nil
}

// loadStorage reads the S3 object storage configuration (DOC-03). Storage stays
// disabled by default: nothing in the API consumes it yet. When enabled, the
// five S3 variables are all required, named individually so a missing one is
// unambiguous; no value read here is ever echoed back in an error.
func loadStorage(cfg *Config) error {
	enabled, err := boolEnv("STORAGE_ENABLED", false)
	if err != nil {
		return err
	}
	cfg.StorageEnabled = enabled

	ttl, err := positiveInt("DOCUMENT_URL_TTL_SECONDS", 300)
	if err != nil {
		return err
	}
	cfg.DocumentURLTTL = time.Duration(ttl) * time.Second

	cfg.S3UsePathStyle, err = boolEnv("S3_USE_PATH_STYLE", false)
	if err != nil {
		return err
	}
	if !enabled {
		return nil
	}
	for _, v := range []struct {
		name string
		dst  *string
	}{
		{"S3_ENDPOINT", &cfg.S3Endpoint},
		{"S3_REGION", &cfg.S3Region},
		{"S3_BUCKET", &cfg.S3Bucket},
		{"S3_ACCESS_KEY", &cfg.S3AccessKey},
		{"S3_SECRET_KEY", &cfg.S3SecretKey},
	} {
		val := os.Getenv(v.name)
		if val == "" {
			return invalid(v.name, "e obrigatoria quando STORAGE_ENABLED=true")
		}
		*v.dst = val
	}
	return nil
}

func boolEnv(name string, def bool) (bool, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return def, nil
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return false, invalid(name, "deve ser true ou false")
	}
	return v, nil
}

func boundedInt(name string, def, maxValue int) (int, error) {
	v, err := positiveInt(name, def)
	if err != nil {
		return 0, err
	}
	if v > maxValue {
		return 0, invalid(name, fmt.Sprintf("deve ser no maximo %d", maxValue))
	}
	return v, nil
}

func invalid(name, reason string) error {
	return fmt.Errorf("%s invalida: %s", name, reason)
}

func getenv(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

func positiveInt(name string, def int) (int, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, invalid(name, "deve ser um inteiro positivo")
	}
	return v, nil
}

func origins(appEnv string) ([]string, error) {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		if appEnv == envDevelopment {
			return []string{"http://localhost:3000"}, nil
		}
		return nil, nil
	}
	var list []string
	for part := range strings.SplitSeq(raw, ",") {
		o := strings.TrimSpace(part)
		u, err := url.Parse(o)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || (u.Path != "" && u.Path != "/") || u.RawQuery != "" {
			return nil, invalid("ALLOWED_ORIGINS", "cada origem deve ter o formato esquema://host[:porta], sem caminho")
		}
		list = append(list, strings.TrimSuffix(o, "/"))
	}
	return list, nil
}
