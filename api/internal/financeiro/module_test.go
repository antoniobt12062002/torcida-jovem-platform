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

// 05-permissoes/T3: FoundationContributions() não declara mais nada de
// financeiro (o placeholder foi removido de identity/app/roles_matrix.go) —
// prova de que a substituição realmente aconteceu, não só que a fundação
// continua válida isoladamente.
func TestFoundationContributionsNoLongerDeclaresFinanceiro(t *testing.T) {
	for _, c := range app.FoundationContributions() {
		if c.Module == "financeiro" {
			t.Fatalf("FoundationContributions() ainda declara financeiro — o placeholder deveria ter sido removido pela T3: %+v", c)
		}
	}
}

// 05-permissoes/T3: a composição real de produção — FoundationContributions()
// (identity, audit) mais financeiro.Contribution() — é aceita por BuildMatrix
// sem erro de duplicata (prova de que o placeholder foi mesmo removido, não
// só que Contribution() funciona isolada) e produz exatamente a matriz
// completa esperada para os 4 papéis que financeiro afeta.
func TestContributionCombinedWithFoundationContributionsBuildsTheFullProductionMatrix(t *testing.T) {
	m, err := app.BuildMatrix(append(app.FoundationContributions(), Contribution())...)
	if err != nil {
		t.Fatalf("BuildMatrix recusou a composição de produção (financeiro ainda duplicado no placeholder?): %v", err)
	}

	wantTesouraria := slices.Clone(operationalNames)
	slices.Sort(wantTesouraria)
	gotTesouraria := slices.Clone(m.Grants[app.RoleTesouraria])
	slices.Sort(gotTesouraria)
	if !slices.Equal(gotTesouraria, wantTesouraria) {
		t.Errorf("TESOURARIA (produção) = %v, esperado %v", gotTesouraria, wantTesouraria)
	}

	// DIRETORIA também recebe identity:user:read de FoundationContributions()
	// (RBAC-01, não relacionado a financeiro) — a composição de produção soma
	// os dois, por isso o esperado aqui não é só readOnlyNames.
	wantDiretoria := append(slices.Clone(readOnlyNames), "identity:user:read")
	slices.Sort(wantDiretoria)
	gotDiretoria := slices.Clone(m.Grants[app.RoleDiretoria])
	slices.Sort(gotDiretoria)
	if !slices.Equal(gotDiretoria, wantDiretoria) {
		t.Errorf("DIRETORIA (produção) = %v, esperado %v", gotDiretoria, wantDiretoria)
	}

	wantConselhoFiscal := append(slices.Clone(readOnlyNames), institutional...)
	wantConselhoFiscal = append(wantConselhoFiscal, "audit:log:read")
	slices.Sort(wantConselhoFiscal)
	gotConselhoFiscal := slices.Clone(m.Grants[app.RoleConselhoFiscal])
	slices.Sort(gotConselhoFiscal)
	if !slices.Equal(gotConselhoFiscal, wantConselhoFiscal) {
		t.Errorf("CONSELHO_FISCAL (produção) = %v, esperado %v", gotConselhoFiscal, wantConselhoFiscal)
	}

	all := append(slices.Clone(operationalNames), institutional...)
	for _, p := range all {
		if !m.Has(app.RolePresidente, p) {
			t.Errorf("PRESIDENTE (produção) deveria ter %s", p)
		}
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
