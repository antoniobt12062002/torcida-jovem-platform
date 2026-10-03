//go:build integration

package financeirohttp

import (
	"net/http"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func createContaBody(tipo, nome string) map[string]any {
	return map[string]any{"tipo": tipo, "nome": nome}
}

// API-01 AC1: quem tem financeiro:conta:create cria uma conta e recebe 201.
func TestCreateContaAnswers201WithTheCreatedConta(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)

	w := e.as(s, http.MethodPost, "/api/v1/financeiro/contas", createContaBody("RECEITA", "Mensalidades"))

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	body := decode(t, w)
	if body["nome"] != "Mensalidades" || body["tipo"] != "RECEITA" || body["ativo"] != true {
		t.Errorf("corpo = %v", body)
	}
}

// API-01 AC2 + Independent Test: listar devolve todas as contas, incluindo
// uma subconta do mesmo tipo do pai.
func TestListContasAnswers200WithEveryContaIncludingASubconta(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	root := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/contas", createContaBody("RECEITA", "Receitas")))
	sub := e.as(s, http.MethodPost, "/api/v1/financeiro/contas", map[string]any{
		"tipo": "RECEITA", "nome": "Mensalidades", "parent_id": root["id"],
	})
	if sub.Code != http.StatusCreated {
		t.Fatalf("criar subconta: %d %s", sub.Code, sub.Body.String())
	}
	if got := decode(t, sub)["parent_id"]; got != root["id"] {
		t.Fatalf("parent_id da subconta criada = %v, esperado %v", got, root["id"])
	}

	w := e.as(s, http.MethodGet, "/api/v1/financeiro/contas", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	items := decode(t, w)["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("items = %v, esperado 2", items)
	}
	var subListed bool
	for _, it := range items {
		m := it.(map[string]any)
		if m["id"] == decode(t, sub)["id"] {
			subListed = true
			if m["parent_id"] != root["id"] {
				t.Errorf("parent_id na listagem = %v, esperado %v", m["parent_id"], root["id"])
			}
		}
	}
	if !subListed {
		t.Fatal("subconta não apareceu na listagem")
	}
}

// API-01 AC3: renomear devolve 200 com o novo nome.
func TestRenameContaAnswers200WithTheRenamedConta(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	created := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/contas", createContaBody("DESPESA", "Nome Antigo")))

	w := e.as(s, http.MethodPatch, "/api/v1/financeiro/contas/"+created["id"].(string), map[string]any{"nome": "Nome Novo"})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	if decode(t, w)["nome"] != "Nome Novo" {
		t.Errorf("corpo = %v", decode(t, w))
	}
}

// API-01 AC4: desativar devolve 204 sem corpo.
func TestDeactivateContaAnswers204(t *testing.T) {
	e := newEnv(t)
	_, s := e.signedIn(domain.RoleTesouraria)
	created := decode(t, e.as(s, http.MethodPost, "/api/v1/financeiro/contas", createContaBody("DESPESA", "Para Desativar")))

	w := e.as(s, http.MethodPost, "/api/v1/financeiro/contas/"+created["id"].(string)+"/deactivate", nil)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, corpo = %s", w.Code, w.Body.String())
	}
	list := decode(t, e.as(s, http.MethodGet, "/api/v1/financeiro/contas", nil))
	found := false
	for _, it := range list["items"].([]any) {
		m := it.(map[string]any)
		if m["id"] == created["id"] {
			found = true
			if m["ativo"] != false {
				t.Errorf("ativo = %v, esperado false", m["ativo"])
			}
		}
	}
	if !found {
		t.Fatal("conta desativada não apareceu na listagem (deveria: desativar nunca remove a linha)")
	}
}

// API-01 AC5: sem a permissão, a resposta é 403 forbidden, para cada uma das
// 4 operações — a Diretoria tem conta:read, mas nenhuma das outras três; o
// Associado não tem nenhuma permissão de financeiro, nem conta:read.
func TestContaOperationsWithoutThePermissionAre403Forbidden(t *testing.T) {
	e := newEnv(t)
	_, diretoria := e.signedIn(domain.RoleDiretoria)
	_, associado := e.signedIn(domain.RoleAssociado)
	_, tesouraria := e.signedIn(domain.RoleTesouraria)
	created := decode(t, e.as(tesouraria, http.MethodPost, "/api/v1/financeiro/contas", createContaBody("DESPESA", "Alvo")))
	id := created["id"].(string)

	cases := map[string]*struct {
		session      session
		method, path string
		body         any
	}{
		"create":     {diretoria, http.MethodPost, "/api/v1/financeiro/contas", createContaBody("RECEITA", "Outra")},
		"rename":     {diretoria, http.MethodPatch, "/api/v1/financeiro/contas/" + id, map[string]any{"nome": "Novo"}},
		"deactivate": {diretoria, http.MethodPost, "/api/v1/financeiro/contas/" + id + "/deactivate", nil},
		"list":       {associado, http.MethodGet, "/api/v1/financeiro/contas", nil},
	}
	for name, tc := range cases {
		w := e.as(tc.session, tc.method, tc.path, tc.body)
		if w.Code != http.StatusForbidden || codeOf(t, w) != "forbidden" {
			t.Errorf("%s: status = %d, code = %q", name, w.Code, codeOf(t, w))
		}
	}
}
