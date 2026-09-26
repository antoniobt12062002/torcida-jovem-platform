//go:build integration

package database_test

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

func newUser(t *testing.T, db *gorm.DB, email string) string {
	t.Helper()
	var id string
	err := db.Raw(`INSERT INTO users (email, name, password_hash) VALUES (?, 'Fulano', 'hash') RETURNING id::text`, email).Scan(&id).Error
	if err != nil || id == "" {
		t.Fatalf("criar usuário %s: %v", email, err)
	}
	return id
}

func TestIdentityTablesExistAndUsersHaveNoAssociadoReference(t *testing.T) {
	_, owner := newAuditDBs(t)

	for _, table := range []string{"users", "roles", "permissions", "role_permissions", "user_roles", "admin_memberships", "sessions", "login_attempts"} {
		var exists bool
		if err := owner.Raw("SELECT to_regclass(?) IS NOT NULL", "public."+table).Scan(&exists).Error; err != nil || !exists {
			t.Errorf("a tabela %s deveria existir (err = %v)", table, err)
		}
	}
	var cols []string
	if err := owner.Raw(`SELECT column_name FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = 'users' AND column_name ILIKE '%associado%'`).Scan(&cols).Error; err != nil {
		t.Fatal(err)
	}
	if len(cols) != 0 {
		t.Errorf("users não pode referenciar associado (o vínculo pertence a associados): %v", cols)
	}
}

func TestUsersEmailIsUniqueIgnoringCaseAndPasswordChangeDefaultsToFalse(t *testing.T) {
	app, _ := newAuditDBs(t)
	id := newUser(t, app, "Ana@Exemplo.com")

	if err := app.Exec(`INSERT INTO users (email, name, password_hash) VALUES ('ana@exemplo.com', 'Outra', 'h')`).Error; err == nil {
		t.Error("e-mail repetido com outra caixa deveria ser recusado")
	}
	var must bool
	var active bool
	if err := app.Raw("SELECT must_change_password, is_active FROM users WHERE id = ?", id).Row().Scan(&must, &active); err != nil || must || !active {
		t.Errorf("must_change_password = %v, is_active = %v, err = %v", must, active, err)
	}
}

func TestSessionsTokenHashIsUnique(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "a@x.com")
	insert := `INSERT INTO sessions (user_id, token_hash, csrf_token, expires_at) VALUES (?, '\x0102', 'c', now() + interval '8 hours')`

	if err := app.Exec(insert, uid).Error; err != nil {
		t.Fatalf("primeira sessão: %v", err)
	}
	if err := app.Exec(insert, uid).Error; err == nil {
		t.Error("token_hash repetido deveria ser recusado")
	}
}

func TestPermissionsNameMustFollowTheModuleResourceActionFormat(t *testing.T) {
	app, _ := newAuditDBs(t)

	for _, bad := range []string{"identity", "identity:user", "Identity:user:read", "identity:user:read:x", "identity-user:read:x"} {
		if err := app.Exec("INSERT INTO permissions (name) VALUES (?)", bad).Error; err == nil {
			t.Errorf("a permissão %q deveria ser recusada pelo banco", bad)
		}
	}
	if err := app.Exec("INSERT INTO permissions (name) VALUES ('identity:user:read')").Error; err != nil {
		t.Errorf("permissão válida recusada: %v", err)
	}
}

func TestMigrationDoesNotSeedRolesOrPermissions(t *testing.T) {
	app, _ := newAuditDBs(t)

	var roles, perms int64
	_ = app.Raw("SELECT count(*) FROM roles").Scan(&roles).Error
	_ = app.Raw("SELECT count(*) FROM permissions").Scan(&perms).Error
	if roles != 0 || perms != 0 {
		t.Errorf("a sincronização cria papéis e permissões; a migração não. roles = %d, permissions = %d", roles, perms)
	}
}

// IDN-06.5: um vínculo ativo por usuário, histórico preservado, sem DELETE.
func TestAdminMembershipsKeepOneActivePerUserAndPreserveHistory(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "a@x.com")
	grant := `INSERT INTO admin_memberships (user_id, reason, granted_by) VALUES (?, 'eleito tesoureiro na assembleia', NULL)`

	if err := app.Exec(grant, uid).Error; err != nil {
		t.Fatalf("primeiro vínculo: %v", err)
	}
	if err := app.Exec(grant, uid).Error; err == nil {
		t.Fatal("um segundo vínculo ativo deveria ser recusado")
	}
	if err := app.Exec(`UPDATE admin_memberships SET revoked_at = now(), revoke_reason = 'fim do mandato' WHERE user_id = ? AND revoked_at IS NULL`, uid).Error; err != nil {
		t.Fatalf("encerrar vínculo: %v", err)
	}
	if err := app.Exec(grant, uid).Error; err != nil {
		t.Fatalf("depois de encerrado, um novo vínculo deveria ser aceito: %v", err)
	}
	var n int64
	if err := app.Raw("SELECT count(*) FROM admin_memberships WHERE user_id = ?", uid).Scan(&n).Error; err != nil || n != 2 {
		t.Errorf("o histórico deveria ter 2 vínculos: n = %d, err = %v", n, err)
	}
}

func TestAdminMembershipsRequireAReasonAndForbidDeleteToTheApplication(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "a@x.com")

	for _, reason := range []string{"", "          ", "curto"} {
		if err := app.Exec(`INSERT INTO admin_memberships (user_id, reason) VALUES (?, ?)`, uid, reason).Error; err == nil {
			t.Errorf("o motivo %q deveria ser recusado", reason)
		}
	}
	if err := app.Exec(`INSERT INTO admin_memberships (user_id, reason) VALUES (?, 'motivo suficiente')`, uid).Error; err != nil {
		t.Fatal(err)
	}
	err := app.Exec("DELETE FROM admin_memberships").Error
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("DELETE pela aplicação: esperava permission denied, veio %v", err)
	}
}

func TestApplicationRoleCanDeleteFromTheOtherIdentityTables(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "a@x.com")

	for _, stmt := range []string{
		"DELETE FROM user_roles", "DELETE FROM role_permissions", "DELETE FROM sessions",
		"DELETE FROM login_attempts", "DELETE FROM roles", "DELETE FROM permissions",
	} {
		if err := app.Exec(stmt).Error; err != nil {
			t.Errorf("%s: %v", stmt, err)
		}
	}
	if err := app.Exec("DELETE FROM users WHERE id = ?", uid).Error; err != nil {
		t.Errorf("DELETE FROM users: %v", err)
	}
}

func TestIdentityDownMigrationRevertsAndUpCanRunAgain(t *testing.T) {
	app, owner := newAuditDBs(t)
	newUser(t, app, "a@x.com")

	if err := execScript(t, owner, migrationFile(t, "000003_identity.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	var exists bool
	if err := owner.Raw("SELECT to_regclass('public.users') IS NOT NULL").Scan(&exists).Error; err != nil || exists {
		t.Fatalf("users deveria ter sido removida: exists = %v, err = %v", exists, err)
	}
	if err := execScript(t, owner, migrationFile(t, "000003_identity.up.sql")); err != nil {
		t.Fatalf("up depois do down: %v", err)
	}
}

func TestIdentityMigrationFailsClearlyWithoutTheApplicationRole(t *testing.T) {
	_, owner := newAuditDBs(t)
	if err := execScript(t, owner, migrationFile(t, "000003_identity.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	script := strings.ReplaceAll(migrationFile(t, "000003_identity.up.sql"), "tj_app", "tj_app_ausente")

	err := execScript(t, owner, script)
	if err == nil || !strings.Contains(err.Error(), "papel tj_app_ausente não existe") {
		t.Errorf("esperava mensagem clara sobre o papel ausente, veio %v", err)
	}
}
