// Package financeiro is the module's composition root.
package financeiro

import (
	"context"

	"gorm.io/gorm"

	finapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
)

// Deps are the platform pieces the module is built on (06-api-http/T1, same
// template as identity.Deps). Documents is already composed by cmd/api/main.go
// (storage.NewS3 + documents.Service) and injected here — financeiro never
// builds its own storage or S3 client (FIN-D-023).
type Deps struct {
	DB         *gorm.DB
	Recorder   *audit.Recorder
	Authorizer *authz.Authorizer
	Documents  *documents.Service
}

// Module is the composed financeiro module: the 2 repositories and the 15
// use cases of 01-04, ready for financeiro/http (06-api-http) to consume.
type Module struct {
	Contas      *infra.ContaRepository
	Lancamentos *infra.LancamentoRepository

	CriarConta     *finapp.CriarConta
	ListarContas   *finapp.ListarContas
	DesativarConta *finapp.DesativarConta
	RenomearConta  *finapp.RenomearConta

	CriarLancamento   *finapp.CriarLancamento
	EditarLancamento  *finapp.EditarLancamento
	CriarDevolucao    *finapp.CriarDevolucao
	ListarLancamentos *finapp.ListarLancamentos

	ReceberLancamento  *finapp.ReceberLancamento
	PagarLancamento    *finapp.PagarLancamento
	CancelarLancamento *finapp.CancelarLancamento
	ConsultarSaldo     *finapp.ConsultarSaldo

	AnexarComprovante    *finapp.AnexarComprovante
	ConsultarComprovante *finapp.ConsultarComprovante
	ListarComprovantes   *finapp.ListarComprovantes
}

// New wires the module. It does no I/O.
func New(d Deps) *Module {
	tx := func(ctx context.Context, fn func(context.Context) error) error { return database.WithTx(ctx, d.DB, fn) }

	contas := infra.NewContaRepository(d.DB)
	lancamentos := infra.NewLancamentoRepository(d.DB)
	existence := infra.NewLancamentoExistenceChecker(d.DB)

	return &Module{
		Contas: contas, Lancamentos: lancamentos,

		CriarConta:   &finapp.CriarConta{Authz: d.Authorizer, Contas: contas, Audit: d.Recorder, Tx: tx},
		ListarContas: &finapp.ListarContas{Authz: d.Authorizer, Contas: contas},
		DesativarConta: &finapp.DesativarConta{
			Authz: d.Authorizer, Contas: contas, Audit: d.Recorder, Tx: tx,
		},
		RenomearConta: &finapp.RenomearConta{
			Authz: d.Authorizer, Contas: contas, Lancamentos: existence, Audit: d.Recorder, Tx: tx,
		},

		CriarLancamento: &finapp.CriarLancamento{
			Authz: d.Authorizer, Contas: contas, Lancamentos: lancamentos, Audit: d.Recorder, Tx: tx,
		},
		EditarLancamento: &finapp.EditarLancamento{
			Authz: d.Authorizer, Contas: contas, Lancamentos: lancamentos, Audit: d.Recorder, Tx: tx,
		},
		CriarDevolucao: &finapp.CriarDevolucao{
			Authz: d.Authorizer, Contas: contas, Lancamentos: lancamentos, Audit: d.Recorder, Tx: tx,
		},
		ListarLancamentos: &finapp.ListarLancamentos{Authz: d.Authorizer, Lancamentos: lancamentos},

		ReceberLancamento: &finapp.ReceberLancamento{
			Authz: d.Authorizer, Lancamentos: lancamentos, Audit: d.Recorder, Tx: tx,
		},
		PagarLancamento: &finapp.PagarLancamento{
			Authz: d.Authorizer, Lancamentos: lancamentos, Audit: d.Recorder, Tx: tx,
		},
		CancelarLancamento: &finapp.CancelarLancamento{
			Authz: d.Authorizer, Lancamentos: lancamentos, Audit: d.Recorder, Tx: tx,
		},
		ConsultarSaldo: &finapp.ConsultarSaldo{Authz: d.Authorizer, Lancamentos: lancamentos},

		AnexarComprovante:    &finapp.AnexarComprovante{Lancamentos: lancamentos, Documents: d.Documents},
		ConsultarComprovante: &finapp.ConsultarComprovante{Documents: d.Documents},
		ListarComprovantes:   &finapp.ListarComprovantes{Documents: d.Documents},
	}
}

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
