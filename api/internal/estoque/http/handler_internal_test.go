package estoquehttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// A tabela única de erros (EST-D-006 + achado mecânico de ErrCampoObrigatorio,
// API-04): cada um dos 8 mapeamentos de domínio/autorização vira o status e
// o code exatos do contrato, embrulhado ou não, mais os 2 mapeamentos
// transversais (service_unavailable, audit_failed) já cobertos pelo switch
// de writeError.
func TestEveryDomainErrorMapsToItsStatusAndCode(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{domain.ErrProdutoNaoEncontrado, http.StatusNotFound, "produto_nao_encontrado"},
		{domain.ErrCodigoDuplicado, http.StatusConflict, "codigo_duplicado"},
		{domain.ErrCampoObrigatorio, http.StatusUnprocessableEntity, "campo_obrigatorio"},
		{domain.ErrDevolucaoInvalida, http.StatusUnprocessableEntity, "devolucao_invalida"},
		{domain.ErrSaldoInsuficiente, http.StatusConflict, "saldo_insuficiente"},
		{domain.ErrQuantidadeInvalida, http.StatusUnprocessableEntity, "quantidade_invalida"},
		{domain.ErrMotivoObrigatorio, http.StatusUnprocessableEntity, "motivo_obrigatorio"},
		{authz.ErrForbidden, http.StatusForbidden, "forbidden"},
		{database.ErrUnavailable, http.StatusServiceUnavailable, "service_unavailable"},
		{fmt.Errorf("%w: disco cheio", audit.ErrWrite), http.StatusInternalServerError, "audit_failed"},
	}
	if len(cases) != 10 {
		t.Fatalf("8 mapeamentos de domínio/autorização + 2 transversais, a tabela do teste tem %d", len(cases))
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

// Nenhum erro de estoque/domain usa seu .Error() (prosa em português) como
// code — a prova é que os codes do contrato são todos snake_case em inglês,
// nunca a mensagem original do sentinel.
func TestDomainSentinelsNeverUseTheirErrorStringAsCode(t *testing.T) {
	prose := []error{
		domain.ErrProdutoNaoEncontrado, domain.ErrCodigoDuplicado, domain.ErrCampoObrigatorio, domain.ErrDevolucaoInvalida,
		domain.ErrSaldoInsuficiente, domain.ErrQuantidadeInvalida, domain.ErrMotivoObrigatorio,
	}
	for _, s := range sentinels {
		for _, p := range prose {
			if errors.Is(s.err, p) && s.code == p.Error() {
				t.Errorf("code de %v não pode ser o .Error() do sentinel (%q)", p, p.Error())
			}
		}
	}
}

// domain.ErrMovimentacaoNaoEncontrada nunca deveria chegar aqui (já traduzido
// para ErrDevolucaoInvalida por RegistrarMovimentacao.Execute); se chegasse,
// cai no 500 internal_error do default, como qualquer erro não mapeado.
func TestMovimentacaoNaoEncontradaFallsBackToInternalError(t *testing.T) {
	h := &Handler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	h.writeError(c, domain.ErrMovimentacaoNaoEncontrada)

	var p struct{ Code string }
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	if w.Code != http.StatusInternalServerError || p.Code != "internal_error" {
		t.Errorf("status = %d, code = %q, esperado 500 internal_error", w.Code, p.Code)
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

// Uma requisição sem sessão no contexto responde 401 unauthenticated, antes
// de qualquer tabela de erro de domínio.
func TestNoSessionInContextAnswersUnauthenticated(t *testing.T) {
	h := &Handler{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	h.writeError(c, errNoSession)

	var p struct{ Code string }
	_ = json.Unmarshal(w.Body.Bytes(), &p)
	if w.Code != http.StatusUnauthorized || p.Code != "unauthenticated" {
		t.Errorf("status = %d, code = %q, esperado 401 unauthenticated", w.Code, p.Code)
	}
}

func TestNewBuildsAHandlerWithTheGivenModuleAndLogger(t *testing.T) {
	h := New(nil, nil)
	if h == nil {
		t.Fatal("New retornou nil")
	}
}
