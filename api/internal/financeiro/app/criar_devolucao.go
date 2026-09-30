package app

import (
	"context"
	"errors"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/money"
)

// LancamentoReaderCreator is what CriarDevolucao needs of
// infra.LancamentoRepository: LancamentoReader to fetch the receita being
// returned, LancamentoCreator to persist the despesa.
type LancamentoReaderCreator interface {
	LancamentoReader
	LancamentoCreator
}

// CriarDevolucao implements LAN-03: creates a despesa linked to a RECEITA
// RECEBIDA it returns (FIN-D-003), unidirectional — the receita is never
// written. It reuses PermLancamentoCreate and ActionLancamentoCreate: a
// devolução is not a distinct business operation (design/02-lancamentos.md;
// the maintainer's decision on this exact point, T4's authorization).
type CriarDevolucao struct {
	Authz       Authorizer
	Contas      ContaReader
	Lancamentos LancamentoReaderCreator
	Audit       Auditor
	Tx          TxFunc
}

// CriarDevolucaoInput carries what Execute needs. Tipo is always DESPESA
// (FIN-D-003: the field can only be filled on a despesa) — not part of the
// input.
type CriarDevolucaoInput struct {
	Actor           authz.Principal
	ContaID         string
	ValorBrutoCents int64
	TaxaCents       int64
	FormaPagamento  domain.FormaPagamento
	DevolucaoDeID   string
}

// Execute checks the permission, validates the despesa's own conta the same
// way CriarLancamento does, validates the referenced lançamento (must be
// RECEITA RECEBIDA), then persists and audits the devolução in one
// transaction. It never touches the referenced lançamento's row.
func (c *CriarDevolucao) Execute(ctx context.Context, in CriarDevolucaoInput) (domain.Lancamento, error) {
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
	if conta.Tipo != domain.TipoDespesa {
		return domain.Lancamento{}, domain.ErrLancamentoTipoIncompativel
	}

	referenciado, err := c.Lancamentos.Buscar(ctx, in.DevolucaoDeID)
	if err != nil {
		if errors.Is(err, domain.ErrLancamentoNaoEncontrado) {
			return domain.Lancamento{}, domain.ErrDevolucaoInvalida
		}
		return domain.Lancamento{}, err
	}
	if referenciado.Tipo != domain.TipoReceita || referenciado.Status != domain.StatusRecebida {
		return domain.Lancamento{}, domain.ErrDevolucaoInvalida
	}

	valorLiquido, err := money.Cents(in.ValorBrutoCents).Sub(money.Cents(in.TaxaCents))
	if err != nil {
		return domain.Lancamento{}, err
	}

	devolucaoDeID := in.DevolucaoDeID
	lancamento := domain.Lancamento{
		Tipo: domain.TipoDespesa, ContaID: in.ContaID, ValorBrutoCents: in.ValorBrutoCents,
		TaxaCents: in.TaxaCents, ValorLiquidoCents: int64(valorLiquido),
		FormaPagamento: in.FormaPagamento, DevolucaoDeID: &devolucaoDeID, CriadoPor: in.Actor.UserID,
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
				"devolucao_de_id": devolucaoDeID,
			},
		})
	})
	if err != nil {
		return domain.Lancamento{}, err
	}
	return criado, nil
}
