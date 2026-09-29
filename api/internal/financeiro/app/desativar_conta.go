package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermContaDeactivate is required to deactivate a conta contábil (PC-03 AC5).
const PermContaDeactivate authz.Permission = "financeiro:conta:deactivate"

// ActionContaDeactivate is the audit action for a deactivated conta (PC-03 AC3).
const ActionContaDeactivate audit.Action = "conta.deactivate"

func init() { audit.Register(ActionContaDeactivate) }

// ContaDeactivator is the narrow shape DesativarConta needs of
// infra.ContaRepository.
type ContaDeactivator interface {
	Desativar(ctx context.Context, id string) error
}

// DesativarConta implements PC-03: deactivates a conta contábil, never
// deleting the row. Deactivating an already-inactive conta is not an error
// (idempotent, Edge Cases).
type DesativarConta struct {
	Authz  Authorizer
	Contas ContaDeactivator
	Audit  Auditor
	Tx     TxFunc
}

type DesativarContaInput struct {
	Actor authz.Principal
	ID    string
}

// Execute checks the permission, then deactivates and audits the conta in
// one transaction.
func (d *DesativarConta) Execute(ctx context.Context, in DesativarContaInput) error {
	if d.Authz == nil || d.Contas == nil || d.Audit == nil || d.Tx == nil {
		return errNotConfigured
	}
	if err := d.Authz.Require(ctx, in.Actor, PermContaDeactivate); err != nil {
		return err
	}

	return d.Tx(ctx, func(ctx context.Context) error {
		if err := d.Contas.Desativar(ctx, in.ID); err != nil {
			return err
		}
		return d.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionContaDeactivate,
			EntityType: "conta_contabil", EntityID: in.ID, Outcome: audit.OutcomeSuccess,
		})
	})
}
