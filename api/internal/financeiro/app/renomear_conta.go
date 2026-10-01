package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermContaUpdate is required to rename a conta contábil (PC-02 AC5).
const PermContaUpdate authz.Permission = "financeiro:conta:update"

// ActionContaRename is the audit action for a renamed conta (PC-02 AC4).
const ActionContaRename audit.Action = "conta.rename"

func init() { audit.Register(ActionContaRename) }

// LancamentoExistenceChecker is the read-only port this sub-spec owns
// (FIN-D-008): whether contaID is referenced by at least one lançamento,
// regardless of status. Its concrete implementation
// (infra.LancamentoExistenceChecker.TemLancamento, built by
// 02-lancamentos/T6) lives in the same financeiro/infra package and
// satisfies this interface structurally — no import between business
// modules (AD-015).
type LancamentoExistenceChecker interface {
	TemLancamento(ctx context.Context, contaID string) (bool, error)
}

// ContaRenamer is the narrow shape RenomearConta needs of
// infra.ContaRepository.
type ContaRenamer interface {
	Buscar(ctx context.Context, id string) (domain.Conta, error)
	Renomear(ctx context.Context, id, nome string) (domain.Conta, error)
}

// RenomearConta implements PC-02: renames a conta contábil only while it has
// never been referenced by any lançamento (FIN-D-008) — not even one that is
// CANCELADA.
type RenomearConta struct {
	Authz       Authorizer
	Contas      ContaRenamer
	Lancamentos LancamentoExistenceChecker
	Audit       Auditor
	Tx          TxFunc
}

type RenomearContaInput struct {
	Actor authz.Principal
	ID    string
	Nome  string
}

// Execute checks the permission, fetches the conta, checks it was never
// used, then persists and audits the rename in one transaction.
func (r *RenomearConta) Execute(ctx context.Context, in RenomearContaInput) (domain.Conta, error) {
	if r.Authz == nil || r.Contas == nil || r.Lancamentos == nil || r.Audit == nil || r.Tx == nil {
		return domain.Conta{}, errNotConfigured
	}
	if err := r.Authz.Require(ctx, in.Actor, PermContaUpdate); err != nil {
		return domain.Conta{}, err
	}

	atual, err := r.Contas.Buscar(ctx, in.ID)
	if err != nil {
		return domain.Conta{}, err
	}

	usada, err := r.Lancamentos.TemLancamento(ctx, in.ID)
	if err != nil {
		return domain.Conta{}, err
	}
	if usada {
		return domain.Conta{}, domain.ErrContaJaUtilizada
	}

	var renomeada domain.Conta
	err = r.Tx(ctx, func(ctx context.Context) error {
		var err error
		renomeada, err = r.Contas.Renomear(ctx, in.ID, in.Nome)
		if err != nil {
			return err
		}
		return r.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionContaRename,
			EntityType: "conta_contabil", EntityID: renomeada.ID, Outcome: audit.OutcomeSuccess,
			Before: map[string]any{"nome": atual.Nome},
			After:  map[string]any{"nome": renomeada.Nome},
		})
	})
	if err != nil {
		return domain.Conta{}, err
	}
	return renomeada, nil
}
