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

func (e env) receberLancamento() *app.ReceberLancamento {
	return &app.ReceberLancamento{Authz: simpleAuthz{}, Lancamentos: e.lancamentos, Audit: e.rec, Tx: e.tx}
}

// receitaCriadaDeTeste creates a RECEITA lançamento, status CRIADA.
func (e env) receitaCriadaDeTeste(t *testing.T, contaID string) domain.Lancamento {
	t.Helper()
	receita, err := e.criarLancamento().Execute(context.Background(), app.CriarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), Tipo: domain.TipoReceita, ContaID: contaID,
		ValorBrutoCents: 10000, FormaPagamento: domain.FormaPix,
	})
	if err != nil {
		t.Fatalf("criar receita de teste: %v", err)
	}
	return receita
}

// WKF-01 AC1: receita CRIADA transiciona para RECEBIDA.
func TestReceberLancamentoTransitionsCriadaReceitaToRecebida(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	receita := e.receitaCriadaDeTeste(t, conta.ID)

	recebida, err := e.receberLancamento().Execute(context.Background(), app.ReceberLancamentoInput{
		Actor: actor(app.PermLancamentoReceive), ID: receita.ID,
	})
	if err != nil {
		t.Fatalf("receber: %v", err)
	}
	if recebida.Status != domain.StatusRecebida {
		t.Errorf("status = %v, esperado RECEBIDA", recebida.Status)
	}
}

// WKF-01 AC2, Independent Test: marcar de novo é recusado.
func TestReceberLancamentoRejectsAnAlreadyRecebidaLancamento(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	receita := e.receitaCriadaDeTeste(t, conta.ID)
	if _, err := e.receberLancamento().Execute(context.Background(), app.ReceberLancamentoInput{
		Actor: actor(app.PermLancamentoReceive), ID: receita.ID,
	}); err != nil {
		t.Fatalf("primeiro recebimento: %v", err)
	}

	_, err := e.receberLancamento().Execute(context.Background(), app.ReceberLancamentoInput{
		Actor: actor(app.PermLancamentoReceive), ID: receita.ID,
	})
	if !errors.Is(err, domain.ErrLancamentoNaoPodeSerRecebido) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoPodeSerRecebido", err)
	}
}

// Edge Case (spec): tentar receber uma DESPESA é recusado estruturalmente.
func TestReceberLancamentoRejectsADespesaLancamento(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	despesa := e.criarLancamentoDeTeste(t, conta.ID)

	_, err := e.receberLancamento().Execute(context.Background(), app.ReceberLancamentoInput{
		Actor: actor(app.PermLancamentoReceive), ID: despesa.ID,
	})
	if !errors.Is(err, domain.ErrLancamentoNaoPodeSerRecebido) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoPodeSerRecebido", err)
	}
}

// WKF-01 AC5: sem a permissão, recusa sem escrever.
func TestReceberLancamentoRequiresThePermissionAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	receita := e.receitaCriadaDeTeste(t, conta.ID)

	_, err := e.receberLancamento().Execute(context.Background(), app.ReceberLancamentoInput{
		Actor: actor(), ID: receita.ID,
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}

	atual, err := e.lancamentos.Buscar(context.Background(), receita.ID)
	if err != nil || atual.Status != domain.StatusCriada {
		t.Errorf("status não deveria ter mudado: %+v, err=%v", atual, err)
	}
}

// WKF-01 AC6: o recebimento audita como lancamento.receive.
func TestReceberLancamentoAuditsAsLancamentoReceive(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	receita := e.receitaCriadaDeTeste(t, conta.ID)

	recebida, err := e.receberLancamento().Execute(context.Background(), app.ReceberLancamentoInput{
		Actor: actor(app.PermLancamentoReceive), ID: receita.ID,
	})
	if err != nil {
		t.Fatalf("receber: %v", err)
	}

	var entityID string
	err = e.db.Raw(`SELECT entity_id FROM audit_log WHERE action = 'lancamento.receive' AND entity_id = ?`, recebida.ID).
		Row().Scan(&entityID)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if entityID != recebida.ID {
		t.Errorf("entity_id = %q, esperado %q", entityID, recebida.ID)
	}
}

// Sem regra de negócio nova: um id inexistente reusa
// domain.ErrLancamentoNaoEncontrado, o mesmo padrão já usado em T2/T3/T4.
func TestReceberLancamentoRejectsANonexistentLancamento(t *testing.T) {
	e := newEnv(t)
	bogus := "00000000-0000-0000-0000-000000000000"

	_, err := e.receberLancamento().Execute(context.Background(), app.ReceberLancamentoInput{
		Actor: actor(app.PermLancamentoReceive), ID: bogus,
	})
	if !errors.Is(err, domain.ErrLancamentoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoEncontrado", err)
	}
}
