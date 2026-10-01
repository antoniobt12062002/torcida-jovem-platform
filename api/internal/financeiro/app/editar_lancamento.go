package app

import (
	"context"
	"errors"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/money"
)

// PermLancamentoUpdate is required to edit a lançamento (LAN-02 AC3).
const PermLancamentoUpdate authz.Permission = "financeiro:lancamento:update"

// ActionLancamentoUpdate is the audit action for an edited lançamento
// (LAN-02 AC4).
const ActionLancamentoUpdate audit.Action = "lancamento.update"

func init() { audit.Register(ActionLancamentoUpdate) }

// LancamentoReader is the narrow shape EditarLancamento needs of
// infra.LancamentoRepository to fetch the lançamento being edited.
type LancamentoReader interface {
	Buscar(ctx context.Context, id string) (domain.Lancamento, error)
}

// LancamentoUpdater is the narrow shape EditarLancamento needs of
// infra.LancamentoRepository to persist the edit.
type LancamentoUpdater interface {
	Atualizar(ctx context.Context, l domain.Lancamento) (domain.Lancamento, error)
}

// LancamentoStore is what EditarLancamento needs to fetch and persist a
// lançamento.
type LancamentoStore interface {
	LancamentoReader
	LancamentoUpdater
}

// EditarLancamento implements LAN-02: edits conta_id, valor_bruto_cents,
// taxa_cents and forma_pagamento, only while status is CRIADA (FIN-D-007).
// It reapplies LAN-01 AC2/AC3's conta validation (exists, ativo, same tipo)
// on every edit — the same invariant FIN-D-015 already requires at creation,
// preserved for the lançamento's whole lifecycle, per the maintainer's
// decision (not a new rule).
type EditarLancamento struct {
	Authz       Authorizer
	Contas      ContaReader
	Lancamentos LancamentoStore
	Audit       Auditor
	Tx          TxFunc
}

// EditarLancamentoInput carries the lançamento's new values for every
// editable field (LAN-02 AC1).
type EditarLancamentoInput struct {
	Actor           authz.Principal
	ID              string
	ContaID         string
	ValorBrutoCents int64
	TaxaCents       int64
	FormaPagamento  domain.FormaPagamento
}

// Execute checks the permission, fetches the lançamento (must be CRIADA),
// validates the new conta the same way CriarLancamento does, recalculates
// valor_liquido_cents, then persists and audits the before/after of every
// changed field in one transaction.
func (e *EditarLancamento) Execute(ctx context.Context, in EditarLancamentoInput) (domain.Lancamento, error) {
	if e.Authz == nil || e.Contas == nil || e.Lancamentos == nil || e.Audit == nil || e.Tx == nil {
		return domain.Lancamento{}, errNotConfigured
	}
	if err := e.Authz.Require(ctx, in.Actor, PermLancamentoUpdate); err != nil {
		return domain.Lancamento{}, err
	}

	atual, err := e.Lancamentos.Buscar(ctx, in.ID)
	if err != nil {
		return domain.Lancamento{}, err
	}
	if atual.Status != domain.StatusCriada {
		return domain.Lancamento{}, domain.ErrLancamentoImutavel
	}

	conta, err := e.Contas.Buscar(ctx, in.ContaID)
	if err != nil {
		if errors.Is(err, domain.ErrContaNaoEncontrada) {
			return domain.Lancamento{}, domain.ErrContaInvalida
		}
		return domain.Lancamento{}, err
	}
	if !conta.Ativo {
		return domain.Lancamento{}, domain.ErrContaInvalida
	}
	if conta.Tipo != domain.TipoConta(atual.Tipo) {
		return domain.Lancamento{}, domain.ErrLancamentoTipoIncompativel
	}

	valorLiquido, err := money.Cents(in.ValorBrutoCents).Sub(money.Cents(in.TaxaCents))
	if err != nil {
		return domain.Lancamento{}, err
	}

	novo := atual
	novo.ContaID = in.ContaID
	novo.ValorBrutoCents = in.ValorBrutoCents
	novo.TaxaCents = in.TaxaCents
	novo.ValorLiquidoCents = int64(valorLiquido)
	novo.FormaPagamento = in.FormaPagamento

	before, after := diffLancamento(atual, novo)

	var editado domain.Lancamento
	err = e.Tx(ctx, func(ctx context.Context) error {
		var err error
		editado, err = e.Lancamentos.Atualizar(ctx, novo)
		if err != nil {
			return err
		}
		return e.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionLancamentoUpdate,
			EntityType: "lancamento", EntityID: editado.ID, Outcome: audit.OutcomeSuccess,
			Before: before, After: after,
		})
	})
	if err != nil {
		return domain.Lancamento{}, err
	}
	return editado, nil
}

// diffLancamento returns the before/after maps of only the fields that
// actually changed (LAN-02 AC4: "the previous and new values of every
// changed field").
func diffLancamento(before, after domain.Lancamento) (map[string]any, map[string]any) {
	prev := map[string]any{}
	next := map[string]any{}
	if before.ContaID != after.ContaID {
		prev["conta_id"], next["conta_id"] = before.ContaID, after.ContaID
	}
	if before.ValorBrutoCents != after.ValorBrutoCents {
		prev["valor_bruto_cents"], next["valor_bruto_cents"] = before.ValorBrutoCents, after.ValorBrutoCents
	}
	if before.TaxaCents != after.TaxaCents {
		prev["taxa_cents"], next["taxa_cents"] = before.TaxaCents, after.TaxaCents
	}
	if before.ValorLiquidoCents != after.ValorLiquidoCents {
		prev["valor_liquido_cents"], next["valor_liquido_cents"] = before.ValorLiquidoCents, after.ValorLiquidoCents
	}
	if before.FormaPagamento != after.FormaPagamento {
		prev["forma_pagamento"], next["forma_pagamento"] = string(before.FormaPagamento), string(after.FormaPagamento)
	}
	return prev, next
}
