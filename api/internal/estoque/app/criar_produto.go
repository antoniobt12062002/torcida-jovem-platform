package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermProdutoCreate is required to create a produto (SKU) de estoque (PRD-01 AC1).
const PermProdutoCreate authz.Permission = "estoque:produto:create"

// ActionProdutoCreate is the audit action for a created produto.
const ActionProdutoCreate audit.Action = "produto.create"

func init() { audit.Register(ActionProdutoCreate) }

// ProdutoCreator is the narrow shape CriarProduto needs of
// infra.ProdutoRepository.
type ProdutoCreator interface {
	Criar(ctx context.Context, p domain.Produto) (domain.Produto, error)
}

// CriarProduto implements PRD-01: creates a produto (SKU) de estoque.
type CriarProduto struct {
	Authz    Authorizer
	Produtos ProdutoCreator
	Audit    Auditor
	Tx       TxFunc
}

// CriarProdutoInput carries what Execute needs.
type CriarProdutoInput struct {
	Actor         authz.Principal
	Codigo        string
	Nome          string
	UnidadeMedida string
}

// Execute checks the permission, validates that no field is blank, then
// persists and audits the produto in one transaction.
func (c *CriarProduto) Execute(ctx context.Context, in CriarProdutoInput) (domain.Produto, error) {
	if c.Authz == nil || c.Produtos == nil || c.Audit == nil || c.Tx == nil {
		return domain.Produto{}, errNotConfigured
	}
	if err := c.Authz.Require(ctx, in.Actor, PermProdutoCreate); err != nil {
		return domain.Produto{}, err
	}
	if motivoBlank(in.Codigo) || motivoBlank(in.Nome) || motivoBlank(in.UnidadeMedida) {
		return domain.Produto{}, domain.ErrCampoObrigatorio
	}

	produto := domain.Produto{Codigo: in.Codigo, Nome: in.Nome, UnidadeMedida: in.UnidadeMedida}
	var criado domain.Produto
	err := c.Tx(ctx, func(ctx context.Context) error {
		var err error
		criado, err = c.Produtos.Criar(ctx, produto)
		if err != nil {
			return err
		}
		return c.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionProdutoCreate,
			EntityType: "produto_estoque", EntityID: criado.ID, Outcome: audit.OutcomeSuccess,
			After: map[string]any{"codigo": criado.Codigo, "nome": criado.Nome, "unidade_medida": criado.UnidadeMedida},
		})
	})
	if err != nil {
		return domain.Produto{}, err
	}
	return criado, nil
}
