package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/money"
)

// PermSaldoRead is required to consult the saldo (WKF-03 AC3). A permission
// of its own, distinct from financeiro:lancamento:read (FIN-D-013).
const PermSaldoRead authz.Permission = "financeiro:saldo:read"

// SaldoReader is the narrow shape ConsultarSaldo needs of
// infra.LancamentoRepository.
type SaldoReader interface {
	Saldo(ctx context.Context) (int64, error)
}

// ConsultarSaldo implements WKF-03: a pure query, in cash-basis regime
// (FIN-D-001) — sum of RECEBIDA minus sum of PAGA, by current status. It
// never writes, never audits, never caches.
type ConsultarSaldo struct {
	Authz       Authorizer
	Lancamentos SaldoReader
}

type ConsultarSaldoInput struct {
	Actor authz.Principal
}

func (c *ConsultarSaldo) Execute(ctx context.Context, in ConsultarSaldoInput) (money.Cents, error) {
	if c.Authz == nil || c.Lancamentos == nil {
		return 0, errNotConfigured
	}
	if err := c.Authz.Require(ctx, in.Actor, PermSaldoRead); err != nil {
		return 0, err
	}
	saldo, err := c.Lancamentos.Saldo(ctx)
	if err != nil {
		return 0, err
	}
	return money.Cents(saldo), nil
}
