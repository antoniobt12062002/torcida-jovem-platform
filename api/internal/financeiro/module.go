// Package financeiro is the module's composition root.
package financeiro

import (
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Contribution publishes financeiro's contribution to the aggregated RBAC
// matrix (identity/app.BuildMatrix), the same mechanism already used by
// identity and audit (FIN-D-017).
//
// For now it carries only the 3 institutional permissions already reserved
// for financeiro since the foundation (FIN-D-014), preserved unchanged from
// identity/app/roles_matrix.go's current placeholder. The 13 operational
// permissions of 01-04 are added by 05/T2; the placeholder itself is removed,
// and this Contribution wired into cmd/api and cmd/bootstrap-admin, only by
// 05/T3 — until then, FoundationContributions() keeps declaring these same
// 3 permissions on its own, so aggregating both together would be rejected by
// BuildMatrix as a duplicate (by design: T1 does not touch the placeholder).
func Contribution() app.Contribution {
	return app.Contribution{
		Module: "financeiro",
		Permissions: []authz.Definition{
			{Permission: "financeiro:prestacao_contas:read", Description: "Ler a prestação de contas"},
			{Permission: "financeiro:prestacao_contas:approve", Description: "Aprovar a prestação de contas"},
			{Permission: "financeiro:parecer:opine", Description: "Emitir parecer"},
		},
		Grants: map[app.Role][]authz.Permission{
			app.RoleConselhoFiscal: {
				"financeiro:prestacao_contas:read", "financeiro:prestacao_contas:approve", "financeiro:parecer:opine",
			},
		},
	}
}
