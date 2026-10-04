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

// MOV-04 AC1: lista todas as movimentações do produto, incluindo AJUSTE, na
// ordem em que ocorreram, sem paginação.
func TestListarMovimentacoesReturnsEveryTipoInCreationOrder(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "L1")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)
	entrada, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 10, Origem: domain.OrigemCompra,
	})
	if err != nil {
		t.Fatalf("entrada: %v", err)
	}
	saida, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Saida, ProdutoID: produtoID, Quantidade: 3, Origem: domain.OrigemVenda,
	})
	if err != nil {
		t.Fatalf("saída: %v", err)
	}
	// Ajuste simulado via Criar direto (AjustarEstoque só existe em
	// 03-ajustes-e-saldo) — prova que a listagem não filtra por tipo.
	ajuste, err := e.movimentacoes.Criar(context.Background(), domain.Movimentacao{
		ProdutoID: produtoID, Tipo: domain.Ajuste, Quantidade: -1, Origem: domain.OrigemAjusteManual, ResponsavelID: actorRM.UserID,
	})
	if err != nil {
		t.Fatalf("ajuste: %v", err)
	}

	lista, err := e.listarMovimentacoes().Execute(context.Background(), app.ListarMovimentacoesInput{
		Actor: e.userActor(t, app.PermMovimentacaoRead), ProdutoID: produtoID,
	})

	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(lista) != 3 || lista[0].ID != entrada.ID || lista[1].ID != saida.ID || lista[2].ID != ajuste.ID {
		t.Errorf("lista = %+v", lista)
	}
}

// MOV-04 AC2: sem a permissão, a resposta é forbidden.
func TestListarMovimentacoesWithoutThePermissionIsForbidden(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "L2")

	_, err := e.listarMovimentacoes().Execute(context.Background(), app.ListarMovimentacoesInput{
		Actor: e.userActor(t), ProdutoID: produtoID,
	})

	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado forbidden", err)
	}
}
