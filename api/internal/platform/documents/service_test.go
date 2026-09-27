//go:build integration

package documents_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/storage"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

const readPermission = "financeiro:comprovante:read"
const writePermission = "financeiro:comprovante:create"

// simpleAuthz forwards to the package-level authz.Require: it decides purely
// by what the Principal holds, with no catalog and no denial hook — the same
// shape documents.Service depends on, proving the service never knows or
// decides the permission itself (AD-015).
type simpleAuthz struct{}

func (simpleAuthz) Require(_ context.Context, p authz.Principal, perm authz.Permission) error {
	return authz.Require(p, perm)
}

// spyStorage wraps a real storage.Storage and counts PresignGet calls, so a
// test can prove no URL was generated (DOC-02 AC2, AC3) without depending on
// implementation internals.
type spyStorage struct {
	storage.Storage
	presignCalls int
}

func (s *spyStorage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	s.presignCalls++
	return s.Storage.PresignGet(ctx, key, ttl)
}

type env struct {
	app *gorm.DB
	s3  *testutil.S3
	st  *spyStorage
	svc *documents.Service
}

func newEnv(t *testing.T) env {
	t.Helper()
	app := testutil.NewTestDB(t)
	s3env := testutil.SharedS3(t)
	s3Storage, err := storage.NewS3(context.Background(), storage.Config{
		Endpoint: s3env.Endpoint, Region: s3env.Region, Bucket: s3env.Bucket,
		AccessKey: s3env.AccessKey, SecretKey: s3env.SecretKey, UsePathStyle: true,
	})
	if err != nil {
		t.Fatalf("storage.NewS3: %v", err)
	}
	st := &spyStorage{Storage: s3Storage}
	rec := audit.NewRecorder(app, logx.New("error", io.Discard))
	svc := &documents.Service{DB: app, Storage: st, Authz: simpleAuthz{}, Audit: rec, URLTTL: 5 * time.Minute}
	return env{app: app, s3: s3env, st: st, svc: svc}
}

func newUser(t *testing.T, db *gorm.DB, email string) uuid.UUID {
	t.Helper()
	var id string
	if err := db.Raw(`INSERT INTO users (email, name, password_hash) VALUES (?, 'Fulano', 'hash') RETURNING id::text`, email).Scan(&id).Error; err != nil {
		t.Fatalf("criar usuário %s: %v", email, err)
	}
	u, err := uuid.Parse(id)
	if err != nil {
		t.Fatalf("uuid do usuário: %v", err)
	}
	return u
}

func principal(uid uuid.UUID, perms ...authz.Permission) authz.Principal {
	m := make(map[authz.Permission]struct{}, len(perms))
	for _, p := range perms {
		m[p] = struct{}{}
	}
	return authz.Principal{UserID: uid.String(), Permissions: m}
}

// Minimal, real fixtures for the four allowed content types, matching Go's
// own net/http sniffing table exactly (verified against net/http/sniff.go).
func pdfBytes() []byte { return []byte("%PDF-1.4\n% conteúdo de teste\n") }
func pngBytes() []byte {
	return append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, []byte("corpo png de teste")...)
}
func jpegBytes() []byte { return append([]byte{0xFF, 0xD8, 0xFF}, []byte("corpo jpeg de teste")...) }
func webpBytes() []byte {
	b := []byte("RIFF")
	b = append(b, 0, 0, 0, 0)
	b = append(b, []byte("WEBPVP")...)
	return append(b, []byte("corpo webp de teste")...)
}
func gifBytes() []byte { return []byte("GIF89a corpo gif, tipo não permitido por esta feição") }

type documentRow struct {
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
}

func readDocument(t *testing.T, db *gorm.DB, id uuid.UUID) documentRow {
	t.Helper()
	var row documentRow
	err := db.Raw(`SELECT owner_type, owner_id, original_filename, content_type, size_bytes, sha256, storage_key,
		version, supersedes_id::text AS supersedes_id, status, uploaded_by::text AS uploaded_by
		FROM documents WHERE id = ?`, id.String()).Scan(&row).Error
	if err != nil {
		t.Fatalf("ler documento %s: %v", id, err)
	}
	return row
}

func objectCount(t *testing.T, env env) int {
	t.Helper()
	client, err := env.s3.Client(context.Background())
	if err != nil {
		t.Fatalf("cliente s3: %v", err)
	}
	out, err := client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(env.s3.Bucket), Prefix: aws.String("documents/"),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2: %v", err)
	}
	return len(out.Contents)
}

func download(t *testing.T, url string) []byte {
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

// panicReader fails the test if Store ever tries to read it — used to prove
// an extension rejection happens before any byte is read (DOC-01 AC10).
type panicReader struct{ t *testing.T }

func (r panicReader) Read([]byte) (int, error) {
	r.t.Fatal("o conteúdo não deveria ser lido: a extensão já deveria ter sido recusada")
	return 0, nil
}

// DOC-01 AC1: grava o objeto e a linha de metadados, com todos os campos.
func TestStoreWritesTheObjectAndTheMetadataRow(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac1@exemplo.com")
	content := pdfBytes()

	doc, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-1", Filename: "comprovante.pdf",
		Content: bytes.NewReader(content), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if doc.StorageKey != "documents/"+doc.ID.String() {
		t.Errorf("storage_key = %q, esperado documents/<uuid>", doc.StorageKey)
	}

	row := readDocument(t, e.app, doc.ID)
	if row.OwnerType != "financeiro.lancamento" || row.OwnerID != "l-1" || row.OriginalFilename != "comprovante.pdf" ||
		row.ContentType != "application/pdf" || row.SizeBytes != int64(len(content)) || row.StorageKey != doc.StorageKey ||
		row.Version != 1 || row.UploadedBy != uid.String() {
		t.Errorf("linha gravada não confere: %+v", row)
	}

	url, err := e.svc.AccessURL(context.Background(), documents.AccessInput{Actor: principal(uid, readPermission), RequiredPermission: readPermission, ID: doc.ID})
	if err != nil {
		t.Fatalf("AccessURL: %v", err)
	}
	if got := download(t, url); !bytes.Equal(got, content) {
		t.Errorf("objeto gravado difere do enviado: got %q", got)
	}
}

// DOC-01 AC2: o SHA-256 é calculado sobre os bytes enviados e gravado em hex.
func TestStoreComputesTheSHA256OverTheStreamedBytes(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac2@exemplo.com")
	content := pngBytes()

	doc, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-2", Filename: "recibo.png",
		Content: bytes.NewReader(content), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	sum := sha256.Sum256(content)
	want := hex.EncodeToString(sum[:])
	if doc.SHA256 != want || len(doc.SHA256) != 64 {
		t.Errorf("sha256 = %q, esperado %q", doc.SHA256, want)
	}
	row := readDocument(t, e.app, doc.ID)
	if row.Sha256 != want {
		t.Errorf("sha256 gravado = %q, esperado %q", row.Sha256, want)
	}
}

// DOC-01 AC3: acima de 10 MiB é recusado antes de gravar qualquer metadado.
func TestStoreRejectsAFileOver10MiB(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac3@exemplo.com")
	before := objectCount(t, e)

	oversized := make([]byte, 10*1024*1024+1)
	copy(oversized, []byte{0xFF, 0xD8, 0xFF})

	_, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-3", Filename: "grande.jpg",
		Content: bytes.NewReader(oversized), UploadedBy: uid,
	})
	if !errors.Is(err, documents.ErrTooLarge) {
		t.Errorf("err = %v, esperado ErrTooLarge", err)
	}
	if after := objectCount(t, e); after != before {
		t.Errorf("nenhum objeto deveria ter sido gravado: antes=%d depois=%d", before, after)
	}
}

// DOC-01 AC4: tipo detectado pelo conteúdo fora da lista é recusado.
func TestStoreRejectsAContentTypeNotInTheAllowList(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac4@exemplo.com")

	_, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-4", Filename: "figura.png",
		Content: bytes.NewReader(gifBytes()), UploadedBy: uid,
	})
	if !errors.Is(err, documents.ErrTypeNotAllowed) {
		t.Errorf("err = %v, esperado ErrTypeNotAllowed", err)
	}
}

// DOC-01 AC5: nova versão referencia a anterior, incrementa a versão e não
// altera o objeto antigo.
func TestStoreVersionsANewUploadOverThePreviousDocument(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac5@exemplo.com")
	v1content := pdfBytes()

	v1, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-5", Filename: "v1.pdf",
		Content: bytes.NewReader(v1content), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store v1: %v", err)
	}

	v2content := append(append([]byte{}, pdfBytes()...), []byte(" v2")...)
	v2, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-5", Filename: "v2.pdf",
		Content: bytes.NewReader(v2content), UploadedBy: uid, Supersedes: &v1.ID,
	})
	if err != nil {
		t.Fatalf("Store v2: %v", err)
	}
	if v2.Version != 2 || v2.SupersedesID == nil || *v2.SupersedesID != v1.ID {
		t.Errorf("v2 = %+v, esperado version=2 supersedes=%s", v2, v1.ID)
	}

	url, err := e.svc.AccessURL(context.Background(), documents.AccessInput{Actor: principal(uid, readPermission), RequiredPermission: readPermission, ID: v1.ID})
	if err != nil {
		t.Fatalf("AccessURL v1: %v", err)
	}
	if got := download(t, url); !bytes.Equal(got, v1content) {
		t.Error("o objeto da v1 não deveria ter mudado depois de gravar a v2")
	}
}

// DOC-01 AC6: nenhuma operação sobrescreve ou apaga um documento gravado.
func TestServiceProvidesNoOverwriteOrDeleteOperation(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac6@exemplo.com")
	doc, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-6", Filename: "doc.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	for name, stmt := range map[string]string{
		"UPDATE": "UPDATE documents SET status = 'INACTIVE' WHERE id = ?",
		"DELETE": "DELETE FROM documents WHERE id = ?",
	} {
		if err := e.app.Exec(stmt, doc.ID.String()).Error; err == nil {
			t.Errorf("%s deveria ser recusado: o serviço não expõe (e o banco não concede) essa operação", name)
		}
	}
}

// DOC-01 AC7: se a gravação dos metadados falhar depois do upload, só o
// objeto recém-criado é removido — provado indiretamente, já que Store não
// devolve a chave num erro: a contagem de objetos no bucket não muda.
func TestStoreRemovesOnlyTheJustCreatedObjectWhenMetadataWriteFails(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac7@exemplo.com")
	before := objectCount(t, e)

	nonexistent := uuid.New()
	_, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-7", Filename: "orfao.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid, Supersedes: &nonexistent,
	})
	if !errors.Is(err, documents.ErrNotFound) {
		t.Fatalf("err = %v, esperado ErrNotFound (supersedes inexistente)", err)
	}
	if after := objectCount(t, e); after != before {
		t.Errorf("o objeto órfão deveria ter sido removido: antes=%d depois=%d", before, after)
	}
}

// DOC-01 AC8: document.create é auditado na mesma operação da gravação.
func TestStoreAuditsDocumentCreate(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac8@exemplo.com")
	doc, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-8", Filename: "doc.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	var action, entityType, entityID, actorID string
	err = e.app.Raw(`SELECT action, entity_type, entity_id, actor_user_id::text FROM audit_log
		WHERE action = 'document.create' AND entity_id = ?`, doc.ID.String()).Row().
		Scan(&action, &entityType, &entityID, &actorID)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if action != "document.create" || entityType != "document" || entityID != doc.ID.String() || actorID != uid.String() {
		t.Errorf("auditoria não confere: action=%q entity_type=%q entity_id=%q actor=%q", action, entityType, entityID, actorID)
	}
}

// DOC-01 AC9: sem a permissão exigida pelo chamador, nada é escrito.
func TestStoreRequiresTheCallerSuppliedPermissionAndWritesNothingWithoutIt(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac9@exemplo.com")
	before := objectCount(t, e)

	_, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-9", Filename: "doc.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	if after := objectCount(t, e); after != before {
		t.Errorf("nada deveria ter sido escrito no storage: antes=%d depois=%d", before, after)
	}
	var n int64
	if err := e.app.Raw("SELECT count(*) FROM documents WHERE owner_id = 'l-9'").Scan(&n).Error; err != nil || n != 0 {
		t.Errorf("nenhuma linha deveria ter sido gravada: n=%d err=%v", n, err)
	}
}

// DOC-01 AC10: extensão fora da lista é recusada antes de ler o conteúdo.
func TestStoreRejectsAnExtensionNotInTheAllowListBeforeReadingContent(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac10@exemplo.com")

	for _, filename := range []string{"malware.exe", "planilha.xlsx", "semextensao"} {
		_, err := e.svc.Store(context.Background(), documents.StoreInput{
			Actor: principal(uid, writePermission), RequiredPermission: writePermission,
			OwnerType: "financeiro.lancamento", OwnerID: "l-10", Filename: filename,
			Content: panicReader{t}, UploadedBy: uid,
		})
		if !errors.Is(err, documents.ErrExtensionNotAllowed) {
			t.Errorf("%s: err = %v, esperado ErrExtensionNotAllowed", filename, err)
		}
	}
	// Maiúsculas também são recusadas ou aceitas ignorando caixa (aqui: aceita).
	_, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-10", Filename: "recibo.PDF",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	})
	if err != nil {
		t.Errorf(".PDF maiúsculo deveria ser aceito (mesma extensão, ignorando caixa): %v", err)
	}
}

// DOC-01 AC11: extensão que não bate com o tipo detectado é recusada.
func TestStoreRejectsWhenTheExtensionDoesNotMatchTheDetectedContentType(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac11@exemplo.com")

	_, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-11", Filename: "disfarcado.pdf",
		Content: bytes.NewReader(pngBytes()), UploadedBy: uid,
	})
	if !errors.Is(err, documents.ErrTypeMismatch) {
		t.Errorf("err = %v, esperado ErrTypeMismatch", err)
	}
}

// DOC-01 AC12: o documento nasce ACTIVE.
func TestStoreSetsStatusToActive(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "ac12@exemplo.com")

	doc, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-12", Filename: "doc.webp",
		Content: bytes.NewReader(webpBytes()), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if doc.Status != "ACTIVE" {
		t.Errorf("status = %q, esperado ACTIVE", doc.Status)
	}
	row := readDocument(t, e.app, doc.ID)
	if row.Status != "ACTIVE" {
		t.Errorf("status gravado = %q, esperado ACTIVE", row.Status)
	}
}

// DOC-02 AC1: URL assinada válida e document.access auditado.
func TestAccessURLReturnsAPresignedURLAndAuditsAccess(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc02ac1@exemplo.com")
	content := jpegBytes()
	doc, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-20", Filename: "foto.jpg",
		Content: bytes.NewReader(content), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	url, err := e.svc.AccessURL(context.Background(), documents.AccessInput{Actor: principal(uid, readPermission), RequiredPermission: readPermission, ID: doc.ID})
	if err != nil {
		t.Fatalf("AccessURL: %v", err)
	}
	if got := download(t, url); !bytes.Equal(got, content) {
		t.Error("o conteúdo baixado pela URL assinada difere do enviado")
	}

	var action, entityID, actorID string
	err = e.app.Raw(`SELECT action, entity_id, actor_user_id::text FROM audit_log
		WHERE action = 'document.access' AND entity_id = ?`, doc.ID.String()).Row().Scan(&action, &entityID, &actorID)
	if err != nil {
		t.Fatalf("ler auditoria de acesso: %v", err)
	}
	if action != "document.access" || entityID != doc.ID.String() || actorID != uid.String() {
		t.Errorf("auditoria de acesso não confere: action=%q entity_id=%q actor=%q", action, entityID, actorID)
	}
}

// DOC-02 AC2: sem a permissão exigida, nenhuma URL é gerada.
func TestAccessURLRequiresTheCallerSuppliedPermissionAndGeneratesNoURL(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc02ac2@exemplo.com")
	doc, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-21", Filename: "doc.pdf",
		Content: bytes.NewReader(pdfBytes()), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	before := e.st.presignCalls
	_, err = e.svc.AccessURL(context.Background(), documents.AccessInput{Actor: principal(uid), RequiredPermission: readPermission, ID: doc.ID})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	if e.st.presignCalls != before {
		t.Error("PresignGet não deveria ter sido chamado sem a permissão exigida")
	}
}

// DOC-02 AC3: id inexistente é não encontrado, sem gerar URL.
func TestAccessURLReturnsNotFoundForANonexistentIDAndGeneratesNoURL(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc02ac3@exemplo.com")

	before := e.st.presignCalls
	_, err := e.svc.AccessURL(context.Background(), documents.AccessInput{Actor: principal(uid, readPermission), RequiredPermission: readPermission, ID: uuid.New()})
	if !errors.Is(err, documents.ErrNotFound) {
		t.Errorf("err = %v, esperado ErrNotFound", err)
	}
	if e.st.presignCalls != before {
		t.Error("PresignGet não deveria ter sido chamado para um id inexistente")
	}
}

// DOC-02 AC6: AccessURL nunca resolve "a versão atual" por conta própria —
// pedir a v1 depois de a v2 existir continua devolvendo a v1.
func TestAccessURLNeverResolvesTheCurrentVersionOnItsOwn(t *testing.T) {
	e := newEnv(t)
	uid := newUser(t, e.app, "doc02ac6@exemplo.com")
	v1content := pdfBytes()
	v1, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-22", Filename: "v1.pdf",
		Content: bytes.NewReader(v1content), UploadedBy: uid,
	})
	if err != nil {
		t.Fatalf("Store v1: %v", err)
	}
	v2content := append(append([]byte{}, pdfBytes()...), []byte(" v2")...)
	if _, err := e.svc.Store(context.Background(), documents.StoreInput{
		Actor: principal(uid, writePermission), RequiredPermission: writePermission,
		OwnerType: "financeiro.lancamento", OwnerID: "l-22", Filename: "v2.pdf",
		Content: bytes.NewReader(v2content), UploadedBy: uid, Supersedes: &v1.ID,
	}); err != nil {
		t.Fatalf("Store v2: %v", err)
	}

	url, err := e.svc.AccessURL(context.Background(), documents.AccessInput{Actor: principal(uid, readPermission), RequiredPermission: readPermission, ID: v1.ID})
	if err != nil {
		t.Fatalf("AccessURL v1: %v", err)
	}
	if got := download(t, url); !bytes.Equal(got, v1content) {
		t.Error("pedir o id da v1 deveria devolver o conteúdo da v1, não o da v2")
	}
}
