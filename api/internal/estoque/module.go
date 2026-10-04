// Package estoque is the module's composition root.
package estoque

import (
	"context"

	"gorm.io/gorm"

	estoqueapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
)

// Deps are the platform pieces the module is built on (05-api-http/T1, same
// template as financeiro.Deps). estoque never uses platform/documents or
// storage — simpler wiring than financeiro's.
type Deps struct {
	DB         *gorm.DB
	Recorder   *audit.Recorder
	Authorizer *authz.Authorizer
}

// Module is the composed estoque module: the 2 repositories and the 6 use
// cases of 01-03, ready for estoque/http (05-api-http) to consume.
type Module struct {
	Produtos      *infra.ProdutoRepository
	Movimentacoes *infra.MovimentacaoRepository

	CriarProduto   *estoqueapp.CriarProduto
	ListarProdutos *estoqueapp.ListarProdutos

	RegistrarMovimentacao *estoqueapp.RegistrarMovimentacao
	ListarMovimentacoes   *estoqueapp.ListarMovimentacoes

	AjustarEstoque *estoqueapp.AjustarEstoque
	ConsultarSaldo *estoqueapp.ConsultarSaldo
}

// New wires the module. It does no I/O.
func New(d Deps) *Module {
	tx := func(ctx context.Context, fn func(context.Context) error) error { return database.WithTx(ctx, d.DB, fn) }

	produtos := infra.NewProdutoRepository(d.DB)
	movimentacoes := infra.NewMovimentacaoRepository(d.DB)

	return &Module{
		Produtos: produtos, Movimentacoes: movimentacoes,

		CriarProduto:   &estoqueapp.CriarProduto{Authz: d.Authorizer, Produtos: produtos, Audit: d.Recorder, Tx: tx},
		ListarProdutos: &estoqueapp.ListarProdutos{Authz: d.Authorizer, Produtos: produtos},

		RegistrarMovimentacao: &estoqueapp.RegistrarMovimentacao{
			Authz: d.Authorizer, Produtos: produtos, Movimentacoes: movimentacoes, Audit: d.Recorder, Tx: tx,
		},
		ListarMovimentacoes: &estoqueapp.ListarMovimentacoes{Authz: d.Authorizer, Movimentacoes: movimentacoes},

		AjustarEstoque: &estoqueapp.AjustarEstoque{
			Authz: d.Authorizer, Produtos: produtos, Movimentacoes: movimentacoes, Audit: d.Recorder, Tx: tx,
		},
		ConsultarSaldo: &estoqueapp.ConsultarSaldo{Authz: d.Authorizer, Produtos: produtos, Movimentacoes: movimentacoes},
	}
}

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
