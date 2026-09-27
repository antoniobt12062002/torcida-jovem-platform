//go:build integration

package storage_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/storage"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

func newS3Storage(t *testing.T) (*storage.S3Storage, *testutil.S3) {
	t.Helper()
	env := testutil.SharedS3(t)
	s, err := storage.NewS3(context.Background(), storage.Config{
		Endpoint: env.Endpoint, Region: env.Region, Bucket: env.Bucket,
		AccessKey: env.AccessKey, SecretKey: env.SecretKey, UsePathStyle: true,
	})
	if err != nil {
		t.Fatalf("storage.NewS3: %v", err)
	}
	return s, env
}

// DOC-01 AC1 / DOC-02 AC1: um objeto gravado com Put é lido de volta através
// da URL assinada por PresignGet — a prova de ponta a ponta do adapter.
func TestS3PutAndPresignGetRoundTrip(t *testing.T) {
	s, _ := newS3Storage(t)
	ctx := context.Background()
	key := "storage/roundtrip.txt"
	body := []byte("conteudo do documento")

	if err := s.Put(ctx, key, bytes.NewReader(body), int64(len(body)), "text/plain"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	presigned, err := s.PresignGet(ctx, key, time.Minute)
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}

	resp, err := http.Get(presigned)
	if err != nil {
		t.Fatalf("GET da url assinada: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, esperado 200", resp.StatusCode)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ler corpo: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("corpo = %q, esperado %q", got, body)
	}
}

// DOC-02.4 (reformulado): o sistema pede ao adapter de storage uma URL com o
// TTL configurado, e a URL gerada carrega esse mesmo valor em
// X-Amz-Expires — decodificado aqui, não observado por comportamento do
// emulador (ver AD-015 e o comentário em testutil/s3_test.go).
func TestS3PresignGetEncodesTheRequestedTTLInTheURL(t *testing.T) {
	s, _ := newS3Storage(t)
	ctx := context.Background()
	key := "storage/ttl.txt"
	if err := s.Put(ctx, key, strings.NewReader("x"), 1, "text/plain"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	const ttl = 42 * time.Second
	presigned, err := s.PresignGet(ctx, key, ttl)
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}

	parsed, err := url.Parse(presigned)
	if err != nil {
		t.Fatalf("parsear url assinada: %v", err)
	}
	got := parsed.Query().Get("X-Amz-Expires")
	if got != "42" {
		t.Errorf("X-Amz-Expires = %q, esperado %q (ttl pedido: %s)", got, "42", ttl)
	}
}

// DOC-02.5: pedir o objeto direto do bucket, sem URL assinada, é recusado
// pelo storage — provado aqui contra o adapter real, não só contra o cliente
// de teste cru (ver testutil.TestSharedS3RejectsAnUnsignedRequest).
func TestS3RejectsAnUnsignedRequest(t *testing.T) {
	s, env := newS3Storage(t)
	ctx := context.Background()
	key := "storage/sem-assinatura.txt"
	if err := s.Put(ctx, key, strings.NewReader("x"), 1, "text/plain"); err != nil {
		t.Fatalf("Put: %v", err)
	}

	resp, err := http.Get(env.Endpoint + "/" + env.Bucket + "/" + key)
	if err != nil {
		t.Fatalf("GET sem assinatura: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusOK {
		t.Errorf("uma requisição sem assinatura deveria ser recusada, veio %d", resp.StatusCode)
	}
}

// DeleteCreated remove só o objeto pedido, nunca outro.
func TestS3DeleteCreatedRemovesOnlyThatObject(t *testing.T) {
	s, _ := newS3Storage(t)
	ctx := context.Background()
	keep := "storage/mantido.txt"
	remove := "storage/removido.txt"
	if err := s.Put(ctx, keep, strings.NewReader("mantido"), 7, "text/plain"); err != nil {
		t.Fatalf("Put keep: %v", err)
	}
	if err := s.Put(ctx, remove, strings.NewReader("removido"), 8, "text/plain"); err != nil {
		t.Fatalf("Put remove: %v", err)
	}

	if err := s.DeleteCreated(ctx, remove); err != nil {
		t.Fatalf("DeleteCreated: %v", err)
	}

	keptURL, err := s.PresignGet(ctx, keep, time.Minute)
	if err != nil {
		t.Fatalf("PresignGet keep: %v", err)
	}
	resp, err := http.Get(keptURL)
	if err != nil {
		t.Fatalf("GET keep: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("objeto mantido deveria continuar acessível, status = %d", resp.StatusCode)
	}

	removedURL, err := s.PresignGet(ctx, remove, time.Minute)
	if err != nil {
		t.Fatalf("PresignGet remove: %v", err)
	}
	resp2, err := http.Get(removedURL)
	if err != nil {
		t.Fatalf("GET remove: %v", err)
	}
	_ = resp2.Body.Close()
	if resp2.StatusCode == http.StatusOK {
		t.Errorf("objeto removido não deveria mais responder 200, veio %d", resp2.StatusCode)
	}
}

// Uma falha de gravação (endpoint inalcançável) é reportada como ErrWrite,
// para que quem chamar Put possa mapear o erro sem depender do texto
// devolvido pelo SDK.
func TestS3PutWrapsAnUnreachableEndpointInErrWrite(t *testing.T) {
	s, err := storage.NewS3(context.Background(), storage.Config{
		Endpoint: "http://127.0.0.1:1", Region: "garage", Bucket: "inexistente",
		AccessKey: "x", SecretKey: "y", UsePathStyle: true,
	})
	if err != nil {
		t.Fatalf("storage.NewS3: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.Put(ctx, "qualquer", strings.NewReader("x"), 1, "text/plain")
	if err == nil {
		t.Fatal("esperava erro ao gravar num endpoint inalcançável")
	}
	if !errors.Is(err, storage.ErrWrite) {
		t.Errorf("erro = %v, esperava que envolvesse storage.ErrWrite", err)
	}
}
