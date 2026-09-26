package app

import (
	"context"
	"errors"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Shared ports of the identity use cases.
type (
	// Authorizer is authz.Authorizer: it decides by effective permissions and
	// audits the denials.
	Authorizer interface {
		Require(ctx context.Context, p authz.Principal, perm authz.Permission) error
	}
	// Auditor records critical entries in the caller's transaction (Record) and
	// security events on their own (RecordSecurity): *audit.Recorder.
	Auditor interface {
		Record(ctx context.Context, e audit.Entry) error
		RecordSecurity(ctx context.Context, e audit.Entry)
	}
	// TxFunc runs fn in one unit of work (database.WithTx).
	TxFunc func(ctx context.Context, fn func(ctx context.Context) error) error
	// PasswordHasher hashes a password with the configured parameters.
	PasswordHasher interface {
		Hash(plain string) (string, error)
	}
	Clock func() time.Time
)

var errNotConfigured = errors.New("identidade: caso de uso mal configurado")
