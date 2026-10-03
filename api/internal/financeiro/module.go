// Package financeiro is the module's composition root.
package financeiro

import (
	finapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Contribution publishes financeiro's contribution to the aggregated RBAC
// matrix (identity/app.BuildMatrix), the same mechanism already used by
// identity and audit (FIN-D-017).
//
// The 13 operational permissions reference the Perm* constants already
// declared by the use cases of 01-04 (financeiro/app) instead of being
// redeclared as independent string literals (FIN-D-019): a rename or typo
// at the declaration site fails to compile here, rather than silently
// drifting from what platform/authz.Require actually checks at runtime.
//
// The 3 institutional permissions are preserved unchanged from
// identity/app/roles_matrix.go's current placeholder (FIN-D-014). The
// placeholder itself, and cmd/api/cmd/bootstrap-admin's wiring to this
// Contribution, are only touched by 05/T3 — until then, FoundationContributions()
// keeps declaring the same 3 institutional names on its own, so aggregating
// both together would be rejected by BuildMatrix as a duplicate (by design).
func Contribution() app.Contribution {
	operational := []authz.Definition{
		{Permission: finapp.PermContaCreate, Description: "Criar conta contábil"},
		{Permission: finapp.PermContaUpdate, Description: "Renomear conta contábil"},
		{Permission: finapp.PermContaDeactivate, Description: "Desativar conta contábil"},
		{Permission: finapp.PermContaRead, Description: "Listar contas contábeis"},
		{Permission: finapp.PermLancamentoCreate, Description: "Criar lançamento (inclui devolução)"},
		{Permission: finapp.PermLancamentoUpdate, Description: "Editar lançamento"},
		{Permission: finapp.PermLancamentoRead, Description: "Listar lançamentos"},
		{Permission: finapp.PermLancamentoReceive, Description: "Receber lançamento"},
		{Permission: finapp.PermLancamentoPay, Description: "Pagar lançamento"},
		{Permission: finapp.PermLancamentoCancel, Description: "Cancelar lançamento"},
		{Permission: finapp.PermSaldoRead, Description: "Consultar saldo"},
		{Permission: finapp.PermComprovanteCreate, Description: "Anexar comprovante"},
		{Permission: finapp.PermComprovanteRead, Description: "Consultar e listar comprovantes"},
	}
	operationalPerms := make([]authz.Permission, len(operational))
	for i, d := range operational {
		operationalPerms[i] = d.Permission
	}
	reads := []authz.Permission{
		finapp.PermContaRead, finapp.PermLancamentoRead, finapp.PermSaldoRead, finapp.PermComprovanteRead,
	}
	institutional := []authz.Permission{
		"financeiro:prestacao_contas:read", "financeiro:prestacao_contas:approve", "financeiro:parecer:opine",
	}

	permissions := []authz.Definition{
		{Permission: "financeiro:prestacao_contas:read", Description: "Ler a prestação de contas"},
		{Permission: "financeiro:prestacao_contas:approve", Description: "Aprovar a prestação de contas"},
		{Permission: "financeiro:parecer:opine", Description: "Emitir parecer"},
	}
	permissions = append(permissions, operational...)

	return app.Contribution{
		Module:      "financeiro",
		Permissions: permissions,
		Grants: map[app.Role][]authz.Permission{
			app.RoleTesouraria:     operationalPerms,
			app.RoleDiretoria:      reads,
			app.RoleConselhoFiscal: append(append([]authz.Permission{}, institutional...), reads...),
		},
	}
}
