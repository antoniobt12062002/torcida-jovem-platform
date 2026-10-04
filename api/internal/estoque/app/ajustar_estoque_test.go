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

// AJS-01 AC1: ajuste negativo é aceito mesmo sem nenhuma movimentação
// anterior (saldo passa a ficar negativo, única válvula de escape autorizada).
func TestAjustarEstoquePersistsANegativeAdjustmentEvenWithNoPriorBalance(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "AJ1")
	actorAJ := e.userActor(t, app.PermMovimentacaoAdjust)

	ajuste, err := e.ajustarEstoque().Execute(context.Background(), app.AjustarEstoqueInput{
		Actor: actorAJ, ProdutoID: produtoID, Quantidade: -3, Motivo: "correção de inventário",
	})

	if err != nil {
		t.Fatalf("ajustar: %v", err)
	}
	if ajuste.Quantidade != -3 || ajuste.Tipo != domain.Ajuste || ajuste.Motivo == nil || *ajuste.Motivo != "correção de inventário" {
		t.Errorf("ajuste = %+v", ajuste)
	}

	lista, _ := e.movimentacoes.ListarPorProduto(context.Background(), produtoID)
	var somado int64
	for _, m := range lista {
		somado += m.Quantidade
	}
	if somado != -3 {
		t.Errorf("saldo = %d, esperado -3", somado)
	}
}

// AJS-01 AC2: motivo vazio é recusado, antes de qualquer leitura/escrita.
func TestAjustarEstoqueWithBlankMotivoIsRejectedBeforeAnyPersistence(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "AJ2")
	actorAJ := e.userActor(t, app.PermMovimentacaoAdjust)

	_, err := e.ajustarEstoque().Execute(context.Background(), app.AjustarEstoqueInput{
		Actor: actorAJ, ProdutoID: produtoID, Quantidade: 5, Motivo: "   ",
	})

	if !errors.Is(err, domain.ErrMotivoObrigatorio) {
		t.Errorf("err = %v, esperado ErrMotivoObrigatorio", err)
	}
	lista, _ := e.movimentacoes.ListarPorProduto(context.Background(), produtoID)
	if len(lista) != 0 {
		t.Errorf("nada deveria ter sido persistido: %d movimentações", len(lista))
	}
}

// AJS-01 AC3: quantidade zero é recusada.
func TestAjustarEstoqueWithZeroQuantidadeIsRejected(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "AJ3")
	actorAJ := e.userActor(t, app.PermMovimentacaoAdjust)

	_, err := e.ajustarEstoque().Execute(context.Background(), app.AjustarEstoqueInput{
		Actor: actorAJ, ProdutoID: produtoID, Quantidade: 0, Motivo: "motivo válido",
	})

	if !errors.Is(err, domain.ErrQuantidadeInvalida) {
		t.Errorf("err = %v, esperado ErrQuantidadeInvalida", err)
	}
}

// AJS-01 AC4: produto inexistente é recusado.
func TestAjustarEstoqueWithAMissingProdutoIsRejected(t *testing.T) {
	e := newEnv(t)
	actorAJ := e.userActor(t, app.PermMovimentacaoAdjust)

	_, err := e.ajustarEstoque().Execute(context.Background(), app.AjustarEstoqueInput{
		Actor: actorAJ, ProdutoID: "00000000-0000-0000-0000-000000000000", Quantidade: 5, Motivo: "motivo válido",
	})

	if !errors.Is(err, domain.ErrProdutoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrProdutoNaoEncontrado", err)
	}
}

// AJS-01 AC5: um ator com só estoque:movimentacao:create (sem adjust) é recusado.
func TestAjustarEstoqueWithOnlyMovimentacaoCreatePermissionIsForbidden(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "AJ4")

	_, err := e.ajustarEstoque().Execute(context.Background(), app.AjustarEstoqueInput{
		Actor: e.userActor(t, app.PermMovimentacaoCreate), ProdutoID: produtoID, Quantidade: 5, Motivo: "motivo válido",
	})

	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado forbidden (movimentacao:create não cobre adjust)", err)
	}
}
