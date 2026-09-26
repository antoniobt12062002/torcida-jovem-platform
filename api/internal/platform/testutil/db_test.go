//go:build integration

package testutil

import (
	"database/sql"
	"strings"
	"sync"
	"testing"

	"gorm.io/gorm"
)

func currentDatabase(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var name string
	if err := db.Raw("SELECT current_database()").Scan(&name).Error; err != nil {
		t.Fatalf("current_database: %v", err)
	}
	return name
}

func TestNewTestDBDataDoesNotLeakBetweenTests(t *testing.T) {
	t.Run("escreve", func(t *testing.T) {
		db := NewTestDB(t)
		owner := OwnerDBFor(t, db)
		if err := owner.Exec("CREATE TABLE isolamento (x int)").Error; err != nil {
			t.Fatalf("criar tabela: %v", err)
		}
		if err := owner.Exec("INSERT INTO isolamento VALUES (1)").Error; err != nil {
			t.Fatalf("inserir: %v", err)
		}
	})

	t.Run("le", func(t *testing.T) {
		db := NewTestDB(t)
		owner := OwnerDBFor(t, db)
		err := owner.Exec("SELECT count(*) FROM isolamento").Error
		if err == nil || !strings.Contains(err.Error(), "isolamento") {
			t.Fatalf("a tabela criada em outro teste não deveria existir aqui, erro = %v", err)
		}
	})
}

func TestNewTestDBPoolUsesTheApplicationRole(t *testing.T) {
	db := NewTestDB(t)

	var user string
	if err := db.Raw("SELECT current_user").Scan(&user).Error; err != nil || user != "tj_app" {
		t.Fatalf("usuário do pool = %q, esperado tj_app (erro %v)", user, err)
	}
	if err := db.Exec("CREATE TABLE nao_deveria (x int)").Error; err == nil {
		t.Error("tj_app não deveria conseguir criar tabelas")
	}

	var ownerUser string
	if err := OwnerDBFor(t, db).Raw("SELECT current_user").Scan(&ownerUser).Error; err != nil || ownerUser != "tj_owner" {
		t.Errorf("usuário do dono = %q, esperado tj_owner (erro %v)", ownerUser, err)
	}
	if currentDatabase(t, OwnerDBFor(t, db)) != currentDatabase(t, db) {
		t.Error("a conexão do dono deveria apontar para o mesmo banco do pool")
	}
}

func TestNewTestDBIsMigratedFromTheTemplate(t *testing.T) {
	db := NewTestDB(t)
	var n int
	if err := OwnerDBFor(t, db).Raw("SELECT count(*) FROM pg_extension WHERE extname = 'pgcrypto'").Scan(&n).Error; err != nil || n != 1 {
		t.Errorf("pgcrypto (migração 000001) deveria estar no clone: n = %d, erro = %v", n, err)
	}
}

func TestNewTestDBDatabaseIsRemovedAfterTheTest(t *testing.T) {
	var name string
	t.Run("cria", func(t *testing.T) {
		name = currentDatabase(t, NewTestDB(t))
	})
	if name == "" {
		t.Fatal("o subteste deveria ter registrado o nome do banco")
	}

	admin, err := sql.Open("pgx", SharedPostgres(t).AdminDSN("tj"))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	var exists bool
	if err := admin.QueryRow("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Errorf("o banco %s deveria ter sido removido ao final do teste", name)
	}
}

func TestNewTestDBSupportsParallelTests(t *testing.T) {
	var mu sync.Mutex
	seen := map[string]bool{}
	t.Run("grupo", func(t *testing.T) {
		for _, n := range []string{"a", "b", "c", "d"} {
			t.Run(n, func(t *testing.T) {
				t.Parallel()
				name := currentDatabase(t, NewTestDB(t))
				mu.Lock()
				defer mu.Unlock()
				if seen[name] {
					t.Errorf("banco %s foi entregue a dois testes", name)
				}
				seen[name] = true
			})
		}
	})
	if len(seen) != 4 {
		t.Errorf("esperava 4 bancos distintos, veio %d", len(seen))
	}
}
