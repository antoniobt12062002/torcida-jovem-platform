package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermLancamentoCancel is required to cancel a lançamento (WKF-02 AC5).
const PermLancamentoCancel authz.Permission = "financeiro:lancamento:cancel"

// ActionLancamentoCancel is the audit action for a cancelled lançamento
// (WKF-02 AC6). platform/audit's requiresReason() already enforces a
// non-empty Reason on any action ending in ".cancel".
const ActionLancamentoCancel audit.Action = "lancamento.cancel"

func init() { audit.Register(ActionLancamentoCancel) }

// LancamentoCanceller is the narrow shape CancelarLancamento needs of
// infra.LancamentoRepository to persist the →CANCELADA transition.
type LancamentoCanceller interface {
	Cancelar(ctx context.Context, id, motivo, canceladoPor string) (domain.Lancamento, error)
}

// LancamentoReaderCanceller is what CancelarLancamento needs to fetch and
// cancel a lançamento.
type LancamentoReaderCanceller interface {
	LancamentoReader
	LancamentoCanceller
}

// CancelarLancamento implements WKF-02: cancels a lançamento from CRIADA,
// RECEBIDA or PAGA — never from CANCELADA again — with a mandatory motivo.
// It never treats cancellation as a devolução and never touches any other
// lançamento row (FIN-D-006: the saldo window this can open is an accepted
// consequence, not mitigated here).
type CancelarLancamento struct {
	Authz       Authorizer
	Lancamentos LancamentoReaderCanceller
	Audit       Auditor
	Tx          TxFunc
}

type CancelarLancamentoInput struct {
	Actor  authz.Principal
	ID     string
	Motivo string
}

// Execute checks the permission, validates the motivo is non-empty (before
// any read or write), fetches the lançamento (must not already be
// CANCELADA), then persists and audits the cancellation in one transaction.
func (c *CancelarLancamento) Execute(ctx context.Context, in CancelarLancamentoInput) (domain.Lancamento, error) {
	if c.Authz == nil || c.Lancamentos == nil || c.Audit == nil || c.Tx == nil {
		return domain.Lancamento{}, errNotConfigured
	}
	if err := c.Authz.Require(ctx, in.Actor, PermLancamentoCancel); err != nil {
		return domain.Lancamento{}, err
	}
	if motivoBlank(in.Motivo) {
		return domain.Lancamento{}, domain.ErrMotivoObrigatorio
	}

	atual, err := c.Lancamentos.Buscar(ctx, in.ID)
	if err != nil {
		return domain.Lancamento{}, err
	}
	if atual.Status == domain.StatusCancelada {
		return domain.Lancamento{}, domain.ErrLancamentoJaCancelado
	}

	var cancelado domain.Lancamento
	err = c.Tx(ctx, func(ctx context.Context) error {
		var err error
		cancelado, err = c.Lancamentos.Cancelar(ctx, in.ID, in.Motivo, in.Actor.UserID)
		if err != nil {
			return err
		}
		return c.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionLancamentoCancel,
			EntityType: "lancamento", EntityID: cancelado.ID, Outcome: audit.OutcomeSuccess,
			Reason: in.Motivo,
		})
	})
	if err != nil {
		return domain.Lancamento{}, err
	}
	return cancelado, nil
}
