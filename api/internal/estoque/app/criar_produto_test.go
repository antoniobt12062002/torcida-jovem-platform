//go:build integration

package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// PRD-01 AC1: cria o SKU com sucesso.
func TestCriarProdutoPersistsAndReturnsTheCreatedSKU(t *testing.T) {
	e := newEnv(t)

	produto, err := e.criarProduto().Execute(context.Background(), app.CriarProdutoInput{
		Actor: actor(app.PermProdutoCreate), Codigo: "CAMISA-TJ-P-PRETA", Nome: "Camisa TJ P Preta", UnidadeMedida: "UN",
	})

	if err != nil {
		t.Fatalf("criar: %v", err)
	}
	if produto.ID == "" || produto.Codigo != "CAMISA-TJ-P-PRETA" || produto.Nome != "Camisa TJ P Preta" || produto.UnidadeMedida != "UN" {
		t.Errorf("produto = %+v", produto)
	}
}

// PRD-01 AC2: codigo duplicado é recusado.
func TestCriarProdutoRejectsADuplicateCodigo(t *testing.T) {
	e := newEnv(t)
	a := actor(app.PermProdutoCreate)
	if _, err := e.criarProduto().Execute(context.Background(), app.CriarProdutoInput{
		Actor: a, Codigo: "DUP", Nome: "x", UnidadeMedida: "UN",
	}); err != nil {
		t.Fatalf("criar primeiro: %v", err)
	}

	_, err := e.criarProduto().Execute(context.Background(), app.CriarProdutoInput{
		Actor: a, Codigo: "DUP", Nome: "y", UnidadeMedida: "UN",
	})

	if !errors.Is(err, domain.ErrCodigoDuplicado) {
		t.Errorf("err = %v, esperado ErrCodigoDuplicado", err)
	}
}

// PRD-01 AC3: campo vazio é recusado, para cada um dos 3 campos.
func TestCriarProdutoRejectsAnyBlankField(t *testing.T) {
	e := newEnv(t)
	a := actor(app.PermProdutoCreate)
	cases := map[string]app.CriarProdutoInput{
		"codigo vazio":         {Actor: a, Codigo: "", Nome: "x", UnidadeMedida: "UN"},
		"nome vazio":           {Actor: a, Codigo: "X", Nome: "", UnidadeMedida: "UN"},
		"unidade_medida vazia": {Actor: a, Codigo: "X", Nome: "x", UnidadeMedida: ""},
	}
	for name, in := range cases {
		_, err := e.criarProduto().Execute(context.Background(), in)
		if !errors.Is(err, domain.ErrCampoObrigatorio) {
			t.Errorf("%s: err = %v, esperado ErrCampoObrigatorio", name, err)
		}
	}
}

// PRD-01 AC4: sem a permissão, a resposta é forbidden, sem revelar se o
// codigo já existe.
func TestCriarProdutoWithoutThePermissionIsForbidden(t *testing.T) {
	e := newEnv(t)

	_, err := e.criarProduto().Execute(context.Background(), app.CriarProdutoInput{
		Actor: actor(), Codigo: "X", Nome: "x", UnidadeMedida: "UN",
	})

	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado forbidden", err)
	}
}
