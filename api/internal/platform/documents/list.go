package documents

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// maxOwnerTypeLen bounds owner_type before it ever reaches a query (DOC-04
// AC3).
const maxOwnerTypeLen = 100

// ownerTypeFormat is the only shape platform/documents knows about
// owner_type: "module.resource". There is no fixed catalog of modules —
// this package cannot know the modules that do not exist yet.
var ownerTypeFormat = regexp.MustCompile(`^[a-z0-9_]+\.[a-z0-9_]+$`)

// ErrOwnerTypeInvalid is document_owner_type_invalid.
var ErrOwnerTypeInvalid = errors.New("document_owner_type_invalid")

func validOwnerType(s string) bool {
	return len(s) > 0 && len(s) <= maxOwnerTypeLen && ownerTypeFormat.MatchString(s)
}

// ListInput carries what ListByOwner needs.
type ListInput struct {
	Actor              authz.Principal
	RequiredPermission authz.Permission
	OwnerType          string
	OwnerID            string
}

type listRow struct {
	ID               string
	OwnerType        string
	OwnerID          string
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	Sha256           string
	StorageKey       string
	Version          int
	SupersedesID     *string
	Status           string
	UploadedBy       string
	UploadedAt       time.Time
}

func (r listRow) document() (Document, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return Document{}, fmt.Errorf("documentos: id inválido no banco: %w", err)
	}
	uploadedBy, err := uuid.Parse(r.UploadedBy)
	if err != nil {
		return Document{}, fmt.Errorf("documentos: uploaded_by inválido no banco: %w", err)
	}
	var supersedes *uuid.UUID
	if r.SupersedesID != nil {
		sid, err := uuid.Parse(*r.SupersedesID)
		if err != nil {
			return Document{}, fmt.Errorf("documentos: supersedes_id inválido no banco: %w", err)
		}
		supersedes = &sid
	}
	return Document{
		ID: id, OwnerType: r.OwnerType, OwnerID: r.OwnerID, OriginalFilename: r.OriginalFilename,
		ContentType: r.ContentType, SizeBytes: r.SizeBytes, SHA256: r.Sha256, StorageKey: r.StorageKey,
		Version: r.Version, SupersedesID: supersedes, Status: r.Status, UploadedBy: uploadedBy, UploadedAt: r.UploadedAt,
	}, nil
}

// ListByOwner returns every version belonging to OwnerType/OwnerID, newest
// first, using documents_owner_idx (T5). It never generates a signed URL and
// never records an audit entry (DOC-04 AC4): only AccessURL discloses a
// document's content, and only that is audited.
func (s *Service) ListByOwner(ctx context.Context, in ListInput) ([]Document, error) {
	if !s.configured() {
		return nil, errNotConfigured
	}
	if err := s.Authz.Require(ctx, in.Actor, in.RequiredPermission); err != nil {
		return nil, err
	}
	if !validOwnerType(in.OwnerType) {
		return nil, ErrOwnerTypeInvalid
	}

	var rows []listRow
	err := conn(ctx, s.DB).Raw(`SELECT id::text AS id, owner_type, owner_id, original_filename, content_type, size_bytes,
		sha256, storage_key, version, supersedes_id::text AS supersedes_id, status, uploaded_by::text AS uploaded_by, uploaded_at
		FROM documents WHERE owner_type = ? AND owner_id = ? ORDER BY uploaded_at DESC`, in.OwnerType, in.OwnerID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("documentos: listar por dono: %w", err)
	}

	docs := make([]Document, 0, len(rows))
	for _, r := range rows {
		d, err := r.document()
		if err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}
