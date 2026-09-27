//go:build integration

package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultS3Image is Garage (Deuxfleurs), chosen 2026-09-27 after checking the
// license and maintenance of the alternatives:
//   - MinIO: community edition archived 2026-04-25 (read-only source, no
//     official binaries or images published since).
//   - LocalStack: community edition support ended 2026-03-23; running it now
//     requires an account, unsuitable for an offline, credential-free test
//     fixture.
//   - SeaweedFS (Apache-2.0, actively maintained, v4.47): tried first for its
//     one-command `weed mini` setup. Chosen candidates keep answering a
//     presigned GET URL after its requested expiry (see the AD-015 note in
//     s3_test.go) — a documented gap across free local S3 emulators, not a
//     defect specific to one implementation.
//   - Garage: AGPL-3.0 (a test-only runtime dependency: we run the published
//     image, never modify or redistribute it, so this does not affect the
//     license of this repository's own code — the same reasoning that
//     already applies to running PostgreSQL, itself under its own license,
//     as a test fixture), actively maintained (v2.4.1, released 2026-09-08,
//     https://hub.docker.com/r/dxflrs/garage/tags). Since v2.3.0 it supports
//     a one-command single-node bootstrap (`--single-node --default-bucket`,
//     https://garagehq.deuxfleurs.fr/documentation/quick-start/), and
//     correctly rejects a request with no valid signature (DOC-02.5,
//     verified below).
const defaultS3Image = "dxflrs/garage:v2.4.1"

const s3Port = "3900/tcp"

// Fixed test credentials and bucket, seeded into the container by Garage's
// --default-bucket bootstrap. Not secrets: the container is ephemeral and
// reachable only from the test process.
const (
	S3TestAccessKey = "tjtestaccesskey0000"
	S3TestSecretKey = "tj-test-secret-key-0123456789abcdef"
	S3TestBucket    = "tj-test-bucket"
	s3TestRegion    = "garage"
)

// garageConfig is the minimal single-node configuration Garage needs. rpc_secret
// only authenticates a node to itself here (single node, no real cluster), so a
// fresh random value per container is enough — nothing depends on its value
// being stable across runs.
const garageConfigTemplate = `
metadata_dir = "/tmp/garage/meta"
data_dir = "/tmp/garage/data"
db_engine = "sqlite"
replication_factor = 1

rpc_bind_addr = "[::]:3901"
rpc_public_addr = "127.0.0.1:3901"
rpc_secret = "%s"

[s3_api]
s3_region = "%s"
api_bind_addr = "[::]:3900"
root_domain = ".s3.garage.localhost"
`

// S3 is a local S3-compatible object storage container for tests.
type S3 struct {
	container testcontainers.Container
	// Endpoint is the S3 API base URL (http://host:port). The emulator only
	// supports path-style addressing.
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Bucket    string
}

var (
	sharedS3Once sync.Once
	sharedS3     *S3
	sharedS3Err  error
)

// SharedS3 starts the emulator once per test binary (one package). The
// testcontainers reaper removes the container when the process exits.
func SharedS3(t testing.TB) *S3 {
	t.Helper()
	sharedS3Once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		sharedS3, sharedS3Err = startS3(ctx)
	})
	if sharedS3Err != nil {
		t.Fatalf("%v", sharedS3Err)
	}
	return sharedS3
}

func rpcSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func startS3(ctx context.Context) (*S3, error) {
	image := os.Getenv("TJ_TEST_S3_IMAGE")
	if image == "" {
		image = defaultS3Image
	}
	secret, err := rpcSecret()
	if err != nil {
		return nil, fmt.Errorf("s3 de teste: gerar rpc_secret: %w", err)
	}
	toml := fmt.Sprintf(garageConfigTemplate, secret, s3TestRegion)

	req := testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{s3Port},
			Env: map[string]string{
				"GARAGE_DEFAULT_ACCESS_KEY": S3TestAccessKey,
				"GARAGE_DEFAULT_SECRET_KEY": S3TestSecretKey,
				"GARAGE_DEFAULT_BUCKET":     S3TestBucket,
			},
			Files: []testcontainers.ContainerFile{{
				Reader:            strings.NewReader(toml),
				ContainerFilePath: "/etc/garage.toml",
				FileMode:          0o644,
			}},
			Cmd:        []string{"/garage", "server", "--single-node", "--default-bucket"},
			WaitingFor: wait.ForListeningPort(s3Port).WithStartupTimeout(90 * time.Second),
		},
	}
	c, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		logs := ""
		if c != nil {
			logs = containerLogs(ctx, c)
			_ = testcontainers.TerminateContainer(c)
		}
		return nil, fmt.Errorf("s3 de teste: falha ao iniciar via Docker (verifique se o Docker está em execução; é necessário com a tag integration): %w\n%s", err, logs)
	}
	host, err := c.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("s3 de teste: endereço do contêiner: %w", err)
	}
	port, err := c.MappedPort(ctx, s3Port)
	if err != nil {
		return nil, fmt.Errorf("s3 de teste: porta do contêiner: %w", err)
	}
	env := &S3{
		container: c,
		Endpoint:  fmt.Sprintf("http://%s:%s", host, port.Port()),
		Region:    s3TestRegion,
		AccessKey: S3TestAccessKey, SecretKey: S3TestSecretKey, Bucket: S3TestBucket,
	}
	// The port opens before --default-bucket finishes creating the key and the
	// bucket; wait until the bucket is actually usable.
	if err := waitBucketReady(ctx, env); err != nil {
		logs := containerLogs(ctx, c)
		_ = testcontainers.TerminateContainer(c)
		return nil, fmt.Errorf("s3 de teste: bucket padrão nunca ficou pronto: %w\n%s", err, logs)
	}
	return env, nil
}

func waitBucketReady(ctx context.Context, env *S3) error {
	client, err := newS3Client(ctx, env)
	if err != nil {
		return err
	}
	var lastErr error
	for range 30 {
		_, lastErr = client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(env.Bucket)})
		if lastErr == nil {
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	return lastErr
}

func newS3Client(ctx context.Context, env *S3) (*s3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(env.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(env.AccessKey, env.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("s3 de teste: config do cliente: %w", err)
	}
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(env.Endpoint)
		o.UsePathStyle = true
	}), nil
}

// Client returns an S3 client configured for this container: path-style, the
// fixed test credentials and region.
func (e *S3) Client(ctx context.Context) (*s3.Client, error) { return newS3Client(ctx, e) }
