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

func (e env) pagarLancamento() *app.PagarLancamento {
	return &app.PagarLancamento{Authz: simpleAuthz{}, Lancamentos: e.lancamentos, Audit: e.rec, Tx: e.tx}
}

// WKF-01 AC3: despesa CRIADA transiciona para PAGA.
func TestPagarLancamentoTransitionsCriadaDespesaToPaga(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	despesa := e.criarLancamentoDeTeste(t, conta.ID)

	paga, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: despesa.ID,
	})
	if err != nil {
		t.Fatalf("pagar: %v", err)
	}
	if paga.Status != domain.StatusPaga {
		t.Errorf("status = %v, esperado PAGA", paga.Status)
	}
}

// WKF-01 AC4, Independent Test: pagar de novo é recusado.
func TestPagarLancamentoRejectsAnAlreadyPagaLancamento(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	despesa := e.criarLancamentoDeTeste(t, conta.ID)
	if _, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: despesa.ID,
	}); err != nil {
		t.Fatalf("primeiro pagamento: %v", err)
	}

	_, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: despesa.ID,
	})
	if !errors.Is(err, domain.ErrLancamentoNaoPodeSerPago) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoPodeSerPago", err)
	}
}

// Edge Case (spec): tentar pagar uma RECEITA é recusado estruturalmente.
func TestPagarLancamentoRejectsAReceitaLancamento(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	receita := e.receitaCriadaDeTeste(t, conta.ID)

	_, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: receita.ID,
	})
	if !errors.Is(err, domain.ErrLancamentoNaoPodeSerPago) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoPodeSerPago", err)
	}
}

// WKF-01 AC4: status diferente de CRIADA (aqui, CANCELADA) é recusado.
func TestPagarLancamentoRejectsANonCriadaStatus(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	despesa := e.criarLancamentoDeTeste(t, conta.ID)
	if err := e.db.Exec("UPDATE lancamentos SET status = 'CANCELADA' WHERE id = ?::uuid", despesa.ID).Error; err != nil {
		t.Fatalf("fixture: marcar CANCELADA: %v", err)
	}

	_, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: despesa.ID,
	})
	if !errors.Is(err, domain.ErrLancamentoNaoPodeSerPago) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoPodeSerPago", err)
	}
}

// WKF-01 AC5: sem a permissão, recusa sem escrever.
func TestPagarLancamentoRequiresThePermissionAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	despesa := e.criarLancamentoDeTeste(t, conta.ID)

	_, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(), ID: despesa.ID,
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}

	atual, err := e.lancamentos.Buscar(context.Background(), despesa.ID)
	if err != nil || atual.Status != domain.StatusCriada {
		t.Errorf("status não deveria ter mudado: %+v, err=%v", atual, err)
	}
}

// WKF-01 AC6: o pagamento audita como lancamento.pay.
func TestPagarLancamentoAuditsAsLancamentoPay(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	despesa := e.criarLancamentoDeTeste(t, conta.ID)

	paga, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: despesa.ID,
	})
	if err != nil {
		t.Fatalf("pagar: %v", err)
	}

	var entityID string
	err = e.db.Raw(`SELECT entity_id FROM audit_log WHERE action = 'lancamento.pay' AND entity_id = ?`, paga.ID).
		Row().Scan(&entityID)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if entityID != paga.ID {
		t.Errorf("entity_id = %q, esperado %q", entityID, paga.ID)
	}
}

// Sem regra de negócio nova: um id inexistente reusa
// domain.ErrLancamentoNaoEncontrado, o mesmo padrão já usado nas tarefas anteriores.
func TestPagarLancamentoRejectsANonexistentLancamento(t *testing.T) {
	e := newEnv(t)
	bogus := "00000000-0000-0000-0000-000000000000"

	_, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: bogus,
	})
	if !errors.Is(err, domain.ErrLancamentoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoEncontrado", err)
	}
}
