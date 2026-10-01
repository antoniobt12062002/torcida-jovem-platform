package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermContaCreate is required to create a conta contábil (PC-01 AC3).
const PermContaCreate authz.Permission = "financeiro:conta:create"

// ActionContaCreate is the audit action for a created conta (PC-01 AC4). The
// domain segment is "conta", not "financeiro.conta": platform/audit's
// Register enforces the two-segment "domain.verb" format (see
// platform/audit/entry.go), the same convention identity and documents
// already follow (e.g. "user.create", "document.create").
const ActionContaCreate audit.Action = "conta.create"

func init() { audit.Register(ActionContaCreate) }

// ContaRepository is the narrow shape CriarConta and ListarContas need of
// infra.ContaRepository.
type ContaRepository interface {
	Buscar(ctx context.Context, id string) (domain.Conta, error)
	Criar(ctx context.Context, c domain.Conta) (domain.Conta, error)
}

// CriarConta implements PC-01: creates a conta contábil, root or subconta.
type CriarConta struct {
	Authz  Authorizer
	Contas ContaRepository
	Audit  Auditor
	Tx     TxFunc
}

// CriarContaInput carries what Execute needs. ParentID is nil for a root
// conta.
type CriarContaInput struct {
	Actor    authz.Principal
	Tipo     domain.TipoConta
	Nome     string
	ParentID *string
}

// Execute checks the permission, validates a given parent (it must exist and
// share the same Tipo, PC-01 AC2), then persists and audits the conta in one
// transaction.
func (c *CriarConta) Execute(ctx context.Context, in CriarContaInput) (domain.Conta, error) {
	if c.Authz == nil || c.Contas == nil || c.Audit == nil || c.Tx == nil {
		return domain.Conta{}, errNotConfigured
	}
	if err := c.Authz.Require(ctx, in.Actor, PermContaCreate); err != nil {
		return domain.Conta{}, err
	}

	if in.ParentID != nil {
		parent, err := c.Contas.Buscar(ctx, *in.ParentID)
		if err != nil {
			return domain.Conta{}, err
		}
		if parent.Tipo != in.Tipo {
			return domain.Conta{}, domain.ErrContaTipoIncompativel
		}
	}

	conta := domain.Conta{Tipo: in.Tipo, Nome: in.Nome, ParentID: in.ParentID}
	var criada domain.Conta
	err := c.Tx(ctx, func(ctx context.Context) error {
		var err error
		criada, err = c.Contas.Criar(ctx, conta)
		if err != nil {
			return err
		}
		return c.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionContaCreate,
			EntityType: "conta_contabil", EntityID: criada.ID, Outcome: audit.OutcomeSuccess,
			After: map[string]any{"id": criada.ID, "tipo": string(criada.Tipo), "nome": criada.Nome},
		})
	})
	if err != nil {
		return domain.Conta{}, err
	}
	return criada, nil
}
