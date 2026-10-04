// Package app holds the identity use cases.
package app

import (
	"fmt"
	"slices"
	"strings"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Contribution is what one module declares: its permissions and which roles
// receive them. A module only declares and grants permissions of its own
// (`<module>:...`), so no module decides about another module's permissions.
type Contribution struct {
	Module      string
	Permissions []authz.Definition
	Grants      map[domain.Role][]authz.Permission
}

// forbiddenToConselhoFiscal are the actions the Conselho Fiscal never receives
// (ADR-005, FIN-001 section 24): it supervises, it does not change. Includes
// both the original write actions and the state-transition/deactivation
// actions a business module may introduce (deactivate, receive, pay from
// financeiro; adjust from estoque) — the invariant is "never operates", not
// just "never writes".
var forbiddenToConselhoFiscal = []string{"create", "update", "delete", "cancel", "deactivate", "receive", "pay", "adjust"}

// BuildMatrix aggregates the contributions of every module into the final
// matrix and validates it. The API refuses to start when it fails.
//
// PRESIDENTE is granted every permission of the aggregated catalog as explicit
// links (there is no wildcard); a new permission reaches PRESIDENTE only
// because the synchronization writes that link.
func BuildMatrix(contributions ...Contribution) (domain.Matrix, error) {
	defs := map[authz.Permission]authz.Definition{}
	grants := map[domain.Role]map[authz.Permission]struct{}{}
	for _, r := range domain.AllRoles {
		grants[r] = map[authz.Permission]struct{}{}
	}

	for _, c := range contributions {
		if strings.TrimSpace(c.Module) == "" {
			return domain.Matrix{}, fmt.Errorf("matriz: contribuição sem módulo")
		}
		own := map[authz.Permission]struct{}{}
		for _, d := range c.Permissions {
			if err := d.Validate(); err != nil {
				return domain.Matrix{}, fmt.Errorf("matriz: módulo %s: %w", c.Module, err)
			}
			if !belongsTo(d.Permission, c.Module) {
				return domain.Matrix{}, fmt.Errorf("matriz: módulo %s declara %s, que é de outro módulo", c.Module, d.Permission)
			}
			if _, dup := defs[d.Permission]; dup {
				return domain.Matrix{}, fmt.Errorf("matriz: permissão %s declarada mais de uma vez", d.Permission)
			}
			defs[d.Permission] = d
			own[d.Permission] = struct{}{}
		}
		for role, perms := range c.Grants {
			if !role.Valid() {
				return domain.Matrix{}, fmt.Errorf("matriz: módulo %s concede a um papel desconhecido: %q", c.Module, role)
			}
			for _, p := range perms {
				if _, ok := own[p]; !ok {
					return domain.Matrix{}, fmt.Errorf("matriz: módulo %s concede %s, que não declarou", c.Module, p)
				}
				if role == domain.RoleConselhoFiscal && slices.Contains(forbiddenToConselhoFiscal, p.Action()) {
					return domain.Matrix{}, fmt.Errorf("matriz: %s nunca recebe a ação %s (%s)", domain.RoleConselhoFiscal, p.Action(), p)
				}
				grants[role][p] = struct{}{}
			}
		}
	}

	for p := range defs {
		grants[domain.RolePresidente][p] = struct{}{}
	}

	m := domain.Matrix{Grants: make(map[domain.Role][]authz.Permission, len(grants))}
	for _, d := range defs {
		m.Definitions = append(m.Definitions, d)
	}
	slices.SortFunc(m.Definitions, func(a, b authz.Definition) int { return strings.Compare(string(a.Permission), string(b.Permission)) })
	for role, set := range grants {
		list := make([]authz.Permission, 0, len(set))
		for p := range set {
			list = append(list, p)
		}
		slices.Sort(list)
		m.Grants[role] = list
	}
	return m, nil
}

func belongsTo(p authz.Permission, module string) bool {
	return strings.HasPrefix(string(p), module+":")
}

// FoundationContributions are the contributions of the foundation modules:
// identity and audit. Every other module (e.g. financeiro) publishes its own
// Contribution from its own root package; the composition root (cmd/api,
// cmd/bootstrap-admin) aggregates all of them together into one BuildMatrix
// call.
func FoundationContributions() []Contribution {
	return []Contribution{
		{
			Module: "identity",
			Permissions: []authz.Definition{
				{Permission: "identity:user:read", Description: "Listar usuários"},
				{Permission: "identity:user:create", Description: "Criar usuários"},
				{Permission: "identity:user:update", Description: "Desativar e reativar usuários"},
				{Permission: "identity:role:assign", Description: "Atribuir papéis a quem tem vínculo de gestão"},
				{Permission: "identity:admin:grant", Description: "Conceder acesso de gestão (promoção)"},
				{Permission: "identity:admin:revoke", Description: "Retirar acesso de gestão"},
				{Permission: "identity:user:reset_password", Description: "Redefinir a senha de outro usuário"},
			},
			Grants: map[domain.Role][]authz.Permission{
				domain.RoleAdminSistema: {
					"identity:user:read", "identity:user:create", "identity:user:update",
					"identity:role:assign", "identity:admin:grant", "identity:admin:revoke", "identity:user:reset_password",
				},
				domain.RoleDiretoria: {"identity:user:read"},
			},
		},
		{
			Module:      "audit",
			Permissions: []authz.Definition{{Permission: "audit:log:read", Description: "Consultar o registro de auditoria"}},
			Grants: map[domain.Role][]authz.Permission{
				domain.RoleAdminSistema:   {"audit:log:read"},
				domain.RoleConselhoFiscal: {"audit:log:read"},
			},
		},
	}
}
