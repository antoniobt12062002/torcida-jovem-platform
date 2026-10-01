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

func (e env) renomearConta() *app.RenomearConta {
	return &app.RenomearConta{Authz: simpleAuthz{}, Contas: e.contas, Lancamentos: e.existenceChecker, Audit: e.rec, Tx: e.tx}
}

// PC-02 AC1: conta nunca usada é renomeada com sucesso.
func TestRenomearContaRenamesAnUnusedConta(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")

	renomeada, err := e.renomearConta().Execute(context.Background(), app.RenomearContaInput{
		Actor: actor(app.PermContaUpdate), ID: conta.ID, Nome: "Produtos Esportivos",
	})
	if err != nil {
		t.Fatalf("renomear: %v", err)
	}
	if renomeada.Nome != "Produtos Esportivos" {
		t.Errorf("nome = %q, esperado %q", renomeada.Nome, "Produtos Esportivos")
	}

	achada, err := e.contas.Buscar(context.Background(), conta.ID)
	if err != nil || achada.Nome != "Produtos Esportivos" {
		t.Errorf("buscar depois de renomear: %+v, err=%v", achada, err)
	}
}

// PC-02 AC2, Edge Case: conta já usada é recusada, mesmo com o único
// lançamento CANCELADA.
func TestRenomearContaRejectsAContaReferencedByALancamentoEvenIfCancelled(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	if err := e.db.Exec("UPDATE lancamentos SET status = 'CANCELADA' WHERE id = ?::uuid", lancamento.ID).Error; err != nil {
		t.Fatalf("fixture: marcar CANCELADA: %v", err)
	}

	_, err := e.renomearConta().Execute(context.Background(), app.RenomearContaInput{
		Actor: actor(app.PermContaUpdate), ID: conta.ID, Nome: "Novo Nome",
	})
	if !errors.Is(err, domain.ErrContaJaUtilizada) {
		t.Errorf("err = %v, esperado ErrContaJaUtilizada", err)
	}

	achada, err := e.contas.Buscar(context.Background(), conta.ID)
	if err != nil || achada.Nome != "Material" {
		t.Errorf("nome não deveria ter mudado: %+v, err=%v", achada, err)
	}
}

// PC-02 AC5: sem a permissão, recusa mesmo para uma conta nunca usada.
func TestRenomearContaRequiresThePermissionAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")

	_, err := e.renomearConta().Execute(context.Background(), app.RenomearContaInput{
		Actor: actor(), ID: conta.ID, Nome: "Outro Nome",
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}

	achada, err := e.contas.Buscar(context.Background(), conta.ID)
	if err != nil || achada.Nome != "Produtos" {
		t.Errorf("nome não deveria ter mudado: %+v, err=%v", achada, err)
	}
}

// PC-02 AC4: a renomeação audita o nome anterior e o novo.
func TestRenomearContaAuditsPreviousAndNewNome(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")

	renomeada, err := e.renomearConta().Execute(context.Background(), app.RenomearContaInput{
		Actor: actor(app.PermContaUpdate), ID: conta.ID, Nome: "Produtos Esportivos",
	})
	if err != nil {
		t.Fatalf("renomear: %v", err)
	}

	var before, after string
	err = e.db.Raw(`SELECT before::text, after::text FROM audit_log WHERE action = 'conta.rename' AND entity_id = ?`, renomeada.ID).
		Row().Scan(&before, &after)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if !strings.Contains(before, "Produtos") || !strings.Contains(after, "Produtos Esportivos") {
		t.Errorf("auditoria não confere: before=%q after=%q", before, after)
	}
}

// PC-02 AC3: a checagem de uso nunca escreve em lancamentos.
func TestRenomearContaNeverWritesToLancamentos(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoReceita, "Produtos")

	if _, err := e.renomearConta().Execute(context.Background(), app.RenomearContaInput{
		Actor: actor(app.PermContaUpdate), ID: conta.ID, Nome: "Produtos Esportivos",
	}); err != nil {
		t.Fatalf("renomear: %v", err)
	}

	var n int64
	if err := e.db.Raw("SELECT count(*) FROM lancamentos").Scan(&n).Error; err != nil || n != 0 {
		t.Errorf("nenhum lançamento deveria existir: n=%d err=%v", n, err)
	}
}

// Sem regra de negócio nova: um id inexistente reusa
// domain.ErrContaNaoEncontrada, o mesmo padrão já usado em T2/T3.
func TestRenomearContaRejectsANonexistentConta(t *testing.T) {
	e := newEnv(t)
	bogus := "00000000-0000-0000-0000-000000000000"

	_, err := e.renomearConta().Execute(context.Background(), app.RenomearContaInput{
		Actor: actor(app.PermContaUpdate), ID: bogus, Nome: "Novo Nome",
	})
	if !errors.Is(err, domain.ErrContaNaoEncontrada) {
		t.Errorf("err = %v, esperado ErrContaNaoEncontrada", err)
	}
}

