package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
)

// PermComprovanteRead is the permission ConsultarComprovante passes as
// documents.AccessInput.RequiredPermission (CMP-02 AC1). Like
// AnexarComprovante, it never checks it itself — platform/documents.Service
// does, and already guarantees refusing without generating a URL (CMP-02
// AC2, DOC-02 AC2).
const PermComprovanteRead authz.Permission = "financeiro:comprovante:read"

// DocumentAccessor is the narrow shape ConsultarComprovante needs of
// *documents.Service (AD-015).
type DocumentAccessor interface {
	AccessURL(ctx context.Context, in documents.AccessInput) (string, error)
}

// ConsultarComprovante implements CMP-02: returns a presigned URL for an
// already attached comprovante, delegating entirely to platform/documents.
// Unlike AnexarComprovante, it has no lançamento existence check — a
// comprovante is identified by its own id, not through the lançamento it
// was attached to.
type ConsultarComprovante struct {
	Documents DocumentAccessor
}

// ConsultarComprovanteInput carries what Execute needs. DocumentID is
// string — financeiro ids are string, documents.AccessInput wants a
// uuid.UUID.
type ConsultarComprovanteInput struct {
	Actor      authz.Principal
	DocumentID string
}

func (c *ConsultarComprovante) Execute(ctx context.Context, in ConsultarComprovanteInput) (string, error) {
	if c.Documents == nil {
		return "", errNotConfigured
	}
	id, err := uuid.Parse(in.DocumentID)
	if err != nil {
		return "", err
	}
	return c.Documents.AccessURL(ctx, documents.AccessInput{
		Actor: in.Actor, RequiredPermission: PermComprovanteRead, ID: id,
	})
}
