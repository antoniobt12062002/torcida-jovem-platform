// Package app holds the estoque use cases: application logic that depends
// only on ports (domain + narrow interfaces), never on concrete
// infrastructure.
package app

import (
	"context"
	"errors"
	"strings"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Shared ports of the estoque use cases.
type (
	// Authorizer is the narrow shape estoque needs of authz.Authorizer
	// (AD-015): estoque never imports the concrete type.
	Authorizer interface {
		Require(ctx context.Context, p authz.Principal, perm authz.Permission) error
	}
	// Auditor is the narrow shape estoque needs of *audit.Recorder.
	Auditor interface {
		Record(ctx context.Context, e audit.Entry) error
	}
	// TxFunc runs fn in one unit of work (database.WithTx).
	TxFunc func(ctx context.Context, fn func(ctx context.Context) error) error
)

var errNotConfigured = errors.New("estoque: caso de uso mal configurado")

func motivoBlank(motivo string) bool { return strings.TrimSpace(motivo) == "" }
