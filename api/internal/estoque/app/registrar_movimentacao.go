package app

import (
	"context"
	"errors"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PermMovimentacaoCreate is required to register an entrada, saída or
// devolução (MOV-01/MOV-02/MOV-03). Devolução reuses it — it is not a
// distinct business operation, same decision as financeiro's FIN-D-003.
const PermMovimentacaoCreate authz.Permission = "estoque:movimentacao:create"

// ActionMovimentacaoCreate is the audit action, reused by the 3 tipos this
// use case handles (ENTRADA, SAIDA, DEVOLUCAO) — see PermMovimentacaoCreate.
const ActionMovimentacaoCreate audit.Action = "movimentacao.create"

func init() { audit.Register(ActionMovimentacaoCreate) }

// ProdutoReader is the narrow shape RegistrarMovimentacao and
// AjustarEstoque need of infra.ProdutoRepository, to validate produto_id.
type ProdutoReader interface {
	Buscar(ctx context.Context, id string) (domain.Produto, error)
}

// MovimentacaoFinder is the narrow shape RegistrarMovimentacao needs to
// fetch the movimentação a devolução references.
type MovimentacaoFinder interface {
	Buscar(ctx context.Context, id string) (domain.Movimentacao, error)
}

// MovimentacaoCreator is the narrow shape RegistrarMovimentacao needs of
// infra.MovimentacaoRepository to persist.
type MovimentacaoCreator interface {
	Criar(ctx context.Context, m domain.Movimentacao) (domain.Movimentacao, error)
}

// SaldoLocker is the narrow shape RegistrarMovimentacao needs to serialize
// a negative-effect write against the current balance.
type SaldoLocker interface {
	SaldoComLock(ctx context.Context, produtoID string) (int64, error)
}

// MovimentacaoStore is what RegistrarMovimentacao needs to fetch, lock and
// persist a movimentação.
type MovimentacaoStore interface {
	MovimentacaoFinder
	MovimentacaoCreator
	SaldoLocker
}

// RegistrarMovimentacao implements MOV-01/MOV-02/MOV-03: entrada, saída and
// devolução, discriminated by Tipo — same granularity decision as
// financeiro's CriarLancamento for RECEITA/DESPESA (EST-D-005).
type RegistrarMovimentacao struct {
	Authz         Authorizer
	Produtos      ProdutoReader
	Movimentacoes MovimentacaoStore
	Audit         Auditor
	Tx            TxFunc
}

// RegistrarMovimentacaoInput carries what Execute needs. Quantidade is
// always a positive integer — the sign is computed internally from Tipo
// (and, for DEVOLUCAO, from the referenced movimentação's Tipo).
type RegistrarMovimentacaoInput struct {
	Actor            authz.Principal
	Tipo             domain.TipoMovimentacao // Entrada | Saida | Devolucao
	ProdutoID        string
	Quantidade       int64
	Origem           domain.OrigemMovimentacao
	MovimentacaoDeID *string // required when Tipo == Devolucao
}

// Execute checks the permission, validates the produto, computes the
// signed quantidade for the given Tipo (EST-D-008/EST-D-009), then —
// inside one transaction — locks and checks the balance only when the
// effective quantidade is negative, persists and audits.
func (r *RegistrarMovimentacao) Execute(ctx context.Context, in RegistrarMovimentacaoInput) (domain.Movimentacao, error) {
	if r.Authz == nil || r.Produtos == nil || r.Movimentacoes == nil || r.Audit == nil || r.Tx == nil {
		return domain.Movimentacao{}, errNotConfigured
	}
	if err := r.Authz.Require(ctx, in.Actor, PermMovimentacaoCreate); err != nil {
		return domain.Movimentacao{}, err
	}
	if in.Quantidade <= 0 {
		return domain.Movimentacao{}, domain.ErrQuantidadeInvalida
	}
	if _, err := r.Produtos.Buscar(ctx, in.ProdutoID); err != nil {
		return domain.Movimentacao{}, err
	}

	var qtd int64
	var movDeID *string
	switch in.Tipo {
	case domain.Entrada:
		qtd = in.Quantidade
	case domain.Saida:
		qtd = -in.Quantidade
	case domain.Devolucao:
		if in.MovimentacaoDeID == nil {
			return domain.Movimentacao{}, domain.ErrDevolucaoInvalida
		}
		ref, err := r.Movimentacoes.Buscar(ctx, *in.MovimentacaoDeID)
		if err != nil {
			if errors.Is(err, domain.ErrMovimentacaoNaoEncontrada) {
				return domain.Movimentacao{}, domain.ErrDevolucaoInvalida
			}
			return domain.Movimentacao{}, err
		}
		if ref.ProdutoID != in.ProdutoID || (ref.Tipo != domain.Entrada && ref.Tipo != domain.Saida) {
			return domain.Movimentacao{}, domain.ErrDevolucaoInvalida
		}
		if ref.Tipo == domain.Saida {
			qtd = in.Quantidade
		} else {
			qtd = -in.Quantidade
		}
		movDeID = in.MovimentacaoDeID
	default:
		return domain.Movimentacao{}, domain.ErrQuantidadeInvalida
	}

	var criada domain.Movimentacao
	err := r.Tx(ctx, func(ctx context.Context) error {
		if qtd < 0 {
			saldo, err := r.Movimentacoes.SaldoComLock(ctx, in.ProdutoID)
			if err != nil {
				return err
			}
			if saldo+qtd < 0 {
				return domain.ErrSaldoInsuficiente
			}
		}
		var err error
		criada, err = r.Movimentacoes.Criar(ctx, domain.Movimentacao{
			ProdutoID: in.ProdutoID, Tipo: in.Tipo, Quantidade: qtd, Origem: in.Origem,
			MovimentacaoDeID: movDeID, ResponsavelID: in.Actor.UserID,
		})
		if err != nil {
			return err
		}
		return r.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionMovimentacaoCreate,
			EntityType: "movimentacao_estoque", EntityID: criada.ID, Outcome: audit.OutcomeSuccess,
			After: map[string]any{"tipo": string(criada.Tipo), "produto_id": criada.ProdutoID, "quantidade": criada.Quantidade},
		})
	})
	if err != nil {
		return domain.Movimentacao{}, err
	}
	return criada, nil
}
