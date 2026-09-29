package app

import (
	"slices"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func foundation(t *testing.T) domain.Matrix {
	t.Helper()
	m, err := BuildMatrix(FoundationContributions()...)
	if err != nil {
		t.Fatalf("a matriz da fundação deveria ser válida: %v", err)
	}
	return m
}

func perms(m domain.Matrix, r domain.Role) []authz.Permission { return m.Grants[r] }

func TestFoundationMatrixHasTheEightRoles(t *testing.T) {
	m := foundation(t)

	if len(domain.AllRoles) != 8 {
		t.Fatalf("papéis = %d", len(domain.AllRoles))
	}
	for _, r := range domain.AllRoles {
		if _, ok := m.Grants[r]; !ok {
			t.Errorf("o papel %s deveria ter entrada na matriz", r)
		}
		if !r.Valid() || r.Description() == "" {
			t.Errorf("o papel %s deveria ser válido e ter descrição", r)
		}
	}
	if len(m.Grants) != 8 {
		t.Errorf("a matriz tem %d papéis", len(m.Grants))
	}
}

// RBAC-01.5 e RBAC-01.6
func TestConselhoFiscalHasTheInstitutionalPermissionsAndNeverCreateUpdateDeleteCancel(t *testing.T) {
	m := foundation(t)

	want := []authz.Permission{
		"audit:log:read", "financeiro:parecer:opine", "financeiro:prestacao_contas:approve", "financeiro:prestacao_contas:read",
	}
	if got := perms(m, domain.RoleConselhoFiscal); !slices.Equal(got, want) {
		t.Errorf("Conselho Fiscal = %v, esperado %v", got, want)
	}
	for _, p := range perms(m, domain.RoleConselhoFiscal) {
		switch p.Action() {
		case "create", "update", "delete", "cancel":
			t.Errorf("o Conselho Fiscal não pode ter %s", p)
		}
	}
}

// RBAC-01.9. A lista cobre tanto as ações originais quanto as que módulos de
// negócio (ex. financeiro) introduzem para transição de estado e desativação
// — o Conselho Fiscal fiscaliza, nunca opera nenhuma delas.
func TestBuildMatrixRefusesAForbiddenActionForConselhoFiscal(t *testing.T) {
	for _, action := range []string{"create", "update", "delete", "cancel", "deactivate", "receive", "pay"} {
		perm := authz.Permission("financeiro:lancamento:" + action)
		_, err := BuildMatrix(Contribution{
			Module:      "financeiro",
			Permissions: []authz.Definition{{Permission: perm}},
			Grants:      map[domain.Role][]authz.Permission{domain.RoleConselhoFiscal: {perm}},
		})
		if err == nil || !strings.Contains(err.Error(), string(domain.RoleConselhoFiscal)) {
			t.Errorf("%s: esperava recusa citando o Conselho Fiscal, veio %v", action, err)
		}
	}
	if _, err := BuildMatrix(Contribution{
		Module:      "financeiro",
		Permissions: []authz.Definition{{Permission: "financeiro:lancamento:read"}, {Permission: "financeiro:parecer:opine"}},
		Grants:      map[domain.Role][]authz.Permission{domain.RoleConselhoFiscal: {"financeiro:lancamento:read", "financeiro:parecer:opine"}},
	}); err != nil {
		t.Errorf("read e opine são permitidos ao Conselho Fiscal: %v", err)
	}
}

// RBAC-01.10: PRESIDENTE recebe tudo, como vínculos explícitos, sem curinga.
func TestPresidenteGetsEveryPermissionOfTheCatalogEvenFromNewContributions(t *testing.T) {
	extra := Contribution{
		Module:      "estoque",
		Permissions: []authz.Definition{{Permission: "estoque:item:create"}, {Permission: "estoque:item:read", CommonRead: true}},
		Grants:      map[domain.Role][]authz.Permission{domain.RoleEstoqueLoja: {"estoque:item:create"}},
	}
	m, err := BuildMatrix(append(FoundationContributions(), extra)...)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := perms(m, domain.RolePresidente), m.Permissions(); !slices.Equal(got, want) {
		t.Errorf("PRESIDENTE = %v, catálogo = %v", got, want)
	}
	for _, p := range []authz.Permission{"estoque:item:create", "estoque:item:read"} {
		if !m.Has(domain.RolePresidente, p) {
			t.Errorf("PRESIDENTE deveria ter %s como vínculo explícito", p)
		}
	}
	if m.Has(domain.RoleTesouraria, "estoque:item:create") || m.Has(domain.RoleConselhoFiscal, "estoque:item:create") {
		t.Error("a permissão nova só vai para os papéis que a contribuição concede")
	}
	if !m.Has(domain.RoleEstoqueLoja, "estoque:item:create") {
		t.Error("ESTOQUE_LOJA deveria receber o que a contribuição concede")
	}
}

func TestAdminSistemaIsASubsetOfPresidente(t *testing.T) {
	m := foundation(t)

	if len(perms(m, domain.RoleAdminSistema)) == 0 {
		t.Fatal("ADMIN_SISTEMA deveria ter as permissões de administração de acesso")
	}
	for _, p := range perms(m, domain.RoleAdminSistema) {
		if !m.Has(domain.RolePresidente, p) {
			t.Errorf("ADMIN_SISTEMA tem %s, que o PRESIDENTE não tem", p)
		}
	}
}

func TestFoundationProvisionalMatrix(t *testing.T) {
	m := foundation(t)

	if got := perms(m, domain.RoleDiretoria); !slices.Equal(got, []authz.Permission{"identity:user:read"}) {
		t.Errorf("DIRETORIA = %v", got)
	}
	for _, r := range []domain.Role{domain.RoleAssociado, domain.RoleTesouraria, domain.RoleEstoqueLoja, domain.RoleEventos} {
		if got := perms(m, r); len(got) != 0 {
			t.Errorf("%s não tem permissões da fundação, tem %v", r, got)
		}
	}
}

// AD-011: ADMIN_SISTEMA recebe exatamente as permissões técnicas da
// fundação (identity e audit) e nenhuma permissão institucional — nem de
// financeiro, nem de qualquer outro módulo de negócio.
func TestAdminSistemaHasExactlyTheTechnicalPermissionsAndNoInstitutionalOne(t *testing.T) {
	m := foundation(t)
	want := []authz.Permission{
		"audit:log:read", "identity:admin:grant", "identity:admin:revoke", "identity:role:assign",
		"identity:user:create", "identity:user:read", "identity:user:reset_password", "identity:user:update",
	}

	got := perms(m, domain.RoleAdminSistema)

	if !slices.Equal(got, want) {
		t.Errorf("ADMIN_SISTEMA = %v, esperado exatamente %v", got, want)
	}
	for _, p := range got {
		if strings.HasPrefix(string(p), "financeiro:") {
			t.Errorf("ADMIN_SISTEMA não pode ter permissão institucional: %s", p)
		}
	}
}

// PRESIDENTE cobre o catálogo inteiro (identity, audit e financeiro), sem
// wildcard: a abrangência vem de cada permissão declarada estar ligada a ele.
func TestPresidenteCoversTheWholeCatalog(t *testing.T) {
	m := foundation(t)
	want := []authz.Permission{
		"audit:log:read", "financeiro:parecer:opine", "financeiro:prestacao_contas:approve", "financeiro:prestacao_contas:read",
		"identity:admin:grant", "identity:admin:revoke", "identity:role:assign", "identity:user:create",
		"identity:user:read", "identity:user:reset_password", "identity:user:update",
	}

	got := perms(m, domain.RolePresidente)

	if !slices.Equal(got, want) {
		t.Errorf("PRESIDENTE = %v, esperado exatamente %v", got, want)
	}
	var defs []authz.Permission
	for _, d := range m.Definitions {
		defs = append(defs, d.Permission)
	}
	slices.Sort(defs)
	if !slices.Equal(got, defs) {
		t.Errorf("PRESIDENTE deveria cobrir exatamente o catálogo de definições: %v x %v", got, defs)
	}
}

func TestBuildMatrixRefusesInvalidContributions(t *testing.T) {
	cases := map[string]Contribution{
		"módulo vazio":                      {Permissions: []authz.Definition{{Permission: "a:b:read"}}},
		"permissão inválida":                {Module: "a", Permissions: []authz.Definition{{Permission: "invalida"}}},
		"permissão de outro módulo":         {Module: "a", Permissions: []authz.Definition{{Permission: "b:c:read"}}},
		"leitura comum em ação de escrita":  {Module: "a", Permissions: []authz.Definition{{Permission: "a:b:create", CommonRead: true}}},
		"duplicada com definição diferente": {Module: "a", Permissions: []authz.Definition{{Permission: "a:b:read"}, {Permission: "a:b:read", CommonRead: true}}},
		"concessão de permissão não declarada": {
			Module: "a", Permissions: []authz.Definition{{Permission: "a:b:read"}},
			Grants: map[domain.Role][]authz.Permission{domain.RoleEventos: {"a:b:create"}},
		},
		"concessão de permissão de outro módulo": {
			Module: "a", Permissions: []authz.Definition{{Permission: "a:b:read"}},
			Grants: map[domain.Role][]authz.Permission{domain.RoleEventos: {"x:y:read"}},
		},
		"concessão a papel desconhecido": {
			Module: "a", Permissions: []authz.Definition{{Permission: "a:b:read"}},
			Grants: map[domain.Role][]authz.Permission{"CHEFE": {"a:b:read"}},
		},
	}
	for name, c := range cases {
		if _, err := BuildMatrix(c); err == nil {
			t.Errorf("%s: a matriz deveria ser recusada", name)
		}
	}
}

func TestBuildMatrixRefusesTheSamePermissionDeclaredByTwoContributions(t *testing.T) {
	c := Contribution{Module: "a", Permissions: []authz.Definition{{Permission: "a:b:read"}}}

	if _, err := BuildMatrix(c, c); err == nil {
		t.Error("a mesma permissão em duas contribuições deveria ser recusada")
	}
}

func TestBuildMatrixIsDeterministicAndSorted(t *testing.T) {
	a := foundation(t)
	b := foundation(t)

	if !slices.Equal(a.Definitions, b.Definitions) {
		t.Error("as definições deveriam ser idênticas entre execuções")
	}
	if !slices.IsSortedFunc(a.Definitions, func(x, y authz.Definition) int { return strings.Compare(string(x.Permission), string(y.Permission)) }) {
		t.Error("as definições deveriam estar ordenadas")
	}
	for r, ps := range a.Grants {
		if !slices.IsSorted(ps) {
			t.Errorf("as permissões de %s deveriam estar ordenadas: %v", r, ps)
		}
		if !slices.Equal(ps, b.Grants[r]) {
			t.Errorf("as permissões de %s mudaram entre execuções", r)
		}
	}
}

func TestBuildMatrixWithoutContributionsIsEmptyButHasEveryRole(t *testing.T) {
	m, err := BuildMatrix()
	if err != nil || len(m.Definitions) != 0 || len(m.Grants) != 8 {
		t.Errorf("matriz = %+v, err = %v", m, err)
	}
}

func TestFoundationDefinitionsMarkNoPermissionAsCommonRead(t *testing.T) {
	for _, d := range foundation(t).Definitions {
		if d.CommonRead {
			t.Errorf("%s não pode ser leitura comum: administração, RBAC e institucional são sempre auditados", d.Permission)
		}
	}
}

// IDN-08.1: a permissão de redefinição administrativa.
func TestResetPasswordPermissionGoesToAdminSistemaAndPresidenteOnly(t *testing.T) {
	m := foundation(t)
	const perm authz.Permission = "identity:user:reset_password"

	var defined bool
	for _, d := range m.Definitions {
		if d.Permission == perm {
			defined = true
			if d.CommonRead {
				t.Error("não pode ser leitura comum")
			}
		}
	}
	if !defined {
		t.Fatal("a permissão deveria estar declarada")
	}
	for _, r := range domain.AllRoles {
		want := r == domain.RoleAdminSistema || r == domain.RolePresidente
		if m.Has(r, perm) != want {
			t.Errorf("%s: Has = %v, esperado %v", r, m.Has(r, perm), want)
		}
	}
}
