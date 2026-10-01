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
	db     *gorm.DB
	rec    *audit.Recorder
	contas *infra.ContaRepository
}

func newEnv(t *testing.T) env {
	t.Helper()
	db := testutil.NewTestDB(t)
	rec := audit.NewRecorder(db, logx.New("error", io.Discard))
	return env{db: db, rec: rec, contas: infra.NewContaRepository(db)}
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
