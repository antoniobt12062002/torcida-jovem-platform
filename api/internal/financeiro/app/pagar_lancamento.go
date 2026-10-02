package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermLancamentoPay is required to mark a DESPESA lançamento as paid
// (WKF-01 AC5).
const PermLancamentoPay authz.Permission = "financeiro:lancamento:pay"

// ActionLancamentoPay is the audit action for a paid lançamento (WKF-01 AC6).
const ActionLancamentoPay audit.Action = "lancamento.pay"

func init() { audit.Register(ActionLancamentoPay) }

// LancamentoPayer is the narrow shape PagarLancamento needs of
// infra.LancamentoRepository to persist the CRIADA→PAGA transition.
type LancamentoPayer interface {
	Pagar(ctx context.Context, id string) (domain.Lancamento, error)
}

// LancamentoReaderPayer is what PagarLancamento needs to fetch and
// transition a lançamento.
type LancamentoReaderPayer interface {
	LancamentoReader
	LancamentoPayer
}

// PagarLancamento implements WKF-01's despesa half: CRIADA→PAGA, only for a
// DESPESA lançamento.
type PagarLancamento struct {
	Authz       Authorizer
	Lancamentos LancamentoReaderPayer
	Audit       Auditor
	Tx          TxFunc
}

type PagarLancamentoInput struct {
	Actor authz.Principal
	ID    string
}

// Execute checks the permission, fetches the lançamento (must be a CRIADA
// DESPESA), then persists and audits the transition in one transaction.
func (p *PagarLancamento) Execute(ctx context.Context, in PagarLancamentoInput) (domain.Lancamento, error) {
	if p.Authz == nil || p.Lancamentos == nil || p.Audit == nil || p.Tx == nil {
		return domain.Lancamento{}, errNotConfigured
	}
	if err := p.Authz.Require(ctx, in.Actor, PermLancamentoPay); err != nil {
		return domain.Lancamento{}, err
	}

	atual, err := p.Lancamentos.Buscar(ctx, in.ID)
	if err != nil {
		return domain.Lancamento{}, err
	}
	if atual.Tipo != domain.TipoDespesa || atual.Status != domain.StatusCriada {
		return domain.Lancamento{}, domain.ErrLancamentoNaoPodeSerPago
	}

	var pago domain.Lancamento
	err = p.Tx(ctx, func(ctx context.Context) error {
		var err error
		pago, err = p.Lancamentos.Pagar(ctx, in.ID)
		if err != nil {
			return err
		}
		return p.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionLancamentoPay,
			EntityType: "lancamento", EntityID: pago.ID, Outcome: audit.OutcomeSuccess,
		})
	})
	if err != nil {
		return domain.Lancamento{}, err
	}
	return pago, nil
}
