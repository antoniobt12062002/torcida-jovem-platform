//go:build integration

package database_test

import (
	"strings"
	"testing"
)

const insertProdutoEstoque = `INSERT INTO produtos_estoque (codigo, nome, unidade_medida) VALUES (?, ?, ?) RETURNING id::text`

// PRD-01 AC1: a linha guarda código, nome e unidade de medida.
func TestProdutosEstoqueRowStoresEveryColumn(t *testing.T) {
	app, _ := newAuditDBs(t)

	var id string
	if err := app.Raw(insertProdutoEstoque, "CAMISA-TJ-P-PRETA", "Camisa TJ P Preta", "UN").Scan(&id).Error; err != nil || id == "" {
		t.Fatalf("inserir produto: id=%q err=%v", id, err)
	}

	var codigo, nome, unidade string
	err := app.Raw(`SELECT codigo, nome, unidade_medida FROM produtos_estoque WHERE id = ?::uuid`, id).
		Row().Scan(&codigo, &nome, &unidade)
	if err != nil {
		t.Fatalf("ler produto gravado: %v", err)
	}
	if codigo != "CAMISA-TJ-P-PRETA" || nome != "Camisa TJ P Preta" || unidade != "UN" {
		t.Errorf("linha gravada não confere: codigo=%q nome=%q unidade=%q", codigo, nome, unidade)
	}
}

// PRD-01 AC2: codigo duplicado é recusado pela UNIQUE do banco.
func TestProdutosEstoqueCodigoIsUnique(t *testing.T) {
	app, _ := newAuditDBs(t)

	if err := app.Exec(`INSERT INTO produtos_estoque (codigo, nome, unidade_medida) VALUES ('X', 'a', 'UN')`).Error; err != nil {
		t.Fatalf("inserir primeiro: %v", err)
	}
	err := app.Exec(`INSERT INTO produtos_estoque (codigo, nome, unidade_medida) VALUES ('X', 'b', 'UN')`).Error
	if err == nil {
		t.Error("codigo duplicado deveria ser recusado")
	}
}

// EST-D-007: tj_app tem INSERT/SELECT, nunca UPDATE nem DELETE — o SKU é
// estável após a criação, sem desativação nem exclusão.
func TestProdutosEstoqueGrantsInsertSelectButNeverUpdateOrDelete(t *testing.T) {
	app, _ := newAuditDBs(t)

	var id string
	if err := app.Raw(insertProdutoEstoque, "Y", "y", "UN").Scan(&id).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := app.Exec("UPDATE produtos_estoque SET nome = 'outro' WHERE id = ?::uuid", id).Error; err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("UPDATE pela aplicação: esperava permission denied, veio %v", err)
	}
	if err := app.Exec("DELETE FROM produtos_estoque WHERE id = ?::uuid", id).Error; err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("DELETE pela aplicação: esperava permission denied, veio %v", err)
	}
}

func TestProdutosEstoqueDownMigrationRevertsAndUpCanRunAgain(t *testing.T) {
	app, owner := newAuditDBs(t)
	if err := app.Raw(insertProdutoEstoque, "Z", "z", "UN").Scan(new(string)).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := execScript(t, owner, migrationFile(t, "000008_estoque_produtos.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	var exists bool
	if err := owner.Raw("SELECT to_regclass('public.produtos_estoque') IS NOT NULL").Scan(&exists).Error; err != nil || exists {
		t.Fatalf("a tabela deveria ter sido removida: exists=%v err=%v", exists, err)
	}
	if err := execScript(t, owner, migrationFile(t, "000008_estoque_produtos.up.sql")); err != nil {
		t.Fatalf("up depois do down: %v", err)
	}
}

func TestProdutosEstoqueMigrationFailsClearlyWithoutTheApplicationRole(t *testing.T) {
	_, owner := newAuditDBs(t)
	if err := execScript(t, owner, migrationFile(t, "000008_estoque_produtos.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	script := strings.ReplaceAll(migrationFile(t, "000008_estoque_produtos.up.sql"), "tj_app", "tj_app_ausente")

	err := execScript(t, owner, script)
	if err == nil || !strings.Contains(err.Error(), "papel tj_app_ausente não existe") {
		t.Errorf("esperava mensagem clara sobre o papel ausente, veio %v", err)
	}
}
