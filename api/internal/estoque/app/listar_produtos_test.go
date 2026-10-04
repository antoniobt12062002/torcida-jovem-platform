//go:build integration

package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PRD-02 AC1: lista todos os SKUs, sem paginação.
func TestListarProdutosReturnsEveryCreatedProdutoNoPagination(t *testing.T) {
	e := newEnv(t)
	actorCreate := actor(app.PermProdutoCreate)
	for _, codigo := range []string{"A", "B", "C"} {
		if _, err := e.criarProduto().Execute(context.Background(), app.CriarProdutoInput{
			Actor: actorCreate, Codigo: codigo, Nome: "produto " + codigo, UnidadeMedida: "UN",
		}); err != nil {
			t.Fatalf("criar %s: %v", codigo, err)
		}
	}

	lista, err := e.listarProdutos().Execute(context.Background(), app.ListarProdutosInput{Actor: actor(app.PermProdutoRead)})

	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(lista) != 3 {
		t.Errorf("lista = %+v, esperado 3 itens", lista)
	}
}

// PRD-02 AC2: sem a permissão, a resposta é forbidden.
func TestListarProdutosWithoutThePermissionIsForbidden(t *testing.T) {
	e := newEnv(t)

	_, err := e.listarProdutos().Execute(context.Background(), app.ListarProdutosInput{Actor: actor()})

	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado forbidden", err)
	}
}
