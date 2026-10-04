package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermSaldoRead is required to consult the saldo (AJS-02 AC1). A permission
// of its own, distinct from estoque:movimentacao:read — the saldo is a
// consulta com regra própria (same decision as financeiro's FIN-D-013).
const PermSaldoRead authz.Permission = "estoque:saldo:read"

// SaldoReader is the narrow shape ConsultarSaldo needs of
// infra.MovimentacaoRepository.
type SaldoReader interface {
	Saldo(ctx context.Context, produtoID string) (int64, error)
}

// ConsultarSaldo implements AJS-02: a pure query — sum of every
// movimentação (entrada positiva, saída negativa, devolução already
// signed, ajuste as given), zero and negative values included, never an
// error for a produto with no movimentações.
type ConsultarSaldo struct {
	Authz         Authorizer
	Produtos      ProdutoReader
	Movimentacoes SaldoReader
}

type ConsultarSaldoInput struct {
	Actor     authz.Principal
	ProdutoID string
}

func (c *ConsultarSaldo) Execute(ctx context.Context, in ConsultarSaldoInput) (int64, error) {
	if c.Authz == nil || c.Produtos == nil || c.Movimentacoes == nil {
		return 0, errNotConfigured
	}
	if err := c.Authz.Require(ctx, in.Actor, PermSaldoRead); err != nil {
		return 0, err
	}
	if _, err := c.Produtos.Buscar(ctx, in.ProdutoID); err != nil {
		return 0, err
	}
	return c.Movimentacoes.Saldo(ctx, in.ProdutoID)
}
