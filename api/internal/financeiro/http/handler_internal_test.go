package financeirohttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/money"
)

// A tabela única de erros (FIN-D-022): cada um dos 21 mapeamentos vira o
// status e o code exatos do contrato, embrulhado ou não.
func TestEveryDomainErrorMapsToItsStatusAndCode(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrContaNaoEncontrada, http.StatusNotFound, "conta_nao_encontrada"},
		{domain.ErrContaTipoIncompativel, http.StatusUnprocessableEntity, "conta_tipo_incompativel"},
		{domain.ErrContaJaUtilizada, http.StatusConflict, "conta_ja_utilizada"},
		{domain.ErrContaInvalida, http.StatusUnprocessableEntity, "conta_invalida"},
		{domain.ErrLancamentoTipoIncompativel, http.StatusUnprocessableEntity, "lancamento_tipo_incompativel"},
		{domain.ErrLancamentoNaoEncontrado, http.StatusNotFound, "lancamento_nao_encontrado"},
		{domain.ErrLancamentoImutavel, http.StatusConflict, "lancamento_imutavel"},
		{domain.ErrDevolucaoInvalida, http.StatusUnprocessableEntity, "devolucao_invalida"},
		{domain.ErrLancamentoNaoPodeSerRecebido, http.StatusConflict, "lancamento_nao_pode_ser_recebido"},
		{domain.ErrLancamentoNaoPodeSerPago, http.StatusConflict, "lancamento_nao_pode_ser_pago"},
		{domain.ErrLancamentoJaCancelado, http.StatusConflict, "lancamento_ja_cancelado"},
		{domain.ErrMotivoObrigatorio, http.StatusUnprocessableEntity, "motivo_obrigatorio"},
		{money.ErrOutOfRange, http.StatusUnprocessableEntity, "amount_out_of_range"},
		{authz.ErrForbidden, http.StatusForbidden, "forbidden"},
		{documents.ErrExtensionNotAllowed, http.StatusUnprocessableEntity, "document_extension_not_allowed"},
		{documents.ErrTypeNotAllowed, http.StatusUnprocessableEntity, "document_type_not_allowed"},
		{documents.ErrTypeMismatch, http.StatusUnprocessableEntity, "document_type_mismatch"},
		{documents.ErrTooLarge, http.StatusRequestEntityTooLarge, "document_too_large"},
		{documents.ErrNotFound, http.StatusNotFound, "document_not_found"},
		{database.ErrUnavailable, http.StatusServiceUnavailable, "service_unavailable"},
		{fmt.Errorf("%w: disco cheio", audit.ErrWrite), http.StatusInternalServerError, "audit_failed"},
	}
	if len(cases) != 21 {
		t.Fatalf("FIN-D-022 define 21 mapeamentos, a tabela do teste tem %d", len(cases))
	}
	h := &Handler{}
	for _, tc := range cases {
		for _, err := range []error{tc.err, fmt.Errorf("caso de uso: %w", tc.err)} {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			h.writeError(c, err)

			var p struct{ Code string }
			_ = json.Unmarshal(w.Body.Bytes(), &p)
			if w.Code != tc.status || p.Code != tc.code || w.Header().Get("Content-Type") != "application/problem+json" {
				t.Errorf("%v: status = %d, code = %q, esperado %d %q", err, w.Code, p.Code, tc.status, tc.code)
			}
		}
	}
}

// documents.ErrOwnerTypeInvalid não deveria ocorrer na prática (FIN-D-022); se
// ocorrer, cai no 500 internal_error do default, como qualquer erro não mapeado.
func TestOwnerTypeInvalidFallsBackToInternalError(t *testing.T) {
	h := &Handler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	h.writeError(c, documents.ErrOwnerTypeInvalid)

	var p struct{ Code string }
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	if w.Code != http.StatusInternalServerError || p.Code != "internal_error" {
		t.Errorf("status = %d, code = %q, esperado 500 internal_error", w.Code, p.Code)
	}
}

// Nenhum erro de financeiro/domain usa seu .Error() (prosa em português) como
// code — a prova é que os codes do contrato são todos snake_case em inglês,
// nunca a mensagem original do sentinel.
func TestDomainSentinelsNeverUseTheirErrorStringAsCode(t *testing.T) {
	prose := []error{
		domain.ErrContaNaoEncontrada, domain.ErrContaTipoIncompativel, domain.ErrContaJaUtilizada,
		domain.ErrContaInvalida, domain.ErrLancamentoTipoIncompativel, domain.ErrLancamentoNaoEncontrado,
		domain.ErrLancamentoImutavel, domain.ErrDevolucaoInvalida, domain.ErrLancamentoNaoPodeSerRecebido,
		domain.ErrLancamentoNaoPodeSerPago, domain.ErrLancamentoJaCancelado, domain.ErrMotivoObrigatorio,
	}
	for _, s := range sentinels {
		for _, p := range prose {
			if errors.Is(s.err, p) && s.code == p.Error() {
				t.Errorf("code de %v não pode ser o .Error() do sentinel (%q)", p, p.Error())
			}
		}
	}
}

// Um erro desconhecido (nunca visto pela tabela) cai no 500 internal_error e
// não vaza detalhe interno no corpo da resposta.
func TestUnknownErrorFallsBackToInternalErrorWithoutLeakingDetail(t *testing.T) {
	h := &Handler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	h.writeError(c, errors.New("segredo-interno: falha do banco"))

	var p struct{ Code string }
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	if w.Code != http.StatusInternalServerError || p.Code != "internal_error" {
		t.Errorf("status = %d, code = %q", w.Code, p.Code)
	}
	if strings.Contains(w.Body.String(), "segredo-interno") {
		t.Errorf("o corpo não pode vazar detalhe interno: %s", w.Body.String())
	}
}

func TestNewBuildsAHandlerWithTheGivenModuleAndLogger(t *testing.T) {
	h := New(nil, nil)
	if h == nil {
		t.Fatal("New retornou nil")
	}
}
