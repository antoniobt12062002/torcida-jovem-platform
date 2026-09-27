package database_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

type fakeNetError struct{}

func (fakeNetError) Error() string   { return "conexão recusada" }
func (fakeNetError) Timeout() bool   { return true }
func (fakeNetError) Temporary() bool { return true }

// Edge case: "IF the database is unavailable during a request THEN 503 with
// code service_unavailable" — this is what recognizes that condition.
func TestUnavailableRecognizesConnectivityFailures(t *testing.T) {
	cases := map[string]error{
		"ErrUnavailable diretamente":         database.ErrUnavailable,
		"embrulhado com %w":                  fmt.Errorf("abrir conexão: %w", database.ErrUnavailable),
		"driver.ErrBadConn":                  driver.ErrBadConn,
		"driver.ErrBadConn embrulhado":       fmt.Errorf("query: %w", driver.ErrBadConn),
		"context.DeadlineExceeded":           context.DeadlineExceeded,
		"net.Error":                          fakeNetError{},
		"net.Error embrulhado":               fmt.Errorf("dial: %w", fakeNetError{}),
		"PgError 08006 (connection_failure)": &pgconn.PgError{Code: "08006"},
		"PgError 08001 (sqlclient_unable_to_establish_sqlconnection)": &pgconn.PgError{Code: "08001"},
		"PgError 57P01 (admin_shutdown)":                              &pgconn.PgError{Code: "57P01"},
		"PgError 57P02 (crash_shutdown)":                              &pgconn.PgError{Code: "57P02"},
		"PgError 57P03 (cannot_connect_now)":                          &pgconn.PgError{Code: "57P03"},
		"pool fechado (database/sql)":                                 errors.New("sql: database is closed"),
		"pool fechado, com prefixo":                                   fmt.Errorf("abrir sessão: %v", errors.New("sql: database is closed")),
	}
	for name, err := range cases {
		if !database.Unavailable(err) {
			t.Errorf("%s: deveria ser reconhecido como indisponibilidade", name)
		}
	}
}

func TestUnavailableRejectsBusinessAndNilErrors(t *testing.T) {
	cases := map[string]error{
		"nil":                                   nil,
		"erro de negócio qualquer":              errors.New("email_taken"),
		"PgError de violação de unicidade":      &pgconn.PgError{Code: "23505"},
		"PgError de dados inválidos":            &pgconn.PgError{Code: "22P02"},
		"contexto cancelado pelo cliente":       context.Canceled,
		"frase parecida, mas não é o sentinela": errors.New("o campo database is closed for edits"),
	}
	for name, err := range cases {
		if database.Unavailable(err) {
			t.Errorf("%s: não deveria ser tratado como indisponibilidade de banco", name)
		}
	}
}

var _ net.Error = fakeNetError{}
