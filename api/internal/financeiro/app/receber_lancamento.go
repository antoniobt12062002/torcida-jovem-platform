package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermLancamentoReceive is required to mark a RECEITA lançamento as received
// (WKF-01 AC5).
const PermLancamentoReceive authz.Permission = "financeiro:lancamento:receive"

// ActionLancamentoReceive is the audit action for a received lançamento
// (WKF-01 AC6).
const ActionLancamentoReceive audit.Action = "lancamento.receive"

func init() { audit.Register(ActionLancamentoReceive) }

// LancamentoReceiver is the narrow shape ReceberLancamento needs of
// infra.LancamentoRepository to persist the CRIADA→RECEBIDA transition.
type LancamentoReceiver interface {
	Receber(ctx context.Context, id string) (domain.Lancamento, error)
}

// LancamentoReaderReceiver is what ReceberLancamento needs to fetch and
// transition a lançamento.
type LancamentoReaderReceiver interface {
	LancamentoReader
	LancamentoReceiver
}

// ReceberLancamento implements WKF-01's receita half: CRIADA→RECEBIDA, only
// for a RECEITA lançamento.
type ReceberLancamento struct {
	Authz       Authorizer
	Lancamentos LancamentoReaderReceiver
	Audit       Auditor
	Tx          TxFunc
}

type ReceberLancamentoInput struct {
	Actor authz.Principal
	ID    string
}

// Execute checks the permission, fetches the lançamento (must be a CRIADA
// RECEITA), then persists and audits the transition in one transaction.
func (r *ReceberLancamento) Execute(ctx context.Context, in ReceberLancamentoInput) (domain.Lancamento, error) {
	if r.Authz == nil || r.Lancamentos == nil || r.Audit == nil || r.Tx == nil {
		return domain.Lancamento{}, errNotConfigured
	}
	if err := r.Authz.Require(ctx, in.Actor, PermLancamentoReceive); err != nil {
		return domain.Lancamento{}, err
	}

	atual, err := r.Lancamentos.Buscar(ctx, in.ID)
	if err != nil {
		return domain.Lancamento{}, err
	}
	if atual.Tipo != domain.TipoReceita || atual.Status != domain.StatusCriada {
		return domain.Lancamento{}, domain.ErrLancamentoNaoPodeSerRecebido
	}

	var recebido domain.Lancamento
	err = r.Tx(ctx, func(ctx context.Context) error {
		var err error
		recebido, err = r.Lancamentos.Receber(ctx, in.ID)
		if err != nil {
			return err
		}
		return r.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionLancamentoReceive,
			EntityType: "lancamento", EntityID: recebido.ID, Outcome: audit.OutcomeSuccess,
		})
	})
	if err != nil {
		return domain.Lancamento{}, err
	}
	return recebido, nil
}
