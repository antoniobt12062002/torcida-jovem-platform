//go:build integration

package estoquehttp

import (
	"net/http"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func createProdutoBody(codigo, nome, unidade string) map[string]any {
	return map[string]any{"codigo": codigo, "nome": nome, "unidade_medida": unidade}
}

// PRD-01 AC1: quem tem estoque:produto:create cria um produto e recebe 201.
func TestCreateProdutoAnswers201WithTheCreatedProduto(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)

	w := e.as(s, http.MethodPost, "/api/v1/estoque/produtos", createProdutoBody("CAMISA-TJ-P", "Camisa P", "UN"))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["codigo"] != "CAMISA-TJ-P" || body["nome"] != "Camisa P" || body["unidade_medida"] != "UN" {
		t.Errorf("corpo = %v", body)
	}
}

// PRD-01 AC3 (achado mecânico de T4): um campo só com espaços passa pelo
// minLength:1 do contrato mas é recusado pelo caso de uso com
// ErrCampoObrigatorio — responde 422 campo_obrigatorio, nunca 500 (a lacuna
// que existiria sem a linha adicionada a handler.go's sentinels nesta task).
func TestCreateProdutoWithAWhitespaceOnlyFieldAnswers422CampoObrigatorio(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)

	w := e.as(s, http.MethodPost, "/api/v1/estoque/produtos", createProdutoBody("   ", "Nome Válido", "UN"))

	if w.Code != http.StatusUnprocessableEntity || codeOf(t, w) != "campo_obrigatorio" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// PRD-01 AC3: código já usado por outro produto responde 409 codigo_duplicado.
func TestCreateProdutoWithARepeatedCodigoAnswers409(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	if w := e.as(s, http.MethodPost, "/api/v1/estoque/produtos", createProdutoBody("CAMISA-TJ-M", "Camisa M", "UN")); w.Code != http.StatusCreated {
		t.Fatalf("setup: %d %s", w.Code, w.Body.String())
	}

	w := e.as(s, http.MethodPost, "/api/v1/estoque/produtos", createProdutoBody("CAMISA-TJ-M", "Outra Camisa M", "UN"))

	if w.Code != http.StatusConflict || codeOf(t, w) != "codigo_duplicado" {
		t.Errorf("status = %d, code = %q", w.Code, codeOf(t, w))
	}
}

// PRD-02 AC1: listar devolve todos os produtos, sem paginação.
func TestListProdutosAnswers200WithEveryProduto(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleEstoqueLoja)
	created := decode(t, e.as(s, http.MethodPost, "/api/v1/estoque/produtos", createProdutoBody("BONE-TJ", "Boné", "UN")))

	w := e.as(s, http.MethodGet, "/api/v1/estoque/produtos", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	items := decode(t, w)["items"].([]any)
	var found bool
	for _, it := range items {
		if it.(map[string]any)["id"] == created["id"] {
			found = true
		}
	}
	if !found {
		t.Fatal("produto criado não apareceu na listagem")
	}
}

// PRD-01/PRD-02: sem a permissão própria, a resposta é 403 forbidden —
// DIRETORIA tem produto:read mas não produto:create; ASSOCIADO não tem nenhuma.
func TestProdutoOperationsWithoutThePermissionAre403Forbidden(t *testing.T) {
	e := newEnv(t)
	_, diretoria := e.signedIn(domain.RoleDiretoria)
	_, associado := e.signedIn(domain.RoleAssociado)

	cases := map[string]*struct {
		session      session
		method, path string
		body         any
	}{
		"create": {diretoria, http.MethodPost, "/api/v1/estoque/produtos", createProdutoBody("X", "X", "UN")},
		"list":   {associado, http.MethodGet, "/api/v1/estoque/produtos", nil},
	}
	for name, tc := range cases {
		w := e.as(tc.session, tc.method, tc.path, tc.body)
		if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
			t.Errorf("%s: status = %d, code = %q", name, w.Code, codeOf(t, w))
		}
	}
}
