//go:build integration

package infra_test

import (
	"context"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// AJS-02: Saldo (sem lock) devolve 0 para um SKU sem nenhuma movimentação.
func TestSaldoReturnsZeroForAProdutoWithNoMovimentacoes(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoID := newTestProduto(t, produtos)

	saldo, err := movs.Saldo(context.Background(), produtoID)

	if err != nil {
		t.Fatalf("saldo: %v", err)
	}
	if saldo != 0 {
		t.Errorf("saldo = %d, esperado 0", saldo)
	}
}

// AJS-02: Saldo reflete entrada+saída+ajuste+devolução combinados.
func TestSaldoReflectsEntradaSaidaAjusteAndDevolucaoCombined(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoID := newTestProduto(t, produtos)
	userID := newTestUser(t, db)
	ctx := context.Background()

	entries := []domain.Movimentacao{
		{ProdutoID: produtoID, Tipo: domain.Entrada, Quantidade: 10, Origem: domain.OrigemCompra, ResponsavelID: userID},
		{ProdutoID: produtoID, Tipo: domain.Saida, Quantidade: -3, Origem: domain.OrigemVenda, ResponsavelID: userID},
		{ProdutoID: produtoID, Tipo: domain.Ajuste, Quantidade: -2, Origem: domain.OrigemAjusteManual, ResponsavelID: userID},
		{ProdutoID: produtoID, Tipo: domain.Devolucao, Quantidade: 1, Origem: domain.OrigemVenda, ResponsavelID: userID},
	}
	for _, m := range entries {
		if _, err := movs.Criar(ctx, m); err != nil {
			t.Fatalf("criar %s: %v", m.Tipo, err)
		}
	}

	saldo, err := movs.Saldo(ctx, produtoID)

	if err != nil {
		t.Fatalf("saldo: %v", err)
	}
	if saldo != 6 { // 10 - 3 - 2 + 1
		t.Errorf("saldo = %d, esperado 6", saldo)
	}
}
