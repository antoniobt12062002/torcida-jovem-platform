package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermMovimentacaoRead is required to list movimentações of a produto (MOV-04 AC1).
const PermMovimentacaoRead authz.Permission = "estoque:movimentacao:read"

// MovimentacaoLister is the narrow shape ListarMovimentacoes needs of
// infra.MovimentacaoRepository.
type MovimentacaoLister interface {
	ListarPorProduto(ctx context.Context, produtoID string) ([]domain.Movimentacao, error)
}

// ListarMovimentacoes implements MOV-04: lists every movimentação of a
// produto, in the order they occurred, of every tipo (including AJUSTE,
// created by 03-ajustes-e-saldo) — no pagination (same deliberate V1
// decision as financeiro's FIN-D-021).
type ListarMovimentacoes struct {
	Authz         Authorizer
	Movimentacoes MovimentacaoLister
}

type ListarMovimentacoesInput struct {
	Actor     authz.Principal
	ProdutoID string
}

func (l *ListarMovimentacoes) Execute(ctx context.Context, in ListarMovimentacoesInput) ([]domain.Movimentacao, error) {
	if l.Authz == nil || l.Movimentacoes == nil {
		return nil, errNotConfigured
	}
	if err := l.Authz.Require(ctx, in.Actor, PermMovimentacaoRead); err != nil {
		return nil, err
	}
	return l.Movimentacoes.ListarPorProduto(ctx, in.ProdutoID)
}
