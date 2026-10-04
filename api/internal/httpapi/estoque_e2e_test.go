//go:build integration

package httpapi

import (
	"net/http"
	"testing"
)

// 05-api-http/T5 (estoque): one true end-to-end test per resource, through
// the real httpapi.NewRouter (estoquehttp.Register, the real chains, every
// response checked against the contracts). PRESIDENTE holds every estoque
// permission automatically.

// API-01: produtos — criar, recusar código repetido, listar.
func TestEstoqueE2EProdutosFlow(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()

	created := e.do(s, http.MethodPost, "/api/v1/estoque/produtos", map[string]any{"codigo": "CAMISA-TJ-P", "nome": "Camisa P", "unidade_medida": "UN"})
	e.expect(created, http.StatusCreated, "")

	dup := e.do(s, http.MethodPost, "/api/v1/estoque/produtos", map[string]any{"codigo": "CAMISA-TJ-P", "nome": "Outra", "unidade_medida": "UN"})
	e.expect(dup, http.StatusConflict, "codigo_duplicado")

	list := e.do(s, http.MethodGet, "/api/v1/estoque/produtos", nil)
	e.expect(list, http.StatusOK, "")
	if items := body(t, list)["items"].([]any); len(items) != 1 {
		t.Fatalf("items = %v, esperado 1", items)
	}
}

// The real router validates estoque request bodies against estoque.yaml
// before any use case runs: a body that violates the contract (minLength,
// request-only enum) answers validation_failed, never the use case's own
// code (campo_obrigatorio, quantidade_invalida) — which is what would come
// back if estoque.yaml were missing from httpx.NewContractValidator, since
// the validator lets through routes absent from every loaded contract.
func TestEstoqueE2ERequestsAreValidatedAgainstTheContract(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()
	produtoID := body(t, e.do(s, http.MethodPost, "/api/v1/estoque/produtos", map[string]any{"codigo": "MEIA", "nome": "Meia", "unidade_medida": "UN"}))["id"].(string)

	e.expect(e.do(s, http.MethodPost, "/api/v1/estoque/produtos", map[string]any{
		"codigo": "", "nome": "Sem código", "unidade_medida": "UN",
	}), http.StatusUnprocessableEntity, "validation_failed")
	e.expect(e.do(s, http.MethodPost, "/api/v1/estoque/movimentacoes", map[string]any{
		"tipo": "AJUSTE", "produto_id": produtoID, "quantidade": 1, "origem": "AJUSTE_MANUAL",
	}), http.StatusUnprocessableEntity, "validation_failed")
}

// API-02: movimentações — entrada, saída acima do saldo recusada, saída,
// devolução da saída, listar na ordem.
func TestEstoqueE2EMovimentacoesFlow(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()
	produtoID := body(t, e.do(s, http.MethodPost, "/api/v1/estoque/produtos", map[string]any{"codigo": "BONE", "nome": "Boné", "unidade_medida": "UN"}))["id"].(string)

	e.expect(e.do(s, http.MethodPost, "/api/v1/estoque/movimentacoes", map[string]any{
		"tipo": "ENTRADA", "produto_id": produtoID, "quantidade": 5, "origem": "INVENTARIO",
	}), http.StatusCreated, "")
	e.expect(e.do(s, http.MethodPost, "/api/v1/estoque/movimentacoes", map[string]any{
		"tipo": "SAIDA", "produto_id": produtoID, "quantidade": 6, "origem": "VENDA",
	}), http.StatusConflict, "saldo_insuficiente")

	saida := e.do(s, http.MethodPost, "/api/v1/estoque/movimentacoes", map[string]any{
		"tipo": "SAIDA", "produto_id": produtoID, "quantidade": 2, "origem": "VENDA",
	})
	e.expect(saida, http.StatusCreated, "")
	if body(t, saida)["quantidade"] != float64(-2) {
		t.Errorf("quantidade da saída = %v, esperado -2 (já assinada)", body(t, saida)["quantidade"])
	}

	e.expect(e.do(s, http.MethodPost, "/api/v1/estoque/movimentacoes", map[string]any{
		"tipo": "DEVOLUCAO", "produto_id": produtoID, "quantidade": 2, "origem": "VENDA", "movimentacao_de_id": body(t, saida)["id"],
	}), http.StatusCreated, "")

	list := e.do(s, http.MethodGet, "/api/v1/estoque/produtos/"+produtoID+"/movimentacoes", nil)
	e.expect(list, http.StatusOK, "")
	if items := body(t, list)["items"].([]any); len(items) != 3 {
		t.Fatalf("items = %v, esperado 3 (a saída recusada nunca é gravada)", items)
	}
}

// API-03/API-04: ajuste com motivo leva o saldo a negativo; sem motivo é
// recusado; o saldo reflete tudo.
func TestEstoqueE2EAjusteESaldoFlow(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()
	produtoID := body(t, e.do(s, http.MethodPost, "/api/v1/estoque/produtos", map[string]any{"codigo": "CACHECOL", "nome": "Cachecol", "unidade_medida": "UN"}))["id"].(string)

	saldo := e.do(s, http.MethodGet, "/api/v1/estoque/produtos/"+produtoID+"/saldo", nil)
	e.expect(saldo, http.StatusOK, "")
	if body(t, saldo)["saldo"] != float64(0) {
		t.Fatalf("saldo inicial = %v, esperado 0", body(t, saldo)["saldo"])
	}

	e.expect(e.do(s, http.MethodPost, "/api/v1/estoque/movimentacoes", map[string]any{
		"tipo": "ENTRADA", "produto_id": produtoID, "quantidade": 10, "origem": "INVENTARIO",
	}), http.StatusCreated, "")
	e.expect(e.do(s, http.MethodPost, "/api/v1/estoque/ajustes", map[string]any{
		"produto_id": produtoID, "quantidade": -12, "motivo": "",
	}), http.StatusUnprocessableEntity, "motivo_obrigatorio")
	e.expect(e.do(s, http.MethodPost, "/api/v1/estoque/ajustes", map[string]any{
		"produto_id": produtoID, "quantidade": -12, "motivo": "perda em inventário físico",
	}), http.StatusCreated, "")

	saldo = e.do(s, http.MethodGet, "/api/v1/estoque/produtos/"+produtoID+"/saldo", nil)
	e.expect(saldo, http.StatusOK, "")
	if body(t, saldo)["saldo"] != float64(-2) {
		t.Errorf("saldo = %v, esperado -2", body(t, saldo)["saldo"])
	}
}
