package app

import (
	"context"
	"errors"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/money"
)

// PermLancamentoCreate is required to create a lançamento (LAN-01 AC4). Also
// used by CriarDevolucao (LAN-03): devolução is not a distinct business
// operation (FIN-D-003).
const PermLancamentoCreate authz.Permission = "financeiro:lancamento:create"

// ActionLancamentoCreate is the audit action for a created lançamento (LAN-01
// AC5). Reused, not duplicated, by CriarDevolucao — see PermLancamentoCreate.
const ActionLancamentoCreate audit.Action = "lancamento.create"

func init() { audit.Register(ActionLancamentoCreate) }

// ContaReader is the narrow shape CriarLancamento needs of
// infra.ContaRepository, to validate conta_id (LAN-01 AC2/AC3): existence,
// ativo and tipo.
type ContaReader interface {
	Buscar(ctx context.Context, id string) (domain.Conta, error)
}

// LancamentoCreator is the narrow shape CriarLancamento needs of
// infra.LancamentoRepository.
type LancamentoCreator interface {
	Criar(ctx context.Context, l domain.Lancamento) (domain.Lancamento, error)
}

// CriarLancamento implements LAN-01: creates a receita or despesa, validating
// its conta and computing valor_liquido_cents once.
type CriarLancamento struct {
	Authz       Authorizer
	Contas      ContaReader
	Lancamentos LancamentoCreator
	Audit       Auditor
	Tx          TxFunc
}

// CriarLancamentoInput carries what Execute needs. CriadoPor is Actor.UserID
// (the lançamento's criado_por is a real foreign key to users).
type CriarLancamentoInput struct {
	Actor           authz.Principal
	Tipo            domain.TipoLancamento
	ContaID         string
	ValorBrutoCents int64
	TaxaCents       int64
	FormaPagamento  domain.FormaPagamento
}

// Execute checks the permission, validates the conta (exists, ativo, same
// tipo), computes valor_liquido_cents, then persists and audits the
// lançamento in one transaction.
func (c *CriarLancamento) Execute(ctx context.Context, in CriarLancamentoInput) (domain.Lancamento, error) {
	if c.Authz == nil || c.Contas == nil || c.Lancamentos == nil || c.Audit == nil || c.Tx == nil {
		return domain.Lancamento{}, errNotConfigured
	}
	if err := c.Authz.Require(ctx, in.Actor, PermLancamentoCreate); err != nil {
		return domain.Lancamento{}, err
	}

	conta, err := c.Contas.Buscar(ctx, in.ContaID)
	if err != nil {
		if errors.Is(err, domain.ErrContaNaoEncontrada) {
			return domain.Lancamento{}, domain.ErrContaInvalida
		}
		return domain.Lancamento{}, err
	}
	if !conta.Ativo {
		return domain.Lancamento{}, domain.ErrContaInvalida
	}
	if conta.Tipo != domain.TipoConta(in.Tipo) {
		return domain.Lancamento{}, domain.ErrLancamentoTipoIncompativel
	}

	valorLiquido, err := money.Cents(in.ValorBrutoCents).Sub(money.Cents(in.TaxaCents))
	if err != nil {
		return domain.Lancamento{}, err
	}

	lancamento := domain.Lancamento{
		Tipo: in.Tipo, ContaID: in.ContaID, ValorBrutoCents: in.ValorBrutoCents,
		TaxaCents: in.TaxaCents, ValorLiquidoCents: int64(valorLiquido),
		FormaPagamento: in.FormaPagamento, CriadoPor: in.Actor.UserID,
	}

	var criado domain.Lancamento
	err = c.Tx(ctx, func(ctx context.Context) error {
		var err error
		criado, err = c.Lancamentos.Criar(ctx, lancamento)
		if err != nil {
			return err
		}
		return c.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionLancamentoCreate,
			EntityType: "lancamento", EntityID: criado.ID, Outcome: audit.OutcomeSuccess,
			After: map[string]any{
				"tipo": string(criado.Tipo), "conta_id": criado.ContaID,
				"valor_bruto_cents": criado.ValorBrutoCents, "forma_pagamento": string(criado.FormaPagamento),
			},
		})
	})
	if err != nil {
		return domain.Lancamento{}, err
	}
	return criado, nil
}
