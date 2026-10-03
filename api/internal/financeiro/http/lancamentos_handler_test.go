//go:build integration

package financeirohttp

import (
	"net/http"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func (e *env) criarContaDeTeste(t *testing.T, s session, tipo, nome string) map[string]any {
	t.Helper()
	w := e.as(s, http.MethodPost, "/api/v1/financeiro/contas", createContaBody(tipo, nome))
	if w.Code != http.StatusCreated {
		t.Fatalf("criar conta de teste: %d %s", w.Code, w.Body.String())
	}
	return decode(t, w)
}

func lancamentoBody(contaID string, bruto, taxa int64) map[string]any {
	return map[string]any{"conta_id": contaID, "valor_bruto_cents": bruto, "taxa_cents": taxa, "forma_pagamento": "PIX"}
}

func createLancamentoBody(tipo, contaID string, bruto, taxa int64) map[string]any {
	m := lancamentoBody(contaID, bruto, taxa)
	m["tipo"] = tipo
	return m
}

// withDevolucaoDeID completes a CreateDevolucaoRequest body: the contract
// requires devolucao_de_id even when the test only cares that the request
// never reaches that far (e.g. a 403 forbidden check) — request-shape
// validation runs before any permission check.
func withDevolucaoDeID(m map[string]any, devolucaoDeID string) map[string]any {
	m["devolucao_de_id"] = devolucaoDeID
	return m
}

// API-02 AC1: criar um lançamento devolve 201 com valor_liquido_cents calculado.
func TestCreateLancamentoAnswers201WithComputedValorLiquido(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	conta := e.criarContaDeTeste(t, s, "RECEITA", "Mensalidades")

	w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("RECEITA", conta["id"].(string), 10000, 300))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["valor_bruto_cents"] != float64(10000) || body["taxa_cents"] != float64(300) || body["valor_liquido_cents"] != float64(9700) {
		t.Errorf("corpo = %v", body)
	}
	if body["status"] != "CRIADA" {
		t.Errorf("status = %v, esperado CRIADA", body["status"])
	}
}

// API-02 AC2: editar enquanto CRIADA substitui todos os campos editáveis e devolve 200.
func TestUpdateLancamentoAnswers200ReplacingEveryEditableField(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	contaA := e.criarContaDeTeste(t, s, "DESPESA", "Conta A")
	contaB := e.criarContaDeTeste(t, s, "DESPESA", "Conta B")
	created := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", contaA["id"].(string), 5000, 100)))

	w := e.as(s, http.MethodPut, "/api/v1/financeiro/lancamentos/"+created["id"].(string), lancamentoBody(contaB["id"].(string), 8000, 200))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["conta_id"] != contaB["id"] || body["valor_bruto_cents"] != float64(8000) || body["valor_liquido_cents"] != float64(7800) {
		t.Errorf("corpo = %v", body)
	}
}

// API-02 AC3: devolução cria uma despesa vinculada e devolve 201.
func TestCreateDevolucaoAnswers201LinkedToTheOriginalReceita(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	contaReceita := e.criarContaDeTeste(t, s, "RECEITA", "Mensalidades")
	contaDespesa := e.criarContaDeTeste(t, s, "DESPESA", "Devoluções")
	receita := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("RECEITA", contaReceita["id"].(string), 10000, 0)))
	recv := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/receive", nil)
	if recv.Code != http.StatusNoContent {
		t.Fatalf("receber receita: %d %s", recv.Code, recv.Body.String())
	}

	body := lancamentoBody(contaDespesa["id"].(string), 10000, 0)
	body["devolucao_de_id"] = receita["id"]
	w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/devolucoes", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	out := decode(t, w)
	if out["tipo"] != "DESPESA" || out["devolucao_de_id"] != receita["id"] {
		t.Errorf("corpo = %v", out)
	}
}

// API-02 AC4: listar devolve todos os lançamentos, incluindo um CANCELADA.
func TestListLancamentosAnswers200WithEveryLancamentoIncludingCancelada(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	conta := e.criarContaDeTeste(t, s, "DESPESA", "Água")
	l1 := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", conta["id"].(string), 1000, 0)))
	l2 := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", conta["id"].(string), 2000, 0)))
	cancel := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+l2["id"].(string)+"/cancel", map[string]any{"reason": "duplicado"})
	if cancel.Code != http.StatusNoContent {
		t.Fatalf("cancelar: %d %s", cancel.Code, cancel.Body.String())
	}

	w := e.as(s, http.MethodGet, "/api/v1/financeiro/lancamentos", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	items := decode(t, w)["items"].([]any)
	var sawCancelada bool
	ids := map[string]bool{}
	for _, it := range items {
		m := it.(map[string]any)
		ids[m["id"].(string)] = true
		if m["id"] == l2["id"] {
			if m["status"] != "CANCELADA" {
				t.Errorf("status do cancelado = %v", m["status"])
			}
			sawCancelada = true
		}
	}
	if !ids[l1["id"].(string)] || !sawCancelada {
		t.Errorf("items = %v", items)
	}
}

// API-02 AC5: editar um lançamento que não está mais CRIADA é 409 lancamento_imutavel.
func TestUpdateLancamentoAfterReceiveIs409LancamentoImutavel(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	conta := e.criarContaDeTeste(t, s, "RECEITA", "Mensalidades")
	created := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("RECEITA", conta["id"].(string), 10000, 0)))
	recv := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+created["id"].(string)+"/receive", nil)
	if recv.Code != http.StatusNoContent {
		t.Fatalf("receber: %d %s", recv.Code, recv.Body.String())
	}

	w := e.as(s, http.MethodPut, "/api/v1/financeiro/lancamentos/"+created["id"].(string), lancamentoBody(conta["id"].(string), 9999, 0))

	if w.Code != http.StatusConflict || codeOf(t, w) != "lancamento_imutavel" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// API-03 AC1/AC2/AC3 + Independent Test: criar uma receita, recebê-la,
// cancelá-la com motivo depois de recebida (permitido, FIN-D-006).
func TestReceivePayAndCancelAnswer204(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	contaReceita := e.criarContaDeTeste(t, s, "RECEITA", "Mensalidades")
	contaDespesa := e.criarContaDeTeste(t, s, "DESPESA", "Água")
	receita := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("RECEITA", contaReceita["id"].(string), 10000, 0)))
	despesa := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", contaDespesa["id"].(string), 3000, 0)))

	if w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/receive", nil); w.Code != http.StatusNoContent {
		t.Fatalf("receive: %d %s", w.Code, w.Body.String())
	}
	if w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+despesa["id"].(string)+"/pay", nil); w.Code != http.StatusNoContent {
		t.Fatalf("pay: %d %s", w.Code, w.Body.String())
	}
	w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/cancel", map[string]any{"reason": "lançamento duplicado"})
	if w.Code != http.StatusNoContent {
		t.Fatalf("cancel depois de recebida: %d %s", w.Code, w.Body.String())
	}
}

// API-03 AC4: cancelar sem motivo é 422 motivo_obrigatorio, sem chamar a
// persistência (o lançamento continua intacto).
func TestCancelLancamentoWithoutReasonIs422AndNeverPersists(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	conta := e.criarContaDeTeste(t, s, "DESPESA", "Água")
	created := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", conta["id"].(string), 1000, 0)))

	w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+created["id"].(string)+"/cancel", map[string]any{"reason": ""})

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "motivo_obrigatorio" {
		t.Fatalf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
	list := decode(t, e.as(s, http.MethodGet, "/api/v1/financeiro/lancamentos", nil))
	for _, it := range list["items"].([]any) {
		m := it.(map[string]any)
		if m["id"] == created["id"] && m["status"] != "CRIADA" {
			t.Errorf("status = %v, esperado CRIADA (nada deveria ter sido persistido)", m["status"])
		}
	}
}

// API-03 AC5: receber/pagar no estado errado e cancelar já cancelado
// devolvem 409 com o code correspondente.
func TestWorkflowTransitionsInTheWrongStateAre409WithTheMatchingCode(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	contaReceita := e.criarContaDeTeste(t, s, "RECEITA", "Mensalidades")
	contaDespesa := e.criarContaDeTeste(t, s, "DESPESA", "Água")
	receita := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("RECEITA", contaReceita["id"].(string), 10000, 0)))
	despesa := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", contaDespesa["id"].(string), 3000, 0)))

	if w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+despesa["id"].(string)+"/receive", nil); w.Code != http.StatusConflict || codeOf(t, w) != "lancamento_nao_pode_ser_recebido" {
		t.Errorf("receive de despesa: status = %d, code = %q", w.Code, codeOf(t, w))
	}
	if w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/pay", nil); w.Code != http.StatusConflict || codeOf(t, w) != "lancamento_nao_pode_ser_pago" {
		t.Errorf("pay de receita: status = %d, code = %q", w.Code, codeOf(t, w))
	}
	cancel1 := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/cancel", map[string]any{"reason": "motivo"})
	if cancel1.Code != http.StatusNoContent {
		t.Fatalf("primeiro cancelamento: %d %s", cancel1.Code, cancel1.Body.String())
	}
	cancel2 := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/cancel", map[string]any{"reason": "de novo"})
	if cancel2.Code != http.StatusConflict || codeOf(t, cancel2) != "lancamento_ja_cancelado" {
		t.Errorf("segundo cancelamento: status = %d, code = %q", cancel2.Code, codeOf(t, cancel2))
	}
}

// Sem a permissão, cada uma das 7 operações é 403 forbidden — a Diretoria
// tem lancamento:read, mas nenhuma das outras; o Associado não tem nenhuma.
func TestLancamentoOperationsWithoutThePermissionAre403Forbidden(t *testing.T) {
	e := newEnv(t)
	_, diretoria := e.signedIn(domain.RoleDiretoria)
	_, associado := e.signedIn(domain.RoleAssociado)
	_, tesouraria := e.signedIn(domain.RoleTesouraria)
	conta := e.criarContaDeTeste(t, tesouraria, "DESPESA", "Água")
	created := decode(t, e.as(tesouraria, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", conta["id"].(string), 1000, 0)))
	id := created["id"].(string)

	cases := map[string]*struct {
		session      session
		method, path string
		body         any
	}{
		"create":    {diretoria, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", conta["id"].(string), 1000, 0)},
		"update":    {diretoria, http.MethodPut, "/api/v1/financeiro/lancamentos/" + id, lancamentoBody(conta["id"].(string), 1000, 0)},
		"devolucao": {diretoria, http.MethodPost, "/api/v1/financeiro/lancamentos/devolucoes", withDevolucaoDeID(lancamentoBody(conta["id"].(string), 1000, 0), id)},
		"receive":   {diretoria, http.MethodPost, "/api/v1/financeiro/lancamentos/" + id + "/receive", nil},
		"pay":       {diretoria, http.MethodPost, "/api/v1/financeiro/lancamentos/" + id + "/pay", nil},
		"cancel":    {diretoria, http.MethodPost, "/api/v1/financeiro/lancamentos/" + id + "/cancel", map[string]any{"reason": "x"}},
		"list":      {associado, http.MethodGet, "/api/v1/financeiro/lancamentos", nil},
	}
	for name, tc := range cases {
		w := e.as(tc.session, tc.method, tc.path, tc.body)
		if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
			t.Errorf("%s: status = %d, code = %q", name, w.Code, codeOf(t, w))
		}
	}
}
