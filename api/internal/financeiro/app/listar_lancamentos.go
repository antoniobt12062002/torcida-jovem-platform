package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermLancamentoRead is required to list lançamentos (LAN-04 AC2).
const PermLancamentoRead authz.Permission = "financeiro:lancamento:read"

// LancamentoLister is the narrow shape ListarLancamentos needs of
// infra.LancamentoRepository.
type LancamentoLister interface {
	Listar(ctx context.Context) ([]domain.Lancamento, error)
}

// ListarLancamentos implements LAN-04: lists every lançamento with every
// field, nothing hidden by default (including CANCELADA).
type ListarLancamentos struct {
	Authz       Authorizer
	Lancamentos LancamentoLister
}

type ListarLancamentosInput struct {
	Actor authz.Principal
}

func (l *ListarLancamentos) Execute(ctx context.Context, in ListarLancamentosInput) ([]domain.Lancamento, error) {
	if l.Authz == nil || l.Lancamentos == nil {
		return nil, errNotConfigured
	}
	if err := l.Authz.Require(ctx, in.Actor, PermLancamentoRead); err != nil {
		return nil, err
	}
	return l.Lancamentos.Listar(ctx)
}
