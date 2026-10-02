package app

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
)

// PermComprovanteCreate is the permission AnexarComprovante passes as
// documents.StoreInput.RequiredPermission (CMP-01 AC1). AnexarComprovante
// never checks it itself — platform/documents.Service.Store does, and
// already guarantees refusing without uploading anything (CMP-01 AC3,
// DOC-01 AC9).
const PermComprovanteCreate authz.Permission = "financeiro:comprovante:create"

// OwnerTypeLancamento is the owner_type financeiro uses for every document
// attached to a lançamento (CMP-01 AC1).
const OwnerTypeLancamento = "financeiro.lancamento"

// DocumentStore is the narrow shape AnexarComprovante needs of
// *documents.Service (AD-015).
type DocumentStore interface {
	Store(ctx context.Context, in documents.StoreInput) (documents.Document, error)
}

// AnexarComprovante implements CMP-01: attaches a document to a lançamento,
// delegating storage entirely to platform/documents. It never reimplements
// storage, versioning, file validation or URL generation.
type AnexarComprovante struct {
	Lancamentos LancamentoReader
	Documents   DocumentStore
}

// AnexarComprovanteInput carries what Execute needs. UploadedBy is derived
// from Actor.UserID — financeiro ids are string, documents.StoreInput wants
// a uuid.UUID.
type AnexarComprovanteInput struct {
	Actor        authz.Principal
	LancamentoID string
	Filename     string
	Content      io.Reader
}

// Execute confirms the lançamento exists (CMP-01 AC2) — before calling
// Store, never after — then delegates to documents.Service.Store with
// Supersedes always nil (FIN-D-016).
func (a *AnexarComprovante) Execute(ctx context.Context, in AnexarComprovanteInput) (documents.Document, error) {
	if a.Lancamentos == nil || a.Documents == nil {
		return documents.Document{}, errNotConfigured
	}
	if _, err := a.Lancamentos.Buscar(ctx, in.LancamentoID); err != nil {
		return documents.Document{}, err
	}

	uploadedBy, err := uuid.Parse(in.Actor.UserID)
	if err != nil {
		return documents.Document{}, err
	}

	return a.Documents.Store(ctx, documents.StoreInput{
		Actor: in.Actor, RequiredPermission: PermComprovanteCreate,
		OwnerType: OwnerTypeLancamento, OwnerID: in.LancamentoID,
		Filename: in.Filename, Content: in.Content, UploadedBy: uploadedBy,
	})
}
