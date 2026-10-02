//go:build integration

package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
)

// fakeDocumentStore is a test double for documents.Service, exercising the
// same AD-015 narrow port (app.DocumentStore) AnexarComprovante depends on —
// no real upload, no testcontainers for S3, nothing of platform/documents'
// own behavior is reimplemented here. Its Store mimics the one behavior
// AnexarComprovante actually relies on: platform/documents' own permission
// check (CMP-01 AC3, DOC-01 AC9) — everything else about Store is out of
// scope for this double.
type fakeDocumentStore struct {
	called bool
	input  documents.StoreInput
	result documents.Document
}

func (f *fakeDocumentStore) Store(ctx context.Context, in documents.StoreInput) (documents.Document, error) {
	f.called = true
	f.input = in
	if err := authz.Require(in.Actor, in.RequiredPermission); err != nil {
		return documents.Document{}, err
	}
	return f.result, nil
}

func (e env) anexarComprovante(store *fakeDocumentStore) *app.AnexarComprovante {
	return &app.AnexarComprovante{Lancamentos: e.lancamentos, Documents: store}
}

// CMP-01 AC1: anexa em lançamento existente, propagando o resultado do
// documents.Service.
func TestAnexarComprovantePersistsThroughDocumentsServiceAndPropagatesResult(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	store := &fakeDocumentStore{result: documents.Document{OriginalFilename: "nota.pdf"}}

	doc, err := e.anexarComprovante(store).Execute(context.Background(), app.AnexarComprovanteInput{
		Actor: e.userActor(t, app.PermComprovanteCreate), LancamentoID: lancamento.ID, Filename: "nota.pdf",
		Content: strings.NewReader("conteudo"),
	})
	if err != nil {
		t.Fatalf("anexar: %v", err)
	}
	if !store.called {
		t.Error("Store deveria ter sido chamado")
	}
	if doc.OriginalFilename != "nota.pdf" {
		t.Errorf("resultado não propagado: %+v", doc)
	}
}

// CMP-01 AC1: OwnerType, OwnerID e RequiredPermission corretos.
func TestAnexarComprovantePassesOwnerTypeOwnerIDAndRequiredPermission(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	store := &fakeDocumentStore{}

	if _, err := e.anexarComprovante(store).Execute(context.Background(), app.AnexarComprovanteInput{
		Actor: e.userActor(t, app.PermComprovanteCreate), LancamentoID: lancamento.ID, Filename: "nota.pdf",
		Content: strings.NewReader("conteudo"),
	}); err != nil {
		t.Fatalf("anexar: %v", err)
	}

	if store.input.OwnerType != "financeiro.lancamento" {
		t.Errorf("OwnerType = %q, esperado %q", store.input.OwnerType, "financeiro.lancamento")
	}
	if store.input.OwnerID != lancamento.ID {
		t.Errorf("OwnerID = %q, esperado %q", store.input.OwnerID, lancamento.ID)
	}
	if store.input.RequiredPermission != app.PermComprovanteCreate {
		t.Errorf("RequiredPermission = %q, esperado %q", store.input.RequiredPermission, app.PermComprovanteCreate)
	}
}

// FIN-D-016: Supersedes nunca é preenchido.
func TestAnexarComprovanteNeverSetsSupersedes(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	store := &fakeDocumentStore{}

	if _, err := e.anexarComprovante(store).Execute(context.Background(), app.AnexarComprovanteInput{
		Actor: e.userActor(t, app.PermComprovanteCreate), LancamentoID: lancamento.ID, Filename: "nota.pdf",
		Content: strings.NewReader("conteudo"),
	}); err != nil {
		t.Fatalf("anexar: %v", err)
	}

	if store.input.Supersedes != nil {
		t.Errorf("Supersedes = %v, esperado nil", store.input.Supersedes)
	}
}

// CMP-01 AC3, DOC-01 AC9: sem a permissão, Store é chamado (com a permissão
// correta) e ele mesmo recusa — AnexarComprovante não duplica a checagem.
func TestAnexarComprovanteDelegatesThePermissionCheckToStore(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	store := &fakeDocumentStore{}

	_, err := e.anexarComprovante(store).Execute(context.Background(), app.AnexarComprovanteInput{
		Actor: e.userActor(t), LancamentoID: lancamento.ID, Filename: "nota.pdf",
		Content: strings.NewReader("conteudo"),
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	if !store.called {
		t.Error("Store deveria ter sido chamado mesmo sem a permissão — quem recusa é o próprio documents.Service")
	}
	if store.input.RequiredPermission != app.PermComprovanteCreate {
		t.Errorf("RequiredPermission repassado = %q, esperado %q", store.input.RequiredPermission, app.PermComprovanteCreate)
	}
}

// CMP-01 AC2: lançamento inexistente é recusado antes de chamar Store.
func TestAnexarComprovanteRejectsANonexistentLancamentoWithoutCallingStore(t *testing.T) {
	e := newEnv(t)
	bogus := "00000000-0000-0000-0000-000000000000"
	store := &fakeDocumentStore{}

	_, err := e.anexarComprovante(store).Execute(context.Background(), app.AnexarComprovanteInput{
		Actor: e.userActor(t, app.PermComprovanteCreate), LancamentoID: bogus, Filename: "nota.pdf",
		Content: strings.NewReader("conteudo"),
	})
	if !errors.Is(err, domain.ErrLancamentoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrLancamentoNaoEncontrado", err)
	}
	if store.called {
		t.Error("Store não deveria ter sido chamado para um lançamento inexistente")
	}
}

// CMP-01 AC4, Edge Case: lançamento CANCELADA aceita anexo sem restrição.
func TestAnexarComprovanteAcceptsACancelledLancamento(t *testing.T) {
	e := newEnv(t)
	conta := e.criarContaDeTeste(t, domain.TipoDespesa, "Material")
	lancamento := e.criarLancamentoDeTeste(t, conta.ID)
	if _, err := e.cancelarLancamento().Execute(context.Background(), app.CancelarLancamentoInput{
		Actor: e.userActor(t, app.PermLancamentoCancel), ID: lancamento.ID, Motivo: "erro",
	}); err != nil {
		t.Fatalf("cancelar: %v", err)
	}
	store := &fakeDocumentStore{result: documents.Document{OriginalFilename: "nota.pdf"}}

	if _, err := e.anexarComprovante(store).Execute(context.Background(), app.AnexarComprovanteInput{
		Actor: e.userActor(t, app.PermComprovanteCreate), LancamentoID: lancamento.ID, Filename: "nota.pdf",
		Content: strings.NewReader("conteudo"),
	}); err != nil {
		t.Errorf("anexar em lançamento CANCELADA deveria ser aceito: %v", err)
	}
	if !store.called {
		t.Error("Store deveria ter sido chamado")
	}
}
