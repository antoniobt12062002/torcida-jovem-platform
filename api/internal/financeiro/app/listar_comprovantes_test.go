//go:build integration

package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
)

// --- unit tests, with a double of documents.Service (AD-015) ---

// fakeDocumentLister mimics the one behavior ListarComprovantes actually
// relies on: platform/documents' own permission check (CMP-03 AC1,
// DOC-04 AC2-like). Everything else about ListByOwner (ordering guarantee,
// storage, no URL/audit) is platform/documents' own responsibility and
// evidence — not reimplemented here.
type fakeDocumentLister struct {
	called bool
	input  documents.ListInput
	result []documents.Document
	err    error
}

func (f *fakeDocumentLister) ListByOwner(ctx context.Context, in documents.ListInput) ([]documents.Document, error) {
	f.called = true
	f.input = in
	if f.err != nil {
		return nil, f.err
	}
	if err := authz.Require(in.Actor, in.RequiredPermission); err != nil {
		return nil, err
	}
	return f.result, nil
}

// CMP-03 AC1: propaga a lista devolvida por ListByOwner, preservando a
// ordem — ListarComprovantes nunca reordena.
func TestListarComprovantesPropagatesTheListPreservingOrder(t *testing.T) {
	newer := documents.Document{ID: uuid.New(), OriginalFilename: "mais-novo.pdf"}
	older := documents.Document{ID: uuid.New(), OriginalFilename: "mais-antigo.pdf"}
	store := &fakeDocumentLister{result: []documents.Document{newer, older}}

	docs, err := (&app.ListarComprovantes{Documents: store}).Execute(context.Background(), app.ListarComprovantesInput{
		Actor: actor(app.PermComprovanteRead), LancamentoID: "l-1",
	})
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(docs) != 2 || docs[0].ID != newer.ID || docs[1].ID != older.ID {
		t.Errorf("ordem não preservada: %+v", docs)
	}
}

// CMP-03 AC1: OwnerType, OwnerID e RequiredPermission corretos.
func TestListarComprovantesPassesOwnerTypeOwnerIDAndRequiredPermission(t *testing.T) {
	store := &fakeDocumentLister{}

	if _, err := (&app.ListarComprovantes{Documents: store}).Execute(context.Background(), app.ListarComprovantesInput{
		Actor: actor(app.PermComprovanteRead), LancamentoID: "l-42",
	}); err != nil {
		t.Fatalf("listar: %v", err)
	}

	if store.input.OwnerType != "financeiro.lancamento" {
		t.Errorf("OwnerType = %q, esperado %q", store.input.OwnerType, "financeiro.lancamento")
	}
	if store.input.OwnerID != "l-42" {
		t.Errorf("OwnerID = %q, esperado %q", store.input.OwnerID, "l-42")
	}
	if store.input.RequiredPermission != app.PermComprovanteRead {
		t.Errorf("RequiredPermission = %q, esperado %q", store.input.RequiredPermission, app.PermComprovanteRead)
	}
}

// Um LancamentoID sem documentos devolve simplesmente a lista vazia que
// ListByOwner devolveria — sem checagem própria de existência.
func TestListarComprovantesReturnsEmptyListWhenThereAreNone(t *testing.T) {
	store := &fakeDocumentLister{result: []documents.Document{}}

	docs, err := (&app.ListarComprovantes{Documents: store}).Execute(context.Background(), app.ListarComprovantesInput{
		Actor: actor(app.PermComprovanteRead), LancamentoID: "00000000-0000-0000-0000-000000000000",
	})
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("esperava lista vazia, veio %+v", docs)
	}
}

// Propaga erros de ListByOwner sem mascarar.
func TestListarComprovantesPropagatesErrors(t *testing.T) {
	sentinel := errors.New("falha de leitura")
	store := &fakeDocumentLister{err: sentinel}

	_, err := (&app.ListarComprovantes{Documents: store}).Execute(context.Background(), app.ListarComprovantesInput{
		Actor: actor(app.PermComprovanteRead), LancamentoID: "l-1",
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, esperado %v", err, sentinel)
	}
}

// CMP-03 AC1: sem a permissão, ListByOwner é chamado (com a permissão
// correta) e ele mesmo recusa — ListarComprovantes não duplica a checagem
// (mesma decisão já adotada em AnexarComprovante/ConsultarComprovante).
func TestListarComprovantesDelegatesThePermissionCheckToListByOwner(t *testing.T) {
	store := &fakeDocumentLister{result: []documents.Document{{ID: uuid.New()}}}

	docs, err := (&app.ListarComprovantes{Documents: store}).Execute(context.Background(), app.ListarComprovantesInput{
		Actor: actor(), LancamentoID: "l-1",
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	if !store.called {
		t.Error("ListByOwner deveria ter sido chamado mesmo sem a permissão — quem recusa é o próprio documents.Service")
	}
	if store.input.RequiredPermission != app.PermComprovanteRead {
		t.Errorf("RequiredPermission repassado = %q, esperado %q", store.input.RequiredPermission, app.PermComprovanteRead)
	}
	if len(docs) != 0 {
		t.Errorf("esperava nenhum documento, veio %+v", docs)
	}
}

// --- minimal real-Postgres tests (no S3): structural + behavioral proof of
// CMP-03 AC3's negative guarantees. ListarComprovantes has no Auditor and no
// LancamentoRepository field at all — it cannot write to audit_log or to
// lancamentos; these tests confirm that architectural guarantee
// behaviorally, without standing up S3/Garage or replicating
// platform/documents' own tests. ---

// CMP-03 AC3: o status do lançamento é idêntico antes e depois da listagem.
func TestListarComprovantesNeverChangesLancamentoStatus(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	store := &fakeDocumentLister{result: []documents.Document{}}

	if _, err := (&app.ListarComprovantes{Documents: store}).Execute(context.Background(), app.ListarComprovantesInput{
		Actor: actor(app.PermComprovanteRead), LancamentoID: lancamento.ID,
	}); err != nil {
		t.Fatalf("listar: %v", err)
	}

	depois, err := e.lancamentos.Buscar(context.Background(), lancamento.ID)
	if err != nil || depois.Status != domain.StatusCriada {
		t.Errorf("status não deveria ter mudado: %+v, err=%v", depois, err)
	}
}

// CMP-03 AC3: nenhuma linha nova em audit_log por causa da listagem.
func TestListarComprovantesNeverWritesAudit(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	store := &fakeDocumentLister{result: []documents.Document{}}

	var antes int64
	if err := e.db.Raw("SELECT count(*) FROM audit_log").Scan(&antes).Error; err != nil {
		t.Fatalf("contar audit_log antes: %v", err)
	}

	if _, err := (&app.ListarComprovantes{Documents: store}).Execute(context.Background(), app.ListarComprovantesInput{
		Actor: actor(app.PermComprovanteRead), LancamentoID: lancamento.ID,
	}); err != nil {
		t.Fatalf("listar: %v", err)
	}

	var depois int64
	if err := e.db.Raw("SELECT count(*) FROM audit_log").Scan(&depois).Error; err != nil {
		t.Fatalf("contar audit_log depois: %v", err)
	}
	if depois != antes {
		t.Errorf("audit_log mudou: antes=%d depois=%d", antes, depois)
	}
}
