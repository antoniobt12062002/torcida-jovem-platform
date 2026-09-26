package config

import "testing"

func TestLoad(t *testing.T) {
	t.Run("usa porta padrao", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://x")
		t.Setenv("PORT", "")
		cfg, err := Load()
		if err != nil || cfg.Port != "8080" || cfg.DatabaseURL != "postgres://x" {
			t.Fatalf("cfg=%+v err=%v", cfg, err)
		}
	})
	t.Run("respeita PORT", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "postgres://x")
		t.Setenv("PORT", "9000")
		cfg, _ := Load()
		if cfg.Port != "9000" {
			t.Fatalf("port=%s", cfg.Port)
		}
	})
	t.Run("exige DATABASE_URL", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")
		if _, err := Load(); err == nil {
			t.Fatal("esperava erro")
		}
	})
}
