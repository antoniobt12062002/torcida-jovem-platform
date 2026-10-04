//go:build integration

package app_test

import (
	"context"
	"io"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// simpleAuthz forwards to the package-level authz.Require: it decides purely
// by what the Principal holds, no catalog — same shape estoque's use cases
// depend on (app.Authorizer). Mesmo padrão de financeiro/app/env_test.go.
type simpleAuthz struct{}

func (simpleAuthz) Require(_ context.Context, p authz.Principal, perm authz.Permission) error {
	return authz.Require(p, perm)
}

type env struct {
	db            *gorm.DB
	rec           *audit.Recorder
	produtos      *infra.ProdutoRepository
	movimentacoes *infra.MovimentacaoRepository
}

func newEnv(t *testing.T) env {
	t.Helper()
	db := testutil.NewTestDB(t)
	rec := audit.NewRecorder(db, logx.New("error", io.Discard))
	return env{
		db: db, rec: rec, produtos: infra.NewProdutoRepository(db),
		movimentacoes: infra.NewMovimentacaoRepository(db),
	}
}

func (e env) tx(ctx context.Context, fn func(context.Context) error) error {
	return database.WithTx(ctx, e.db, fn)
}

func principal(uid string, perms ...authz.Permission) authz.Principal {
	m := make(map[authz.Permission]struct{}, len(perms))
	for _, p := range perms {
		m[p] = struct{}{}
	}
	return authz.Principal{UserID: uid, Permissions: m}
}

// actor is a syntactically valid UUID with the given permissions.
// produtos_estoque has no foreign key to users, so no real user row is needed.
func actor(perms ...authz.Permission) authz.Principal {
	return principal(uuid.New().String(), perms...)
}

func (e env) criarProduto() *app.CriarProduto {
	return &app.CriarProduto{Authz: simpleAuthz{}, Produtos: e.produtos, Audit: e.rec, Tx: e.tx}
}

func (e env) listarProdutos() *app.ListarProdutos {
	return &app.ListarProdutos{Authz: simpleAuthz{}, Produtos: e.produtos}
}

// newUser inserts a real users row and returns its id: movimentacoes_estoque
// .responsavel_id is a real foreign key, unlike produtos_estoque, so tests
// that create a movimentação need an actor backed by an existing user.
func newUser(t *testing.T, e env) string {
	t.Helper()
	var id string
	err := e.db.Raw(`INSERT INTO users (email, name, password_hash) VALUES (?, 'Fulano', 'hash') RETURNING id::text`,
		uuid.New().String()+"@exemplo.com").Scan(&id).Error
	if err != nil || id == "" {
		t.Fatalf("criar usuário: id=%q err=%v", id, err)
	}
	return id
}

// userActor is like actor, but backed by a real users row (see newUser).
func (e env) userActor(t *testing.T, perms ...authz.Permission) authz.Principal {
	return principal(newUser(t, e), perms...)
}

func (e env) criarProdutoDeTeste(t *testing.T, codigo string) string {
	t.Helper()
	p, err := e.criarProduto().Execute(context.Background(), app.CriarProdutoInput{
		Actor: actor(app.PermProdutoCreate), Codigo: codigo, Nome: "produto " + codigo, UnidadeMedida: "UN",
	})
	if err != nil {
		t.Fatalf("criar produto de teste: %v", err)
	}
	return p.ID
}

func (e env) registrarMovimentacao() *app.RegistrarMovimentacao {
	return &app.RegistrarMovimentacao{
		Authz: simpleAuthz{}, Produtos: e.produtos, Movimentacoes: e.movimentacoes, Audit: e.rec, Tx: e.tx,
	}
}
