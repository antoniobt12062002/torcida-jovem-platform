// Package app holds the identity use cases.
package app

import "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"

// Role is a public alias of identity/domain.Role — the identical type, never a
// second definition. It exists because architecture_test.go lets a module use
// another module's published app and root package, but never its domain, infra
// or http: a module publishing a Contribution (identity/app.Contribution) needs
// the role vocabulary to build its Grants map, and identity/domain is off
// limits to it. Contribution.Grants stays typed map[domain.Role]... unchanged —
// Role being an alias (not a defined type) makes the two interchangeable.
type Role = domain.Role

// The 8 roles provisioned by the synchronization (identity/domain/role.go is
// still the only place that spells their names), re-exported so other modules
// can reference them without importing identity/domain.
const (
	RoleAssociado      = domain.RoleAssociado
	RolePresidente     = domain.RolePresidente
	RoleDiretoria      = domain.RoleDiretoria
	RoleTesouraria     = domain.RoleTesouraria
	RoleEstoqueLoja    = domain.RoleEstoqueLoja
	RoleEventos        = domain.RoleEventos
	RoleConselhoFiscal = domain.RoleConselhoFiscal
	RoleAdminSistema   = domain.RoleAdminSistema
)
