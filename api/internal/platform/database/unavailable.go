package database

import (
	"context"
	"database/sql/driver"
	"errors"
	"net"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrUnavailable marks a database connectivity failure — connection refused,
// timeout, exhausted pool, or a PostgreSQL error of the connection-exception
// class or an admin/crash shutdown — as opposed to a business error or a bug.
// A use case maps it to 503 service_unavailable, without the driver's message.
var ErrUnavailable = errors.New("database: indisponível")

// pgConnectionExceptionClass is the PostgreSQL error class for connection
// failures (SQLSTATE 08xxx): https://www.postgresql.org/docs/current/errcodes-appendix.html
const pgConnectionExceptionClass = "08"

// pgAdminOrCrashShutdown are SQLSTATEs for a server going down mid-request:
// 57P01 admin shutdown, 57P02 crash shutdown, 57P03 cannot connect now.
var pgAdminOrCrashShutdown = map[string]bool{"57P01": true, "57P02": true, "57P03": true}

// errDBClosedText is the stable, unexported message database/sql returns once
// a *sql.DB's Close has run: a closed pool is no different, for the caller,
// from a database that stopped answering. Matched as a suffix, since our own
// error wrapping only ever prepends a static prefix ("abrir sessão: ..."),
// never driver output — so this can never leak a DSN or a query.
const errDBClosedText = "sql: database is closed"

// Unavailable reports whether err denotes a database connectivity failure.
func Unavailable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrUnavailable) || errors.Is(err, driver.ErrBadConn) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if strings.HasSuffix(err.Error(), errDBClosedText) {
		return true
	}
	// pgconn.ConnectError always wraps the underlying dial failure, which
	// implements net.Error, so unwrapping to net.Error already covers it — a
	// separate *pgconn.ConnectError check would only ever be dead code.
	if _, ok := errors.AsType[net.Error](err); ok {
		return true
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgAdminOrCrashShutdown[pgErr.Code] {
			return true
		}
		if len(pgErr.Code) >= 2 && pgErr.Code[:2] == pgConnectionExceptionClass {
			return true
		}
	}
	return false
}
