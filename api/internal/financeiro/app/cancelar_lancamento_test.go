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

func (e env) cancelarLancamento() *app.CancelarLancamento {
	return &app.CancelarLancamento{Authz: simpleAuthz{}, Lancamentos: e.lancamentos, Audit: e.rec, Tx: e.tx}
}

// despesaPagaDeTeste creates a DESPESA lançamento and marks it PAGA (real
// transition, via PagarLancamento).
func (e env) despesaPagaDeTeste(t *testing.T, contaID string) domain.Lancamento {
	t.Helper()
	despesa := e.criarLancamentoDeTeste(t, contaID)
	paga, err := e.pagarLancamento().Execute(context.Background(), app.PagarLancamentoInput{
		Actor: actor(app.PermLancamentoPay), ID: despesa.ID,
	})
	if err != nil {
		t.Fatalf("pagar despesa de teste: %v", err)
	}
	return paga
}

// WKF-02 AC1: cancela a partir de CRIADA.
func TestCancelarLancamentoCancelsFromCriada(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)

	cancelado, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: lancamento.ID, Motivo: "erro de digitação",
	})
	if err != nil {
		t.Fatalf("cancelar: %v", err)
	}
	if cancelado.Status != domain.StatusCancelada {
		t.Errorf("status = %v, esperado CANCELADA", cancelado.Status)
	}
}

// WKF-02 AC1: cancela a partir de RECEBIDA.
func TestCancelarLancamentoCancelsFromRecebida(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	receita := e.receitaRecebidaDeTeste(t, conta.ID)

	cancelado, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: receita.ID, Motivo: "recebimento incorreto",
	})
	if err != nil {
		t.Fatalf("cancelar: %v", err)
	}
	if cancelado.Status != domain.StatusCancelada {
		t.Errorf("status = %v, esperado CANCELADA", cancelado.Status)
	}
}

// WKF-02 AC1: cancela a partir de PAGA.
func TestCancelarLancamentoCancelsFromPaga(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	despesa := e.despesaPagaDeTeste(t, conta.ID)

	cancelado, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: despesa.ID, Motivo: "pagamento incorreto",
	})
	if err != nil {
		t.Fatalf("cancelar: %v", err)
	}
	if cancelado.Status != domain.StatusCancelada {
		t.Errorf("status = %v, esperado CANCELADA", cancelado.Status)
	}
}

// WKF-02 AC3: já CANCELADA é recusado.
func TestCancelarLancamentoRejectsAnAlreadyCancelledLancamento(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	if _, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: lancamento.ID, Motivo: "primeiro cancelamento",
	}); err != nil {
		t.Fatalf("primeiro cancelamento: %v", err)
	}

	_, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: lancamento.ID, Motivo: "segunda tentativa",
	})
	if !errors.Is(err, domain.ErrLancamentoJaCancelado) {
		t.Errorf("err = %v, esperado ErrLancamentoJaCancelado", err)
	}
}

// WKF-02 AC2: motivo vazio é recusado, sem escrever.
func TestCancelarLancamentoRejectsAnEmptyMotivoAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)

	_, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: lancamento.ID, Motivo: "   ",
	})
	if !errors.Is(err, domain.ErrMotivoObrigatorio) {
		t.Errorf("err = %v, esperado ErrMotivoObrigatorio", err)
	}

	atual, err := e.lancamentos.Buscar(context.Background(), lancamento.ID)
	if err != nil || atual.Status != domain.StatusCriada {
		t.Errorf("status não deveria ter mudado: %+v, err=%v", atual, err)
	}
}

// WKF-02 AC5: sem a permissão, recusa sem escrever.
func TestCancelarLancamentoRequiresThePermissionAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)

	_, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t), ID: lancamento.ID, Motivo: "motivo válido",
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}

	atual, err := e.lancamentos.Buscar(context.Background(), lancamento.ID)
	if err != nil || atual.Status != domain.StatusCriada {
		t.Errorf("status não deveria ter mudado: %+v, err=%v", atual, err)
	}
}

// WKF-02: persiste motivo_cancelamento, cancelado_por e cancelado_em.
func TestCancelarLancamentoPersistsMotivoCanceladoPorAndCanceladoEm(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	ator := e.userActor(t, app.PermLancamentoCancel)

	cancelado, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: ator, ID: lancamento.ID, Motivo: "erro de digitação",
	})
	if err != nil {
		t.Fatalf("cancelar: %v", err)
	}
	if cancelado.MotivoCancelamento == nil || *cancelado.MotivoCancelamento != "erro de digitação" {
		t.Errorf("MotivoCancelamento = %v, esperado %q", cancelado.MotivoCancelamento, "erro de digitação")
	}
	if cancelado.CanceladoPor == nil || *cancelado.CanceladoPor != ator.UserID {
		t.Errorf("CanceladoPor = %v, esperado %q", cancelado.CanceladoPor, ator.UserID)
	}
	if cancelado.CanceladoEm == nil {
		t.Error("CanceladoEm não deveria ser nil")
	}
}

// WKF-02 AC6: audita como lancamento.cancel, com o motivo no Reason.
func TestCancelarLancamentoAuditsAsLancamentoCancelWithReason(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)

	cancelado, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: lancamento.ID, Motivo: "erro de digitação",
	})
	if err != nil {
		t.Fatalf("cancelar: %v", err)
	}

	var entityID, reason string
	err = e.db.Raw(`SELECT entity_id, reason FROM audit_log WHERE action = 'lancamento.cancel' AND entity_id = ?`, cancelado.ID).
		Row().Scan(&entityID, &reason)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if entityID != cancelado.ID || reason != "erro de digitação" {
		t.Errorf("auditoria não confere: entity_id=%q reason=%q", entityID, reason)
	}
}

// WKF-02 AC4: nunca escreve em outro lançamento.
func TestCancelarLancamentoNeverModifiesAnotherLancamento(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	outro := e.criarLancamentoDeTeste(t, conta.ID)

	if _, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: lancamento.ID, Motivo: "erro de digitação",
	}); err != nil {
		t.Fatalf("cancelar: %v", err)
	}

	depois, err := e.lancamentos.Buscar(context.Background(), outro.ID)
	if err != nil {
		t.Fatalf("buscar outro lançamento: %v", err)
	}
	if depois.Status != domain.StatusCriada || depois.MotivoCancelamento != nil {
		t.Errorf("outro lançamento foi alterado: %+v", depois)
	}
}

// Sem regra de negócio nova: um id inexistente reusa
// domain.ErrLancamentoNaoEncontrado, o mesmo padrão já usado nas tarefas anteriores.
func TestCancelarLancamentoRejectsANonexistentLancamento(t *testing.T) {
	e := newEnv(t)
	bogus := "00000000-0000-0000-0000-000000000000"

	_, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: bogus, Motivo: "motivo válido",
	})
	if !errors.Is(err, domain.ErrLancamentoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoEncontrado", err)
	}
}
