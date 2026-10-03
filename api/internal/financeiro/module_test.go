package financeiro

import (
	"slices"
	"testing"

	finapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// The 3 institutional permissions, exactly as already reserved to financeiro
// since the foundation (FIN-D-014) — PERM-02 AC1, preserved unchanged by T2.
var institutional = []authz.Permission{
	"financeiro:prestacao_contas:read", "financeiro:prestacao_contas:approve", "financeiro:parecer:opine",
}

// The 13 operational permissions of 01-04 (PERM-01 AC1, PERM-03), as the
// consolidated matrix in financeiro/STATE.md names them.
var operationalNames = []authz.Permission{
	"financeiro:conta:create", "financeiro:conta:update", "financeiro:conta:deactivate", "financeiro:conta:read",
	"financeiro:lancamento:create", "financeiro:lancamento:update", "financeiro:lancamento:read",
	"financeiro:lancamento:receive", "financeiro:lancamento:pay", "financeiro:lancamento:cancel",
	"financeiro:saldo:read",
	"financeiro:comprovante:create", "financeiro:comprovante:read",
}

var readOnlyNames = []authz.Permission{
	"financeiro:conta:read", "financeiro:lancamento:read", "financeiro:saldo:read", "financeiro:comprovante:read",
}

var forbiddenToConselhoFiscalActions = []string{"create", "update", "deactivate", "receive", "pay", "cancel"}

func permissionSet(defs []authz.Definition) map[authz.Permission]authz.Definition {
	out := make(map[authz.Permission]authz.Definition, len(defs))
	for _, d := range defs {
		out[d.Permission] = d
	}
	return out
}

// PERM-01 AC1: as 13 permissões operacionais, declaradas exatamente uma vez
// cada, com os nomes exatos consolidados em financeiro/STATE.md — mais as 3
// institucionais, 16 no total, sem nenhum nome fora dessas duas listas.
func TestContributionDeclaresTheThirteenOperationalPermissionsExactlyOnce(t *testing.T) {
	c := Contribution()

	if len(c.Permissions) != 16 {
		t.Fatalf("Permissions tem %d entradas, esperado 16 (13 operacionais + 3 institucionais)", len(c.Permissions))
	}
	byName := permissionSet(c.Permissions)
	if len(byName) != 16 {
		t.Fatalf("Permissions tem nomes duplicados: %d únicos de 16 entradas", len(byName))
	}
	for _, want := range operationalNames {
		if _, ok := byName[want]; !ok {
			t.Errorf("faltando a permissão operacional %s", want)
		}
	}
	for _, want := range institutional {
		if _, ok := byName[want]; !ok {
			t.Errorf("faltando a permissão institucional %s", want)
		}
	}
}

// PERM-01 AC2: nenhuma das 16 permissões (operacionais ou institucionais) é
// CommonRead — financeiro nunca usa leitura comum.
func TestContributionNeverDeclaresCommonRead(t *testing.T) {
	for _, d := range Contribution().Permissions {
		if d.CommonRead {
			t.Errorf("%s declarada como CommonRead — financeiro nunca usa leitura comum", d.Permission)
		}
	}
}

// PERM-03 AC1: TESOURARIA recebe exatamente as 13 operacionais (nenhuma
// institucional).
func TestContributionGrantsTesourariaExactlyTheThirteenOperationalPermissions(t *testing.T) {
	got := slices.Clone(Contribution().Grants[app.RoleTesouraria])
	slices.Sort(got)
	want := slices.Clone(operationalNames)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Grants[TESOURARIA] = %v, esperado %v", got, want)
	}
}

// PERM-03 AC2: DIRETORIA recebe exatamente as 4 permissões de leitura.
func TestContributionGrantsDiretoriaExactlyTheFourReadPermissions(t *testing.T) {
	got := slices.Clone(Contribution().Grants[app.RoleDiretoria])
	slices.Sort(got)
	want := slices.Clone(readOnlyNames)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Grants[DIRETORIA] = %v, esperado %v", got, want)
	}
}

// PERM-03 AC3: CONSELHO_FISCAL recebe exatamente as 4 de leitura mais as 3
// institucionais — nunca nenhuma de escrita/transição.
func TestContributionGrantsConselhoFiscalExactlyFourReadsPlusThreeInstitutional(t *testing.T) {
	got := slices.Clone(Contribution().Grants[app.RoleConselhoFiscal])
	slices.Sort(got)
	want := append(slices.Clone(readOnlyNames), institutional...)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("Grants[CONSELHO_FISCAL] = %v, esperado %v", got, want)
	}
}

// PERM-01 AC3: prova própria de financeiro (belt-and-suspenders, sem depender
// só de BuildMatrix) de que CONSELHO_FISCAL nunca recebe create/update/
// deactivate/receive/pay/cancel.
func TestContributionNeverGrantsConselhoFiscalAForbiddenAction(t *testing.T) {
	for _, p := range Contribution().Grants[app.RoleConselhoFiscal] {
		if slices.Contains(forbiddenToConselhoFiscalActions, p.Action()) {
			t.Errorf("CONSELHO_FISCAL nunca deveria receber %s (ação %q)", p, p.Action())
		}
	}
}

// PERM-03 AC4: nenhum outro papel recebe nenhuma permissão de financeiro —
// só TESOURARIA, DIRETORIA e CONSELHO_FISCAL aparecem em Grants (PRESIDENTE
// é automático via BuildMatrix, nunca declarado aqui).
func TestContributionGrantsNoFinanceiroPermissionToOtherRoles(t *testing.T) {
	c := Contribution()

	if len(c.Grants) != 3 {
		t.Fatalf("Grants tem %d papéis, esperado exatamente 3 (TESOURARIA, DIRETORIA, CONSELHO_FISCAL): %v", len(c.Grants), c.Grants)
	}
	for _, r := range []app.Role{app.RoleAssociado, app.RoleEstoqueLoja, app.RoleEventos, app.RoleAdminSistema, app.RolePresidente} {
		if perms, ok := c.Grants[r]; ok && len(perms) > 0 {
			t.Errorf("%s não deveria receber nenhuma permissão de financeiro nesta Contribution, recebeu %v", r, perms)
		}
	}
}

// PERM-01 AC1/AC4: a Contribution com as 16 permissões é aceita por
// BuildMatrix sem erro.
func TestContributionIsAcceptedAloneByBuildMatrix(t *testing.T) {
	if _, err := app.BuildMatrix(Contribution()); err != nil {
		t.Fatalf("BuildMatrix recusou a Contribution de financeiro: %v", err)
	}
}

// PERM-01 AC4: PRESIDENTE recebe automaticamente as 16 permissões quando a
// matriz é construída — comportamento já existente de BuildMatrix, sem ação
// extra de financeiro.
func TestPresidenteReceivesAllSixteenPermissionsAutomatically(t *testing.T) {
	m, err := app.BuildMatrix(Contribution())
	if err != nil {
		t.Fatalf("BuildMatrix: %v", err)
	}
	all := append(slices.Clone(operationalNames), institutional...)
	for _, p := range all {
		if !m.Has(app.RolePresidente, p) {
			t.Errorf("PRESIDENTE deveria receber %s automaticamente", p)
		}
	}
}

// PERM-02 AC1/AC3: as 3 permissões institucionais e seu grant a
// CONSELHO_FISCAL continuam exatamente iguais ao placeholder da fundação —
// comparação parcial (só o subconjunto institucional), nunca uma igualdade
// integral entre Contribution() e FoundationContributions(): a partir da T2
// as duas declaram deliberadamente conjuntos diferentes (16 vs. 3).
func TestContributionPreservesTheThreeInstitutionalPermissionsExactlyLikeThePlaceholder(t *testing.T) {
	var placeholder app.Contribution
	found := false
	for _, c := range app.FoundationContributions() {
		if c.Module == "financeiro" {
			placeholder = c
			found = true
		}
	}
	if !found {
		t.Fatal("FoundationContributions() não tem mais o placeholder de financeiro — T2 não deveria alterar isso")
	}

	real := Contribution()

	realInstitutional := make([]authz.Definition, 0, 3)
	for _, d := range real.Permissions {
		if slices.Contains(institutional, d.Permission) {
			realInstitutional = append(realInstitutional, d)
		}
	}
	gotPerms := slices.Clone(realInstitutional)
	slices.SortFunc(gotPerms, func(a, b authz.Definition) int { return int(a.Permission[0]) - int(b.Permission[0]) })
	wantPerms := slices.Clone(placeholder.Permissions)
	slices.SortFunc(wantPerms, func(a, b authz.Definition) int { return int(a.Permission[0]) - int(b.Permission[0]) })
	if !slices.Equal(gotPerms, wantPerms) {
		t.Errorf("subconjunto institucional de Permissions = %+v, placeholder tinha %+v", gotPerms, wantPerms)
	}

	realInstitutionalGrants := make([]authz.Permission, 0, 3)
	for _, p := range real.Grants[app.RoleConselhoFiscal] {
		if slices.Contains(institutional, p) {
			realInstitutionalGrants = append(realInstitutionalGrants, p)
		}
	}
	gotGrants := slices.Clone(realInstitutionalGrants)
	slices.Sort(gotGrants)
	wantGrants := slices.Clone(placeholder.Grants[app.RoleConselhoFiscal])
	slices.Sort(wantGrants)
	if !slices.Equal(gotGrants, wantGrants) {
		t.Errorf("subconjunto institucional de Grants[CONSELHO_FISCAL] = %v, placeholder tinha %v", gotGrants, wantGrants)
	}
}

// T2 não mexe em FoundationContributions(): a matriz da fundação continua
// válida e com o placeholder, exatamente como antes desta tarefa — a
// substituição é exclusiva da T3.
func TestFoundationContributionsIsUnchangedByT2(t *testing.T) {
	if _, err := app.BuildMatrix(app.FoundationContributions()...); err != nil {
		t.Fatalf("a matriz da fundação deveria continuar válida: %v", err)
	}
}

// As 13 permissões operacionais declaradas em Contribution() usam exatamente
// as mesmas constantes que os casos de uso de 01-04 checam em tempo de
// execução (FIN-D-019) — não literais redigitados. Este teste prova a
// igualdade de valor; a referência direta ao identificador (em module.go) já
// garante, em tempo de compilação, que não há uma segunda fonte divergente.
func TestContributionOperationalPermissionsMatchTheUseCasesConstantsExactly(t *testing.T) {
	cases := map[authz.Permission]authz.Permission{
		"financeiro:conta:create":       finapp.PermContaCreate,
		"financeiro:conta:update":       finapp.PermContaUpdate,
		"financeiro:conta:deactivate":   finapp.PermContaDeactivate,
		"financeiro:conta:read":         finapp.PermContaRead,
		"financeiro:lancamento:create":  finapp.PermLancamentoCreate,
		"financeiro:lancamento:update":  finapp.PermLancamentoUpdate,
		"financeiro:lancamento:read":    finapp.PermLancamentoRead,
		"financeiro:lancamento:receive": finapp.PermLancamentoReceive,
		"financeiro:lancamento:pay":     finapp.PermLancamentoPay,
		"financeiro:lancamento:cancel":  finapp.PermLancamentoCancel,
		"financeiro:saldo:read":         finapp.PermSaldoRead,
		"financeiro:comprovante:create": finapp.PermComprovanteCreate,
		"financeiro:comprovante:read":   finapp.PermComprovanteRead,
	}
	for want, got := range cases {
		if got != want {
			t.Errorf("constante de financeiro/app = %q, esperado %q", got, want)
		}
	}
}
