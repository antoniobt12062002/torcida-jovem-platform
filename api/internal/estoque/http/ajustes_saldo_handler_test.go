//go:build integration

package estoquehttp

import (
	"net/http"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func createAjusteBody(produtoID string, quantidade int64, motivo string) map[string]any {
	return map[string]any{"produto_id": produtoID, "quantidade": quantidade, "motivo": motivo}
}

// AJS-01 AC1: ajuste negativo é aceito mesmo levando o saldo a negativo,
// responde 201.
func TestCreateAjusteBelowZeroAnswers201(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "AJUSTE-1")

	w := e.as(s, http.MethodPost, "/api/v1/estoque/ajustes", createAjusteBody(produtoID, -3, "inventário físico"))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["tipo"] != "AJUSTE" || body["quantidade"] != float64(-3) || body["motivo"] != "inventário físico" {
		t.Errorf("corpo = %v", body)
	}

	saldo := decode(t, e.as(s, http.MethodGet, "/api/v1/estoque/produtos/"+produtoID+"/saldo", nil))
	if saldo["saldo"] != float64(-3) {
		t.Errorf("saldo = %v, esperado -3", saldo["saldo"])
	}
}

// AJS-01 AC2: motivo vazio responde 422 motivo_obrigatorio, antes de
// qualquer leitura/escrita.
func TestCreateAjusteWithoutMotivoAnswers422(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "AJUSTE-2")

	w := e.as(s, http.MethodPost, "/api/v1/estoque/ajustes", createAjusteBody(produtoID, 5, ""))

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "motivo_obrigatorio" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// AJS-01 AC3: quantidade zero responde 422 quantidade_invalida.
func TestCreateAjusteWithZeroQuantidadeAnswers422(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "AJUSTE-3")

	w := e.as(s, http.MethodPost, "/api/v1/estoque/ajustes", createAjusteBody(produtoID, 0, "motivo qualquer"))

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "quantidade_invalida" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// AJS-01 AC4: produto inexistente responde 404 produto_nao_encontrado.
func TestCreateAjusteForAnUnknownProdutoAnswers404(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)

	w := e.as(s, http.MethodPost, "/api/v1/estoque/ajustes", createAjusteBody("00000000-0000-0000-0000-000000000000", 1, "motivo"))

	if w.Code != http.StatusNotFound || codeOf(t, w) != "produto_nao_encontrado" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// AJS-01 AC5: um ator com movimentacao:create mas sem movimentacao:adjust é
// recusado com 403 forbidden — as duas permissões são distintas.
func TestCreateAjusteWithOnlyMovimentacaoCreatePermissionIsForbidden(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "AJUSTE-4")
	_, diretoria := e.signedIn(domain.RoleDiretoria)

	w := e.as(diretoria, http.MethodPost, "/api/v1/estoque/ajustes", createAjusteBody(produtoID, 1, "motivo"))

	if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// AJS-02 AC1: saldo consultado reflete entrada+saída combinadas, zero incluído.
func TestGetSaldoAnswers200WithTheCombinedBalance(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "SALDO-1")
	e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("ENTRADA", produtoID, 10, "INVENTARIO", nil))
	e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("SAIDA", produtoID, 3, "VENDA", nil))

	w := e.as(s, http.MethodGet, "/api/v1/estoque/produtos/"+produtoID+"/saldo", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if decode(t, w)["saldo"] != float64(7) {
		t.Errorf("saldo = %v, esperado 7", decode(t, w)["saldo"])
	}
}

// AJS-02 AC3: um ator com só movimentacao:read (não saldo:read) é recusado
// com 403 forbidden — permissão própria, mesma lógica de FIN-D-013.
func TestGetSaldoWithOnlyMovimentacaoReadPermissionIsForbidden(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "SALDO-2")
	_, associado := e.signedIn(domain.RoleAssociado)

	w := e.as(associado, http.MethodGet, "/api/v1/estoque/produtos/"+produtoID+"/saldo", nil)

	if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// AJS-02 AC2: produto inexistente responde 404 produto_nao_encontrado.
func TestGetSaldoForAnUnknownProdutoAnswers404(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleDiretoria)

	w := e.as(s, http.MethodGet, "/api/v1/estoque/produtos/00000000-0000-0000-0000-000000000000/saldo", nil)

	if w.Code != http.StatusNotFound || codeOf(t, w) != "produto_nao_encontrado" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}
