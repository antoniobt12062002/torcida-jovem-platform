package estoque

import (
	"slices"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// The 6 permissions of 01-03, exactly as EST-D-004 fixed them.
var operationalNames = []authz.Permission{
	"estoque:produto:create", "estoque:produto:read",
	"estoque:movimentacao:create", "estoque:movimentacao:adjust", "estoque:movimentacao:read",
	"estoque:saldo:read",
}

var readOnlyNames = []authz.Permission{
	"estoque:produto:read", "estoque:movimentacao:read", "estoque:saldo:read",
}

func permissionSet(defs []authz.Definition) map[authz.Permission]authz.Definition {
	out := make(map[authz.Permission]authz.Definition, len(defs))
	for _, d := range defs {
		out[d.Permission] = d
	}
	return out
}

// PERM-01 AC1: as 6 permissões, declaradas exatamente uma vez cada, com os
// nomes exatos fixados em EST-D-004.
func TestContributionDeclaresTheSixPermissionsExactlyOnce(t *testing.T) {
	c := Contribution()

	if len(c.Permissions) != 6 {
		t.Fatalf("Permissions tem %d entradas, esperado 6", len(c.Permissions))
	}
	byName := permissionSet(c.Permissions)
	if len(byName) != 6 {
		t.Fatalf("Permissions tem nomes duplicados: %d únicos de 6 entradas", len(byName))
	}
	for _, want := range operationalNames {
		if _, ok := byName[want]; !ok {
			t.Errorf("faltando a permissão %s", want)
		}
	}
}

// Nenhuma das 6 permissões é CommonRead — estoque afeta o razão de
// movimentações, sempre auditado (mesma decisão de financeiro).
func TestContributionNeverDeclaresCommonRead(t *testing.T) {
	for _, d := range Contribution().Permissions {
		if d.CommonRead {
			t.Errorf("%s declarada como CommonRead — estoque nunca usa leitura comum", d.Permission)
		}
	}
}

// PERM-01 AC2: ESTOQUE_LOJA recebe exatamente as 6.
func TestContributionGrantsEstoqueLojaExactlyTheSixPermissions(t *testing.T) {
	got := slices.Clone(Contribution().Grants[app.RoleEstoqueLoja])
	slices.Sort(got)
	want := slices.Clone(operationalNames)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Grants[ESTOQUE_LOJA] = %v, esperado %v", got, want)
	}
}

// PERM-01 AC2: DIRETORIA recebe exatamente as 3 de leitura.
func TestContributionGrantsDiretoriaExactlyTheThreeReadPermissions(t *testing.T) {
	got := slices.Clone(Contribution().Grants[app.RoleDiretoria])
	slices.Sort(got)
	want := slices.Clone(readOnlyNames)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Grants[DIRETORIA] = %v, esperado %v", got, want)
	}
}

// PERM-01 AC2: CONSELHO_FISCAL recebe exatamente as 3 de leitura — nunca
// create/adjust.
func TestContributionGrantsConselhoFiscalExactlyTheThreeReadPermissions(t *testing.T) {
	got := slices.Clone(Contribution().Grants[app.RoleConselhoFiscal])
	slices.Sort(got)
	want := slices.Clone(readOnlyNames)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Grants[CONSELHO_FISCAL] = %v, esperado %v", got, want)
	}
}

// Prova própria de estoque (belt-and-suspenders, sem depender só de
// BuildMatrix/forbiddenToConselhoFiscal) de que CONSELHO_FISCAL nunca
// recebe create nem adjust.
func TestContributionNeverGrantsConselhoFiscalCreateOrAdjust(t *testing.T) {
	for _, p := range Contribution().Grants[app.RoleConselhoFiscal] {
		action := p.Action()
		if action == "create" || action == "adjust" {
			t.Errorf("CONSELHO_FISCAL nunca deveria receber %s (ação %q)", p, action)
		}
	}
}

// Nenhum outro papel recebe nenhuma permissão de estoque — só ESTOQUE_LOJA,
// DIRETORIA e CONSELHO_FISCAL aparecem em Grants (PRESIDENTE é automático
// via BuildMatrix, nunca declarado aqui).
func TestContributionGrantsNoEstoquePermissionToOtherRoles(t *testing.T) {
	c := Contribution()

	if len(c.Grants) != 3 {
		t.Fatalf("Grants tem %d papéis, esperado exatamente 3 (ESTOQUE_LOJA, DIRETORIA, CONSELHO_FISCAL): %v", len(c.Grants), c.Grants)
	}
	for _, r := range []app.Role{app.RoleAssociado, app.RoleTesouraria, app.RoleEventos, app.RoleAdminSistema, app.RolePresidente} {
		if perms, ok := c.Grants[r]; ok && len(perms) > 0 {
			t.Errorf("%s não deveria receber nenhuma permissão de estoque nesta Contribution, recebeu %v", r, perms)
		}
	}
}

// A Contribution com as 6 permissões é aceita por BuildMatrix sem erro.
func TestContributionIsAcceptedAloneByBuildMatrix(t *testing.T) {
	if _, err := app.BuildMatrix(Contribution()); err != nil {
		t.Fatalf("BuildMatrix recusou a Contribution de estoque: %v", err)
	}
}

// PRESIDENTE recebe automaticamente as 6 permissões quando a matriz é
// construída — comportamento já existente de BuildMatrix, sem ação extra
// de estoque.
func TestPresidenteReceivesAllSixPermissionsAutomatically(t *testing.T) {
	m, err := app.BuildMatrix(Contribution())
	if err != nil {
		t.Fatalf("BuildMatrix: %v", err)
	}
	for _, p := range operationalNames {
		if !m.Has(app.RolePresidente, p) {
			t.Errorf("PRESIDENTE deveria receber %s automaticamente", p)
		}
	}
}
