//go:build integration

package app_test

import (
	"context"
	"io"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// simpleAuthz forwards to the package-level authz.Require: it decides purely
// by what the Principal holds, no catalog, no denial hook — the same shape
// financeiro's use cases depend on (app.Authorizer). financeiro does not own
// a permission catalog in this test: 05-permissoes builds the real
// Contribution separately; these tests only need to prove each use case
// checks the exact permission it is supposed to.
type simpleAuthz struct{}

func (simpleAuthz) Require(_ context.Context, p authz.Principal, perm authz.Permission) error {
	return authz.Require(p, perm)
}

type env struct {
	db          *gorm.DB
	rec         *audit.Recorder
	contas      *infra.ContaRepository
	lancamentos *infra.LancamentoRepository
}

func newEnv(t *testing.T) env {
	t.Helper()
	db := testutil.NewTestDB(t)
	rec := audit.NewRecorder(db, logx.New("error", io.Discard))
	return env{db: db, rec: rec, contas: infra.NewContaRepository(db), lancamentos: infra.NewLancamentoRepository(db)}
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

// actor is a syntactically valid UUID with the given permissions. contas_contabeis
// has no foreign key to users, so no real user row is needed.
func actor(perms ...authz.Permission) authz.Principal {
	return principal(uuid.New().String(), perms...)
}

// newUser inserts a real users row and returns its id: lancamentos.criado_por
// (and cancelado_por, in 03-workflow-e-saldo) is a real foreign key, unlike
// contas_contabeis, so tests that create a lançamento need an actor backed
// by an existing user.
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
