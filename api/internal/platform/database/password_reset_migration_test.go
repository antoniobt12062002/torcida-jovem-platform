//go:build integration

package database_test

import (
	"strings"
	"testing"
)

func TestPasswordResetTablesExistAndNeverStoreTheToken(t *testing.T) {
	_, owner := newAuditDBs(t)

	for _, table := range []string{"password_reset_tokens", "password_reset_requests", "password_change_attempts"} {
		var exists bool
		if err := owner.Raw("SELECT to_regclass(?) IS NOT NULL", "public."+table).Scan(&exists).Error; err != nil || !exists {
			t.Errorf("a tabela %s deveria existir (err = %v)", table, err)
		}
	}
	var cols []string
	if err := owner.Raw(`SELECT column_name FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name IN ('password_reset_tokens', 'password_reset_requests')
		  AND (column_name = 'token' OR column_name ILIKE '%email' OR column_name ILIKE '%link%' OR column_name ILIKE '%password%')`).Scan(&cols).Error; err != nil {
		t.Fatal(err)
	}
	if len(cols) != 0 {
		t.Errorf("nenhuma coluna pode guardar o token, o e-mail ou o link: %v", cols)
	}
}

// IDN-07.2: só o hash do token, e único.
func TestResetTokenHashIsUniqueAndTheApplicationCannotDeleteTokens(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "ana@exemplo.com")
	insert := `INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES (?::uuid, '\x0102', now() + interval '30 minutes')`

	if err := app.Exec(insert, uid).Error; err != nil {
		t.Fatalf("primeiro token: %v", err)
	}
	if err := app.Exec(insert, uid).Error; err == nil {
		t.Error("token_hash repetido deveria ser recusado")
	}
	if err := app.Exec("UPDATE password_reset_tokens SET used_at = now()").Error; err != nil {
		t.Errorf("a aplicação precisa marcar o token como usado: %v", err)
	}
	err := app.Exec("DELETE FROM password_reset_tokens").Error
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("DELETE pela aplicação: esperava permission denied, veio %v", err)
	}
}

func TestRequestAndAttemptCountersAcceptInsertsAndCleanup(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "ana@exemplo.com")

	if err := app.Exec(`INSERT INTO password_reset_requests (email_hash) VALUES ('\xabcd')`).Error; err != nil {
		t.Errorf("solicitação: %v", err)
	}
	if err := app.Exec(`INSERT INTO password_change_attempts (user_id) VALUES (?::uuid)`, uid).Error; err != nil {
		t.Errorf("tentativa de troca: %v", err)
	}
	for _, stmt := range []string{"DELETE FROM password_reset_requests", "DELETE FROM password_change_attempts"} {
		if err := app.Exec(stmt).Error; err != nil {
			t.Errorf("%s: %v", stmt, err)
		}
	}
	if err := app.Exec(`INSERT INTO password_change_attempts (user_id) VALUES ('0f8fad5b-d9cb-469f-a165-70867728950e')`).Error; err == nil {
		t.Error("a tentativa de troca de senha exige um usuário existente")
	}
}

func TestPasswordResetDownMigrationRevertsAndUpCanRunAgain(t *testing.T) {
	_, owner := newAuditDBs(t)

	if err := execScript(t, owner, migrationFile(t, "000004_password_reset.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	var exists bool
	if err := owner.Raw("SELECT to_regclass('public.password_reset_tokens') IS NOT NULL").Scan(&exists).Error; err != nil || exists {
		t.Fatalf("a tabela deveria ter sido removida: exists = %v, err = %v", exists, err)
	}
	if err := execScript(t, owner, migrationFile(t, "000004_password_reset.up.sql")); err != nil {
		t.Fatalf("up depois do down: %v", err)
	}
}

func TestPasswordResetMigrationFailsClearlyWithoutTheApplicationRole(t *testing.T) {
	_, owner := newAuditDBs(t)
	if err := execScript(t, owner, migrationFile(t, "000004_password_reset.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	script := strings.ReplaceAll(migrationFile(t, "000004_password_reset.up.sql"), "tj_app", "tj_app_ausente")

	err := execScript(t, owner, script)
	if err == nil || !strings.Contains(err.Error(), "papel tj_app_ausente não existe") {
		t.Errorf("esperava mensagem clara sobre o papel ausente, veio %v", err)
	}
}
