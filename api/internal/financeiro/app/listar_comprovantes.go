package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
)

// DocumentLister is the narrow shape ListarComprovantes needs of
// *documents.Service (AD-015).
type DocumentLister interface {
	ListByOwner(ctx context.Context, in documents.ListInput) ([]documents.Document, error)
}

// ListarComprovantes implements CMP-03: lists every comprovante attached to
// a lançamento, delegating entirely to platform/documents. It has no
// Authz (permission is delegated, CMP-03 AC1, same decision as
// AnexarComprovante/ConsultarComprovante), no Auditor and no
// LancamentoRepository — it cannot record an audit entry or write to
// lancamentos even by mistake (CMP-03 AC3), and performs no existence
// check of its own (a LancamentoID with no documents simply yields the
// empty list ListByOwner already returns).
type ListarComprovantes struct {
	Documents DocumentLister
}

type ListarComprovantesInput struct {
	Actor        authz.Principal
	LancamentoID string
}

// Execute delegates to documents.Service.ListByOwner and returns its result
// exactly as received — no reordering, no filtering, no URL generation.
func (l *ListarComprovantes) Execute(ctx context.Context, in ListarComprovantesInput) ([]documents.Document, error) {
	if l.Documents == nil {
		return nil, errNotConfigured
	}
	return l.Documents.ListByOwner(ctx, documents.ListInput{
		Actor: in.Actor, RequiredPermission: PermComprovanteRead,
		OwnerType: OwnerTypeLancamento, OwnerID: in.LancamentoID,
	})
}
