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

func (e env) listarContas() *app.ListarContas {
	return &app.ListarContas{Authz: simpleAuthz{}, Contas: e.contas}
}

// PC-04 AC1: lista com hierarquia e status ativo.
func TestListarContasReturnsHierarchyAndAtivoStatus(t *testing.T) {
	e := newEnv(t)
	raiz, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Produtos",
	})
	if err != nil {
		t.Fatalf("criar conta raiz: %v", err)
	}
	sub, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Camisetas", ParentID: &raiz.ID,
	})
	if err != nil {
		t.Fatalf("criar subconta: %v", err)
	}

	contas, err := e.listarContas().Execute(context.Background(), app.ListarContasInput{Actor: actor(app.PermContaRead)})
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	var achouRaiz, achouSub bool
	for _, c := range contas {
		if c.ID == raiz.ID {
			achouRaiz = true
			if c.ParentID != nil || !c.Ativo {
				t.Errorf("raiz = %+v", c)
			}
		}
		if c.ID == sub.ID {
			achouSub = true
			if c.ParentID == nil || *c.ParentID != raiz.ID {
				t.Errorf("subconta.ParentID = %v, esperado %q", c.ParentID, raiz.ID)
			}
		}
	}
	if !achouRaiz || !achouSub {
		t.Errorf("esperava ver as duas contas na listagem, veio %+v", contas)
	}
}

// PC-04 AC2: sem a permissão, nenhuma linha.
func TestListarContasRequiresThePermission(t *testing.T) {
	e := newEnv(t)
	if _, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Produtos",
	}); err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	contas, err := e.listarContas().Execute(context.Background(), app.ListarContasInput{Actor: actor()})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	if len(contas) != 0 {
		t.Errorf("esperava nenhuma linha, veio %d", len(contas))
	}
}
