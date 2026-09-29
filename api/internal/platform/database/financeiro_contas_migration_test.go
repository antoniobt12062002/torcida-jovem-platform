//go:build integration

package database_test

import (
	"strings"
	"testing"
)

const insertConta = `INSERT INTO contas_contabeis (tipo, nome, parent_id) VALUES (?, ?, ?) RETURNING id::text`

// PC-01 AC1: a linha guarda tipo, nome, hierarquia e nasce ativa.
func TestContasContabeisRowStoresEveryColumnAndDefaultsToAtivo(t *testing.T) {
	app, _ := newAuditDBs(t)

	var id string
	if err := app.Raw(insertConta, "RECEITA", "Produtos", nil).Scan(&id).Error; err != nil || id == "" {
		t.Fatalf("inserir conta raiz: id=%q err=%v", id, err)
	}

	var tipo, nome string
	var ativo bool
	var parentID *string
	err := app.Raw(`SELECT tipo, nome, ativo, parent_id::text FROM contas_contabeis WHERE id = ?::uuid`, id).
		Row().Scan(&tipo, &nome, &ativo, &parentID)
	if err != nil {
		t.Fatalf("ler conta gravada: %v", err)
	}
	if tipo != "RECEITA" || nome != "Produtos" || !ativo || parentID != nil {
		t.Errorf("linha gravada não confere: tipo=%q nome=%q ativo=%v parent_id=%v", tipo, nome, ativo, parentID)
	}

	var subID string
	if err := app.Raw(insertConta, "RECEITA", "Camisetas", id).Scan(&subID).Error; err != nil {
		t.Fatalf("inserir subconta: %v", err)
	}
	var subParent string
	if err := app.Raw("SELECT parent_id::text FROM contas_contabeis WHERE id = ?::uuid", subID).Row().Scan(&subParent); err != nil || subParent != id {
		t.Errorf("parent_id da subconta = %q, esperado %q (err=%v)", subParent, id, err)
	}
}

// tipo só aceita RECEITA ou DESPESA.
func TestContasContabeisTipoIsRestrictedByCheck(t *testing.T) {
	app, _ := newAuditDBs(t)

	if err := app.Exec(`INSERT INTO contas_contabeis (tipo, nome) VALUES ('OUTRO', 'x')`).Error; err == nil {
		t.Error("tipo fora de RECEITA/DESPESA deveria ser recusado")
	}
	for _, tipo := range []string{"RECEITA", "DESPESA"} {
		if err := app.Exec(`INSERT INTO contas_contabeis (tipo, nome) VALUES (?, ?)`, tipo, "conta "+tipo).Error; err != nil {
			t.Errorf("tipo %s deveria ser aceito: %v", tipo, err)
		}
	}
}

// PC-01 AC2 (nível de schema): parent_id é uma FK de verdade — um id
// inexistente é recusado. A regra de "mesma tipo do pai" é validada em
// aplicação (T2), não aqui.
func TestContasContabeisParentIDIsARealForeignKey(t *testing.T) {
	app, _ := newAuditDBs(t)

	const bogus = "00000000-0000-0000-0000-000000000000"
	if err := app.Exec(`INSERT INTO contas_contabeis (tipo, nome, parent_id) VALUES ('RECEITA', 'x', ?::uuid)`, bogus).Error; err == nil {
		t.Error("parent_id inexistente deveria ser recusado pela FK")
	}
}

// PC-03 AC1/AC4: tj_app tem INSERT/SELECT/UPDATE, nunca DELETE — desativação
// (UPDATE de ativo) é possível, exclusão nunca é.
func TestContasContabeisGrantsInsertSelectUpdateButNeverDelete(t *testing.T) {
	app, _ := newAuditDBs(t)

	var id string
	if err := app.Raw(insertConta, "DESPESA", "Material", nil).Scan(&id).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := app.Exec("UPDATE contas_contabeis SET ativo = false WHERE id = ?::uuid", id).Error; err != nil {
		t.Errorf("UPDATE (desativação) deveria ser permitido: %v", err)
	}
	var ativo bool
	if err := app.Raw("SELECT ativo FROM contas_contabeis WHERE id = ?::uuid", id).Row().Scan(&ativo); err != nil || ativo {
		t.Errorf("ativo deveria ser false após o UPDATE: ativo=%v err=%v", ativo, err)
	}

	err := app.Exec("DELETE FROM contas_contabeis WHERE id = ?::uuid", id).Error
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("DELETE pela aplicação: esperava permission denied, veio %v", err)
	}
	var n int64
	if err := app.Raw("SELECT count(*) FROM contas_contabeis WHERE id = ?::uuid", id).Scan(&n).Error; err != nil || n != 1 {
		t.Errorf("a conta deveria continuar intacta: n=%d err=%v", n, err)
	}
}

func TestContasContabeisDownMigrationRevertsAndUpCanRunAgain(t *testing.T) {
	app, owner := newAuditDBs(t)
	if err := app.Raw(insertConta, "RECEITA", "x", nil).Scan(new(string)).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := execScript(t, owner, migrationFile(t, "000006_financeiro_contas.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	var exists bool
	if err := owner.Raw("SELECT to_regclass('public.contas_contabeis') IS NOT NULL").Scan(&exists).Error; err != nil || exists {
		t.Fatalf("a tabela deveria ter sido removida: exists=%v err=%v", exists, err)
	}
	if err := execScript(t, owner, migrationFile(t, "000006_financeiro_contas.up.sql")); err != nil {
		t.Fatalf("up depois do down: %v", err)
	}
}

func TestContasContabeisMigrationFailsClearlyWithoutTheApplicationRole(t *testing.T) {
	_, owner := newAuditDBs(t)
	if err := execScript(t, owner, migrationFile(t, "000006_financeiro_contas.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	script := strings.ReplaceAll(migrationFile(t, "000006_financeiro_contas.up.sql"), "tj_app", "tj_app_ausente")

	err := execScript(t, owner, script)
	if err == nil || !strings.Contains(err.Error(), "papel tj_app_ausente não existe") {
		t.Errorf("esperava mensagem clara sobre o papel ausente, veio %v", err)
	}
}
