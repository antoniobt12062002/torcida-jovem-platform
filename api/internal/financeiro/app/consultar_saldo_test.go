//go:build integration

package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/money"
)

func (e env) consultarSaldo() *app.ConsultarSaldo {
	return &app.ConsultarSaldo{Authz: simpleAuthz{}, Lancamentos: e.lancamentos}
}

// WKF-03 AC1, Independent Test: receita RECEBIDA de 100 + despesa PAGA de 40
// => saldo 60.
func TestConsultarSaldoSumsRecebidaMinusPaga(t *testing.T) {
	e := newEnv(t)
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")

	receita, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoReceita, ContaID: contaReceita.ID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if err != nil {
		t.Fatalf("criar receita: %v", err)
	}
	if _, err := e.receberLancamento().Execute(context.Background(), app.ReceberLancamentoInput{
		Actor: actor(app.PermLancamentoReceive), ID: receita.ID,
	}); err != nil {
		t.Fatalf("receber: %v", err)
	}

	despesa, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoDespesa, ContaID: contaDespesa.ID,
		ValorBrutoCents: 4000, FormaPagamento: domain.FormaPix,
	})
	if err != nil {
		t.Fatalf("criar despesa: %v", err)
	}
	if _, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: despesa.ID,
	}); err != nil {
		t.Fatalf("pagar: %v", err)
	}

	saldo, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermSaldoRead)})
	if err != nil {
		t.Fatalf("consultar saldo: %v", err)
	}
	if saldo != money.Cents(6000) {
		t.Errorf("saldo = %d, esperado 6000 (receita 10000 - despesa 4000)", saldo)
	}
}

// WKF-03 AC2: lançamento CRIADA não altera o saldo.
func TestConsultarSaldoIgnoresCriada(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	e.receitaCriadaDeTeste(t, conta.ID) // permanece CRIADA

	saldo, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermSaldoRead)})
	if err != nil {
		t.Fatalf("consultar saldo: %v", err)
	}
	if saldo != 0 {
		t.Errorf("saldo = %d, esperado 0", saldo)
	}
}

// WKF-03 AC2: lançamento CANCELADA (direto de CRIADA) não altera o saldo.
func TestConsultarSaldoIgnoresCancelled(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	if _, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: lancamento.ID, Motivo: "erro",
	}); err != nil {
		t.Fatalf("cancelar: %v", err)
	}

	saldo, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermSaldoRead)})
	if err != nil {
		t.Fatalf("consultar saldo: %v", err)
	}
	if saldo != 0 {
		t.Errorf("saldo = %d, esperado 0", saldo)
	}
}

// WKF-03, FIN-D-006: receita RECEBIDA e depois CANCELADA deixa de compor o saldo.
func TestConsultarSaldoExcludesAReceitaCancelledAfterBeingRecebida(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	receita := e.receitaRecebidaDeTeste(t, conta.ID)

	saldoAntes, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermSaldoRead)})
	if err != nil {
		t.Fatalf("consultar saldo antes: %v", err)
	}
	if saldoAntes != money.Cents(receita.ValorLiquidoCents) {
		t.Fatalf("saldo antes = %d, esperado %d", saldoAntes, receita.ValorLiquidoCents)
	}

	if _, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: receita.ID, Motivo: "recebimento incorreto",
	}); err != nil {
		t.Fatalf("cancelar: %v", err)
	}

	saldoDepois, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermSaldoRead)})
	if err != nil {
		t.Fatalf("consultar saldo depois: %v", err)
	}
	if saldoDepois != 0 {
		t.Errorf("saldo depois = %d, esperado 0 — a receita cancelada não deveria mais compor o saldo", saldoDepois)
	}
}

// WKF-03, FIN-D-006: despesa PAGA e depois CANCELADA deixa de compor o saldo.
func TestConsultarSaldoExcludesADespesaCancelledAfterBeingPaga(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	despesa := e.despesaPagaDeTeste(t, conta.ID)

	saldoAntes, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermSaldoRead)})
	if err != nil {
		t.Fatalf("consultar saldo antes: %v", err)
	}
	if saldoAntes != money.Cents(-despesa.ValorLiquidoCents) {
		t.Fatalf("saldo antes = %d, esperado %d", saldoAntes, -despesa.ValorLiquidoCents)
	}

	if _, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: despesa.ID, Motivo: "pagamento incorreto",
	}); err != nil {
		t.Fatalf("cancelar: %v", err)
	}

	saldoDepois, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermSaldoRead)})
	if err != nil {
		t.Fatalf("consultar saldo depois: %v", err)
	}
	if saldoDepois != 0 {
		t.Errorf("saldo depois = %d, esperado 0 — a despesa cancelada não deveria mais compor o saldo", saldoDepois)
	}
}

// WKF-03 AC3: sem financeiro:saldo:read, recusa sem retornar valor. Não
// reaproveita financeiro:lancamento:read (FIN-D-013).
func TestConsultarSaldoRequiresThePermission(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	e.receitaRecebidaDeTeste(t, conta.ID)

	_, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermLancamentoRead)})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden (financeiro:lancamento:read não deveria bastar)", err)
	}
}

// Conjunto vazio: nenhum lançamento => saldo zero.
func TestConsultarSaldoReturnsZeroWithNoLancamentos(t *testing.T) {
	e := newEnv(t)

	saldo, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actor(app.PermSaldoRead)})
	if err != nil {
		t.Fatalf("consultar saldo: %v", err)
	}
	if saldo != 0 {
		t.Errorf("saldo = %d, esperado 0", saldo)
	}
}
