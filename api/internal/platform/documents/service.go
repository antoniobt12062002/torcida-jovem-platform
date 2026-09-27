// Package documents guarda, versiona e dá acesso a documentos: a fundação de
// documentos (ADR-006, protocolo S3, sem provedor definido). Não possui
// nenhuma permissão de negócio própria — cada chamada recebe do módulo
// consumidor o ator e a permissão exigida, e só repassa para authz.Require
// (AD-015): quem decide o nome e o dono da permissão é sempre quem chama.
package documents

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/storage"
)

// maxSize is the 10 MiB upload limit (DOC-01 AC3).
const maxSize = 10 * 1024 * 1024

// allowedByExt maps each accepted extension (lowercase) to the content type
// its bytes must sniff to. There is no entry for an extension this feature
// does not accept.
var allowedByExt = map[string]string{
	".pdf":  "application/pdf",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

func allowedContentType(ct string) bool {
	switch ct {
	case "application/pdf", "image/jpeg", "image/png", "image/webp":
		return true
	}
	return false
}

// Errors of the documents service; their messages are the API error codes
// (see design.md's Error Handling Strategy). A forbidden or not-found
// document uses the shared authz.ErrForbidden and ErrNotFound below instead
// of a named code, on purpose: neither reveals more than "access denied" or
// "not found".
var (
	ErrExtensionNotAllowed = errors.New("document_extension_not_allowed")
	ErrTypeNotAllowed      = errors.New("document_type_not_allowed")
	ErrTypeMismatch        = errors.New("document_type_mismatch")
	ErrTooLarge            = errors.New("document_too_large")
	// ErrNotFound is returned when the document id (or, in Store, the id it
	// is meant to supersede) does not exist.
	ErrNotFound = errors.New("document_not_found")

	errNotConfigured = errors.New("documentos: serviço mal configurado")
)

// Audit actions this package owns and registers for itself (every module
// registers its own, per platform/audit's convention).
const (
	ActionCreate audit.Action = "document.create"
	ActionAccess audit.Action = "document.access"
)

func init() { audit.Register(ActionCreate, ActionAccess) }

// Document is one stored version. SupersedesID is nil for a document's first
// version. Status is always ACTIVE in this feature (see Data Models): the
// column exists for a future state, but nothing here transitions it.
type Document struct {
	ID               uuid.UUID
	OwnerType        string
	OwnerID          string
	OriginalFilename string
	ContentType      string
	SizeBytes        int64
	SHA256           string
	StorageKey       string
	Version          int
	SupersedesID     *uuid.UUID
	Status           string
	UploadedBy       uuid.UUID
	UploadedAt       time.Time
}

// Authorizer is the narrow shape documents needs of authz.Authorizer (the
// same pattern as platform/audit's query.go and identity/app's use cases):
// documents never imports the concrete type, and never declares a permission
// of its own.
type Authorizer interface {
	Require(ctx context.Context, p authz.Principal, perm authz.Permission) error
}

// Auditor is the narrow shape documents needs of *audit.Recorder.
type Auditor interface {
	Record(ctx context.Context, e audit.Entry) error
}

// StoreInput carries what Store needs. Actor and RequiredPermission are
// decided entirely by the caller (AD-015): this package never knows their
// name or shape. Supersedes is the id of the version this upload replaces,
// or nil for a document's first version — the service never resolves "the
// current version" on the caller's behalf (see design.md's Tech Decisions).
type StoreInput struct {
	Actor              authz.Principal
	RequiredPermission authz.Permission
	OwnerType          string
	OwnerID            string
	Filename           string
	Content            io.Reader
	UploadedBy         uuid.UUID
	Supersedes         *uuid.UUID
}

// AccessInput carries what AccessURL needs. ID always names one exact
// version; there is no "latest" to ask for.
type AccessInput struct {
	Actor              authz.Principal
	RequiredPermission authz.Permission
	ID                 uuid.UUID
}

// Service stores, versions and gives access to documents.
type Service struct {
	DB      *gorm.DB
	Storage storage.Storage
	Authz   Authorizer
	Audit   Auditor
	// URLTTL is the validity of a presigned URL from AccessURL.
	URLTTL time.Duration
}

func (s *Service) configured() bool {
	return s.DB != nil && s.Storage != nil && s.Authz != nil && s.Audit != nil
}

// conn returns the transaction carried by ctx (database.WithTx), or the pool
// when there is none.
func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := database.TxFrom(ctx); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
}

// readWithLimit reads r fully, rejecting anything past max bytes before it is
// ever written to storage or the database (DOC-01 AC3).
func readWithLimit(r io.Reader, max int64) ([]byte, error) {
	buf, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		return nil, fmt.Errorf("documentos: ler conteúdo: %w", err)
	}
	if int64(len(buf)) > max {
		return nil, ErrTooLarge
	}
	return buf, nil
}

// Store validates, uploads and records one document version. Order: the
// permission the caller supplied, the filename's extension (before any byte
// is read), the size limit, the content type detected from the bytes, the
// upload to storage, then one transaction with the metadata row and the
// audit entry. If the transaction fails after the upload succeeded, only the
// object just created is removed (DOC-01 AC7) — never anything referenced
// before.
func (s *Service) Store(ctx context.Context, in StoreInput) (Document, error) {
	if !s.configured() {
		return Document{}, errNotConfigured
	}
	if err := s.Authz.Require(ctx, in.Actor, in.RequiredPermission); err != nil {
		return Document{}, err
	}

	ext := strings.ToLower(filepath.Ext(in.Filename))
	expectedContentType, ok := allowedByExt[ext]
	if !ok {
		return Document{}, ErrExtensionNotAllowed
	}

	buf, err := readWithLimit(in.Content, maxSize)
	if err != nil {
		return Document{}, err
	}

	detected := http.DetectContentType(buf)
	if !allowedContentType(detected) {
		return Document{}, ErrTypeNotAllowed
	}
	if detected != expectedContentType {
		return Document{}, ErrTypeMismatch
	}

	sum := sha256.Sum256(buf)
	doc := Document{
		ID: uuid.New(), OwnerType: in.OwnerType, OwnerID: in.OwnerID, OriginalFilename: in.Filename,
		ContentType: detected, SizeBytes: int64(len(buf)), SHA256: hex.EncodeToString(sum[:]),
		Version: 1, SupersedesID: in.Supersedes, Status: "ACTIVE", UploadedBy: in.UploadedBy,
	}
	doc.StorageKey = "documents/" + doc.ID.String()

	if err := s.Storage.Put(ctx, doc.StorageKey, bytes.NewReader(buf), int64(len(buf)), detected); err != nil {
		return Document{}, err
	}

	err = database.WithTx(ctx, s.DB, func(ctx context.Context) error {
		if in.Supersedes != nil {
			var prevVersion int
			row := conn(ctx, s.DB).Raw("SELECT version FROM documents WHERE id = ?", in.Supersedes.String()).Row()
			if scanErr := row.Scan(&prevVersion); scanErr != nil {
				if errors.Is(scanErr, sql.ErrNoRows) {
					return ErrNotFound
				}
				return fmt.Errorf("documentos: consultar versão anterior: %w", scanErr)
			}
			doc.Version = prevVersion + 1
		}

		row := conn(ctx, s.DB).Raw(`INSERT INTO documents
			(id, owner_type, owner_id, original_filename, content_type, size_bytes, sha256, storage_key, version, supersedes_id, status, uploaded_by)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			RETURNING uploaded_at`,
			doc.ID.String(), doc.OwnerType, doc.OwnerID, doc.OriginalFilename, doc.ContentType, doc.SizeBytes,
			doc.SHA256, doc.StorageKey, doc.Version, nullableUUID(doc.SupersedesID), doc.Status, doc.UploadedBy.String(),
		).Row()
		if scanErr := row.Scan(&doc.UploadedAt); scanErr != nil {
			return fmt.Errorf("documentos: gravar metadados: %w", scanErr)
		}

		return s.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionCreate,
			EntityType: "document", EntityID: doc.ID.String(), Outcome: audit.OutcomeSuccess,
			After: map[string]any{
				"owner_type": doc.OwnerType, "owner_id": doc.OwnerID, "version": doc.Version, "storage_key": doc.StorageKey,
			},
		})
	})
	if err != nil {
		if delErr := s.Storage.DeleteCreated(ctx, doc.StorageKey); delErr != nil {
			return Document{}, fmt.Errorf("%w (e falhou ao desfazer o objeto órfão: %v)", err, delErr)
		}
		return Document{}, err
	}
	return doc, nil
}

// AccessURL returns a presigned GET URL for the exact document id given. It
// never resolves "the current version" (DOC-02 AC6): the caller always names
// one specific version.
func (s *Service) AccessURL(ctx context.Context, in AccessInput) (string, error) {
	if !s.configured() {
		return "", errNotConfigured
	}
	if err := s.Authz.Require(ctx, in.Actor, in.RequiredPermission); err != nil {
		return "", err
	}

	var key string
	err := database.WithTx(ctx, s.DB, func(ctx context.Context) error {
		row := conn(ctx, s.DB).Raw("SELECT storage_key FROM documents WHERE id = ?", in.ID.String()).Row()
		if scanErr := row.Scan(&key); scanErr != nil {
			if errors.Is(scanErr, sql.ErrNoRows) {
				return ErrNotFound
			}
			return fmt.Errorf("documentos: consultar documento: %w", scanErr)
		}
		return s.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: ActionAccess,
			EntityType: "document", EntityID: in.ID.String(), Outcome: audit.OutcomeSuccess,
		})
	})
	if err != nil {
		return "", err
	}

	return s.Storage.PresignGet(ctx, key, s.URLTTL)
}
