package domain

import (
	"slices"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Matrix is the aggregated, validated result of every module's contribution:
// the permission catalog and the permissions of each role. It is what the
// synchronization writes to the database.
type Matrix struct {
	// Definitions, sorted by permission.
	Definitions []authz.Definition
	// Grants has an entry for every role in AllRoles; permissions sorted.
	Grants map[Role][]authz.Permission
}

// Has reports whether the role is granted the permission.
func (m Matrix) Has(r Role, perm authz.Permission) bool {
	return slices.Contains(m.Grants[r], perm)
}

// Permissions lists every permission of the catalog, sorted.
func (m Matrix) Permissions() []authz.Permission {
	out := make([]authz.Permission, len(m.Definitions))
	for i, d := range m.Definitions {
		out[i] = d.Permission
	}
	return out
}

// PermissionsOf is the union of the permissions granted to the given roles,
// sorted and without repetition.
func (m Matrix) PermissionsOf(roles []Role) []authz.Permission {
	var out []authz.Permission
	for _, r := range roles {
		for _, p := range m.Grants[r] {
			if !slices.Contains(out, p) {
				out = append(out, p)
			}
		}
	}
	slices.Sort(out)
	return out
}
