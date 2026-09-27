//go:build integration

package database_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// A real pgconn.ConnectError — the shape pgx actually returns when a connection
// attempt fails — is recognized. The port is closed on loopback, so the refusal
// is immediate, with no Docker and no timeout to wait out.
func TestUnavailableRecognizesARealConnectError(t *testing.T) {
	_, err := pgconn.Connect(context.Background(), "postgres://tj_app:x@127.0.0.1:1/tj?sslmode=disable&connect_timeout=2")
	if err == nil {
		t.Fatal("a conexão deveria falhar (porta fechada)")
	}
	if _, ok := err.(*pgconn.ConnectError); !ok {
		t.Fatalf("o fixture não é o *pgconn.ConnectError esperado: %T: %v", err, err)
	}

	if !database.Unavailable(err) {
		t.Errorf("um erro real de conexão recusada deveria ser reconhecido como indisponibilidade")
	}
}
