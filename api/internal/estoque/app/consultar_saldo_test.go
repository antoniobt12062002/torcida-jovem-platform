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

// AJS-02 AC1: saldo correto, inclusive zero e negativo.
func TestConsultarSaldoReturnsTheExactSumIncludingZeroAndNegative(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "CS1")
	readActor := e.userActor(t, app.PermSaldoRead)

	zero, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: readActor, ProdutoID: produtoID})
	if err != nil || zero != 0 {
		t.Fatalf("saldo inicial = %d, err = %v, esperado 0", zero, err)
	}

	writeActor := e.userActor(t, app.PermMovimentacaoAdjust)
	if _, err := e.ajustarEstoque().Execute(context.Background(), app.AjustarEstoqueInput{
		Actor: writeActor, ProdutoID: produtoID, Quantidade: -5, Motivo: "correção",
	}); err != nil {
		t.Fatalf("ajustar: %v", err)
	}

	negativo, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: readActor, ProdutoID: produtoID})
	if err != nil {
		t.Fatalf("consultar: %v", err)
	}
	if negativo != -5 {
		t.Errorf("saldo = %d, esperado -5", negativo)
	}
}

// AJS-02 AC2: produto inexistente é recusado.
func TestConsultarSaldoWithAMissingProdutoIsRejected(t *testing.T) {
	e := newEnv(t)

	_, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{
		Actor: e.userActor(t, app.PermSaldoRead), ProdutoID: "00000000-0000-0000-0000-000000000000",
	})

	if !errors.Is(err, domain.ErrProdutoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrProdutoNaoEncontrado", err)
	}
}

// AJS-02 AC3: estoque:saldo:read é permissão distinta de
// estoque:movimentacao:read — um ator com só a segunda é recusado.
func TestConsultarSaldoWithOnlyMovimentacaoReadPermissionIsForbidden(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "CS2")

	_, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{
		Actor: e.userActor(t, app.PermMovimentacaoRead), ProdutoID: produtoID,
	})

	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado forbidden (movimentacao:read não cobre saldo:read)", err)
	}
}
