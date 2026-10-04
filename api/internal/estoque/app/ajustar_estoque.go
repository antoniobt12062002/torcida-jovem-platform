package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermMovimentacaoAdjust is required to register an ajuste (AJS-01 AC1).
// Distinct from PermMovimentacaoCreate — ajuste is an exceptional
// operation with its own authorization (decisão do mantenedor, EST-D-004).
const PermMovimentacaoAdjust authz.Permission = "estoque:movimentacao:adjust"

// ActionMovimentacaoAdjust is the audit action for an ajuste.
const ActionMovimentacaoAdjust audit.Action = "movimentacao.adjust"

func init() { audit.Register(ActionMovimentacaoAdjust) }

// MovimentacaoCreatorOnly is the narrow shape AjustarEstoque needs of
// infra.MovimentacaoRepository — deliberately without SaldoComLock: the
// absence of that method in the interface is the structural guarantee
// that this use case can never check the balance, even by mistake (same
// spirit as financeiro/app.ListarComprovantes never having Lancamentos or
// Auditor on its struct).
type MovimentacaoCreatorOnly interface {
	Criar(ctx context.Context, m domain.Movimentacao) (domain.Movimentacao, error)
}

// AjustarEstoque implements AJS-01: registers an AJUSTE — the authorized
// escape valve that may push the balance negative, always with a
// mandatory motivo, checked before any read or write (same order as
// financeiro/app.CancelarLancamento.Execute).
type AjustarEstoque struct {
	Authz         Authorizer
	Produtos      ProdutoReader
	Movimentacoes MovimentacaoCreatorOnly
	Audit         Auditor
	Tx            TxFunc
}

type AjustarEstoqueInput struct {
	Actor      authz.Principal
	ProdutoID  string
	Quantidade int64 // signed delta, never zero
	Motivo     string
}

func (a *AjustarEstoque) Execute(ctx context.Context, in AjustarEstoqueInput) (domain.Movimentacao, error) {
	if a.Authz == nil || a.Produtos == nil || a.Movimentacoes == nil || a.Audit == nil || a.Tx == nil {
		return domain.Movimentacao{}, errNotConfigured
	}
	if err := a.Authz.Require(ctx, in.Actor, PermMovimentacaoAdjust); err != nil {
		return domain.Movimentacao{}, err
	}
	if in.Quantidade == 0 {
		return domain.Movimentacao{}, domain.ErrQuantidadeInvalida
	}
	if motivoBlank(in.Motivo) {
		return domain.Movimentacao{}, domain.ErrMotivoObrigatorio
	}
	if _, err := a.Produtos.Buscar(ctx, in.ProdutoID); err != nil {
		return domain.Movimentacao{}, err
	}

	var criada domain.Movimentacao
	err := a.Tx(ctx, func(ctx context.Context) error {
		motivo := in.Motivo
		var err error
		criada, err = a.Movimentacoes.Criar(ctx, domain.Movimentacao{
			ProdutoID: in.ProdutoID, Tipo: domain.Ajuste, Quantidade: in.Quantidade,
			Origem: domain.OrigemAjusteManual, Motivo: &motivo, ResponsavelID: in.Actor.UserID,
		})
		if err != nil {
			return err
		}
		return a.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionMovimentacaoAdjust,
			EntityType: "movimentacao_estoque", EntityID: criada.ID, Outcome: audit.OutcomeSuccess,
			Reason: in.Motivo,
		})
	})
	if err != nil {
		return domain.Movimentacao{}, err
	}
	return criada, nil
}
