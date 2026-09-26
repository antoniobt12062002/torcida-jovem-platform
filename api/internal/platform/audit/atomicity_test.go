//go:build integration

package audit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// createUser is a stand-in for a use case: it changes data and audits the
// change in the same unit of work.
func createUser(ctx context.Context, e env, name string) error {
	return database.WithTx(ctx, e.app, func(ctx context.Context) error {
		tx, _ := database.TxFrom(ctx)
		if err := tx.Exec("INSERT INTO fixture_users (name) VALUES (?)", name).Error; err != nil {
			return err
		}
		return e.rec.Record(ctx, audit.Entry{
			Action: audit.UserCreate, EntityType: "user", EntityID: name, Outcome: audit.OutcomeSuccess,
			After: map[string]any{"name": name},
		})
	})
}

func atomicityEnv(t *testing.T) env {
	t.Helper()
	e := newEnv(t)
	if err := e.owner.Exec(`CREATE TABLE fixture_users (name text PRIMARY KEY);
		GRANT SELECT, INSERT ON fixture_users TO tj_app`).Error; err != nil {
		t.Fatalf("criar tabela de teste: %v", err)
	}
	return e
}

func countUsers(t *testing.T, e env) int64 {
	t.Helper()
	var n int64
	if err := e.owner.Raw("SELECT count(*) FROM fixture_users").Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

// AUD-01.1: dado e auditoria são gravados juntos.
func TestChangeAndAuditAreWrittenTogetherOnSuccess(t *testing.T) {
	e := atomicityEnv(t)

	if err := createUser(ctxWith("req-12345678", actorID), e, "ana"); err != nil {
		t.Fatalf("caso de uso: %v", err)
	}

	if n := countUsers(t, e); n != 1 {
		t.Errorf("usuários = %d", n)
	}
	rows := e.rows(t)
	if len(rows) != 1 || rows[0].EntityID != "ana" {
		t.Errorf("auditoria = %+v", rows)
	}
}

// AUD-01.2: se a auditoria falha, a alteração é desfeita.
func TestAuditFailureRollsBackTheChange(t *testing.T) {
	e := atomicityEnv(t)
	if err := e.owner.Exec("ALTER TABLE audit_log RENAME TO audit_log_off").Error; err != nil {
		t.Fatal(err)
	}

	err := createUser(ctxWith("req-12345678", actorID), e, "ana")

	if !errors.Is(err, audit.ErrWrite) {
		t.Errorf("err = %v, esperava ErrWrite (mapeável para 500 audit_failed)", err)
	}
	if n := countUsers(t, e); n != 0 {
		t.Errorf("a alteração deveria ter sido desfeita, usuários = %d", n)
	}
}

func TestInvalidAuditEntryAlsoRollsBackTheChange(t *testing.T) {
	e := atomicityEnv(t)

	err := database.WithTx(ctxWith("req-12345678", actorID), e.app, func(ctx context.Context) error {
		tx, _ := database.TxFrom(ctx)
		if err := tx.Exec("INSERT INTO fixture_users (name) VALUES ('ana')").Error; err != nil {
			return err
		}
		return e.rec.Record(ctx, audit.Entry{Action: "user.inventada", EntityType: "user", EntityID: "ana", Outcome: audit.OutcomeSuccess})
	})

	if !errors.Is(err, audit.ErrInvalidEntry) || countUsers(t, e) != 0 || len(e.rows(t)) != 0 {
		t.Errorf("err = %v, usuários = %d, auditoria = %d", err, countUsers(t, e), len(e.rows(t)))
	}
}

// AUD-01.1, no outro sentido: se a alteração falha depois de auditar, o registro
// de auditoria também é desfeito (não fica auditoria de algo que não aconteceu).
func TestAuditIsRolledBackWhenTheChangeFailsAfterwards(t *testing.T) {
	e := atomicityEnv(t)
	boom := errors.New("boom")

	err := database.WithTx(ctxWith("req-12345678", actorID), e.app, func(ctx context.Context) error {
		if err := e.rec.Record(ctx, audit.Entry{
			Action: audit.UserCreate, EntityType: "user", EntityID: "ana", Outcome: audit.OutcomeSuccess,
		}); err != nil {
			return err
		}
		return boom
	})

	if !errors.Is(err, boom) || len(e.rows(t)) != 0 {
		t.Errorf("err = %v, auditoria = %d", err, len(e.rows(t)))
	}
}
