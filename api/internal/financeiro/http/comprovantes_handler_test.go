//go:build integration

package financeirohttp

import (
	"net/http"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

// pdfBytes is the same minimal, sniffable PDF fixture platform/documents's
// own tests use (service_test.go) — http.DetectContentType recognizes the
// "%PDF-" prefix as application/pdf.
func pdfBytes() string { return "%PDF-1.4\n% conteúdo de teste\n" }

func (e *env) criarLancamentoDeTeste(t *testing.T, s session) string {
	t.Helper()
	conta := e.criarContaDeTeste(t, s, "DESPESA", "Água")
	created := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/lancamentos", createLancamentoBody("DESPESA", conta["id"].(string), 1000, 0)))
	return created["id"].(string)
}

// API-05 AC1 + Independent Test: anexar, consultar a URL assinada e listar
// — tudo via HTTP.
func TestCreateComprovanteAnswers201AndCanBeRetrievedAndListed(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	lancamentoID := e.criarLancamentoDeTeste(t, s)

	w := e.uploadFile(s, "/api/v1/financeiro/lancamentos/"+lancamentoID+"/comprovantes", "nota.pdf", pdfBytes())
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	created := decode(t, w)
	if created["content_type"] != "application/pdf" || created["original_filename"] != "nota.pdf" {
		t.Errorf("corpo = %v", created)
	}

	urlResp := e.as(s, http.MethodGet, "/api/v1/financeiro/comprovantes/"+created["id"].(string)+"/url", nil)
	if urlResp.Code != http.StatusOK {
		t.Fatalf("url: status = %d, corpo = %s", urlResp.Code, urlResp.Body.String())
	}
	if urlResp.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q, esperado no-store", urlResp.Header().Get("Cache-Control"))
	}
	if decode(t, urlResp)["url"] == "" {
		t.Error("url vazia")
	}

	list := decode(t, e.as(s, http.MethodGet, "/api/v1/financeiro/lancamentos/"+lancamentoID+"/comprovantes", nil))
	items := list["items"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["id"] != created["id"] {
		t.Errorf("items = %v", items)
	}
}

// API-05 AC2: anexar a um lançamento inexistente é 404, antes de qualquer
// chamada a documents.Service.Store.
func TestCreateComprovanteToAMissingLancamentoIs404(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)

	w := e.uploadFile(s, "/api/v1/financeiro/lancamentos/00000000-0000-0000-0000-000000000000/comprovantes", "nota.pdf", pdfBytes())

	if w.Code != http.StatusNotFound || codeOf(t, w) != "lancamento_nao_encontrado" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// API-05 AC5: os erros de platform/documents chegam com o code exato que o
// próprio pacote usa, sem nenhuma regra nova de financeiro — uma extensão
// não aceita responde 422 document_extension_not_allowed.
func TestCreateComprovanteWithADisallowedExtensionIs422WithDocumentsOwnCode(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	lancamentoID := e.criarLancamentoDeTeste(t, s)

	w := e.uploadFile(s, "/api/v1/financeiro/lancamentos/"+lancamentoID+"/comprovantes", "nota.exe", pdfBytes())

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "document_extension_not_allowed" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// Sem a permissão, cada uma das 3 operações é 403 forbidden — permissão
// delegada inteiramente a platform/documents (nenhum Authz em financeiro/app).
func TestComprovanteOperationsWithoutThePermissionAre403Forbidden(t *testing.T) {
	e := newEnv(t)
	_, tesouraria := e.signedIn(domain.RoleTesouraria)
	_, associado := e.signedIn(domain.RoleAssociado)
	lancamentoID := e.criarLancamentoDeTeste(t, tesouraria)
	upload := e.uploadFile(tesouraria, "/api/v1/financeiro/lancamentos/"+lancamentoID+"/comprovantes", "nota.pdf", pdfBytes())
	if upload.Code != http.StatusCreated {
		t.Fatalf("anexar: %d %s", upload.Code, upload.Body.String())
	}
	documentID := decode(t, upload)["id"].(string)

	createW := e.uploadFile(associado, "/api/v1/financeiro/lancamentos/"+lancamentoID+"/comprovantes", "nota.pdf", pdfBytes())
	if createW.Code != http.StatusForbidden || codeOf(t, createW) != "forbidden" {
		t.Errorf("create: status = %d, code = %q", createW.Code, codeOf(t, createW))
	}
	urlW := e.as(associado, http.MethodGet, "/api/v1/financeiro/comprovantes/"+documentID+"/url", nil)
	if urlW.Code != http.StatusForbidden || codeOf(t, urlW) != "forbidden" {
		t.Errorf("url: status = %d, code = %q", urlW.Code, codeOf(t, urlW))
	}
	listW := e.as(associado, http.MethodGet, "/api/v1/financeiro/lancamentos/"+lancamentoID+"/comprovantes", nil)
	if listW.Code != http.StatusForbidden || codeOf(t, listW) != "forbidden" {
		t.Errorf("list: status = %d, code = %q", listW.Code, codeOf(t, listW))
	}
}
