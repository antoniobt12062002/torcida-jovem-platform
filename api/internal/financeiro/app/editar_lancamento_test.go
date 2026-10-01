//go:build integration

package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func (e env) editarLancamento() *app.EditarLancamento {
	return &app.EditarLancamento{Authz: simpleAuthz{}, Contas: e.contas, Lancamentos: e.lancamentos, Audit: e.rec, Tx: e.tx}
}

func (e env) criarLancamentoDeTeste(t *testing.T, contaID string) domain.Lancamento {
	t.Helper()
	lancamento, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoDespesa, ContaID: contaID,
		ValorBrutoCents: 10000, TaxaCents: 400, FormaPagamento: domain.FormaPix,
	})
	if err != nil {
		t.Fatalf("criar lançamento de teste: %v", err)
	}
	return lancamento
}

// LAN-02 AC1: edita os campos permitidos em CRIADA e recalcula valor_liquido_cents.
func TestEditarLancamentoUpdatesFieldsWhileCriadaAndRecalculatesValorLiquido(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	outraConta := e.criarContaDeTeste(t, domain.TipoDespesa, "Serviços")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)

	editado, err := e.editarLancamento().Execute(context.Background(), app.EditarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoUpdate), ID: lancamento.ID, ContaID: outraConta.ID,
		ValorBrutoCents: 20000, TaxaCents: 1000, FormaPagamento: domain.FormaCartao,
	})
	if err != nil {
		t.Fatalf("editar lançamento: %v", err)
	}
	if editado.ContaID != outraConta.ID || editado.ValorBrutoCents != 20000 || editado.TaxaCents != 1000 ||
		editado.ValorLiquidoCents != 19000 || editado.FormaPagamento != domain.FormaCartao {
		t.Errorf("lançamento editado = %+v", editado)
	}
}

// LAN-02 AC2: fora de CRIADA, recusa com lancamento_imutavel. Como
// 03-workflow-e-saldo (dona da transição CRIADA→PAGA) ainda não existe, o
// teste monta o cenário com um UPDATE direto de fixture (mesma técnica já
// registrada em tasks/02-lancamentos.md para a T4) — não é uma dependência
// de código entre 02 e 03.
func TestEditarLancamentoRejectsEditOutsideCriada(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	if err := e.db.Exec("UPDATE lancamentos SET status = 'PAGA' WHERE id = ?::uuid", lancamento.ID).Error; err != nil {
		t.Fatalf("fixture: marcar PAGA: %v", err)
	}

	_, err := e.editarLancamento().Execute(context.Background(), app.EditarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoUpdate), ID: lancamento.ID, ContaID: conta.ID,
		ValorBrutoCents: 20000, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, domain.ErrLancamentoImutavel) {
		t.Errorf("err = %v, esperado ErrLancamentoImutavel", err)
	}
}

// Consequência técnica de LAN-01 AC2/AC3 preservada na edição (decisão do
// mantenedor): trocar conta_id para uma conta inativa é recusado.
func TestEditarLancamentoRejectsAnInactiveConta(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	outraConta := e.criarContaDeTeste(t, domain.TipoDespesa, "Serviços")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	if err := e.desativarConta().Execute(context.Background(), app.DesativarContaInput{
		Actor: actor(app.PermContaDeactivate), ID: outraConta.ID,
	}); err != nil {
		t.Fatalf("desativar conta: %v", err)
	}

	_, err := e.editarLancamento().Execute(context.Background(), app.EditarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoUpdate), ID: lancamento.ID, ContaID: outraConta.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, domain.ErrContaInvalida) {
		t.Errorf("err = %v, esperado ErrContaInvalida", err)
	}
}

// Idem, tipo incompatível.
func TestEditarLancamentoRejectsATipoIncompatibleConta(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)

	_, err := e.editarLancamento().Execute(context.Background(), app.EditarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoUpdate), ID: lancamento.ID, ContaID: contaReceita.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, domain.ErrLancamentoTipoIncompativel) {
		t.Errorf("err = %v, esperado ErrLancamentoTipoIncompativel", err)
	}
}

// LAN-02 AC3: sem a permissão, recusa sem escrever.
func TestEditarLancamentoRequiresThePermissionAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)

	_, err := e.editarLancamento().Execute(context.Background(), app.EditarLancamentoInput{
		Actor: e.userActor(t), ID: lancamento.ID, ContaID: conta.ID,
		ValorBrutoCents: 99999, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}

	atual, err := e.lancamentos.Buscar(context.Background(), lancamento.ID)
	if err != nil {
		t.Fatalf("buscar: %v", err)
	}
	if atual.ValorBrutoCents != lancamento.ValorBrutoCents {
		t.Errorf("lançamento não deveria ter sido alterado: %+v", atual)
	}
}

// LAN-02 AC4: audita antes/depois de cada campo alterado.
func TestEditarLancamentoAuditsPreviousAndNewValuesOfChangedFields(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)

	_, err := e.editarLancamento().Execute(context.Background(), app.EditarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoUpdate), ID: lancamento.ID, ContaID: conta.ID,
		ValorBrutoCents: 20000, TaxaCents: lancamento.TaxaCents, FormaPagamento: lancamento.FormaPagamento,
	})
	if err != nil {
		t.Fatalf("editar lançamento: %v", err)
	}

	var before, after string
	err = e.db.Raw(`SELECT before::text, after::text FROM audit_log WHERE action = 'lancamento.update' AND entity_id = ?`, lancamento.ID).
		Row().Scan(&before, &after)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if !strings.Contains(before, "10000") || !strings.Contains(after, "20000") {
		t.Errorf("auditoria não confere: before=%q after=%q", before, after)
	}
	if strings.Contains(before, "\"forma_pagamento\"") || strings.Contains(after, "\"forma_pagamento\"") {
		t.Errorf("forma_pagamento não mudou, não deveria estar na auditoria: before=%q after=%q", before, after)
	}
}

// Sem regra de negócio nova aqui, mesmo padrão de T2/T3: um id inexistente
// reusa domain.ErrLancamentoNaoEncontrado.
func TestEditarLancamentoRejectsANonexistentLancamento(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	bogus := "00000000-0000-0000-0000-000000000000"

	_, err := e.editarLancamento().Execute(context.Background(), app.EditarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoUpdate), ID: bogus, ContaID: conta.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, domain.ErrLancamentoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoEncontrado", err)
	}
}
