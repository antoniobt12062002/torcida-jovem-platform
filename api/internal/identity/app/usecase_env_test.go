//go:build integration

package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// ucEnv is a database with the matrix synchronized and every repository, for
// the use case tests.
type ucEnv struct {
	db, owner  *gorm.DB
	users      *infra.UserRepository
	sessions   *infra.SessionRepository
	members    *infra.AdminMembershipRepository
	recovery   *infra.RecoveryRepository
	roles      *infra.RoleRepository
	rec        *audit.Recorder
	authorizer *authz.Authorizer
	matrix     domain.Matrix
	hasher     *password.Hasher
	deny       *password.Denylist
	logs       *bytes.Buffer
	now        time.Time
	ctx        context.Context
	sequence   int
}

func newUCEnv(t *testing.T) *ucEnv {
	t.Helper()
	db := testutil.NewTestDB(t)
	logs := &bytes.Buffer{}
	rec := audit.NewRecorder(db, logx.New("info", logs))
	m, err := app.BuildMatrix(app.FoundationContributions()...)
	if err != nil {
		t.Fatal(err)
	}
	hasher, err := password.NewHasher(password.Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1})
	if err != nil {
		t.Fatal(err)
	}
	authorizer, err := authz.NewAuthorizer(m.Definitions, audit.DeniedHook(rec))
	if err != nil {
		t.Fatal(err)
	}
	e := &ucEnv{
		db: db, owner: testutil.OwnerDBFor(t, db), rec: rec, matrix: m, hasher: hasher, logs: logs, now: t0,
		users: infra.NewUserRepository(db), sessions: infra.NewSessionRepository(db), members: infra.NewAdminMembershipRepository(db),
		recovery: infra.NewRecoveryRepository(db), roles: infra.NewRoleRepository(db, rec),
		authorizer: authorizer, deny: password.NewDenylist("12345678\nsenhacomum1234\n"),
		ctx: httpx.WithRequestID(context.Background(), "req-12345678"),
	}
	if _, err := e.roles.Sync(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	return e
}

func (e *ucEnv) tx(ctx context.Context, fn func(context.Context) error) error {
	return database.WithTx(ctx, e.db, fn)
}

func (e *ucEnv) clock() time.Time { return e.now }

// actor creates an active user with the given roles (a membership when any is
// administrative) and returns it with its principal, as the session service would.
func (e *ucEnv) actor(t *testing.T, roles ...domain.Role) (domain.User, authz.Principal) {
	t.Helper()
	e.sequence++
	hash, _ := e.hasher.Hash("senha-do-ator-" + string(rune('a'+e.sequence)))
	u, err := e.users.Create(context.Background(), domain.User{
		Email: "u" + itoa(e.sequence) + "@exemplo.com", Name: "Usuário " + itoa(e.sequence), PasswordHash: hash, Active: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) == 0 {
		roles = []domain.Role{domain.RoleAssociado}
	}
	if !slices.Contains(roles, domain.RoleAssociado) {
		roles = append(roles, domain.RoleAssociado)
	}
	if err := e.users.SetRoles(context.Background(), u.ID, roles); err != nil {
		t.Fatal(err)
	}
	for _, r := range roles {
		if r != domain.RoleAssociado {
			m, _ := domain.NewAdminMembership(u.ID, "membro da diretoria eleita", nil, e.now)
			if _, err := e.members.Grant(context.Background(), m); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	return u, e.principal(t, u.ID)
}

func (e *ucEnv) principal(t *testing.T, userID string) authz.Principal {
	t.Helper()
	perms, err := e.roles.EffectivePermissions(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	roles, _ := e.roles.RolesOf(context.Background(), userID)
	p := authz.Principal{UserID: userID, Permissions: map[authz.Permission]struct{}{}}
	for _, r := range roles {
		p.Roles = append(p.Roles, string(r))
	}
	for _, perm := range perms {
		p.Permissions[perm] = struct{}{}
	}
	return p
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

type auditEvent struct {
	Action     string
	Outcome    string
	ActorType  string
	ActorID    *string
	EntityType string
	EntityID   string
	Reason     *string
	Before     *string
	After      *string
	Context    string
}

func (e *ucEnv) events(t *testing.T, action string) []auditEvent {
	t.Helper()
	var rows []auditEvent
	err := e.owner.Raw(`SELECT action, outcome, actor_type, actor_user_id::text AS actor_id, entity_type, entity_id, reason,
		before::text AS before, after::text AS after, context::text AS context
		FROM audit_log WHERE action = ? ORDER BY occurred_at, id`, action).Scan(&rows).Error
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func (e *ucEnv) count(t *testing.T, table string) int64 {
	t.Helper()
	var n int64
	if err := e.owner.Raw("SELECT count(*) FROM " + table).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func (e *ucEnv) breakAudit(t *testing.T) {
	t.Helper()
	if err := e.owner.Exec("ALTER TABLE audit_log RENAME TO audit_log_off").Error; err != nil {
		t.Fatal(err)
	}
}

func (e *ucEnv) sessionFor(t *testing.T, userID string) (token string, s domain.Session) {
	t.Helper()
	token, hash, _ := domain.NewSessionToken()
	csrf, _ := domain.NewCSRFToken()
	s, err := e.sessions.Create(context.Background(), userID, hash, csrf, e.now, e.now.Add(8*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	return token, s
}

func (e *ucEnv) activeSessions(t *testing.T, userID string) int64 {
	t.Helper()
	var n int64
	_ = e.owner.Raw("SELECT count(*) FROM sessions WHERE user_id = ?::uuid AND revoked_at IS NULL", userID).Scan(&n).Error
	return n
}

func (e *ucEnv) rolesOf(t *testing.T, userID string) []domain.Role {
	t.Helper()
	r, err := e.roles.RolesOf(context.Background(), userID)
	if err != nil {
		t.Fatal(err)
	}
	return r
}
