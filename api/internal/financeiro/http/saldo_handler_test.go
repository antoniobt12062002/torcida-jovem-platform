//go:build integration

package financeirohttp

import (
	"net/http"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

// API-04 AC1 + Independent Test: zero lançamentos é saldo zero; depois de
// criar e receber uma receita, o saldo reflete o valor líquido.
func TestGetSaldoAnswers200WithTheExactValueIncludingZero(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)

	w0 := e.as(s, http.MethodGet, "/api/v1/financeiro/saldo", nil)
	if w0.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w0.Code, w0.Body.String())
	}
	if decode(t, w0)["saldo_cents"] != float64(0) {
		t.Fatalf("saldo inicial = %v, esperado 0", decode(t, w0)["saldo_cents"])
	}

	conta := e.criarContaDeTeste(t, s, "RECEITA", "Mensalidades")
	receita := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("RECEITA", conta["id"].(string), 10000, 300)))
	if w := e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos/"+receita["id"].(string)+"/receive", nil); w.Code != http.StatusNoContent {
		t.Fatalf("receber: %d %s", w.Code, w.Body.String())
	}

	w1 := e.as(s, http.MethodGet, "/api/v1/financeiro/saldo", nil)
	if w1.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w1.Code, w1.Body.String())
	}
	if decode(t, w1)["saldo_cents"] != float64(9700) {
		t.Errorf("saldo = %v, esperado 9700 (10000-300)", decode(t, w1)["saldo_cents"])
	}
}

// Sem financeiro:saldo:read, a resposta é 403 forbidden — permissão própria,
// distinta de lancamento:read (FIN-D-013): o Associado não tem nenhuma.
func TestGetSaldoWithoutThePermissionIs403Forbidden(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleAssociado)

	w := e.as(s, http.MethodGet, "/api/v1/financeiro/saldo", nil)

	if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}
