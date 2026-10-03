//go:build integration

package main

import (
	"errors"
	"io"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/config"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// baseTestConfig is a config.Config built directly (not through config.Load,
// which reads the environment) with fast, test-only parameters — same
// pattern already used by internal/httpapi/e2e_test.go.
func baseTestConfig(t *testing.T) config.Config {
	t.Helper()
	db := testutil.NewTestDB(t)
	s3 := testutil.SharedS3(t)
	return config.Config{
		AppEnv:            "development",
		LogLevel:          "error",
		Port:              "0",
		DatabaseURL:       testutil.AppDSN(t, db),
		AuthHashKey:       []byte("boot-test-hash-key-0123456789ab"),
		Argon2MemoryKiB:   64,
		Argon2Iterations:  1,
		Argon2Parallelism: 1,
		PasswordResetTTL:  3600_000_000_000, // 1h, in time.Duration nanoseconds
		EmailProvider:     "disabled",
		AppBaseURL:        "http://localhost:3000",
		StorageEnabled:    true,
		S3Endpoint:        s3.Endpoint,
		S3Region:          s3.Region,
		S3Bucket:          s3.Bucket,
		S3AccessKey:       s3.AccessKey,
		S3SecretKey:       s3.SecretKey,
		S3UsePathStyle:    true,
		DocumentURLTTL:    300_000_000_000, // 5min
	}
}

// 06-api-http/T1: financeiro's composition root (financeiro.New) is wired
// into cmd/api's own boot sequence, alongside identity's, with no error and
// a real, usable handler — the first test cmd/api has ever had.
func TestBootWiresFinanceiroAlongsideIdentityWithStorageEnabled(t *testing.T) {
	cfg := baseTestConfig(t)
	logger := logx.New("error", io.Discard)

	handler, err := boot(cfg, logger)

	if err != nil {
		t.Fatalf("boot: %v", err)
	}
	if handler == nil {
		t.Fatal("boot devolveu handler nil sem erro")
	}
}

// FIN-D-023: a API se recusa a compor financeiro sem armazenamento
// configurado — nunca um 503 por requisição, a falha é no boot.
func TestBootFailsWithoutStorageEnabled(t *testing.T) {
	cfg := baseTestConfig(t)
	cfg.StorageEnabled = false
	logger := logx.New("error", io.Discard)

	_, err := boot(cfg, logger)

	if !errors.Is(err, errStorageRequired) {
		t.Fatalf("err = %v, esperado errStorageRequired", err)
	}
}
