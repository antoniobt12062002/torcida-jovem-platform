//go:build integration

package httpapi

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// uploadFile sends a real multipart/form-data request (field "file")
// through the real router — financeiro's one such route.
func (e *e2e) uploadFile(s *sess, path, filename, content string) *httptest.ResponseRecorder {
	e.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		e.t.Fatal(err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		e.t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		e.t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Origin", appOrigin)
	if s != nil {
		if s.cookie != nil {
			req.AddCookie(s.cookie)
		}
		if s.csrf != "" {
			req.Header.Set("X-CSRF-Token", s.csrf)
		}
	}
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	e.contract.ValidateResponse(e.t, req, w)
	return w
}

// 06-api-http/T8: one true end-to-end test per resource, through the real
// httpapi.NewRouter (financeirohttp.Register, the real chains) — distinct
// from T4-T7's own tests, which exercise the handlers directly through the
// manual harness of financeiro/http/env_test.go (FIN-D-029), never through
// this real wiring. PRESIDENTE holds every permission automatically.

// API-01: contas — criar uma raiz e uma subconta, listar as duas, renomear,
// desativar.
func TestFinanceiroE2EContasFlow(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()

	root := e.do(s, http.MethodPost, "/api/v1/financeiro/contas", map[string]any{"tipo": "RECEITA", "nome": "Receitas"})
	e.expect(root, http.StatusCreated, "")
	rootID := body(t, root)["id"].(string)

	sub := e.do(s, http.MethodPost, "/api/v1/financeiro/contas", map[string]any{"tipo": "RECEITA", "nome": "Mensalidades", "parent_id": rootID})
	e.expect(sub, http.StatusCreated, "")
	subID := body(t, sub)["id"].(string)

	list := e.do(s, http.MethodGet, "/api/v1/financeiro/contas", nil)
	e.expect(list, http.StatusOK, "")
	if items := body(t, list)["items"].([]any); len(items) != 2 {
		t.Fatalf("items = %v, esperado 2", items)
	}

	rename := e.do(s, http.MethodPatch, "/api/v1/financeiro/contas/"+subID, map[string]any{"nome": "Mensalidades 2026"})
	e.expect(rename, http.StatusOK, "")
	if body(t, rename)["nome"] != "Mensalidades 2026" {
		t.Errorf("nome = %v", body(t, rename)["nome"])
	}

	deactivate := e.do(s, http.MethodPost, "/api/v1/financeiro/contas/"+rootID+"/deactivate", nil)
	e.expect(deactivate, http.StatusNoContent, "")
}

// API-02: lançamentos — criar, editar enquanto CRIADA, listar.
func TestFinanceiroE2ELancamentosFlow(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()
	conta := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/contas", map[string]any{"tipo": "DESPESA", "nome": "Água"}))

	created := e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos", map[string]any{
		"tipo": "DESPESA", "conta_id": conta["id"], "valor_bruto_cents": 5000, "taxa_cents": 100, "forma_pagamento": "PIX",
	})
	e.expect(created, http.StatusCreated, "")
	id := body(t, created)["id"].(string)
	if body(t, created)["valor_liquido_cents"] != float64(4900) {
		t.Errorf("valor_liquido_cents = %v", body(t, created)["valor_liquido_cents"])
	}

	edited := e.do(s, http.MethodPut, "/api/v1/financeiro/lancamentos/"+id, map[string]any{
		"conta_id": conta["id"], "valor_bruto_cents": 6000, "taxa_cents": 100, "forma_pagamento": "PIX",
	})
	e.expect(edited, http.StatusOK, "")
	if body(t, edited)["valor_bruto_cents"] != float64(6000) {
		t.Errorf("valor_bruto_cents = %v", body(t, edited)["valor_bruto_cents"])
	}

	list := e.do(s, http.MethodGet, "/api/v1/financeiro/lancamentos", nil)
	e.expect(list, http.StatusOK, "")
	if items := body(t, list)["items"].([]any); len(items) != 1 {
		t.Fatalf("items = %v, esperado 1", items)
	}
}

// API-03: workflow — receber, pagar, cancelar (depois de liquidado, FIN-D-006).
func TestFinanceiroE2EWorkflowFlow(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()
	contaReceita := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/contas", map[string]any{"tipo": "RECEITA", "nome": "Mensalidades"}))
	contaDespesa := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/contas", map[string]any{"tipo": "DESPESA", "nome": "Água"}))
	receita := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos", map[string]any{
		"tipo": "RECEITA", "conta_id": contaReceita["id"], "valor_bruto_cents": 10000, "taxa_cents": 0, "forma_pagamento": "PIX",
	}))
	despesa := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos", map[string]any{
		"tipo": "DESPESA", "conta_id": contaDespesa["id"], "valor_bruto_cents": 3000, "taxa_cents": 0, "forma_pagamento": "PIX",
	}))

	e.expect(e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/receive", nil), http.StatusNoContent, "")
	e.expect(e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+despesa["id"].(string)+"/pay", nil), http.StatusNoContent, "")
	e.expect(e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/cancel", map[string]any{"reason": "lançamento duplicado"}), http.StatusNoContent, "")
}

// API-04: saldo — zero lançamentos é 0; depois de receber, reflete o líquido.
func TestFinanceiroE2ESaldoReflectsReceivedLancamentos(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()

	e.expect(e.do(s, http.MethodGet, "/api/v1/financeiro/saldo", nil), http.StatusOK, "")

	conta := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/contas", map[string]any{"tipo": "RECEITA", "nome": "Mensalidades"}))
	lancamento := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos", map[string]any{
		"tipo": "RECEITA", "conta_id": conta["id"], "valor_bruto_cents": 10000, "taxa_cents": 300, "forma_pagamento": "PIX",
	}))
	e.expect(e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+lancamento["id"].(string)+"/receive", nil), http.StatusNoContent, "")

	saldo := e.do(s, http.MethodGet, "/api/v1/financeiro/saldo", nil)
	e.expect(saldo, http.StatusOK, "")
	if body(t, saldo)["saldo_cents"] != float64(9700) {
		t.Errorf("saldo_cents = %v, esperado 9700", body(t, saldo)["saldo_cents"])
	}
}

// API-05: comprovantes — anexar, consultar a URL assinada, listar.
func TestFinanceiroE2EComprovantesFlow(t *testing.T) {
	e := newE2E(t)
	s := e.adminSession()
	conta := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/contas", map[string]any{"tipo": "DESPESA", "nome": "Água"}))
	lancamento := body(t, e.do(s, http.MethodPost, "/api/v1/financeiro/lancamentos", map[string]any{
		"tipo": "DESPESA", "conta_id": conta["id"], "valor_bruto_cents": 1000, "taxa_cents": 0, "forma_pagamento": "PIX",
	}))
	lancamentoID := lancamento["id"].(string)

	created := e.uploadFile(s, "/api/v1/financeiro/lancamentos/"+lancamentoID+"/comprovantes", "nota.pdf", "%PDF-1.4\n% conteúdo de teste\n")
	e.expect(created, http.StatusCreated, "")
	documentID := body(t, created)["id"].(string)

	url := e.do(s, http.MethodGet, "/api/v1/financeiro/comprovantes/"+documentID+"/url", nil)
	e.expect(url, http.StatusOK, "")
	if url.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q", url.Header().Get("Cache-Control"))
	}

	list := e.do(s, http.MethodGet, "/api/v1/financeiro/lancamentos/"+lancamentoID+"/comprovantes", nil)
	e.expect(list, http.StatusOK, "")
	if items := body(t, list)["items"].([]any); len(items) != 1 {
		t.Fatalf("items = %v, esperado 1", items)
	}
}
