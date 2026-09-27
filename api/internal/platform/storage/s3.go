package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config points an S3Storage at an endpoint. It only ever uses the S3
// protocol: no field here names or depends on a specific provider (ADR-006).
type Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	// UsePathStyle is required by local S3 emulators.
	UsePathStyle bool
}

// S3Storage implements Storage against any S3-compatible endpoint via the AWS
// SDK for Go v2 — the SDK's name is a historical artifact of AWS having
// defined the protocol; nothing here is AWS-specific.
type S3Storage struct {
	client *s3.Client
	bucket string
}

var _ Storage = (*S3Storage)(nil)

// NewS3 builds the client. It does no I/O: constructing an aws.Config from
// static credentials is a local computation, so a bad endpoint or bucket only
// surfaces on the first call.
func NewS3(ctx context.Context, cfg Config) (*S3Storage, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("storage: configurar cliente s3: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = cfg.UsePathStyle
	})
	return &S3Storage{client: client, bucket: cfg.Bucket}, nil
}

// Put uploads r, streamed, under key. A failure (network, credentials, bucket
// unreachable) is wrapped in ErrWrite; the wrap keeps the underlying detail
// for logs, but a caller that maps errors through errors.Is (as this codebase
// always does — see identity/http's writeError) never repeats that detail
// back to a client.
func (s *S3Storage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), Key: aws.String(key), Body: r, ContentLength: aws.Int64(size), ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWrite, err)
	}
	return nil
}

// PresignGet returns a GET URL signed to expire after ttl. Presigning is a
// local computation (no network round trip), so this practically never fails;
// when it does, the error is returned as is, since it is always a
// programming or configuration mistake, not a transient storage failure.
func (s *S3Storage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	presigner := s3.NewPresignClient(s.client)
	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("storage: assinar url: %w", err)
	}
	return req.URL, nil
}

// DeleteCreated removes key. See the Storage interface: this only ever undoes
// an upload the caller just made, never a document that was recorded.
func (s *S3Storage) DeleteCreated(ctx context.Context, key string) error {
	if _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}); err != nil {
		return fmt.Errorf("%w: %w", ErrWrite, err)
	}
	return nil
}
