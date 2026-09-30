//go:build integration

package database_test

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

// valor_liquido_cents é NOT NULL sem default: a aplicação (T2) sempre o
// calcula e grava na criação, então todo INSERT de teste o fornece.
const insertLancamento = `INSERT INTO lancamentos (tipo, conta_id, valor_bruto_cents, valor_liquido_cents, forma_pagamento, criado_por)
	VALUES (?, ?, ?, ?, ?, ?) RETURNING id::text`

func newConta(t *testing.T, db *gorm.DB, tipo, nome string) string {
	t.Helper()
	var id string
	if err := db.Raw(insertConta, tipo, nome, nil).Scan(&id).Error; err != nil || id == "" {
		t.Fatalf("criar conta %s: id=%q err=%v", nome, id, err)
	}
	return id
}

// LAN-01 AC1: a linha guarda tipo, conta, valores e nasce CRIADA.
func TestLancamentosRowStoresEveryColumnAndDefaultsToCriada(t *testing.T) {
	app, _ := newAuditDBs(t)
	contaID := newConta(t, app, "RECEITA", "Produtos")
	userID := newUser(t, app, "tesouraria@exemplo.com")

	var id string
	if err := app.Raw(insertLancamento, "RECEITA", contaID, 10000, 10000, "PIX", userID).Scan(&id).Error; err != nil || id == "" {
		t.Fatalf("inserir lançamento: id=%q err=%v", id, err)
	}

	var tipo, status, formaPagamento string
	var valorBruto, taxa, valorLiquido int64
	err := app.Raw(`SELECT tipo, status, forma_pagamento, valor_bruto_cents, taxa_cents, valor_liquido_cents
		FROM lancamentos WHERE id = ?::uuid`, id).
		Row().Scan(&tipo, &status, &formaPagamento, &valorBruto, &taxa, &valorLiquido)
	if err != nil {
		t.Fatalf("ler lançamento gravado: %v", err)
	}
	if tipo != "RECEITA" || status != "CRIADA" || formaPagamento != "PIX" || valorBruto != 10000 || taxa != 0 || valorLiquido != 10000 {
		t.Errorf("linha gravada não confere: tipo=%q status=%q forma=%q bruto=%d taxa=%d liquido=%d",
			tipo, status, formaPagamento, valorBruto, taxa, valorLiquido)
	}
}

// tipo, forma_pagamento e status só aceitam os valores válidos.
func TestLancamentosCheckConstraintsRestrictTipoFormaPagamentoAndStatus(t *testing.T) {
	app, _ := newAuditDBs(t)
	contaID := newConta(t, app, "RECEITA", "Produtos")
	userID := newUser(t, app, "tesouraria2@exemplo.com")

	if err := app.Exec(insertLancamento, "OUTRO", contaID, 100, 100, "PIX", userID).Error; err == nil {
		t.Error("tipo fora de RECEITA/DESPESA deveria ser recusado")
	}
	if err := app.Exec(insertLancamento, "RECEITA", contaID, 100, 100, "BOLETO", userID).Error; err == nil {
		t.Error("forma_pagamento fora do conjunto válido deveria ser recusado")
	}
	if err := app.Exec(`INSERT INTO lancamentos (tipo, conta_id, valor_bruto_cents, valor_liquido_cents, forma_pagamento, status, criado_por)
		VALUES ('RECEITA', ?, 100, 100, 'PIX', 'QUEBRADA', ?)`, contaID, userID).Error; err == nil {
		t.Error("status fora do conjunto válido deveria ser recusado")
	}
	for _, forma := range []string{"PIX", "CARTAO", "DINHEIRO", "TRANSFERENCIA", "OUTROS"} {
		if err := app.Exec(insertLancamento, "RECEITA", contaID, 100, 100, forma, userID).Error; err != nil {
			t.Errorf("forma_pagamento %s deveria ser aceita: %v", forma, err)
		}
	}
}

// valor_bruto_cents > 0 e taxa_cents >= 0.
func TestLancamentosValorBrutoMustBePositiveAndTaxaCannotBeNegative(t *testing.T) {
	app, _ := newAuditDBs(t)
	contaID := newConta(t, app, "RECEITA", "Produtos")
	userID := newUser(t, app, "tesouraria3@exemplo.com")

	if err := app.Exec(insertLancamento, "RECEITA", contaID, 0, 0, "PIX", userID).Error; err == nil {
		t.Error("valor_bruto_cents = 0 deveria ser recusado")
	}
	if err := app.Exec(`INSERT INTO lancamentos (tipo, conta_id, valor_bruto_cents, taxa_cents, valor_liquido_cents, forma_pagamento, criado_por)
		VALUES ('RECEITA', ?, 100, -1, 101, 'PIX', ?)`, contaID, userID).Error; err == nil {
		t.Error("taxa_cents negativa deveria ser recusada")
	}
}

// conta_id é NOT NULL e uma FK de verdade para contas_contabeis.
func TestLancamentosContaIDIsRequiredAndIsARealForeignKey(t *testing.T) {
	app, _ := newAuditDBs(t)
	userID := newUser(t, app, "tesouraria4@exemplo.com")

	if err := app.Exec(`INSERT INTO lancamentos (tipo, conta_id, valor_bruto_cents, valor_liquido_cents, forma_pagamento, criado_por)
		VALUES ('RECEITA', NULL, 100, 100, 'PIX', ?)`, userID).Error; err == nil {
		t.Error("conta_id NULL deveria ser recusado")
	}
	const bogus = "00000000-0000-0000-0000-000000000000"
	if err := app.Exec(insertLancamento, "RECEITA", bogus, 100, 100, "PIX", userID).Error; err == nil {
		t.Error("conta_id inexistente deveria ser recusado pela FK")
	}
}

// devolucao_de_id é uma auto-referência nullable, FK de verdade quando preenchida.
func TestLancamentosDevolucaoDeIDIsANullableSelfReference(t *testing.T) {
	app, _ := newAuditDBs(t)
	contaReceita := newConta(t, app, "RECEITA", "Produtos")
	contaDespesa := newConta(t, app, "DESPESA", "Devoluções")
	userID := newUser(t, app, "tesouraria5@exemplo.com")

	var receitaID string
	if err := app.Raw(insertLancamento, "RECEITA", contaReceita, 10000, 10000, "PIX", userID).Scan(&receitaID).Error; err != nil {
		t.Fatalf("inserir receita: %v", err)
	}

	var devolucaoID string
	err := app.Raw(`INSERT INTO lancamentos (tipo, conta_id, valor_bruto_cents, valor_liquido_cents, forma_pagamento, devolucao_de_id, criado_por)
		VALUES ('DESPESA', ?, 10000, 10000, 'PIX', ?, ?) RETURNING id::text`, contaDespesa, receitaID, userID).Scan(&devolucaoID).Error
	if err != nil || devolucaoID == "" {
		t.Fatalf("inserir devolução: id=%q err=%v", devolucaoID, err)
	}

	const bogus = "00000000-0000-0000-0000-000000000000"
	err = app.Exec(`INSERT INTO lancamentos (tipo, conta_id, valor_bruto_cents, valor_liquido_cents, forma_pagamento, devolucao_de_id, criado_por)
		VALUES ('DESPESA', ?, 10000, 10000, 'PIX', ?::uuid, ?)`, contaDespesa, bogus, userID).Error
	if err == nil {
		t.Error("devolucao_de_id inexistente deveria ser recusado pela FK")
	}
}

// PC-01 AC1 (nível de schema, dona: tj_app): INSERT/SELECT/UPDATE, nunca DELETE.
func TestLancamentosGrantsInsertSelectUpdateButNeverDelete(t *testing.T) {
	app, _ := newAuditDBs(t)
	contaID := newConta(t, app, "DESPESA", "Material")
	userID := newUser(t, app, "tesouraria6@exemplo.com")

	var id string
	if err := app.Raw(insertLancamento, "DESPESA", contaID, 9600, 9600, "PIX", userID).Scan(&id).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := app.Exec("UPDATE lancamentos SET status = 'CANCELADA' WHERE id = ?::uuid", id).Error; err != nil {
		t.Errorf("UPDATE deveria ser permitido: %v", err)
	}

	err := app.Exec("DELETE FROM lancamentos WHERE id = ?::uuid", id).Error
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("DELETE pela aplicação: esperava permission denied, veio %v", err)
	}
	var n int64
	if err := app.Raw("SELECT count(*) FROM lancamentos WHERE id = ?::uuid", id).Scan(&n).Error; err != nil || n != 1 {
		t.Errorf("o lançamento deveria continuar intacto: n=%d err=%v", n, err)
	}
}

func TestLancamentosDownMigrationRevertsAndUpCanRunAgain(t *testing.T) {
	app, owner := newAuditDBs(t)
	contaID := newConta(t, app, "RECEITA", "x")
	userID := newUser(t, app, "tesouraria7@exemplo.com")
	if err := app.Raw(insertLancamento, "RECEITA", contaID, 100, 100, "PIX", userID).Scan(new(string)).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := execScript(t, owner, migrationFile(t, "000007_financeiro_lancamentos.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	var exists bool
	if err := owner.Raw("SELECT to_regclass('public.lancamentos') IS NOT NULL").Scan(&exists).Error; err != nil || exists {
		t.Fatalf("a tabela deveria ter sido removida: exists=%v err=%v", exists, err)
	}
	if err := execScript(t, owner, migrationFile(t, "000007_financeiro_lancamentos.up.sql")); err != nil {
		t.Fatalf("up depois do down: %v", err)
	}
}

func TestLancamentosMigrationFailsClearlyWithoutTheApplicationRole(t *testing.T) {
	_, owner := newAuditDBs(t)
	if err := execScript(t, owner, migrationFile(t, "000007_financeiro_lancamentos.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	script := strings.ReplaceAll(migrationFile(t, "000007_financeiro_lancamentos.up.sql"), "tj_app", "tj_app_ausente")

	err := execScript(t, owner, script)
	if err == nil || !strings.Contains(err.Error(), "papel tj_app_ausente não existe") {
		t.Errorf("esperava mensagem clara sobre o papel ausente, veio %v", err)
	}
}
