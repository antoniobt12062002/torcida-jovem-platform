package config

import (
	"errors"
	"fmt"
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
}

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
	return cfg, nil
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
