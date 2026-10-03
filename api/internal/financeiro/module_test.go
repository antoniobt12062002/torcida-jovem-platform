package financeiro

import (
	"slices"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// The 3 institutional permissions, exactly as already reserved to financeiro
// since the foundation (FIN-D-014) — PERM-02 AC1.
var institutional = []authz.Permission{
	"financeiro:prestacao_contas:read", "financeiro:prestacao_contas:approve", "financeiro:parecer:opine",
}

func TestContributionDeclaresExactlyTheThreeInstitutionalPermissions(t *testing.T) {
	c := Contribution()

	if c.Module != "financeiro" {
		t.Errorf("Module = %q, esperado %q", c.Module, "financeiro")
	}
	if len(c.Permissions) != 3 {
		t.Fatalf("Permissions tem %d entradas, esperado 3 (T1 não introduz nenhuma operacional)", len(c.Permissions))
	}
	for _, want := range institutional {
		if !slices.ContainsFunc(c.Permissions, func(d authz.Definition) bool { return d.Permission == want }) {
			t.Errorf("faltando a permissão institucional %s", want)
		}
	}
}

// PERM-01 AC2: nenhuma permissão de financeiro é CommonRead.
func TestContributionNeverDeclaresCommonRead(t *testing.T) {
	for _, d := range Contribution().Permissions {
		if d.CommonRead {
			t.Errorf("%s declarada como CommonRead — financeiro nunca usa leitura comum", d.Permission)
		}
	}
}

// PERM-02 AC1: as 3 institucionais são concedidas só ao Conselho Fiscal.
func TestContributionGrantsTheThreeInstitutionalPermissionsOnlyToConselhoFiscal(t *testing.T) {
	c := Contribution()

	if len(c.Grants) != 1 {
		t.Fatalf("Grants concede a %d papéis, esperado só CONSELHO_FISCAL", len(c.Grants))
	}
	got := slices.Clone(c.Grants[app.RoleConselhoFiscal])
	slices.Sort(got)
	want := slices.Clone(institutional)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Grants[CONSELHO_FISCAL] = %v, esperado %v", got, want)
	}
}

// PERM-02 AC1/AC3: a Contribution real preserva exatamente o que o placeholder
// da fundação já concede hoje — mesma permissão, mesma descrição, mesmo papel.
// Comparação direta "antes" (placeholder, ainda vivo em FoundationContributions)
// vs. "depois" (Contribution real) sem jamais agregar os dois juntos no mesmo
// BuildMatrix — eles declaram os mesmos nomes, então BuildMatrix recusaria a
// duplicata (T3, não T1, é quem efetivamente troca um pelo outro).
func TestContributionMatchesTheFoundationPlaceholderExactly(t *testing.T) {
	var placeholder app.Contribution
	found := false
	for _, c := range app.FoundationContributions() {
		if c.Module == "financeiro" {
			placeholder = c
			found = true
		}
	}
	if !found {
		t.Fatal("FoundationContributions() não tem mais o placeholder de financeiro — T1 não deveria alterar isso")
	}

	real := Contribution()

	gotPerms := slices.Clone(real.Permissions)
	slices.SortFunc(gotPerms, func(a, b authz.Definition) int { return int(a.Permission[0]) - int(b.Permission[0]) })
	wantPerms := slices.Clone(placeholder.Permissions)
	slices.SortFunc(wantPerms, func(a, b authz.Definition) int { return int(a.Permission[0]) - int(b.Permission[0]) })
	if !slices.Equal(gotPerms, wantPerms) {
		t.Errorf("Permissions = %+v, placeholder tinha %+v", gotPerms, wantPerms)
	}

	gotGrants := slices.Clone(real.Grants[app.RoleConselhoFiscal])
	slices.Sort(gotGrants)
	wantGrants := slices.Clone(placeholder.Grants[app.RoleConselhoFiscal])
	slices.Sort(wantGrants)
	if !slices.Equal(gotGrants, wantGrants) {
		t.Errorf("Grants[CONSELHO_FISCAL] = %v, placeholder tinha %v", gotGrants, wantGrants)
	}
}

// PERM-01 AC1/AC4: a Contribution isolada (sem o placeholder ainda em
// FoundationContributions, que duplicaria os mesmos nomes) é aceita por
// BuildMatrix sem erro, e PRESIDENTE recebe as 3 automaticamente.
func TestContributionIsAcceptedAloneByBuildMatrix(t *testing.T) {
	m, err := app.BuildMatrix(Contribution())
	if err != nil {
		t.Fatalf("BuildMatrix recusou a Contribution de financeiro: %v", err)
	}
	for _, p := range institutional {
		if !m.Has(app.RolePresidente, p) {
			t.Errorf("PRESIDENTE deveria receber %s automaticamente", p)
		}
		if !m.Has(app.RoleConselhoFiscal, p) {
			t.Errorf("CONSELHO_FISCAL deveria receber %s", p)
		}
	}
}

// T1 não mexe em FoundationContributions(): a matriz da fundação continua
// válida e com o placeholder, exatamente como antes desta tarefa.
func TestFoundationContributionsIsUnchangedByT1(t *testing.T) {
	if _, err := app.BuildMatrix(app.FoundationContributions()...); err != nil {
		t.Fatalf("a matriz da fundação deveria continuar válida: %v", err)
	}
}
