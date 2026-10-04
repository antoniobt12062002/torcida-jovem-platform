// Package estoque is the module's composition root.
package estoque

import (
	estoqueapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Contribution publishes estoque's contribution to the aggregated RBAC
// matrix (identity/app.BuildMatrix), same mechanism already used by
// identity and financeiro (FIN-D-017).
//
// The 6 permissions reference the Perm* constants already declared by the
// use cases of 01-03 (estoque/app) instead of being redeclared as
// independent string literals (same decision as financeiro's FIN-D-019): a
// rename or typo at the declaration site fails to compile here, rather
// than silently drifting from what platform/authz.Require actually checks
// at runtime.
func Contribution() app.Contribution {
	operational := []authz.Definition{
		{Permission: estoqueapp.PermProdutoCreate, Description: "Criar produto (SKU) de estoque"},
		{Permission: estoqueapp.PermProdutoRead, Description: "Listar produtos (SKUs) de estoque"},
		{Permission: estoqueapp.PermMovimentacaoCreate, Description: "Registrar entrada, saída ou devolução"},
		{Permission: estoqueapp.PermMovimentacaoAdjust, Description: "Registrar ajuste de estoque"},
		{Permission: estoqueapp.PermMovimentacaoRead, Description: "Listar movimentações de estoque"},
		{Permission: estoqueapp.PermSaldoRead, Description: "Consultar saldo de estoque"},
	}
	operationalPerms := make([]authz.Permission, len(operational))
	for i, d := range operational {
		operationalPerms[i] = d.Permission
	}
	reads := []authz.Permission{
		estoqueapp.PermProdutoRead, estoqueapp.PermMovimentacaoRead, estoqueapp.PermSaldoRead,
	}

	return app.Contribution{
		Module:      "estoque",
		Permissions: operational,
		Grants: map[app.Role][]authz.Permission{
			app.RoleEstoqueLoja:    operationalPerms,
			app.RoleDiretoria:      reads,
			app.RoleConselhoFiscal: reads,
		},
	}
}
