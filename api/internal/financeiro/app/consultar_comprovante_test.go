//go:build integration

package app_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/storage"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// --- unit tests, with a double of documents.Service (AD-015) ---

// fakeDocumentAccessor mimics the one behavior ConsultarComprovante actually
// relies on: platform/documents' own permission check (CMP-02 AC2,
// DOC-02 AC2). Everything else about AccessURL (signature, expiry, storage)
// is out of scope for this double.
type fakeDocumentAccessor struct {
	called bool
	input  documents.AccessInput
	url    string
	err    error
}

func (f *fakeDocumentAccessor) AccessURL(ctx context.Context, in documents.AccessInput) (string, error) {
	f.called = true
	f.input = in
	if f.err != nil {
		return "", f.err
	}
	if err := authz.Require(in.Actor, in.RequiredPermission); err != nil {
		return "", err
	}
	return f.url, nil
}

// CMP-02 AC1: propaga a URL devolvida por AccessURL.
func TestConsultarComprovantePropagatesTheURL(t *testing.T) {
	store := &fakeDocumentAccessor{url: "https://example.test/signed"}
	id := uuid.New()

	url, err := (&app.ConsultarComprovante{Documents: store}).Execute(context.Background(), app.ConsultarComprovanteInput{
		Actor: actor(app.PermComprovanteRead), DocumentID: id.String(),
	})
	if err != nil {
		t.Fatalf("consultar: %v", err)
	}
	if url != "https://example.test/signed" {
		t.Errorf("url = %q, esperado %q", url, "https://example.test/signed")
	}
}

// CMP-02 AC1: RequiredPermission e ID corretos.
func TestConsultarComprovantePassesRequiredPermissionAndConvertsDocumentID(t *testing.T) {
	store := &fakeDocumentAccessor{}
	id := uuid.New()

	if _, err := (&app.ConsultarComprovante{Documents: store}).Execute(context.Background(), app.ConsultarComprovanteInput{
		Actor: actor(app.PermComprovanteRead), DocumentID: id.String(),
	}); err != nil {
		t.Fatalf("consultar: %v", err)
	}

	if store.input.RequiredPermission != app.PermComprovanteRead {
		t.Errorf("RequiredPermission = %q, esperado %q", store.input.RequiredPermission, app.PermComprovanteRead)
	}
	if store.input.ID != id {
		t.Errorf("ID = %v, esperado %v", store.input.ID, id)
	}
}

// Propaga erros de AccessURL (ex.: documento inexistente) sem mascarar.
func TestConsultarComprovantePropagatesAccessURLErrors(t *testing.T) {
	sentinel := errors.New("documento inexistente")
	store := &fakeDocumentAccessor{err: sentinel}

	_, err := (&app.ConsultarComprovante{Documents: store}).Execute(context.Background(), app.ConsultarComprovanteInput{
		Actor: actor(app.PermComprovanteRead), DocumentID: uuid.New().String(),
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, esperado %v", err, sentinel)
	}
}

// CMP-02 AC2, DOC-02 AC2: sem a permissão, AccessURL é chamado (com a
// permissão correta) e ele mesmo recusa — ConsultarComprovante não duplica
// a checagem (mesma decisão já adotada em AnexarComprovante, T1).
func TestConsultarComprovanteDelegatesThePermissionCheckToAccessURL(t *testing.T) {
	store := &fakeDocumentAccessor{url: "https://example.test/signed"}

	_, err := (&app.ConsultarComprovante{Documents: store}).Execute(context.Background(), app.ConsultarComprovanteInput{
		Actor: actor(), DocumentID: uuid.New().String(),
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	if !store.called {
		t.Error("AccessURL deveria ter sido chamado mesmo sem a permissão — quem recusa é o próprio documents.Service")
	}
	if store.input.RequiredPermission != app.PermComprovanteRead {
		t.Errorf("RequiredPermission repassado = %q, esperado %q", store.input.RequiredPermission, app.PermComprovanteRead)
	}
}

// --- single real-infrastructure test: proves the financeiro composition,
// not platform/documents' own already-validated behavior (signature
// validity, expiration, PresignGet, storage, versioning, Supersedes, file
// validation remain platform/documents' responsibility and evidence) ---

func newComprovanteEnv(t *testing.T) (*gorm.DB, *documents.Service) {
	t.Helper()
	db := testutil.NewTestDB(t)
	s3env := testutil.SharedS3(t)
	s3Storage, err := storage.NewS3(context.Background(), storage.Config{
		Endpoint: s3env.Endpoint, Region: s3env.Region, Bucket: s3env.Bucket,
		AccessKey: s3env.AccessKey, SecretKey: s3env.SecretKey, UsePathStyle: true,
	})
	if err != nil {
		t.Fatalf("storage.NewS3: %v", err)
	}
	rec := audit.NewRecorder(db, logx.New("error", io.Discard))
	svc := &documents.Service{DB: db, Storage: s3Storage, Authz: simpleAuthz{}, Audit: rec, URLTTL: 5 * time.Minute}
	return db, svc
}

func newUserForComprovante(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	var id string
	err := db.Raw(`INSERT INTO users (email, name, password_hash) VALUES (?, 'Fulano', 'hash') RETURNING id::text`,
		uuid.New().String()+"@exemplo.com").Scan(&id).Error
	if err != nil || id == "" {
		t.Fatalf("criar usuário: id=%q err=%v", id, err)
	}
	u, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("uuid do usuário: %v", err)
	}
	return u
}

func downloadReal(t *testing.T, url string) []byte {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s: status = %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ler corpo: %v", err)
	}
	return body
}

// ConsultarComprovante -> documents.Service.AccessURL -> URL assinada ->
// download HTTP real -> mesmo conteúdo que acabou de ser armazenado.
func TestConsultarComprovanteRoundTripsToTheStoredContent(t *testing.T) {
	db, svc := newComprovanteEnv(t)
	uid := newUserForComprovante(t, db)
	content := []byte("%PDF-1.4\n conteudo de teste\n")

	doc, err := svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid.String(), app.PermComprovanteCreate), RequiredPermission: app.PermComprovanteCreate,
		OwnerType: app.OwnerTypeLancamento, OwnerID: "l-roundtrip", Filename: "comprovante.pdf",
		Content: bytes.NewReader(content), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store (fixture): %v", err)
	}

	url, err := (&app.ConsultarComprovante{Documents: svc}).Execute(context.Background(), app.ConsultarComprovanteInput{
		Actor: principal(uid.String(), app.PermComprovanteRead), DocumentID: doc.ID.String(),
	})
	if err != nil {
		t.Fatalf("ConsultarComprovante: %v", err)
	}

	got := downloadReal(t, url)
	if !bytes.Equal(got, content) {
		t.Errorf("conteúdo baixado difere do armazenado: got %q, want %q", got, content)
	}
}
