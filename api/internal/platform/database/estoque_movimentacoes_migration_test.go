//go:build integration

package database_test

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

const insertMovimentacao = `INSERT INTO movimentacoes_estoque (produto_id, tipo, quantidade, origem, responsavel_id)
	VALUES (?, ?, ?, ?, ?) RETURNING id::text`

func newProdutoEstoque(t *testing.T, db *gorm.DB, codigo string) string {
	t.Helper()
	var id string
	if err := db.Raw(insertProdutoEstoque, codigo, "produto "+codigo, "UN").Scan(&id).Error; err != nil || id == "" {
		t.Fatalf("criar produto %s: id=%q err=%v", codigo, id, err)
	}
	return id
}

// MOV-01 AC1: a linha guarda produto, tipo, quantidade assinada e origem.
func TestMovimentacoesEstoqueRowStoresEveryColumn(t *testing.T) {
	app, _ := newAuditDBs(t)
	produtoID := newProdutoEstoque(t, app, "A1")
	userID := newUser(t, app, "estoque1@exemplo.com")

	var id string
	if err := app.Raw(insertMovimentacao, produtoID, "ENTRADA", 10, "COMPRA", userID).Scan(&id).Error; err != nil || id == "" {
		t.Fatalf("inserir movimentação: id=%q err=%v", id, err)
	}

	var tipo, origem string
	var quantidade int64
	err := app.Raw(`SELECT tipo, quantidade, origem FROM movimentacoes_estoque WHERE id = ?::uuid`, id).
		Row().Scan(&tipo, &quantidade, &origem)
	if err != nil {
		t.Fatalf("ler movimentação gravada: %v", err)
	}
	if tipo != "ENTRADA" || quantidade != 10 || origem != "COMPRA" {
		t.Errorf("linha gravada não confere: tipo=%q quantidade=%d origem=%q", tipo, quantidade, origem)
	}
}

// tipo, origem e quantidade só aceitam os valores válidos.
func TestMovimentacoesEstoqueCheckConstraintsRestrictTipoOrigemAndQuantidade(t *testing.T) {
	app, _ := newAuditDBs(t)
	produtoID := newProdutoEstoque(t, app, "A2")
	userID := newUser(t, app, "estoque2@exemplo.com")

	if err := app.Exec(insertMovimentacao, produtoID, "OUTRO", 10, "COMPRA", userID).Error; err == nil {
		t.Error("tipo fora do conjunto válido deveria ser recusado")
	}
	if err := app.Exec(insertMovimentacao, produtoID, "ENTRADA", 10, "OUTRO", userID).Error; err == nil {
		t.Error("origem fora do conjunto válido deveria ser recusada")
	}
	if err := app.Exec(insertMovimentacao, produtoID, "ENTRADA", 0, "COMPRA", userID).Error; err == nil {
		t.Error("quantidade = 0 deveria ser recusada")
	}
	if err := app.Exec(insertMovimentacao, produtoID, "AJUSTE", -5, "AJUSTE_MANUAL", userID).Error; err != nil {
		t.Errorf("quantidade negativa deveria ser aceita (ajuste): %v", err)
	}
	for _, tipo := range []string{"ENTRADA", "SAIDA", "AJUSTE", "DEVOLUCAO"} {
		if err := app.Exec(insertMovimentacao, produtoID, tipo, 1, "AJUSTE_MANUAL", userID).Error; err != nil {
			t.Errorf("tipo %s deveria ser aceito: %v", tipo, err)
		}
	}
}

// produto_id e responsavel_id são FKs de verdade.
func TestMovimentacoesEstoqueProdutoIDAndResponsavelIDAreRealForeignKeys(t *testing.T) {
	app, _ := newAuditDBs(t)
	produtoID := newProdutoEstoque(t, app, "A3")
	userID := newUser(t, app, "estoque3@exemplo.com")
	const bogus = "00000000-0000-0000-0000-000000000000"

	if err := app.Exec(insertMovimentacao, bogus, "ENTRADA", 10, "COMPRA", userID).Error; err == nil {
		t.Error("produto_id inexistente deveria ser recusado pela FK")
	}
	if err := app.Exec(insertMovimentacao, produtoID, "ENTRADA", 10, "COMPRA", bogus).Error; err == nil {
		t.Error("responsavel_id inexistente deveria ser recusado pela FK")
	}
}

// movimentacao_de_id é uma auto-referência nullable, FK de verdade quando preenchida.
func TestMovimentacoesEstoqueMovimentacaoDeIDIsANullableSelfReference(t *testing.T) {
	app, _ := newAuditDBs(t)
	produtoID := newProdutoEstoque(t, app, "A4")
	userID := newUser(t, app, "estoque4@exemplo.com")

	var saidaID string
	if err := app.Raw(insertMovimentacao, produtoID, "SAIDA", -5, "VENDA", userID).Scan(&saidaID).Error; err != nil {
		t.Fatalf("inserir saída: %v", err)
	}

	var devolucaoID string
	err := app.Raw(`INSERT INTO movimentacoes_estoque (produto_id, tipo, quantidade, origem, movimentacao_de_id, responsavel_id)
		VALUES (?, 'DEVOLUCAO', 5, 'VENDA', ?, ?) RETURNING id::text`, produtoID, saidaID, userID).Scan(&devolucaoID).Error
	if err != nil || devolucaoID == "" {
		t.Fatalf("inserir devolução: id=%q err=%v", devolucaoID, err)
	}

	const bogus = "00000000-0000-0000-0000-000000000000"
	err = app.Exec(`INSERT INTO movimentacoes_estoque (produto_id, tipo, quantidade, origem, movimentacao_de_id, responsavel_id)
		VALUES (?, 'DEVOLUCAO', 5, 'VENDA', ?::uuid, ?)`, produtoID, bogus, userID).Error
	if err == nil {
		t.Error("movimentacao_de_id inexistente deveria ser recusado pela FK")
	}
}

// AD-009/EST-D-010: tj_app tem INSERT/SELECT, nunca UPDATE nem DELETE —
// append-only, correção é sempre uma nova movimentação.
func TestMovimentacoesEstoqueGrantsInsertSelectButNeverUpdateOrDelete(t *testing.T) {
	app, _ := newAuditDBs(t)
	produtoID := newProdutoEstoque(t, app, "A5")
	userID := newUser(t, app, "estoque5@exemplo.com")

	var id string
	if err := app.Raw(insertMovimentacao, produtoID, "ENTRADA", 10, "COMPRA", userID).Scan(&id).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := app.Exec("UPDATE movimentacoes_estoque SET quantidade = 20 WHERE id = ?::uuid", id).Error; err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("UPDATE pela aplicação: esperava permission denied, veio %v", err)
	}
	if err := app.Exec("DELETE FROM movimentacoes_estoque WHERE id = ?::uuid", id).Error; err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("DELETE pela aplicação: esperava permission denied, veio %v", err)
	}
}

func TestMovimentacoesEstoqueDownMigrationRevertsAndUpCanRunAgain(t *testing.T) {
	app, owner := newAuditDBs(t)
	produtoID := newProdutoEstoque(t, app, "A6")
	userID := newUser(t, app, "estoque6@exemplo.com")
	if err := app.Raw(insertMovimentacao, produtoID, "ENTRADA", 10, "COMPRA", userID).Scan(new(string)).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := execScript(t, owner, migrationFile(t, "000009_estoque_movimentacoes.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	var exists bool
	if err := owner.Raw("SELECT to_regclass('public.movimentacoes_estoque') IS NOT NULL").Scan(&exists).Error; err != nil || exists {
		t.Fatalf("a tabela deveria ter sido removida: exists=%v err=%v", exists, err)
	}
	if err := execScript(t, owner, migrationFile(t, "000009_estoque_movimentacoes.up.sql")); err != nil {
		t.Fatalf("up depois do down: %v", err)
	}
}

func TestMovimentacoesEstoqueMigrationFailsClearlyWithoutTheApplicationRole(t *testing.T) {
	_, owner := newAuditDBs(t)
	if err := execScript(t, owner, migrationFile(t, "000009_estoque_movimentacoes.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	script := strings.ReplaceAll(migrationFile(t, "000009_estoque_movimentacoes.up.sql"), "tj_app", "tj_app_ausente")

	err := execScript(t, owner, script)
	if err == nil || !strings.Contains(err.Error(), "papel tj_app_ausente não existe") {
		t.Errorf("esperava mensagem clara sobre o papel ausente, veio %v", err)
	}
}
