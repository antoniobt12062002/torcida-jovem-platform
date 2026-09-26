//go:build integration

package database_test

import (
	"fmt"
	"slices"
	"testing"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// schemaViolations lists the columns that break the money rules of ADR-003:
// no real, double precision or numeric column, and every column ending in
// _cents must be bigint.
func schemaViolations(db *gorm.DB) ([]string, error) {
	var rows []struct {
		TableName  string
		ColumnName string
		DataType   string
	}
	err := db.Raw(`
		SELECT table_name, column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND (data_type IN ('real', 'double precision', 'numeric')
		       OR (column_name LIKE '%\_cents' AND data_type <> 'bigint'))
		ORDER BY table_name, column_name`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, fmt.Sprintf("%s.%s (%s)", r.TableName, r.ColumnName, r.DataType))
	}
	return out, nil
}

// MNY-02.7 e MNY-02.8: o esquema produzido por todas as migrações respeita as
// regras de dinheiro. Toda migração nova passa por este teste.
func TestSchemaGuardPassesOnCurrentMigrations(t *testing.T) {
	owner := testutil.OwnerDBFor(t, testutil.NewTestDB(t))

	violations, err := schemaViolations(owner)
	if err != nil {
		t.Fatalf("consultar o esquema: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("o esquema atual viola as regras de dinheiro: %v", violations)
	}
}

func TestSchemaGuardRejectsForbiddenColumnTypes(t *testing.T) {
	owner := testutil.OwnerDBFor(t, testutil.NewTestDB(t))
	if err := owner.Exec(`CREATE TABLE fixture_types (
		price real,
		fee double precision,
		amount numeric(12,2),
		ratio numeric
	)`).Error; err != nil {
		t.Fatalf("criar fixture: %v", err)
	}

	violations, err := schemaViolations(owner)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"fixture_types.amount (numeric)",
		"fixture_types.fee (double precision)",
		"fixture_types.price (real)",
		"fixture_types.ratio (numeric)",
	}
	if !slices.Equal(violations, want) {
		t.Errorf("violações = %v, esperado %v", violations, want)
	}
}

func TestSchemaGuardRejectsCentsColumnsThatAreNotBigint(t *testing.T) {
	owner := testutil.OwnerDBFor(t, testutil.NewTestDB(t))
	if err := owner.Exec(`CREATE TABLE fixture_cents (
		gross_amount_cents integer,
		fee_cents smallint,
		net_amount_cents bigint,
		note text
	)`).Error; err != nil {
		t.Fatalf("criar fixture: %v", err)
	}

	violations, err := schemaViolations(owner)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"fixture_cents.fee_cents (smallint)",
		"fixture_cents.gross_amount_cents (integer)",
	}
	if !slices.Equal(violations, want) {
		t.Errorf("violações = %v, esperado %v (bigint e colunas sem o sufixo não podem ser reprovadas)", violations, want)
	}
}
