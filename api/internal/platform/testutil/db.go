//go:build integration

package testutil

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
)

const templateDB = "tj_template"

var (
	templateOnce sync.Once
	templateErr  error

	// cloneMu serializes CREATE DATABASE ... TEMPLATE, which PostgreSQL does not
	// allow to run concurrently against the same template.
	cloneMu   sync.Mutex
	dbCounter atomic.Int64

	databases sync.Map // *gorm.DB (application pool) -> databaseInfo
)

type databaseInfo struct {
	pg   *Postgres
	name string
}

// NewTestDB returns a pool, connected as tj_app, to a private clone of a
// database migrated once per package. The clone is dropped when the test ends.
func NewTestDB(t testing.TB) *gorm.DB {
	t.Helper()
	pg := SharedPostgres(t)
	templateOnce.Do(func() { templateErr = createTemplate(pg) })
	if templateErr != nil {
		t.Fatalf("banco-modelo dos testes: %v", templateErr)
	}

	name := fmt.Sprintf("t_%d_%d", os.Getpid(), dbCounter.Add(1))
	cloneMu.Lock()
	err := execAdmin(pg, fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s OWNER tj_owner", name, templateDB))
	cloneMu.Unlock()
	if err != nil {
		t.Fatalf("clonar banco-modelo: %v", err)
	}

	db, err := database.Open(pg.AppDSN(name), logx.New("error", io.Discard))
	if err != nil {
		_ = execAdmin(pg, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", name))
		t.Fatalf("abrir pool da aplicação: %v", err)
	}
	databases.Store(db, databaseInfo{pg: pg, name: name})

	t.Cleanup(func() {
		closePool(db)
		databases.Delete(db)
		if err := execAdmin(pg, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", name)); err != nil {
			t.Errorf("remover banco %s: %v", name, err)
		}
	})
	return db
}

// OwnerDBFor opens a tj_owner connection to the same database as db, which must
// come from NewTestDB. Tests use it for what the application role cannot do,
// such as DDL or checking that database rules also bind the owner.
func OwnerDBFor(t testing.TB, db *gorm.DB) *gorm.DB {
	t.Helper()
	v, ok := databases.Load(db)
	if !ok {
		t.Fatal("OwnerDBFor: o pool não veio de NewTestDB")
	}
	info := v.(databaseInfo)
	owner, err := database.Open(info.pg.OwnerDSN(info.name), logx.New("error", io.Discard))
	if err != nil {
		t.Fatalf("abrir conexão do dono: %v", err)
	}
	t.Cleanup(func() { closePool(owner) })
	return owner
}

func createTemplate(pg *Postgres) error {
	if err := execAdmin(pg, fmt.Sprintf("CREATE DATABASE %s OWNER tj_owner", templateDB)); err != nil {
		return err
	}
	return database.Migrate(pg.OwnerDSN(templateDB), migrationsDir())
}

func execAdmin(pg *Postgres, statement string) error {
	admin, err := sql.Open("pgx", pg.AdminDSN("tj"))
	if err != nil {
		return err
	}
	defer func() { _ = admin.Close() }()
	_, err = admin.Exec(statement)
	return err
}

func closePool(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func migrationsDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations")
}
