//go:build integration

package database_test

import (
	"strings"
	"testing"
)

const insertDocument = `INSERT INTO documents
	(owner_type, owner_id, original_filename, content_type, size_bytes, sha256, storage_key, version, supersedes_id, uploaded_by)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?::uuid)
	RETURNING id::text`

// DOC-01 AC1: a linha de metadados guarda todos os campos previstos.
func TestDocumentsRowStoresEveryMetadataColumn(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "documentos-ac1@exemplo.com")

	var id string
	err := app.Raw(insertDocument,
		"financeiro.lancamento", "l-1", "comprovante.pdf", "application/pdf", 1024,
		strings.Repeat("a", 64), "documents/11111111-1111-1111-1111-111111111111", 1, nil, uid,
	).Scan(&id).Error
	if err != nil || id == "" {
		t.Fatalf("inserir documento: id=%q err=%v", id, err)
	}

	var (
		ownerType, ownerID, filename, contentType, sha, key, status string
		size                                                        int64
		version                                                     int
	)
	err = app.Raw(`SELECT owner_type, owner_id, original_filename, content_type, size_bytes, sha256, storage_key, version, status
		FROM documents WHERE id = ?::uuid`, id).
		Row().Scan(&ownerType, &ownerID, &filename, &contentType, &size, &sha, &key, &version, &status)
	if err != nil {
		t.Fatalf("ler documento gravado: %v", err)
	}
	if ownerType != "financeiro.lancamento" || ownerID != "l-1" || filename != "comprovante.pdf" ||
		contentType != "application/pdf" || size != 1024 || len(sha) != 64 ||
		key != "documents/11111111-1111-1111-1111-111111111111" || version != 1 {
		t.Errorf("linha gravada não confere: owner_type=%q owner_id=%q filename=%q content_type=%q size=%d sha=%q key=%q version=%d",
			ownerType, ownerID, filename, contentType, size, sha, key, version)
	}
}

// DOC-01 AC12: o documento nasce ACTIVE, sem informar status.
func TestDocumentsStatusDefaultsToActive(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "documentos-ac12@exemplo.com")

	var id string
	err := app.Raw(insertDocument,
		"financeiro.lancamento", "l-2", "recibo.png", "image/png", 2048,
		strings.Repeat("b", 64), "documents/22222222-2222-2222-2222-222222222222", 1, nil, uid,
	).Scan(&id).Error
	if err != nil {
		t.Fatalf("inserir documento: %v", err)
	}

	var status string
	if err := app.Raw("SELECT status FROM documents WHERE id = ?::uuid", id).Row().Scan(&status); err != nil {
		t.Fatalf("ler status: %v", err)
	}
	if status != "ACTIVE" {
		t.Errorf("status = %q, esperado ACTIVE", status)
	}
}

// DOC-01 AC5: supersedes_id único impede que duas versões declarem suceder a
// mesma versão anterior (bifurcação da cadeia).
func TestDocumentsSupersedesIDIsUniquePreventingFork(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "documentos-ac5@exemplo.com")

	var v1 string
	if err := app.Raw(insertDocument,
		"financeiro.lancamento", "l-3", "v1.pdf", "application/pdf", 100,
		strings.Repeat("c", 64), "documents/33333333-3333-3333-3333-333333333333", 1, nil, uid,
	).Scan(&v1).Error; err != nil {
		t.Fatalf("inserir v1: %v", err)
	}

	var v2 string
	if err := app.Raw(insertDocument,
		"financeiro.lancamento", "l-3", "v2.pdf", "application/pdf", 200,
		strings.Repeat("d", 64), "documents/44444444-4444-4444-4444-444444444444", 2, v1, uid,
	).Scan(&v2).Error; err != nil {
		t.Fatalf("inserir v2 sucedendo v1: %v", err)
	}

	err := app.Raw(insertDocument,
		"financeiro.lancamento", "l-3", "v2b.pdf", "application/pdf", 210,
		strings.Repeat("e", 64), "documents/55555555-5555-5555-5555-555555555555", 2, v1, uid,
	).Scan(new(string)).Error
	if err == nil {
		t.Error("uma segunda versão sucedendo a mesma v1 deveria ser recusada (bifurcação)")
	}
}

// DOC-01 AC6 / AC12: a aplicação só tem INSERT e SELECT — nenhuma operação
// sobrescreve ou apaga um documento já gravado, e status nunca muda.
func TestDocumentsGrantsTheApplicationRoleOnlyInsertAndSelect(t *testing.T) {
	app, _ := newAuditDBs(t)
	uid := newUser(t, app, "documentos-ac6@exemplo.com")

	var id string
	if err := app.Raw(insertDocument,
		"financeiro.lancamento", "l-4", "doc.pdf", "application/pdf", 300,
		strings.Repeat("f", 64), "documents/66666666-6666-6666-6666-666666666666", 1, nil, uid,
	).Scan(&id).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	for name, stmt := range map[string]string{
		"UPDATE": "UPDATE documents SET status = 'INACTIVE'",
		"DELETE": "DELETE FROM documents",
	} {
		err := app.Exec(stmt).Error
		if err == nil || !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("%s pela aplicação: esperava permission denied, veio %v", name, err)
		}
	}

	var n int64
	if err := app.Raw("SELECT count(*) FROM documents WHERE id = ?::uuid", id).Scan(&n).Error; err != nil || n != 1 {
		t.Errorf("o documento deveria continuar intacto e legível: n=%d err=%v", n, err)
	}
}

func TestDocumentsDownMigrationRevertsAndUpCanRunAgain(t *testing.T) {
	app, owner := newAuditDBs(t)
	uid := newUser(t, app, "documentos-down@exemplo.com")
	if err := app.Raw(insertDocument,
		"financeiro.lancamento", "l-5", "doc.pdf", "application/pdf", 400,
		strings.Repeat("g", 64), "documents/77777777-7777-7777-7777-777777777777", 1, nil, uid,
	).Scan(new(string)).Error; err != nil {
		t.Fatalf("inserir: %v", err)
	}

	if err := execScript(t, owner, migrationFile(t, "000005_documents.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	var exists bool
	if err := owner.Raw("SELECT to_regclass('public.documents') IS NOT NULL").Scan(&exists).Error; err != nil || exists {
		t.Fatalf("a tabela deveria ter sido removida: exists=%v err=%v", exists, err)
	}
	if err := execScript(t, owner, migrationFile(t, "000005_documents.up.sql")); err != nil {
		t.Fatalf("up depois do down: %v", err)
	}
}

func TestDocumentsMigrationFailsClearlyWithoutTheApplicationRole(t *testing.T) {
	_, owner := newAuditDBs(t)
	if err := execScript(t, owner, migrationFile(t, "000005_documents.down.sql")); err != nil {
		t.Fatalf("down: %v", err)
	}
	script := strings.ReplaceAll(migrationFile(t, "000005_documents.up.sql"), "tj_app", "tj_app_ausente")

	err := execScript(t, owner, script)
	if err == nil || !strings.Contains(err.Error(), "papel tj_app_ausente não existe") {
		t.Errorf("esperava mensagem clara sobre o papel ausente, veio %v", err)
	}
}
