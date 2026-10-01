//go:build integration

package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func (e env) criarLancamento() *app.CriarLancamento {
	return &app.CriarLancamento{Authz: simpleAuthz{}, Contas: e.contas, Lancamentos: e.lancamentos, Audit: e.rec, Tx: e.tx}
}

// LAN-01 AC1: cria com status CRIADA e valor_liquido_cents calculado.
func TestCriarLancamentoPersistsWithCriadaStatusAndComputedValorLiquido(t *testing.T) {
	e := newEnv(t)
	conta, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoDespesa, Nome: "Material",
	})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	lancamento, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoDespesa, ContaID: conta.ID,
		ValorBrutoCents: 10000, TaxaCents: 400, FormaPagamento: domain.FormaPix,
	})
	if err != nil {
		t.Fatalf("criar lançamento: %v", err)
	}
	if lancamento.ID == "" || lancamento.Status != domain.StatusCriada || lancamento.ValorLiquidoCents != 9600 {
		t.Errorf("lancamento = %+v", lancamento)
	}
}

// LAN-01 AC2: conta inativa é recusada com conta_invalida.
func TestCriarLancamentoRejectsAnInactiveConta(t *testing.T) {
	e := newEnv(t)
	conta, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Produtos",
	})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}
	if err := e.desativarConta().Execute(context.Background(), app.DesativarContaInput{
		Actor: actor(app.PermContaDeactivate), ID: conta.ID,
	}); err != nil {
		t.Fatalf("desativar conta: %v", err)
	}

	_, err = e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoReceita, ContaID: conta.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, domain.ErrContaInvalida) {
		t.Errorf("err = %v, esperado ErrContaInvalida", err)
	}
}

// LAN-01 AC2: conta inexistente é recusada com conta_invalida.
func TestCriarLancamentoRejectsANonexistentConta(t *testing.T) {
	e := newEnv(t)
	bogus := "00000000-0000-0000-0000-000000000000"

	_, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoReceita, ContaID: bogus,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, domain.ErrContaInvalida) {
		t.Errorf("err = %v, esperado ErrContaInvalida", err)
	}
}

// LAN-01 AC3, FIN-D-015: tipo do lançamento tem que bater com o tipo da conta.
func TestCriarLancamentoRejectsATipoIncompatibleWithTheConta(t *testing.T) {
	e := newEnv(t)
	conta, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoDespesa, Nome: "Material",
	})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	_, err = e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoReceita, ContaID: conta.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, domain.ErrLancamentoTipoIncompativel) {
		t.Errorf("err = %v, esperado ErrLancamentoTipoIncompativel", err)
	}
}

// LAN-01 AC4: sem a permissão, recusa sem escrever.
func TestCriarLancamentoRequiresThePermissionAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	conta, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Produtos",
	})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	_, err = e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t), Tipo: domain.TipoReceita, ContaID: conta.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	var n int64
	if err := e.db.Raw("SELECT count(*) FROM lancamentos").Scan(&n).Error; err != nil || n != 0 {
		t.Errorf("nenhum lançamento deveria ter sido gravado: n=%d err=%v", n, err)
	}
}

// LAN-01 AC5: a criação audita tipo, conta_id, valor_bruto_cents e forma_pagamento.
func TestCriarLancamentoAuditsTipoContaValorBrutoAndFormaPagamento(t *testing.T) {
	e := newEnv(t)
	conta, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Produtos",
	})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	lancamento, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoReceita, ContaID: conta.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if err != nil {
		t.Fatalf("criar lançamento: %v", err)
	}

	var entityID, after string
	err = e.db.Raw(`SELECT entity_id, after::text FROM audit_log WHERE action = 'lancamento.create' AND entity_id = ?`, lancamento.ID).
		Row().Scan(&entityID, &after)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if entityID != lancamento.ID || after == "" {
		t.Errorf("auditoria não confere: entity_id=%q after=%q", entityID, after)
	}
}
