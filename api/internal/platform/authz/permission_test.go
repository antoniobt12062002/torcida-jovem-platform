package authz

import "testing"

// RBAC-01.1: a permissão tem a forma módulo:recurso:ação.
func TestParsePermissionAcceptsModuleResourceAction(t *testing.T) {
	for _, s := range []string{
		"financeiro:lancamento:read", "identity:user:create", "audit:log:read", "financeiro:prestacao_contas:approve",
	} {
		p, err := ParsePermission(s)
		if err != nil || string(p) != s {
			t.Errorf("ParsePermission(%q) = %q, %v", s, p, err)
		}
	}
}

func TestParsePermissionRejectsOtherForms(t *testing.T) {
	for _, s := range []string{
		"", "identity", "identity:user", "identity:user:read:extra", "Identity:user:read",
		"identity:user:Read", "identity:us-er:read", "identity::read", ":user:read", "identity:user:", "identity:user:read ", "identity:user:re4d",
	} {
		if p, err := ParsePermission(s); err == nil {
			t.Errorf("ParsePermission(%q) deveria falhar, veio %q", s, p)
		}
	}
}

func TestActionIsTheLastSegment(t *testing.T) {
	if got := Permission("financeiro:parecer:opine").Action(); got != "opine" {
		t.Errorf("Action = %q", got)
	}
}

func TestDefinitionRequiresAValidPermission(t *testing.T) {
	if err := (Definition{Permission: "identity:user:read"}).Validate(); err != nil {
		t.Errorf("definição válida recusada: %v", err)
	}
	if err := (Definition{Permission: "invalida"}).Validate(); err == nil {
		t.Error("definição com permissão inválida deveria ser recusada")
	}
}

func TestCommonReadIsOnlyAllowedForReadActions(t *testing.T) {
	if err := (Definition{Permission: "associados:cadastro:read", CommonRead: true}).Validate(); err != nil {
		t.Errorf("leitura comum de ação read recusada: %v", err)
	}
	for _, p := range []Permission{"identity:user:create", "financeiro:parecer:opine", "identity:role:assign", "audit:log:export"} {
		if err := (Definition{Permission: p, CommonRead: true}).Validate(); err == nil {
			t.Errorf("CommonRead em %s deveria ser recusado", p)
		}
	}
}
