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

func (e env) desativarConta() *app.DesativarConta {
	return &app.DesativarConta{Authz: simpleAuthz{}, Contas: e.contas, Audit: e.rec, Tx: e.tx}
}

// PC-03 AC1/AC4: desativa sem excluir a linha.
func TestDesativarContaSetsAtivoFalseWithoutDeleting(t *testing.T) {
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
		t.Fatalf("desativar: %v", err)
	}

	achada, err := e.contas.Buscar(context.Background(), conta.ID)
	if err != nil {
		t.Fatalf("buscar após desativar: %v", err)
	}
	if achada.Ativo {
		t.Errorf("esperava ativo=false, veio %+v", achada)
	}
}

// Edge case (spec): desativar uma conta já inativa não é erro.
func TestDesativarContaIsIdempotent(t *testing.T) {
	e := newEnv(t)
	conta, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoDespesa, Nome: "Material",
	})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}
	in := app.DesativarContaInput{Actor: actor(app.PermContaDeactivate), ID: conta.ID}
	if err := e.desativarConta().Execute(context.Background(), in); err != nil {
		t.Fatalf("primeira desativação: %v", err)
	}
	if err := e.desativarConta().Execute(context.Background(), in); err != nil {
		t.Errorf("segunda desativação deveria ser idempotente, veio err=%v", err)
	}

	achada, err := e.contas.Buscar(context.Background(), conta.ID)
	if err != nil {
		t.Fatalf("buscar: %v", err)
	}
	if achada.Ativo {
		t.Errorf("esperava ativo=false, veio %+v", achada)
	}
}

// PC-03 AC5: sem a permissão, recusa sem escrever.
func TestDesativarContaRequiresThePermissionAndWritesNothing(t *testing.T) {
	e := newEnv(t)
	conta, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Produtos",
	})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	err = e.desativarConta().Execute(context.Background(), app.DesativarContaInput{
		Actor: actor(), ID: conta.ID,
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}

	achada, err := e.contas.Buscar(context.Background(), conta.ID)
	if err != nil {
		t.Fatalf("buscar: %v", err)
	}
	if !achada.Ativo {
		t.Errorf("conta não deveria ter sido desativada")
	}
}

// PC-03 AC3: desativação audita.
func TestDesativarContaAudits(t *testing.T) {
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
		t.Fatalf("desativar: %v", err)
	}

	var action, entityID string
	err = e.db.Raw(`SELECT action, entity_id FROM audit_log WHERE action = 'conta.deactivate' AND entity_id = ?`, conta.ID).
		Row().Scan(&action, &entityID)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if entityID != conta.ID {
		t.Errorf("entity_id = %q, esperado %q", entityID, conta.ID)
	}
}

// Sem regra de negócio nova aqui: um id inexistente reusa
// domain.ErrContaNaoEncontrada, o mesmo padrão já usado por Buscar/Criar
// (T2) — consequência técnica direta de já existir esse sentinel, não uma
// nova regra sobre desativação.
func TestDesativarContaRejectsANonexistentConta(t *testing.T) {
	e := newEnv(t)
	bogus := "00000000-0000-0000-0000-000000000000"

	err := e.desativarConta().Execute(context.Background(), app.DesativarContaInput{
		Actor: actor(app.PermContaDeactivate), ID: bogus,
	})
	if !errors.Is(err, domain.ErrContaNaoEncontrada) {
		t.Errorf("err = %v, esperado ErrContaNaoEncontrada", err)
	}
}
