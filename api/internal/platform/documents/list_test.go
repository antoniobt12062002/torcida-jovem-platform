//go:build integration

package documents_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
)

// DOC-04 AC1: toda versão do dono, mais nova primeiro, sem gerar URL.
func TestListByOwnerReturnsEveryVersionNewestFirstWithoutGeneratingAURL(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc04ac1@exemplo.com")

	v1, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-30", Filename: "v1.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store v1: %v", err)
	}
	v2content := append(append([]byte{}, pdfBytes()...), []byte(" v2")...)
	v2, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-30", Filename: "v2.pdf",
		Content: bytes.NewReader(v2content), UploadedBy: uid, Supersedes: &v1.ID,
	})
	if err != nil {
		t.Fatalf("Store v2: %v", err)
	}
	// De outro dono: não deve aparecer na listagem do dono l-30.
	if _, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-31", Filename: "outro.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	}); err != nil {
		t.Fatalf("Store outro dono: %v", err)
	}

	before := e.st.presignCalls
	docs, err := e.svc.ListByOwner(context.Background(), documents.ListInput{
		Actor: principal(uid, readPermission), RequiredPermission: readPermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-30",
	})
	if err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}
	if e.st.presignCalls != before {
		t.Error("ListByOwner não deveria gerar nenhuma URL assinada")
	}
	if len(docs) != 2 || docs[0].ID != v2.ID || docs[1].ID != v1.ID {
		t.Fatalf("esperava [v2, v1] (mais nova primeiro), veio %+v", docs)
	}
}

// DOC-04 AC2: sem a permissão exigida, nenhuma linha é devolvida.
func TestListByOwnerRequiresTheCallerSuppliedPermissionAndReturnsNoRows(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc04ac2@exemplo.com")
	if _, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-32", Filename: "doc.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	}); err != nil {
		t.Fatalf("Store: %v", err)
	}

	docs, err := e.svc.ListByOwner(context.Background(), documents.ListInput{
		Actor: principal(uid), RequiredPermission: readPermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-32",
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	if len(docs) != 0 {
		t.Errorf("esperava nenhuma linha, veio %d", len(docs))
	}
}

// DOC-04 AC3: owner_type fora do formato técnico é recusado sem consultar o
// banco — provado com um contexto já cancelado: se a consulta fosse
// executada, o erro devolvido seria de contexto cancelado, não
// ErrOwnerTypeInvalid.
func TestListByOwnerRejectsAnInvalidOwnerTypeWithoutQueryingTheDatabase(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc04ac3@exemplo.com")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for name, ownerType := range map[string]string{
		"vazio":               "",
		"sem ponto":           "financeirolancamento",
		"maiúscula":           "Financeiro.Lancamento",
		"maior que 100 chars": "financeiro." + string(make([]byte, 100)),
	} {
		_, err := e.svc.ListByOwner(ctx, documents.ListInput{
			Actor: principal(uid, readPermission), RequiredPermission: readPermission,
			OwnerType: ownerType, OwnerID: "l-33",
		})
		if !errors.Is(err, documents.ErrOwnerTypeInvalid) {
			t.Errorf("%s: err = %v, esperado ErrOwnerTypeInvalid (banco não deveria ter sido consultado)", name, err)
		}
	}
}

// DOC-04 AC4: a listagem não grava nenhum evento de auditoria.
func TestListByOwnerDoesNotRecordAnAuditEntry(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc04ac4@exemplo.com")
	if _, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-34", Filename: "doc.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	}); err != nil {
		t.Fatalf("Store: %v", err)
	}

	var before int64
	if err := e.app.Raw("SELECT count(*) FROM audit_log").Scan(&before).Error; err != nil {
		t.Fatalf("contar auditoria antes: %v", err)
	}

	if _, err := e.svc.ListByOwner(context.Background(), documents.ListInput{
		Actor: principal(uid, readPermission), RequiredPermission: readPermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-34",
	}); err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}

	var after int64
	if err := e.app.Raw("SELECT count(*) FROM audit_log").Scan(&after).Error; err != nil {
		t.Fatalf("contar auditoria depois: %v", err)
	}
	if after != before {
		t.Errorf("ListByOwner não deveria gravar auditoria: antes=%d depois=%d", before, after)
	}
}

// Nenhum documento de outro dono aparece; um dono sem documentos devolve uma
// lista vazia, não um erro.
func TestListByOwnerReturnsAnEmptyListForAnOwnerWithNoDocuments(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc04-vazio@exemplo.com")

	docs, err := e.svc.ListByOwner(context.Background(), documents.ListInput{
		Actor: principal(uid, readPermission), RequiredPermission: readPermission,
		OwnerType: "financeiro.lancamento", OwnerID: "sem-documentos",
	})
	if err != nil {
		t.Fatalf("ListByOwner: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("esperava lista vazia, veio %d", len(docs))
	}
}

// DOC-04 AC3: o limite de 100 caracteres é exato — 100 é aceito, 101 (no
// mesmo formato válido) é recusado.
func TestListByOwnerEnforcesTheOwnerTypeLengthLimitExactly(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc04-limite@exemplo.com")

	exactly100 := strings.Repeat("x", 49) + "." + strings.Repeat("x", 50)
	if len(exactly100) != 100 {
		t.Fatalf("fixture errada: len = %d, esperado 100", len(exactly100))
	}
	if _, err := e.svc.ListByOwner(context.Background(), documents.ListInput{
		Actor: principal(uid, readPermission), RequiredPermission: readPermission,
		OwnerType: exactly100, OwnerID: "x",
	}); err != nil {
		t.Errorf("owner_type com exatamente 100 caracteres deveria ser aceito: %v", err)
	}

	exactly101 := strings.Repeat("x", 49) + "." + strings.Repeat("x", 51)
	if len(exactly101) != 101 {
		t.Fatalf("fixture errada: len = %d, esperado 101", len(exactly101))
	}
	_, err := e.svc.ListByOwner(context.Background(), documents.ListInput{
		Actor: principal(uid, readPermission), RequiredPermission: readPermission,
		OwnerType: exactly101, OwnerID: "x",
	})
	if !errors.Is(err, documents.ErrOwnerTypeInvalid) {
		t.Errorf("owner_type com 101 caracteres: err = %v, esperado ErrOwnerTypeInvalid", err)
	}
}

// owner_type no formato técnico é aceito mesmo sem estar em nenhum catálogo
// fixo de módulos (DOC-04 AC3: "nenhuma lista fixa de valores aceitos").
func TestListByOwnerAcceptsAnyTechnicallyValidOwnerTypeWithNoFixedCatalog(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc04-catalogo@exemplo.com")

	_, err := e.svc.ListByOwner(context.Background(), documents.ListInput{
		Actor: principal(uid, readPermission), RequiredPermission: readPermission,
		OwnerType: "modulo_futuro.recurso_novo", OwnerID: "x",
	})
	if err != nil {
		t.Errorf("um owner_type tecnicamente válido de um módulo ainda inexistente deveria ser aceito: %v", err)
	}
}
