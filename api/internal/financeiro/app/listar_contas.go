package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermContaRead is required to list contas contábeis (PC-04 AC2).
const PermContaRead authz.Permission = "financeiro:conta:read"

// ContaLister is the narrow shape ListarContas needs of infra.ContaRepository.
type ContaLister interface {
	Listar(ctx context.Context) ([]domain.Conta, error)
}

// ListarContas implements PC-04: lists every conta contábil with its
// hierarchy and ativo status.
type ListarContas struct {
	Authz  Authorizer
	Contas ContaLister
}

type ListarContasInput struct {
	Actor authz.Principal
}

func (l *ListarContas) Execute(ctx context.Context, in ListarContasInput) ([]domain.Conta, error) {
	if l.Authz == nil || l.Contas == nil {
		return nil, errNotConfigured
	}
	if err := l.Authz.Require(ctx, in.Actor, PermContaRead); err != nil {
		return nil, err
	}
	return l.Contas.Listar(ctx)
}
