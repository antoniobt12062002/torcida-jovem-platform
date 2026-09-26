//go:build integration

package database_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

const insertAudit = `INSERT INTO audit_log (actor_type, actor_user_id, action, entity_type, entity_id, outcome)
	VALUES (?, ?, 'user.create', 'user', 'u-1', ?)`

func newAuditDBs(t *testing.T) (app, owner *gorm.DB) {
	t.Helper()
	app = testutil.NewTestDB(t)
	return app, testutil.OwnerDBFor(t, app)
}

func seedAudit(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec(insertAudit, "system", nil, "success").Error; err != nil {
		t.Fatalf("inserir registro de auditoria: %v", err)
	}
}

func migrationFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "migrations", name))
	if err != nil {
		t.Fatalf("ler %s: %v", name, err)
	}
	return string(b)
}

func execScript(t *testing.T, db *gorm.DB, script string) error {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	_, err = sqlDB.Exec(script)
	return err
}

// AUD-02.1: UPDATE, DELETE e TRUNCATE são recusados pelo banco, até para o dono.
func TestAuditLogRejectsUpdateDeleteAndTruncateForTheOwner(t *testing.T) {
	app, owner := newAuditDBs(t)
	seedAudit(t, app)

	for name, stmt := range map[string]string{
		"UPDATE":   "UPDATE audit_log SET action = 'x'",
		"DELETE":   "DELETE FROM audit_log",
		"TRUNCATE": "TRUNCATE audit_log",
	} {
		err := owner.Exec(stmt).Error
		if err == nil || !strings.Contains(err.Error(), "audit_log é imutável") {
			t.Errorf("%s pelo dono: esperava a exceção de imutabilidade, veio %v", name, err)
		}
	}
	var n int64
	if err := owner.Raw("SELECT count(*) FROM audit_log").Scan(&n).Error; err != nil || n != 1 {
		t.Errorf("o registro deveria continuar intacto: n=%d err=%v", n, err)
	}
}

// AUD-02.2: a aplicação tem só INSERT e SELECT.
func TestAuditLogGrantsTheApplicationRoleOnlyInsertAndSelect(t *testing.T) {
	app, _ := newAuditDBs(t)
	seedAudit(t, app)

	var n int64
	if err := app.Raw("SELECT count(*) FROM audit_log").Scan(&n).Error; err != nil || n != 1 {
		t.Fatalf("SELECT pela aplicação: n=%d err=%v", n, err)
	}
	for name, stmt := range map[string]string{
		"UPDATE":   "UPDATE audit_log SET action = 'x'",
		"DELETE":   "DELETE FROM audit_log",
		"TRUNCATE": "TRUNCATE audit_log",
	} {
		err := app.Exec(stmt).Error
		if err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s pela aplicação: esperava permission denied, veio %v", name, err)
		}
	}
}

func TestAuditLogRejectsInconsistentActorAndUnknownValues(t *testing.T) {
	app, _ := newAuditDBs(t)
	const uid = "0f8fad5b-d9cb-469f-a165-70867728950e"

	cases := []struct {
		name      string
		actorType string
		actorID   any
		outcome   string
	}{
		{"usuário sem id", "user", nil, "success"},
		{"anônimo com id", "anonymous", uid, "failure"},
		{"sistema com id", "system", uid, "success"},
		{"tipo de ator desconhecido", "robot", nil, "success"},
		{"resultado desconhecido", "system", nil, "maybe"},
	}
	for _, tc := range cases {
		if err := app.Exec(insertAudit, tc.actorType, tc.actorID, tc.outcome).Error; err == nil {
			t.Errorf("%s: o banco deveria recusar", tc.name)
		}
	}
	valid := []struct {
		actorType string
		actorID   any
		outcome   string
	}{
		{"user", uid, "success"}, {"anonymous", nil, "failure"}, {"system", nil, "denied"},
	}
	for _, v := range valid {
		if err := app.Exec(insertAudit, v.actorType, v.actorID, v.outcome).Error; err != nil {
			t.Errorf("combinação válida %v recusada: %v", v, err)
		}
	}
}

func TestAuditLogDefaultsContextToAnEmptyObject(t *testing.T) {
	app, _ := newAuditDBs(t)
	seedAudit(t, app)

	var ctx string
	if err := app.Raw("SELECT context::text FROM audit_log").Scan(&ctx).Error; err != nil || ctx != "{}" {
		t.Errorf("context = %q, err = %v", ctx, err)
	}
}

func TestAuditLogDownMigrationRevertsAndUpCanRunAgain(t *testing.T) {
	app, owner := newAuditDBs(t)
	seedAudit(t, app)

	if err := execScript(t, owner, migrationFile(t, "000002_audit_log.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	var exists bool
	if err := owner.Raw("SELECT to_regclass('public.audit_log') IS NOT NULL").Scan(&exists).Error; err != nil || exists {
		t.Fatalf("a tabela deveria ter sido removida: exists=%v err=%v", exists, err)
	}
	if err := execScript(t, owner, migrationFile(t, "000002_audit_log.up.sql")); err != nil {
		t.Fatalf("up depois do down: %v", err)
	}
}

func TestAuditLogMigrationFailsClearlyWithoutTheApplicationRole(t *testing.T) {
	_, owner := newAuditDBs(t)
	if err := execScript(t, owner, migrationFile(t, "000002_audit_log.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	script := strings.ReplaceAll(migrationFile(t, "000002_audit_log.up.sql"), "tj_app", "tj_app_ausente")

	err := execScript(t, owner, script)
	if err == nil || !strings.Contains(err.Error(), "papel tj_app_ausente não existe") {
		t.Errorf("esperava mensagem clara sobre o papel ausente, veio %v", err)
	}
}
