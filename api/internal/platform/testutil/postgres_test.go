//go:build integration

package testutil

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func openSQL(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("abrir conexão: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestPostgresStartsVersion16WithRoles(t *testing.T) {
	pg := SharedPostgres(t)
	admin := openSQL(t, pg.AdminDSN("tj"))

	var version string
	if err := admin.QueryRow("SHOW server_version").Scan(&version); err != nil {
		t.Fatalf("consultar versão: %v", err)
	}
	if !strings.HasPrefix(version, "16.") {
		t.Errorf("versão do PostgreSQL = %q, esperado 16.x", version)
	}

	for _, role := range []string{"tj_owner", "tj_app"} {
		var super bool
		if err := admin.QueryRow("SELECT rolsuper FROM pg_roles WHERE rolname = $1", role).Scan(&super); err != nil {
			t.Fatalf("papel %s não existe: %v", role, err)
		}
		if super {
			t.Errorf("papel %s não deveria ser superusuário", role)
		}
	}

	app := openSQL(t, pg.AppDSN("tj"))
	var user string
	if err := app.QueryRow("SELECT current_user").Scan(&user); err != nil || user != "tj_app" {
		t.Fatalf("conexão da aplicação: usuário = %q, erro = %v", user, err)
	}
	if _, err := app.Exec("CREATE TABLE fumaca_app (id int)"); err == nil {
		t.Error("tj_app não deveria conseguir criar tabelas (sem DDL)")
	}

	owner := openSQL(t, pg.OwnerDSN("tj"))
	if _, err := owner.Exec("CREATE TABLE fumaca_dono (id int)"); err != nil {
		t.Errorf("tj_owner deveria conseguir criar tabelas: %v", err)
	}
	_, _ = owner.Exec("DROP TABLE IF EXISTS fumaca_dono")
}

func TestPostgresInstanceIsSharedWithinThePackage(t *testing.T) {
	if SharedPostgres(t) != SharedPostgres(t) {
		t.Error("SharedPostgres deveria devolver a mesma instância na mesma execução de pacote")
	}
}

// Uma falha ao iniciar o contêiner deve reprovar o teste nomeando o Docker,
// nunca pular em silêncio. O processo filho roda o teste de fumaça com uma
// imagem inexistente (TJ_TEST_POSTGRES_IMAGE). O testcontainers volta ao socket
// padrão quando DOCKER_HOST é inválido, então "Docker parado" não é simulável
// sem derrubar o daemon; este teste cobre a mensagem e a reprovação.
func TestPostgresFailsNamingDockerWhenStartFails(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPostgresStartsVersion16WithRoles$", "-test.count=1", "-test.v")
	cmd.Env = append(os.Environ(), "TJ_TEST_POSTGRES_IMAGE=postgres:tag-inexistente-tj")
	out, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("o processo filho deveria falhar, mas passou:\n%s", out)
	}
	text := string(out)
	if !strings.Contains(text, "--- FAIL") || !strings.Contains(text, "Docker") {
		t.Errorf("esperava reprovação com mensagem citando Docker:\n%s", text)
	}
	if strings.Contains(text, "--- SKIP") {
		t.Errorf("o teste não pode ser pulado em silêncio:\n%s", text)
	}
}
