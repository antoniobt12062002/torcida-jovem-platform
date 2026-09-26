package domain

import (
	"slices"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func TestRolesAreValidAndHaveDescriptions(t *testing.T) {
	if len(AllRoles) != 8 {
		t.Fatalf("papéis = %d", len(AllRoles))
	}
	seen := map[Role]bool{}
	for _, r := range AllRoles {
		if !r.Valid() || r.Description() == "" || seen[r] {
			t.Errorf("papel %q inválido, sem descrição ou repetido", r)
		}
		seen[r] = true
	}
	if Role("CHEFE").Valid() || Role("").Valid() || Role("presidente").Valid() {
		t.Error("papéis fora da lista não são válidos")
	}
}

func TestMatrixHasAndPermissions(t *testing.T) {
	m := Matrix{
		Definitions: []authz.Definition{{Permission: "a:b:read"}, {Permission: "a:b:update"}},
		Grants:      map[Role][]authz.Permission{RoleEventos: {"a:b:read"}},
	}

	if !m.Has(RoleEventos, "a:b:read") || m.Has(RoleEventos, "a:b:update") || m.Has(RoleTesouraria, "a:b:read") {
		t.Error("Has não reflete as concessões")
	}
	if got := m.Permissions(); !slices.Equal(got, []authz.Permission{"a:b:read", "a:b:update"}) {
		t.Errorf("Permissions = %v", got)
	}
}

func TestPermissionsOfIsTheSortedUnionWithoutRepetition(t *testing.T) {
	m := Matrix{Grants: map[Role][]authz.Permission{
		RoleEventos:    {"a:b:read", "a:b:update"},
		RoleTesouraria: {"a:b:read", "c:d:create"},
	}}

	got := m.PermissionsOf([]Role{RoleTesouraria, RoleEventos, RoleAssociado})

	if !slices.Equal(got, []authz.Permission{"a:b:read", "a:b:update", "c:d:create"}) {
		t.Errorf("PermissionsOf = %v", got)
	}
	if got := m.PermissionsOf(nil); len(got) != 0 {
		t.Errorf("sem papéis: %v", got)
	}
}
