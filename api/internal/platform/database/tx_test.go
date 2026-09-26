//go:build integration

package database_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

func txFixture(t *testing.T) *gorm.DB {
	t.Helper()
	app := testutil.NewTestDB(t)
	owner := testutil.OwnerDBFor(t, app)
	if err := owner.Exec(`CREATE TABLE fixture_items (name text PRIMARY KEY);
		GRANT SELECT, INSERT, DELETE ON fixture_items TO tj_app`).Error; err != nil {
		t.Fatalf("criar tabela de teste: %v", err)
	}
	return app
}

func countItems(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := db.Raw("SELECT count(*) FROM fixture_items").Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func insertItem(ctx context.Context, name string) error {
	tx, ok := database.TxFrom(ctx)
	if !ok {
		return errors.New("sem transação no contexto")
	}
	return tx.Exec("INSERT INTO fixture_items (name) VALUES (?)", name).Error
}

// AUD-01.1: as escritas do callback e a transação que as contém andam juntas.
func TestWithTxCommitsWhenTheCallbackSucceeds(t *testing.T) {
	db := txFixture(t)

	err := database.WithTx(context.Background(), db, func(ctx context.Context) error {
		return insertItem(ctx, "a")
	})

	if err != nil || countItems(t, db) != 1 {
		t.Errorf("err = %v, itens = %d", err, countItems(t, db))
	}
}

func TestWithTxRollsBackAndReturnsTheCallbackError(t *testing.T) {
	db := txFixture(t)
	boom := errors.New("boom")

	err := database.WithTx(context.Background(), db, func(ctx context.Context) error {
		if err := insertItem(ctx, "a"); err != nil {
			return err
		}
		return boom
	})

	if !errors.Is(err, boom) {
		t.Errorf("err = %v, esperava o erro do callback", err)
	}
	if n := countItems(t, db); n != 0 {
		t.Errorf("a escrita deveria ter sido revertida, itens = %d", n)
	}
}

func TestWithTxRollsBackAndRepanicsWhenTheCallbackPanics(t *testing.T) {
	db := txFixture(t)

	func() {
		defer func() {
			if recover() == nil {
				t.Error("o panic deveria ser propagado")
			}
		}()
		_ = database.WithTx(context.Background(), db, func(ctx context.Context) error {
			_ = insertItem(ctx, "a")
			panic("falha inesperada")
		})
	}()

	if n := countItems(t, db); n != 0 {
		t.Errorf("a escrita deveria ter sido revertida, itens = %d", n)
	}
}

// AUD-01.6: sem transação no contexto, quem exige transação consegue saber.
func TestTxFromReportsWhenThereIsNoTransaction(t *testing.T) {
	if tx, ok := database.TxFrom(context.Background()); ok || tx != nil {
		t.Errorf("TxFrom sem transação = (%v, %v)", tx, ok)
	}
}

func TestTxFromReturnsTheTransactionInsideWithTx(t *testing.T) {
	db := txFixture(t)

	_ = database.WithTx(context.Background(), db, func(ctx context.Context) error {
		if tx, ok := database.TxFrom(ctx); !ok || tx == nil || tx == db {
			t.Errorf("TxFrom dentro da transação = (%v, %v)", tx, ok)
		}
		return nil
	})
}

// Casos de uso compostos reaproveitam a transação em andamento: uma só
// unidade de trabalho por caso de uso.
func TestWithTxInsideATransactionJoinsIt(t *testing.T) {
	db := txFixture(t)
	boom := errors.New("boom")

	err := database.WithTx(context.Background(), db, func(ctx context.Context) error {
		inner := database.WithTx(ctx, db, func(ctx context.Context) error {
			return insertItem(ctx, "a")
		})
		if inner != nil {
			return inner
		}
		return boom
	})

	if !errors.Is(err, boom) || countItems(t, db) != 0 {
		t.Errorf("a transação interna deveria ter sido revertida junto: err = %v, itens = %d", err, countItems(t, db))
	}
}
