//go:build integration

package database_test

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync/atomic"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

const migrationsDir = "../../../migrations"

var dbCounter atomic.Int64

func freshDatabase(t *testing.T, pg *testutil.Postgres) string {
	t.Helper()
	name := fmt.Sprintf("migrate_%d_%d", os.Getpid(), dbCounter.Add(1))
	admin, err := sql.Open("pgx", pg.AdminDSN("tj"))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if _, err := admin.Exec(fmt.Sprintf("CREATE DATABASE %s OWNER tj_owner", name)); err != nil {
		t.Fatalf("criar banco: %v", err)
	}
	t.Cleanup(func() {
		a, err := sql.Open("pgx", pg.AdminDSN("tj"))
		if err != nil {
			return
		}
		defer a.Close()
		_, _ = a.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", name))
	})
	return name
}

func latestMigrationVersion(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`^(\d+)_.*\.up\.sql$`)
	latest := 0
	for _, e := range entries {
		if m := re.FindStringSubmatch(filepath.Base(e.Name())); m != nil {
			v, _ := strconv.Atoi(m[1])
			latest = max(latest, v)
		}
	}
	if latest == 0 {
		t.Fatal("nenhuma migração encontrada")
	}
	return latest
}

func schemaVersion(t *testing.T, pg *testutil.Postgres, name string) (version int, dirty bool) {
	t.Helper()
	db, err := sql.Open("pgx", pg.OwnerDSN(name))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatalf("ler schema_migrations: %v", err)
	}
	return version, dirty
}

func TestMigrateAppliesEveryMigrationAndEnablesPgcrypto(t *testing.T) {
	pg := testutil.SharedPostgres(t)
	name := freshDatabase(t, pg)

	if err := database.Migrate(pg.OwnerDSN(name), migrationsDir); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	version, dirty := schemaVersion(t, pg, name)
	if version != latestMigrationVersion(t) || dirty {
		t.Errorf("versão = %d (sujo: %v), esperado %d limpo", version, dirty, latestMigrationVersion(t))
	}

	db, err := sql.Open("pgx", pg.OwnerDSN(name))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var ext string
	if err := db.QueryRow("SELECT extname FROM pg_extension WHERE extname = 'pgcrypto'").Scan(&ext); err != nil {
		t.Fatalf("pgcrypto deveria estar instalado: %v", err)
	}
	var id string
	if err := db.QueryRow("SELECT gen_random_uuid()::text").Scan(&id); err != nil || len(id) != 36 {
		t.Errorf("gen_random_uuid() = %q, erro = %v", id, err)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	pg := testutil.SharedPostgres(t)
	name := freshDatabase(t, pg)

	if err := database.Migrate(pg.OwnerDSN(name), migrationsDir); err != nil {
		t.Fatalf("primeira execução: %v", err)
	}
	before, _ := schemaVersion(t, pg, name)

	if err := database.Migrate(pg.OwnerDSN(name), migrationsDir); err != nil {
		t.Fatalf("segunda execução deveria ser idempotente: %v", err)
	}
	after, dirty := schemaVersion(t, pg, name)
	if before != after || dirty {
		t.Errorf("versão mudou ou ficou suja: antes %d, depois %d, sujo %v", before, after, dirty)
	}
}
