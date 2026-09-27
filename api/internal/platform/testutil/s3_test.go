//go:build integration

package testutil_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// s3Client builds a real S3 client (path-style, the fixed test credentials)
// against the shared emulator — independent of platform/storage, which does
// not exist yet: this proves the chosen emulator itself behaves like S3.
func s3Client(t *testing.T) (*s3.Client, *testutil.S3) {
	t.Helper()
	env := testutil.SharedS3(t)
	client, err := env.Client(context.Background())
	if err != nil {
		t.Fatalf("cliente S3: %v", err)
	}
	return client, env
}

// T3 smoke test (DOC-02 base): grava e lê um objeto pelo protocolo S3 real.
func TestSharedS3StoresAndReadsAnObject(t *testing.T) {
	client, env := s3Client(t)
	ctx := context.Background()
	key := "smoke/hello.txt"
	body := []byte("ola, garage")

	if _, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(env.Bucket), Key: aws.String(key), Body: bytes.NewReader(body), ContentLength: aws.Int64(int64(len(body))),
	}); err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	out, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(env.Bucket), Key: aws.String(key)})
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	defer func() { _ = out.Body.Close() }()
	got, err := io.ReadAll(out.Body)
	if err != nil {
		t.Fatalf("ler o corpo: %v", err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("corpo lido = %q, esperado %q", got, body)
	}
}

// AD-015: nenhum teste aqui afirma que o emulador recusa uma URL assinada
// expirada. Verificado em MinIO (descontinuado antes de testar), LocalStack
// (descartado antes de testar por exigir conta), SeaweedFS 4.47 e Garage
// v2.4.1: nenhum aplica a janela de expiração de uma URL assinada, só a
// assinatura em si (uma URL de 3s de validade continuava respondendo 200
// depois de 7s de espera nos dois candidatos testados de verdade). A
// aplicação efetiva da expiração é responsabilidade do provedor S3 de
// produção; o que esta aplicação controla — pedir e gerar a URL com o TTL
// certo — é verificado em platform/storage (T4), decodificando o
// X-Amz-Expires da própria URL, não observando o comportamento do emulador.

// DOC-02.5: pedir o objeto direto do bucket, sem assinatura, é recusado.
func TestSharedS3RejectsAnUnsignedRequest(t *testing.T) {
	client, env := s3Client(t)
	ctx := context.Background()
	key := "smoke/sem-assinatura.txt"
	if _, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(env.Bucket), Key: aws.String(key), Body: bytes.NewReader([]byte("x")), ContentLength: aws.Int64(1),
	}); err != nil {
		t.Fatalf("PutObject: %v", err)
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
