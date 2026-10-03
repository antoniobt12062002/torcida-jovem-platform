//go:build integration

package infra_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"slices"
	"sync"
	"testing"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

type env struct {
	app, owner *gorm.DB
	repo       *infra.RoleRepository
}

func newEnv(t *testing.T) env {
	t.Helper()
	db := testutil.NewTestDB(t)
	rec := audit.NewRecorder(db, logx.New("error", io.Discard))
	return env{app: db, owner: testutil.OwnerDBFor(t, db), repo: infra.NewRoleRepository(db, rec)}
}

// matrix mirrors the real production composition (cmd/api, cmd/bootstrap-admin:
// FoundationContributions() + financeiro.Contribution()) so that role/permission
// invariants tested here (e.g. CONSELHO_FISCAL's effective permissions) stay
// accurate to what the system actually grants — 05-permissoes/T3 moved
// financeiro out of FoundationContributions() alone.
func matrix(t *testing.T, extra ...app.Contribution) domain.Matrix {
	t.Helper()
	contributions := append([]app.Contribution{}, app.FoundationContributions()...)
	contributions = append(contributions, financeiro.Contribution())
	contributions = append(contributions, extra...)
	m, err := app.BuildMatrix(contributions...)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// removableFixture is a synthetic contribution with exactly 3 permissions,
// used to simulate "a module's permissions left the code" (a module no
// longer declaring them, so Sync must deactivate/reactivate without ever
// deleting). Before 05-permissoes/T3, these tests reused financeiro's
// foundation placeholder for this; since T3 moved financeiro out of
// FoundationContributions() entirely, they need their own disposable
// fixture — unrelated to any real module.
func removableFixture() app.Contribution {
	return app.Contribution{
		Module: "estoque",
		Permissions: []authz.Definition{
			{Permission: "estoque:item:create"},
			{Permission: "estoque:item:update"},
			{Permission: "estoque:item:read", CommonRead: true},
		},
		Grants: map[domain.Role][]authz.Permission{
			domain.RoleEstoqueLoja: {"estoque:item:create", "estoque:item:update", "estoque:item:read"},
		},
	}
}

func count(t *testing.T, db *gorm.DB, table string) int64 {
	t.Helper()
	var n int64
	if err := db.Raw("SELECT count(*) FROM " + table).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

type syncEvent struct {
	ActorType string
	Before    string
	After     string
	Context   string
}

func syncEvents(t *testing.T, e env) []syncEvent {
	t.Helper()
	var rows []syncEvent
	err := e.owner.Raw(`SELECT actor_type, before::text AS before, after::text AS after, context::text AS context
		FROM audit_log WHERE action = 'rbac.sync' ORDER BY occurred_at, id`).Scan(&rows).Error
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

func decode(t *testing.T, s string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("JSON inválido %q: %v", s, err)
	}
	return m
}

func grantsTotal(m domain.Matrix) (n int64) {
	for _, ps := range m.Grants {
		n += int64(len(ps))
	}
	return n
}

// RBAC-01.2: a primeira sincronização cria papéis, permissões e vínculos.
func TestFirstSyncCreatesRolesPermissionsAndLinksAndAuditsTheChange(t *testing.T) {
	e := newEnv(t)
	m := matrix(t)

	res, err := e.repo.Sync(context.Background(), m)

	if err != nil || !res.Changed {
		t.Fatalf("Sync = %+v, %v", res, err)
	}
	if got := count(t, e.owner, "roles"); got != int64(len(domain.AllRoles)) {
		t.Errorf("papéis = %d", got)
	}
	if got := count(t, e.owner, "permissions"); got != int64(len(m.Definitions)) {
		t.Errorf("permissões = %d", got)
	}
	if got := count(t, e.owner, "role_permissions"); got != grantsTotal(m) {
		t.Errorf("vínculos = %d, esperado %d", got, grantsTotal(m))
	}
	events := syncEvents(t, e)
	if len(events) != 1 || events[0].ActorType != "system" {
		t.Fatalf("eventos = %+v", events)
	}
	ctx := decode(t, events[0].Context)
	added, _ := ctx["permissions_added"].([]any)
	if len(added) != len(m.Definitions) {
		t.Errorf("permissions_added = %v", ctx["permissions_added"])
	}
	before := decode(t, events[0].Before)
	if perms, _ := before["permissions"].([]any); len(perms) != 0 {
		t.Errorf("antes da primeira sincronização não havia permissões: %v", before)
	}
	after := decode(t, events[0].After)
	if perms, _ := after["permissions"].([]any); len(perms) != len(m.Definitions) {
		t.Errorf("after.permissions = %v", after["permissions"])
	}
}

// RBAC-01.2 e RBAC-01.3: sincronizar de novo não muda nada nem grava auditoria.
func TestSecondSyncIsIdempotentAndWritesNoAudit(t *testing.T) {
	e := newEnv(t)
	m := matrix(t)
	if _, err := e.repo.Sync(context.Background(), m); err != nil {
		t.Fatal(err)
	}

	res, err := e.repo.Sync(context.Background(), m)

	if err != nil || res.Changed {
		t.Fatalf("Sync = %+v, %v", res, err)
	}
	if n := len(syncEvents(t, e)); n != 1 {
		t.Errorf("eventos rbac.sync = %d, esperado 1", n)
	}
	if got := count(t, e.owner, "role_permissions"); got != grantsTotal(m) {
		t.Errorf("vínculos = %d", got)
	}
}

// RBAC-01.8: permissão que sai do código fica inativa, não é apagada.
func TestRemovedPermissionBecomesInactiveAndIsNeverDeleted(t *testing.T) {
	e := newEnv(t)
	full := matrix(t, removableFixture())
	if _, err := e.repo.Sync(context.Background(), full); err != nil {
		t.Fatal(err)
	}
	reduced := matrix(t)

	res, err := e.repo.Sync(context.Background(), reduced)

	if err != nil || !res.Changed {
		t.Fatalf("Sync = %+v, %v", res, err)
	}
	if got := count(t, e.owner, "permissions"); got != int64(len(full.Definitions)) {
		t.Errorf("nenhuma permissão pode ser apagada: %d de %d", got, len(full.Definitions))
	}
	var active bool
	if err := e.owner.Raw("SELECT is_active FROM permissions WHERE name = 'estoque:item:create'").Scan(&active).Error; err != nil || active {
		t.Errorf("a permissão removida deveria estar inativa: active = %v, err = %v", active, err)
	}
	var links int64
	_ = e.owner.Raw(`SELECT count(*) FROM role_permissions rp JOIN permissions p ON p.id = rp.permission_id WHERE p.name LIKE 'estoque:%'`).Scan(&links).Error
	if links != 0 {
		t.Errorf("uma permissão inativa não mantém vínculos com papéis: %d", links)
	}
	events := syncEvents(t, e)
	ctx := decode(t, events[1].Context)
	if d, _ := ctx["permissions_deactivated"].([]any); len(d) != 3 {
		t.Errorf("permissions_deactivated = %v", ctx["permissions_deactivated"])
	}
}

func TestPermissionReturnedToTheCodeIsReactivated(t *testing.T) {
	e := newEnv(t)
	full := matrix(t, removableFixture())
	reduced := matrix(t)
	for _, m := range []domain.Matrix{full, reduced, full} {
		if _, err := e.repo.Sync(context.Background(), m); err != nil {
			t.Fatal(err)
		}
	}

	var active bool
	if err := e.owner.Raw("SELECT is_active FROM permissions WHERE name = 'estoque:item:create'").Scan(&active).Error; err != nil || !active {
		t.Errorf("deveria ter sido reativada: active = %v, err = %v", active, err)
	}
	events := syncEvents(t, e)
	if len(events) != 3 {
		t.Fatalf("eventos = %d", len(events))
	}
	if r, _ := decode(t, events[2].Context)["permissions_reactivated"].([]any); len(r) != 3 {
		t.Errorf("permissions_reactivated = %v", decode(t, events[2].Context)["permissions_reactivated"])
	}
	if got := count(t, e.owner, "role_permissions"); got != grantsTotal(full) {
		t.Errorf("os vínculos deveriam ter voltado: %d de %d", got, grantsTotal(full))
	}
}

// RBAC-01.10: a permissão nova chega ao PRESIDENTE por um vínculo explícito criado na sincronização.
func TestNewContributionPermissionReachesPresidenteOnlyThroughTheSync(t *testing.T) {
	e := newEnv(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	extra := app.Contribution{
		Module:      "estoque",
		Permissions: []authz.Definition{{Permission: "estoque:item:create"}},
		Grants:      map[domain.Role][]authz.Permission{domain.RoleEstoqueLoja: {"estoque:item:create"}},
	}
	var n int64
	q := `SELECT count(*) FROM role_permissions rp JOIN roles r ON r.id = rp.role_id JOIN permissions p ON p.id = rp.permission_id
		WHERE r.name = 'PRESIDENTE' AND p.name = 'estoque:item:create'`
	_ = e.owner.Raw(q).Scan(&n).Error
	if n != 0 {
		t.Fatal("antes da sincronização o PRESIDENTE não pode ter a permissão nova")
	}

	if _, err := e.repo.Sync(context.Background(), matrix(t, extra)); err != nil {
		t.Fatal(err)
	}

	_ = e.owner.Raw(q).Scan(&n).Error
	if n != 1 {
		t.Errorf("o PRESIDENTE deveria ter o vínculo explícito, n = %d", n)
	}
	ctx := decode(t, syncEvents(t, e)[1].Context)
	links, _ := ctx["links_added"].([]any)
	if len(links) != 2 { // ESTOQUE_LOJA e PRESIDENTE
		t.Errorf("links_added = %v", ctx["links_added"])
	}
}

func TestRemovedGrantIsListedAndDeleted(t *testing.T) {
	e := newEnv(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	var contributions []app.Contribution
	for _, c := range app.FoundationContributions() {
		if c.Module == "identity" {
			delete(c.Grants, domain.RoleDiretoria)
		}
		contributions = append(contributions, c)
	}
	contributions = append(contributions, financeiro.Contribution())
	m, _ := app.BuildMatrix(contributions...)

	if _, err := e.repo.Sync(context.Background(), m); err != nil {
		t.Fatal(err)
	}

	ctx := decode(t, syncEvents(t, e)[1].Context)
	removed, _ := ctx["links_removed"].([]any)
	if len(removed) != 1 {
		t.Fatalf("links_removed = %v", ctx["links_removed"])
	}
	if got := count(t, e.owner, "role_permissions"); got != grantsTotal(m) {
		t.Errorf("vínculos = %d, esperado %d", got, grantsTotal(m))
	}
}

// A auditoria permite reconstruir a história: o estado anterior de cada evento é o posterior do evento anterior.
func TestEachSyncEventBeforeEqualsThePreviousEventAfter(t *testing.T) {
	e := newEnv(t)
	full := matrix(t, removableFixture())
	reduced := matrix(t)
	for _, m := range []domain.Matrix{full, reduced, full} {
		if _, err := e.repo.Sync(context.Background(), m); err != nil {
			t.Fatal(err)
		}
	}

	events := syncEvents(t, e)
	for i := 1; i < len(events); i++ {
		if !reflect.DeepEqual(decode(t, events[i].Before), decode(t, events[i-1].After)) {
			t.Errorf("before do evento %d difere do after do evento %d", i, i-1)
		}
	}
	if reflect.DeepEqual(decode(t, events[1].Before), decode(t, events[1].After)) {
		t.Error("um evento que mudou algo não pode ter before igual a after")
	}
}

// AUD-01.2: se a auditoria falhar, a sincronização inteira é revertida.
func TestSyncIsRolledBackWhenTheAuditWriteFails(t *testing.T) {
	e := newEnv(t)
	if err := e.owner.Exec("ALTER TABLE audit_log RENAME TO audit_log_off").Error; err != nil {
		t.Fatal(err)
	}

	_, err := e.repo.Sync(context.Background(), matrix(t))

	if !errors.Is(err, audit.ErrWrite) {
		t.Errorf("err = %v, esperava ErrWrite", err)
	}
	for _, table := range []string{"roles", "permissions", "role_permissions"} {
		if n := count(t, e.owner, table); n != 0 {
			t.Errorf("%s deveria ter sido revertida, linhas = %d", table, n)
		}
	}
}

// Instâncias subindo juntas: o lock consultivo evita duplicar o trabalho e o evento.
func TestConcurrentSyncsDoNotDuplicateRowsOrEvents(t *testing.T) {
	e := newEnv(t)
	m := matrix(t)
	const instances = 4

	var wg sync.WaitGroup
	results := make([]bool, instances)
	errs := make([]error, instances)
	for i := 0; i < instances; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := e.repo.Sync(context.Background(), m)
			results[i], errs[i] = res.Changed, err
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("instância %d: %v", i, err)
		}
	}
	changed := 0
	for _, c := range results {
		if c {
			changed++
		}
	}
	if changed != 1 {
		t.Errorf("exatamente uma instância deveria ter alterado algo: %d", changed)
	}
	if n := len(syncEvents(t, e)); n != 1 {
		t.Errorf("eventos rbac.sync = %d, esperado 1", n)
	}
	if got := count(t, e.owner, "role_permissions"); got != grantsTotal(m) {
		t.Errorf("vínculos = %d", got)
	}
}

func assignRoles(t *testing.T, e env, email string, roles ...domain.Role) string {
	t.Helper()
	var id string
	if err := e.app.Raw(`INSERT INTO users (email, name, password_hash) VALUES (?, 'Fulano', 'h') RETURNING id::text`, email).Scan(&id).Error; err != nil {
		t.Fatal(err)
	}
	for _, r := range roles {
		if err := e.app.Exec(`INSERT INTO user_roles (user_id, role_id) SELECT ?::uuid, id FROM roles WHERE name = ?`, id, string(r)).Error; err != nil {
			t.Fatal(err)
		}
	}
	return id
}

// RBAC-01.7: a permissão efetiva é a união dos papéis; só contam as permissões ativas.
func TestEffectivePermissionsAreTheUnionOfTheActiveRolesPermissions(t *testing.T) {
	e := newEnv(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	id := assignRoles(t, e, "a@x.com", domain.RoleConselhoFiscal, domain.RoleDiretoria)

	got, err := e.repo.EffectivePermissions(context.Background(), id)

	want := []authz.Permission{
		"audit:log:read",
		"financeiro:comprovante:read", "financeiro:conta:read", "financeiro:lancamento:read",
		"financeiro:parecer:opine", "financeiro:prestacao_contas:approve", "financeiro:prestacao_contas:read",
		"financeiro:saldo:read", "identity:user:read",
	}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("permissões = %v, %v; esperado %v", got, err, want)
	}
}

func TestEffectivePermissionsExcludeInactiveOnesAndHandleUsersWithoutRoles(t *testing.T) {
	e := newEnv(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	id := assignRoles(t, e, "a@x.com", domain.RoleAdminSistema)
	if err := e.owner.Exec("UPDATE permissions SET is_active = false WHERE name = 'audit:log:read'").Error; err != nil {
		t.Fatal(err)
	}

	got, err := e.repo.EffectivePermissions(context.Background(), id)

	if err != nil || slices.Contains(got, "audit:log:read") || !slices.Contains(got, "identity:user:read") {
		t.Errorf("permissões = %v, %v", got, err)
	}
	none := assignRoles(t, e, "b@x.com")
	if got, err := e.repo.EffectivePermissions(context.Background(), none); err != nil || len(got) != 0 {
		t.Errorf("sem papéis: %v, %v", got, err)
	}
	if got, err := e.repo.EffectivePermissions(context.Background(), "0f8fad5b-d9cb-469f-a165-70867728950e"); err != nil || len(got) != 0 {
		t.Errorf("usuário inexistente: %v, %v", got, err)
	}
}

func TestRolesOfReturnsTheRoleNamesOfTheUser(t *testing.T) {
	e := newEnv(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	id := assignRoles(t, e, "a@x.com", domain.RoleTesouraria, domain.RoleAssociado)

	got, err := e.repo.RolesOf(context.Background(), id)

	want := []domain.Role{domain.RoleAssociado, domain.RoleTesouraria}
	if err != nil || !slices.Equal(got, want) {
		t.Errorf("papéis = %v, %v; esperado %v", got, err, want)
	}
}
