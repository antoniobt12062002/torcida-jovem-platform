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

func (e env) listarLancamentos() *app.ListarLancamentos {
	return &app.ListarLancamentos{Authz: simpleAuthz{}, Lancamentos: e.lancamentos}
}

// LAN-04 AC1: lista com todos os campos, incluindo status e devolucao_de_id
// quando presente.
func TestListarLancamentosReturnsEveryFieldIncludingStatusAndDevolucaoDeID(t *testing.T) {
	e := newEnv(t)
	contaReceita := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")
	contaDespesa := e.criarContaDeTeste(t, domain.TipoDespesa, "Devoluções")
	receita := e.receitaRecebidaDeTeste(t, contaReceita.ID)
	devolucao, err := e.criarDevolucao().Execute(context.Background(), app.CriarDevolucaoInput{
		Actor: e.userActor(t, app.PermLancamentoCreate), ContaID: contaDespesa.ID,
		ValorBrutoCents: 5000, FormaPagamento: domain.FormaPix, DevolucaoDeID: receita.ID,
	})
	if err != nil {
		t.Fatalf("criar devolução: %v", err)
	}

	lancamentos, err := e.listarLancamentos().Execute(context.Background(), app.ListarLancamentosInput{Actor: actor(app.PermLancamentoRead)})
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	var achouReceita, achouDevolucao bool
	for _, l := range lancamentos {
		if l.ID == receita.ID {
			achouReceita = true
			if l.Status != domain.StatusRecebida || l.DevolucaoDeID != nil {
				t.Errorf("receita = %+v", l)
			}
		}
		if l.ID == devolucao.ID {
			achouDevolucao = true
			if l.DevolucaoDeID == nil || *l.DevolucaoDeID != receita.ID || l.ValorBrutoCents != 5000 {
				t.Errorf("devolução = %+v", l)
			}
		}
	}
	if !achouReceita || !achouDevolucao {
		t.Errorf("esperava ver os dois lançamentos na listagem, veio %+v", lancamentos)
	}
}

// LAN-04 Independent Test: cancelados não são ocultados por padrão. Fixture
// via SQL direto porque 03-workflow-e-saldo (dona da transição de
// cancelamento) ainda não existe.
func TestListarLancamentosIncludesCancelled(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	if err := e.db.Exec("UPDATE lancamentos SET status = 'CANCELADA' WHERE id = ?::uuid", lancamento.ID).Error; err != nil {
		t.Fatalf("fixture: marcar CANCELADA: %v", err)
	}

	lancamentos, err := e.listarLancamentos().Execute(context.Background(), app.ListarLancamentosInput{Actor: actor(app.PermLancamentoRead)})
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	var achou bool
	for _, l := range lancamentos {
		if l.ID == lancamento.ID {
			achou = true
			if l.Status != domain.StatusCancelada {
				t.Errorf("status = %v, esperado CANCELADA", l.Status)
			}
		}
	}
	if !achou {
		t.Errorf("esperava ver o lançamento cancelado na listagem, veio %+v", lancamentos)
	}
}

// LAN-04 AC2: sem a permissão, nenhuma linha.
func TestListarLancamentosRequiresThePermission(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	e.criarLancamentoDeTeste(t, conta.ID)

	lancamentos, err := e.listarLancamentos().Execute(context.Background(), app.ListarLancamentosInput{Actor: actor()})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	if len(lancamentos) != 0 {
		t.Errorf("esperava nenhuma linha, veio %d", len(lancamentos))
	}
}
