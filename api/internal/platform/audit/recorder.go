package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

var (
	// ErrNoTransaction is returned by Record outside a database transaction.
	ErrNoTransaction = errors.New("audit: Record exige a transação de negócio no contexto")
	// ErrWrite wraps a failure to write the entry; a use case maps it to 500 audit_failed.
	ErrWrite = errors.New("audit: falha ao gravar o registro")
)

type actorKey struct{}

// WithActor returns a context that names the authenticated user, so entries
// recorded in it get that user as their actor.
func WithActor(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, actorKey{}, userID)
}

// Recorder writes audit entries.
type Recorder struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewRecorder(db *gorm.DB, log *slog.Logger) *Recorder {
	return &Recorder{db: db, log: log}
}

// Record writes the entry in the transaction carried by ctx (see
// database.WithTx), so the change and its audit trail commit or roll back
// together. Outside a transaction it returns ErrNoTransaction and writes
// nothing. Use it for critical business changes: if it fails, the use case must
// fail too.
func (r *Recorder) Record(ctx context.Context, e Entry) error {
	tx, ok := database.TxFrom(ctx)
	if !ok {
		return ErrNoTransaction
	}
	return insert(ctx, tx, e)
}

// RecordSecurity writes a security or telemetry event in its own transaction,
// independent of any business transaction in ctx, so it survives a rollback of
// the operation it describes. It never fails the caller: if the write fails,
// it logs an operational incident and returns.
func (r *Recorder) RecordSecurity(ctx context.Context, e Entry) {
	if err := insert(ctx, r.db, e); err != nil {
		r.log.ErrorContext(ctx, "falha ao gravar evento de segurança na auditoria",
			slog.String("incident", "audit_security_write_failed"),
			slog.String("action", string(e.Action)),
			slog.String("request_id", httpx.RequestIDFrom(ctx)),
			slog.String("error", err.Error()))
	}
}

func insert(ctx context.Context, db *gorm.DB, e Entry) error {
	if e.ActorType == "" {
		if id, _ := ctx.Value(actorKey{}).(string); id != "" {
			e.ActorType, e.ActorID = ActorUser, id
		}
	}
	if err := e.Validate(); err != nil {
		return err
	}
	e = e.Redacted()

	before, err := jsonOrNil(e.Before)
	if err != nil {
		return err
	}
	after, err := jsonOrNil(e.After)
	if err != nil {
		return err
	}
	contextJSON, err := json.Marshal(nonNil(e.Context))
	if err != nil {
		return fmt.Errorf("%w: contexto: %v", ErrInvalidEntry, err)
	}

	var actorUserID, reason, requestID any
	if e.ActorType == ActorUser {
		actorUserID = e.ActorID
	}
	if e.Reason != "" {
		reason = e.Reason
	}
	if id := httpx.RequestIDFrom(ctx); id != "" {
		requestID = id
	}

	err = db.WithContext(ctx).Exec(`INSERT INTO audit_log
		(actor_type, actor_user_id, action, entity_type, entity_id, before, after, outcome, reason, context, request_id)
		VALUES (?, ?, ?, ?, ?, ?::jsonb, ?::jsonb, ?, ?, ?::jsonb, ?)`,
		string(e.ActorType), actorUserID, string(e.Action), e.EntityType, e.EntityID,
		before, after, string(e.Outcome), reason, string(contextJSON), requestID).Error
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWrite, err)
	}
	return nil
}

func jsonOrNil(m map[string]any) (any, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidEntry, err)
	}
	return string(b), nil
}

func nonNil(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}
