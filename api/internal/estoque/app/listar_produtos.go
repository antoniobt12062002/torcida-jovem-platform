package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermProdutoRead is required to list produtos (SKUs) de estoque (PRD-02 AC1).
const PermProdutoRead authz.Permission = "estoque:produto:read"

// ProdutoLister is the narrow shape ListarProdutos needs of
// infra.ProdutoRepository.
type ProdutoLister interface {
	Listar(ctx context.Context) ([]domain.Produto, error)
}

// ListarProdutos implements PRD-02: lists every produto (SKU) de estoque,
// no pagination, filter or ordering (same deliberate V1 decision as
// financeiro's FIN-D-021 — never implemented here).
type ListarProdutos struct {
	Authz    Authorizer
	Produtos ProdutoLister
}

type ListarProdutosInput struct {
	Actor authz.Principal
}

func (l *ListarProdutos) Execute(ctx context.Context, in ListarProdutosInput) ([]domain.Produto, error) {
	if l.Authz == nil || l.Produtos == nil {
		return nil, errNotConfigured
	}
	if err := l.Authz.Require(ctx, in.Actor, PermProdutoRead); err != nil {
		return nil, err
	}
	return l.Produtos.Listar(ctx)
}
