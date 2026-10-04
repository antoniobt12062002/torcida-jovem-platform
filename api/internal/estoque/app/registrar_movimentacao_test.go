//go:build integration

package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// MOV-01 AC1: entrada registra quantidade positiva.
func TestRegistrarMovimentacaoEntradaPersistsAPositiveQuantidade(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "E1")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)

	criada, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 10, Origem: domain.OrigemCompra,
	})

	if err != nil {
		t.Fatalf("registrar entrada: %v", err)
	}
	if criada.Quantidade != 10 || criada.Tipo != domain.Entrada {
		t.Errorf("movimentação = %+v", criada)
	}
}

// MOV-01 AC2: produto_id inexistente é recusado.
func TestRegistrarMovimentacaoWithAMissingProdutoIsRejected(t *testing.T) {
	e := newEnv(t)
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)

	_, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: uuid.NewString(), Quantidade: 10, Origem: domain.OrigemCompra,
	})

	if !errors.Is(err, domain.ErrProdutoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrProdutoNaoEncontrado", err)
	}
}

// MOV-01 AC3: quantidade zero ou negativa é recusada.
func TestRegistrarMovimentacaoWithZeroOrNegativeQuantidadeIsRejected(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "E2")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)

	for _, qtd := range []int64{0, -1} {
		_, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
			Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: qtd, Origem: domain.OrigemCompra,
		})
		if !errors.Is(err, domain.ErrQuantidadeInvalida) {
			t.Errorf("quantidade=%d: err = %v, esperado ErrQuantidadeInvalida", qtd, err)
		}
	}
}

// MOV-01 AC4: sem a permissão, forbidden.
func TestRegistrarMovimentacaoWithoutThePermissionIsForbidden(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "E3")

	_, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: e.userActor(t), Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 10, Origem: domain.OrigemCompra,
	})

	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado forbidden", err)
	}
}

// MOV-02 AC1: saída registra quantidade negativa quando o saldo é suficiente.
func TestRegistrarMovimentacaoSaidaPersistsANegativeQuantidadeWhenBalanceIsSufficient(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "S1")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)
	if _, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 10, Origem: domain.OrigemCompra,
	}); err != nil {
		t.Fatalf("entrada inicial: %v", err)
	}

	saida, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Saida, ProdutoID: produtoID, Quantidade: 4, Origem: domain.OrigemVenda,
	})

	if err != nil {
		t.Fatalf("registrar saída: %v", err)
	}
	if saida.Quantidade != -4 {
		t.Errorf("quantidade = %d, esperado -4", saida.Quantidade)
	}
}

// MOV-02 AC2: saída com saldo insuficiente é recusada, sem registrar nada.
func TestRegistrarMovimentacaoSaidaWithInsufficientBalanceIsRejected(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "S2")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)
	if _, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 3, Origem: domain.OrigemCompra,
	}); err != nil {
		t.Fatalf("entrada inicial: %v", err)
	}

	_, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Saida, ProdutoID: produtoID, Quantidade: 4, Origem: domain.OrigemVenda,
	})

	if !errors.Is(err, domain.ErrSaldoInsuficiente) {
		t.Errorf("err = %v, esperado ErrSaldoInsuficiente", err)
	}
	lista, _ := e.movimentacoes.ListarPorProduto(context.Background(), produtoID)
	if len(lista) != 1 {
		t.Errorf("nada deveria ter sido registrado além da entrada: %d movimentações", len(lista))
	}
}

// MOV-03 AC1: devolução de uma SAIDA repõe estoque (soma positiva, nunca bloqueada).
func TestRegistrarMovimentacaoDevolucaoOfASaidaRepõeEstoque(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "D1")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)
	if _, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 10, Origem: domain.OrigemCompra,
	}); err != nil {
		t.Fatalf("entrada: %v", err)
	}
	saida, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Saida, ProdutoID: produtoID, Quantidade: 4, Origem: domain.OrigemVenda,
	})
	if err != nil {
		t.Fatalf("saída: %v", err)
	}

	devolucao, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Devolucao, ProdutoID: produtoID, Quantidade: 2, Origem: domain.OrigemVenda,
		MovimentacaoDeID: &saida.ID,
	})

	if err != nil {
		t.Fatalf("devolução: %v", err)
	}
	if devolucao.Quantidade != 2 {
		t.Errorf("quantidade = %d, esperado +2 (repõe estoque)", devolucao.Quantidade)
	}
	lista, _ := e.movimentacoes.ListarPorProduto(context.Background(), produtoID)
	var somado int64
	for _, m := range lista {
		somado += m.Quantidade
	}
	if somado != 8 { // 10 - 4 + 2
		t.Errorf("soma das movimentações = %d, esperado 8", somado)
	}
}

// MOV-03 AC1 (direção inversa): devolução de uma ENTRADA remove estoque,
// sujeita à mesma regra de saldo insuficiente que SAIDA.
func TestRegistrarMovimentacaoDevolucaoOfAnEntradaRemovesEstoqueAndCanBeRejectedByBalance(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "D2")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)
	entrada, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 5, Origem: domain.OrigemCompra,
	})
	if err != nil {
		t.Fatalf("entrada: %v", err)
	}
	if _, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Saida, ProdutoID: produtoID, Quantidade: 3, Origem: domain.OrigemVenda,
	}); err != nil {
		t.Fatalf("saída: %v", err)
	}
	// Saldo atual = 2. Devolver 3 da entrada original removeria 3, ficando
	// negativo — deve ser recusado pela mesma regra de saldo insuficiente.
	_, err = e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Devolucao, ProdutoID: produtoID, Quantidade: 3, Origem: domain.OrigemCompra,
		MovimentacaoDeID: &entrada.ID,
	})
	if !errors.Is(err, domain.ErrSaldoInsuficiente) {
		t.Errorf("err = %v, esperado ErrSaldoInsuficiente", err)
	}

	// Devolver 2 (dentro do saldo disponível) deve ser aceito, com sinal negativo.
	devolucao, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Devolucao, ProdutoID: produtoID, Quantidade: 2, Origem: domain.OrigemCompra,
		MovimentacaoDeID: &entrada.ID,
	})
	if err != nil {
		t.Fatalf("devolução dentro do saldo: %v", err)
	}
	if devolucao.Quantidade != -2 {
		t.Errorf("quantidade = %d, esperado -2 (remove estoque)", devolucao.Quantidade)
	}
}

// MOV-03 AC2: devolução referenciando AJUSTE, outra DEVOLUCAO, id
// inexistente ou produto diferente é recusada com devolucao_invalida.
func TestRegistrarMovimentacaoDevolucaoWithAnInvalidReferenceIsRejected(t *testing.T) {
	e := newEnv(t)
	produtoA := e.criarProdutoDeTeste(t, "D3A")
	produtoB := e.criarProdutoDeTeste(t, "D3B")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)
	entradaA, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoA, Quantidade: 10, Origem: domain.OrigemCompra,
	})
	if err != nil {
		t.Fatalf("entrada A: %v", err)
	}

	bogus := uuid.NewString()
	cases := map[string]*string{
		"inexistente":       &bogus,
		"produto_diferente": &entradaA.ID, // referenciado existe, mas produtoID do input será B
	}
	for name, ref := range cases {
		produtoAlvo := produtoA
		if name == "produto_diferente" {
			produtoAlvo = produtoB
		}
		_, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
			Actor: actorRM, Tipo: domain.Devolucao, ProdutoID: produtoAlvo, Quantidade: 1, Origem: domain.OrigemVenda,
			MovimentacaoDeID: ref,
		})
		if !errors.Is(err, domain.ErrDevolucaoInvalida) {
			t.Errorf("%s: err = %v, esperado ErrDevolucaoInvalida", name, err)
		}
	}

	// Referenciar um AJUSTE (simulado diretamente via Criar, já que
	// AjustarEstoque só existe em 03-ajustes-e-saldo).
	ajuste, err := e.movimentacoes.Criar(context.Background(), domain.Movimentacao{
		ProdutoID: produtoA, Tipo: domain.Ajuste, Quantidade: 5, Origem: domain.OrigemAjusteManual, ResponsavelID: actorRM.UserID,
	})
	if err != nil {
		t.Fatalf("criar ajuste de teste: %v", err)
	}
	_, err = e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Devolucao, ProdutoID: produtoA, Quantidade: 1, Origem: domain.OrigemVenda,
		MovimentacaoDeID: &ajuste.ID,
	})
	if !errors.Is(err, domain.ErrDevolucaoInvalida) {
		t.Errorf("referenciar ajuste: err = %v, esperado ErrDevolucaoInvalida", err)
	}

	// Referenciar outra DEVOLUCAO (EST-D-009: nunca devolução de devolução).
	devolucao, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Devolucao, ProdutoID: produtoA, Quantidade: 1, Origem: domain.OrigemCompra,
		MovimentacaoDeID: &entradaA.ID,
	})
	if err != nil {
		t.Fatalf("devolução válida de teste: %v", err)
	}
	_, err = e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Devolucao, ProdutoID: produtoA, Quantidade: 1, Origem: domain.OrigemCompra,
		MovimentacaoDeID: &devolucao.ID,
	})
	if !errors.Is(err, domain.ErrDevolucaoInvalida) {
		t.Errorf("referenciar outra devolução: err = %v, esperado ErrDevolucaoInvalida", err)
	}
}

// MOV-03 AC4: a movimentação original nunca é alterada por uma devolução.
func TestRegistrarMovimentacaoDevolucaoNeverChangesTheOriginal(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "D4")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)
	saida, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 10, Origem: domain.OrigemCompra,
	})
	if err != nil {
		t.Fatalf("entrada: %v", err)
	}

	if _, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Devolucao, ProdutoID: produtoID, Quantidade: 3, Origem: domain.OrigemCompra,
		MovimentacaoDeID: &saida.ID,
	}); err != nil {
		t.Fatalf("devolução: %v", err)
	}

	original, err := e.movimentacoes.Buscar(context.Background(), saida.ID)
	if err != nil {
		t.Fatalf("buscar original: %v", err)
	}
	if original.Quantidade != saida.Quantidade {
		t.Errorf("original foi alterada: quantidade = %d, esperado %d", original.Quantidade, saida.Quantidade)
	}
}
