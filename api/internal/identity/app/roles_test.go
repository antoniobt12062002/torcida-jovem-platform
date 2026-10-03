package app

import (
	"reflect"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

// acceptsDomainRole only exists so the test below can pass a Role value where
// a domain.Role parameter is expected: compiles only if the two are the exact
// same type, since a distinct defined type would require an explicit
// conversion at the call site.
func acceptsDomainRole(r domain.Role) domain.Role { return r }

// Role must be the exact same type as domain.Role (a type alias), never a second
// definition: other modules build Contribution.Grants (map[domain.Role]...) with
// this type, so any mismatch would fail to compile at the call site.
func TestRoleIsAnAliasOfDomainRoleNotANewType(t *testing.T) {
	d := acceptsDomainRole(RoleConselhoFiscal)
	if d != domain.RoleConselhoFiscal {
		t.Errorf("RoleConselhoFiscal passado a domain.Role = %q, esperado %q", d, domain.RoleConselhoFiscal)
	}

	if reflect.TypeOf(RoleConselhoFiscal) != reflect.TypeOf(domain.RoleConselhoFiscal) {
		t.Errorf("reflect.TypeOf(app.RoleConselhoFiscal) = %v, esperado idêntico a reflect.TypeOf(domain.RoleConselhoFiscal) = %v",
			reflect.TypeOf(RoleConselhoFiscal), reflect.TypeOf(domain.RoleConselhoFiscal))
	}
}

// The 8 re-exported constants must carry exactly the values already defined in
// identity/domain/role.go — never a duplicated, independently-spelled value.
func TestReexportedRoleConstantsMatchDomainValuesExactly(t *testing.T) {
	cases := []struct {
		name string
		got  Role
		want domain.Role
	}{
		{"RoleAssociado", RoleAssociado, domain.RoleAssociado},
		{"RolePresidente", RolePresidente, domain.RolePresidente},
		{"RoleDiretoria", RoleDiretoria, domain.RoleDiretoria},
		{"RoleTesouraria", RoleTesouraria, domain.RoleTesouraria},
		{"RoleEstoqueLoja", RoleEstoqueLoja, domain.RoleEstoqueLoja},
		{"RoleEventos", RoleEventos, domain.RoleEventos},
		{"RoleConselhoFiscal", RoleConselhoFiscal, domain.RoleConselhoFiscal},
		{"RoleAdminSistema", RoleAdminSistema, domain.RoleAdminSistema},
	}
	if len(cases) != len(domain.AllRoles) {
		t.Fatalf("re-export cobre %d papéis, domain.AllRoles tem %d", len(cases), len(domain.AllRoles))
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, esperado %q", c.name, c.got, c.want)
		}
	}
}
