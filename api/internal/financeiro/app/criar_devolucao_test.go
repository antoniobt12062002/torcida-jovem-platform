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

func (e env) criarDevolucao() *app.CriarDevolucao {
	return &app.CriarDevolucao{Authz: simpleAuthz{}, Contas: e.contas, Lancamentos: e.lancamentos, Audit: e.rec, Tx: e.tx}
}

// receitaRecebidaDeTeste creates a RECEITA lançamento and marks it RECEBIDA
// via fixture (see (env).marcarRecebida).
func (e env) receitaRecebidaDeTeste(t *testing.T, contaReceita string) domain.Lancamento {
	t.Helper()
	receita, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoReceita, ContaID: contaReceita,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if err != nil {
		t.Fatalf("criar receita de teste: %v", err)
	}
	e.marcarRecebida(t, receita.ID)
	receita.Status = domain.StatusRecebida
	return receita
}

// LAN-03 AC1: cria como despesa vinculada à receita original.
func TestCriarDevolucaoPersistsAsDespesaLinkedToTheReceitaOriginal(t *testing.T) {
	e := newEnv(t)
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	receita := e.receitaRecebidaDeTeste(t, contaReceita.ID)

	devolucao, err := e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), ContaID: contaDespesa.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix, DevolucaoDeID: receita.ID,
	})
	if err != nil {
		t.Fatalf("criar devolução: %v", err)
	}
	if devolucao.Tipo != domain.TipoDespesa || devolucao.Status != domain.StatusCriada ||
		devolucao.DevolucaoDeID == nil || *devolucao.DevolucaoDeID != receita.ID {
		t.Errorf("devolução = %+v", devolucao)
	}
}

// LAN-03 AC2: referenciado que não é RECEITA é recusado, mesmo com status
// RECEBIDA — isola a checagem de tipo da checagem de status (a próxima
// abaixo). Nada hoje impede uma DESPESA de chegar a RECEBIDA no schema (essa
// correlação é de 03-workflow-e-saldo, ainda inexistente); a fixture só
// força esse estado para provar que CriarDevolucao rejeita pelo tipo, não
// por acidente de status.
func TestCriarDevolucaoRequiresTheReferencedToBeReceita(t *testing.T) {
	e := newEnv(t)
	contaDespesaRef := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	outraDespesa := e.criarLancamentoDeTeste(t, contaDespesaRef.ID)
	e.marcarRecebida(t, outraDespesa.ID)

	_, err := e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), ContaID: contaDespesa.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix, DevolucaoDeID: outraDespesa.ID,
	})
	if !errors.Is(err, domain.ErrDevolucaoInvalida) {
		t.Errorf("err = %v, esperado ErrDevolucaoInvalida", err)
	}
}

// LAN-03 AC2: receita referenciada que ainda não está RECEBIDA é recusada.
func TestCriarDevolucaoRequiresTheReferencedToBeRecebida(t *testing.T) {
	e := newEnv(t)
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	receitaCriada, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoReceita, ContaID: contaReceita.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if err != nil {
		t.Fatalf("criar receita: %v", err)
	}

	_, err = e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), ContaID: contaDespesa.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix, DevolucaoDeID: receitaCriada.ID,
	})
	if !errors.Is(err, domain.ErrDevolucaoInvalida) {
		t.Errorf("err = %v, esperado ErrDevolucaoInvalida", err)
	}
}

// LAN-03 AC3: a receita original nunca é escrita.
func TestCriarDevolucaoNeverModifiesTheOriginalReceita(t *testing.T) {
	e := newEnv(t)
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	receita := e.receitaRecebidaDeTeste(t, contaReceita.ID)

	if _, err := e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), ContaID: contaDespesa.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix, DevolucaoDeID: receita.ID,
	}); err != nil {
		t.Fatalf("criar devolução: %v", err)
	}

	depois, err := e.lancamentos.Buscar(context.Background(), receita.ID)
	if err != nil {
		t.Fatalf("buscar receita: %v", err)
	}
	if depois.Status != domain.StatusRecebida || depois.ValorBrutoCents != receita.ValorBrutoCents ||
		depois.ContaID != receita.ContaID || depois.DevolucaoDeID != nil {
		t.Errorf("receita original foi alterada: antes=%+v depois=%+v", receita, depois)
	}
}

// LAN-03: reusa financeiro:lancamento:create — sem ela, recusa sem escrever.
func TestCriarDevolucaoUsesThePermissionOfLancamentoCreate(t *testing.T) {
	e := newEnv(t)
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	receita := e.receitaRecebidaDeTeste(t, contaReceita.ID)

	_, err := e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t), ContaID: contaDespesa.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix, DevolucaoDeID: receita.ID,
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	var n int64
	if err := e.db.Raw("SELECT count(*) FROM lancamentos WHERE devolucao_de_id = ?::uuid", receita.ID).Scan(&n).Error; err != nil || n != 0 {
		t.Errorf("nenhuma devolução deveria ter sido gravada: n=%d err=%v", n, err)
	}
}

// Sem regra de negócio nova: um devolucao_de_id inexistente reusa
// ErrDevolucaoInvalida (mesmo padrão de conta_invalida em T2).
func TestCriarDevolucaoRejectsANonexistentReferencedLancamento(t *testing.T) {
	e := newEnv(t)
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	bogus := "00000000-0000-0000-0000-000000000000"

	_, err := e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), ContaID: contaDespesa.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix, DevolucaoDeID: bogus,
	})
	if !errors.Is(err, domain.ErrDevolucaoInvalida) {
		t.Errorf("err = %v, esperado ErrDevolucaoInvalida", err)
	}
}

// A devolução audita como lancamento.create (LAN-01 AC5, reusada — decisão
// do mantenedor), com devolucao_de_id no payload para diferenciá-la.
func TestCriarDevolucaoAuditsAsLancamentoCreateWithDevolucaoDeIDInPayload(t *testing.T) {
	e := newEnv(t)
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	receita := e.receitaRecebidaDeTeste(t, contaReceita.ID)

	devolucao, err := e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), ContaID: contaDespesa.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix, DevolucaoDeID: receita.ID,
	})
	if err != nil {
		t.Fatalf("criar devolução: %v", err)
	}

	var entityID, after string
	err = e.db.Raw(`SELECT entity_id, after::text FROM audit_log WHERE action = 'lancamento.create' AND entity_id = ?`, devolucao.ID).
		Row().Scan(&entityID, &after)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if entityID != devolucao.ID || !strings.Contains(after, receita.ID) {
		t.Errorf("auditoria não confere: entity_id=%q after=%q (esperava conter %q)", entityID, after, receita.ID)
	}
}

// A devolução é sempre DESPESA: reusa a validação de conta de LAN-01
// AC2/AC3 (mesmo tipo, ativa) para a conta da despesa em si.
func TestCriarDevolucaoRejectsAnInactiveConta(t *testing.T) {
	e := newEnv(t)
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	receita := e.receitaRecebidaDeTeste(t, contaReceita.ID)
	if err := e.desativarConta().Execute(context.Background(), app.DesativarContaInput{
		Actor: actor(app.PermContaDeactivate), ID: contaDespesa.ID,
	}); err != nil {
		t.Fatalf("desativar conta: %v", err)
	}

	_, err := e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), ContaID: contaDespesa.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix, DevolucaoDeID: receita.ID,
	})
	if !errors.Is(err, domain.ErrContaInvalida) {
		t.Errorf("err = %v, esperado ErrContaInvalida", err)
	}
}
