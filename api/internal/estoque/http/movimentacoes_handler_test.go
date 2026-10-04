//go:build integration

package estoquehttp

import (
	"net/http"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func (e *env) createProduto(s session, codigo string) string {
	e.t.Helper()
	w := e.as(s, http.MethodPost, "/api/v1/estoque/produtos", createProdutoBody(codigo, codigo, "UN"))
	if w.Code != http.StatusCreated {
		e.t.Fatalf("criar produto: %d %s", w.Code, w.Body.String())
	}
	return decode(e.t, w)["id"].(string)
}

func createMovimentacaoBody(tipo, produtoID string, quantidade int64, origem string, movimentacaoDeID *string) map[string]any {
	body := map[string]any{"tipo": tipo, "produto_id": produtoID, "quantidade": quantidade, "origem": origem}
	if movimentacaoDeID != nil {
		body["movimentacao_de_id"] = *movimentacaoDeID
	}
	return body
}

// MOV-01 AC1: entrada registrada responde 201 com a movimentação.
func TestCreateMovimentacaoEntradaAnswers201(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "ENTRADA-1")

	w := e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("ENTRADA", produtoID, 10, "INVENTARIO", nil))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["tipo"] != "ENTRADA" || body["quantidade"] != float64(10) || body["produto_id"] != produtoID {
		t.Errorf("corpo = %v", body)
	}
}

// MOV-02 AC: saída que deixaria o saldo negativo responde 409 saldo_insuficiente.
func TestCreateMovimentacaoSaidaAboveBalanceAnswers409(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "SAIDA-1")

	w := e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("SAIDA", produtoID, 5, "VENDA", nil))

	if w.Code != http.StatusConflict || codeOf(t, w) != "saldo_insuficiente" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// MOV-03 AC: devolução de uma saída repõe estoque e responde 201.
func TestCreateMovimentacaoDevolucaoOfASaidaAnswers201(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "DEVOLUCAO-1")
	if w := e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("ENTRADA", produtoID, 10, "INVENTARIO", nil)); w.Code != http.StatusCreated {
		t.Fatalf("setup entrada: %d %s", w.Code, w.Body.String())
	}
	saida := decode(t, e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("SAIDA", produtoID, 4, "VENDA", nil)))
	saidaID := saida["id"].(string)

	w := e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("DEVOLUCAO", produtoID, 4, "VENDA", &saidaID))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["movimentacao_de_id"] != saidaID || body["tipo"] != "DEVOLUCAO" {
		t.Errorf("corpo = %v", body)
	}
}

// MOV-03: devolução sem movimentacao_de_id responde 422 devolucao_invalida.
func TestCreateMovimentacaoDevolucaoWithoutReferenceAnswers422(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "DEVOLUCAO-2")

	w := e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("DEVOLUCAO", produtoID, 1, "VENDA", nil))

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "devolucao_invalida" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// API-02 AC5 / MOV-03 AC2 (via HTTP): referência inexistente, de outro SKU,
// a um AJUSTE ou a outra DEVOLUCAO — todas respondem 422 devolucao_invalida.
func TestCreateMovimentacaoDevolucaoWithAnInvalidReferenceAnswers422(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoA := e.createProduto(s, "DEV-INV-A")
	produtoB := e.createProduto(s, "DEV-INV-B")
	entrada := decode(t, e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("ENTRADA", produtoA, 10, "INVENTARIO", nil)))["id"].(string)
	ajuste := decode(t, e.as(s, http.MethodPost, "/api/v1/estoque/ajustes", map[string]any{"produto_id": produtoA, "quantidade": 1, "motivo": "contagem"}))["id"].(string)
	devolucao := decode(t, e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("DEVOLUCAO", produtoA, 1, "COMPRA", &entrada)))["id"].(string)
	inexistente := "00000000-0000-0000-0000-000000000000"

	cases := map[string]struct {
		produtoID string
		ref       string
	}{
		"inexistente":     {produtoA, inexistente},
		"outro_sku":       {produtoB, entrada},
		"ajuste":          {produtoA, ajuste},
		"outra_devolucao": {produtoA, devolucao},
	}
	for name, tc := range cases {
		w := e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("DEVOLUCAO", tc.produtoID, 1, "COMPRA", &tc.ref))
		if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "devolucao_invalida" {
			t.Errorf("%s: status = %d, code = %q", name, w.Code, codeOf(t, w))
		}
	}
}

// Quantidade zero responde 422 quantidade_invalida.
func TestCreateMovimentacaoWithZeroQuantidadeAnswers422(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "ZERO-1")

	w := e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("ENTRADA", produtoID, 0, "INVENTARIO", nil))

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "quantidade_invalida" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// Produto inexistente responde 404 produto_nao_encontrado.
func TestCreateMovimentacaoForAnUnknownProdutoAnswers404(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)

	w := e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("ENTRADA", "00000000-0000-0000-0000-000000000000", 1, "INVENTARIO", nil))

	if w.Code != http.StatusNotFound || codeOf(t, w) != "produto_nao_encontrado" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// MOV-04 AC1: listar devolve todas as movimentações do produto, na ordem.
func TestListMovimentacoesAnswers200WithEveryMovimentacao(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(s, "LISTA-1")
	e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("ENTRADA", produtoID, 10, "INVENTARIO", nil))
	e.as(s, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("SAIDA", produtoID, 3, "VENDA", nil))

	w := e.as(s, http.MethodGet, "/api/v1/estoque/produtos/"+produtoID+"/movimentacoes", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	items := decode(t, w)["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("items = %v, esperado 2", items)
	}
}

// Sem a permissão própria, a resposta é 403 forbidden — DIRETORIA tem
// movimentacao:read mas não movimentacao:create.
func TestMovimentacaoOperationsWithoutThePermissionAre403Forbidden(t *testing.T) {
	e := newEnv(t)
	_, owners := e.signedIn(domain.RoleEstoqueLoja)
	produtoID := e.createProduto(owners, "FORBIDDEN-1")
	_, diretoria := e.signedIn(domain.RoleDiretoria)
	_, associado := e.signedIn(domain.RoleAssociado)

	w := e.as(diretoria, http.MethodPost, "/api/v1/estoque/movimentacoes", createMovimentacaoBody("ENTRADA", produtoID, 1, "INVENTARIO", nil))
	if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
		t.Errorf("create: status = %d, code = %q", w.Code, codeOf(t, w))
	}
	w = e.as(associado, http.MethodGet, "/api/v1/estoque/produtos/"+produtoID+"/movimentacoes", nil)
	if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
		t.Errorf("list: status = %d, code = %q", w.Code, codeOf(t, w))
	}
}
