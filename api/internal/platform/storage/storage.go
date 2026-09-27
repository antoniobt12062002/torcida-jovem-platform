// Package storage is a minimal object storage abstraction over the S3
// protocol (ADR-006). No specific provider is assumed: implementations talk
// to anything that speaks the S3 protocol, chosen later.
package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrNotFound is returned when the requested object does not exist.
var ErrNotFound = errors.New("storage: objeto não encontrado")

// ErrWrite wraps a failure to write an object (network, credentials, bucket
// unreachable). It carries no detail from the underlying client, so a caller
// mapping it to HTTP never risks leaking storage internals.
var ErrWrite = errors.New("storage: falha ao gravar")

// Storage is what platform/documents needs of an S3-compatible object store.
// There is deliberately no operation that overwrites or deletes an object once
// it has been referenced by a document: DeleteCreated only ever undoes an
// upload that was never committed to the database, in the same request that
// created it.
type Storage interface {
	// Put uploads r (exactly size bytes) under key, streamed, with contentType.
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error

	// PresignGet returns a short-lived, signed GET URL for key, valid for ttl.
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)

	// DeleteCreated removes the object at key. It exists only to undo an
	// upload the caller just made when persisting its metadata failed in the
	// same operation — never to delete a document that was ever recorded.
	DeleteCreated(ctx context.Context, key string) error
}
