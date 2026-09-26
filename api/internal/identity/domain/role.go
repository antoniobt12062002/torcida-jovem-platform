// Package domain holds the identity entities and rules, free of infrastructure.
package domain

// Role groups permissions. Authorization never decides by role name: it uses the
// effective permissions. This file is the only place that spells the role names
// (a source-tree test enforces it).
type Role string

const (
	RoleAssociado      Role = "ASSOCIADO"
	RolePresidente     Role = "PRESIDENTE"
	RoleDiretoria      Role = "DIRETORIA"
	RoleTesouraria     Role = "TESOURARIA"
	RoleEstoqueLoja    Role = "ESTOQUE_LOJA"
	RoleEventos        Role = "EVENTOS"
	RoleConselhoFiscal Role = "CONSELHO_FISCAL"
	RoleAdminSistema   Role = "ADMIN_SISTEMA"
)

// AllRoles lists the roles provisioned by the synchronization, in a stable order.
var AllRoles = []Role{
	RoleAssociado, RolePresidente, RoleDiretoria, RoleTesouraria,
	RoleEstoqueLoja, RoleEventos, RoleConselhoFiscal, RoleAdminSistema,
}

var descriptions = map[Role]string{
	RoleAssociado:      "Usuário autenticado sem acesso administrativo",
	RolePresidente:     "Presidência: todas as permissões do catálogo",
	RoleDiretoria:      "Diretoria",
	RoleTesouraria:     "Tesouraria",
	RoleEstoqueLoja:    "Estoque e loja",
	RoleEventos:        "Eventos",
	RoleConselhoFiscal: "Conselho Fiscal: fiscalização, prestação de contas e parecer",
	RoleAdminSistema:   "Administração técnica do sistema",
}

// Valid reports whether r is one of the provisioned roles.
func (r Role) Valid() bool {
	_, ok := descriptions[r]
	return ok
}

// Description is the human description stored with the role.
func (r Role) Description() string { return descriptions[r] }
