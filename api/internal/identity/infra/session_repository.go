package infra

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

// SessionRepository persists server-side sessions. Only the SHA-256 of the
// token is stored.
type SessionRepository struct{ db *gorm.DB }

func NewSessionRepository(db *gorm.DB) *SessionRepository { return &SessionRepository{db: db} }

type sessionRow struct {
	ID         string
	UserID     string
	CSRFToken  string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
}

func (r sessionRow) session() domain.Session {
	return domain.Session{
		ID: r.ID, UserID: r.UserID, CSRFToken: r.CSRFToken, CreatedAt: r.CreatedAt,
		LastSeenAt: r.LastSeenAt, ExpiresAt: r.ExpiresAt, RevokedAt: r.RevokedAt,
	}
}

const sessionColumns = `id::text AS id, user_id::text AS user_id, csrf_token, created_at, last_seen_at, expires_at, revoked_at`

// Create stores a session under the SHA-256 of its token, created and last seen
// at the given instant (the caller's clock, so tests can control time).
func (r *SessionRepository) Create(ctx context.Context, userID string, tokenHash []byte, csrfToken string, at, expiresAt time.Time) (domain.Session, error) {
	var row sessionRow
	err := conn(ctx, r.db).Raw(`INSERT INTO sessions (user_id, token_hash, csrf_token, created_at, last_seen_at, expires_at)
		VALUES (?::uuid, ?, ?, ?, ?, ?) RETURNING `+sessionColumns, userID, bytesParam(tokenHash), csrfToken, at, at, expiresAt).Scan(&row).Error
	if err != nil {
		return domain.Session{}, fmt.Errorf("sessão: criar: %w", err)
	}
	return row.session(), nil
}

// FindByToken looks the session up by the hash of the token. It returns
// revoked and expired sessions too; the session service decides validity.
func (r *SessionRepository) FindByToken(ctx context.Context, token string) (domain.Session, error) {
	var rows []sessionRow
	err := conn(ctx, r.db).Raw("SELECT "+sessionColumns+" FROM sessions WHERE token_hash = ?", bytesParam(domain.HashSessionToken(token))).Scan(&rows).Error
	if err != nil {
		return domain.Session{}, fmt.Errorf("sessão: consultar: %w", err)
	}
	if len(rows) == 0 {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return rows[0].session(), nil
}

// Touch moves last_seen_at to at, but only when the current value is older than
// minInterval, so a busy client does not cause a write per request. It never
// touches a revoked session. It reports whether it wrote.
func (r *SessionRepository) Touch(ctx context.Context, sessionID string, at time.Time, minInterval time.Duration) (bool, error) {
	if !uuidFormat.MatchString(sessionID) {
		return false, nil
	}
	res := conn(ctx, r.db).Exec(`UPDATE sessions SET last_seen_at = ?
		WHERE id = ?::uuid AND revoked_at IS NULL AND last_seen_at <= ?`, at, sessionID, at.Add(-minInterval))
	if res.Error != nil {
		return false, fmt.Errorf("sessão: atualizar uso: %w", res.Error)
	}
	return res.RowsAffected > 0, nil
}

// Revoke revokes one session; revoking twice, or an unknown id, does nothing.
func (r *SessionRepository) Revoke(ctx context.Context, sessionID string) error {
	if !uuidFormat.MatchString(sessionID) {
		return nil
	}
	err := conn(ctx, r.db).Exec("UPDATE sessions SET revoked_at = now() WHERE id = ?::uuid AND revoked_at IS NULL", sessionID).Error
	if err != nil {
		return fmt.Errorf("sessão: revogar: %w", err)
	}
	return nil
}

// RevokeAllForUser revokes every active session of the user and returns how
// many. Inside a use case it joins the transaction of the context, so it commits
// or rolls back with the change that caused it.
func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID string) (int64, error) {
	return r.revokeAll(ctx, userID, "")
}

// RevokeAllForUserExcept revokes the user's active sessions but the given one.
func (r *SessionRepository) RevokeAllForUserExcept(ctx context.Context, userID, keepSessionID string) (int64, error) {
	if !uuidFormat.MatchString(keepSessionID) {
		return 0, fmt.Errorf("sessão: id a preservar inválido")
	}
	return r.revokeAll(ctx, userID, keepSessionID)
}

func (r *SessionRepository) revokeAll(ctx context.Context, userID, except string) (int64, error) {
	if !uuidFormat.MatchString(userID) {
		return 0, nil
	}
	q := "UPDATE sessions SET revoked_at = now() WHERE user_id = ?::uuid AND revoked_at IS NULL"
	args := []any{userID}
	if except != "" {
		q += " AND id <> ?::uuid"
		args = append(args, except)
	}
	res := conn(ctx, r.db).Exec(q, args...)
	if res.Error != nil {
		return 0, fmt.Errorf("sessão: revogar todas: %w", res.Error)
	}
	return res.RowsAffected, nil
}
