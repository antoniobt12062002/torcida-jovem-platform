package infra

import (
	"context"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

// maxResetRequestsPerHour is the number of recovery requests per e-mail hash
// that still create a token.
const (
	maxResetRequestsPerHour = 3
	resetRequestWindow      = time.Hour
)

// RecoveryRepository persists the password recovery: single-use tokens (only
// their SHA-256), the count of requests per e-mail hash, and the count of wrong
// current passwords in password changes (separate from the login attempts).
type RecoveryRepository struct{ db *gorm.DB }

func NewRecoveryRepository(db *gorm.DB) *RecoveryRepository { return &RecoveryRepository{db: db} }

// CreateToken stores a token by its hash and invalidates the user's other
// pending tokens first, in the same statement sequence of the caller's unit of work.
func (r *RecoveryRepository) CreateToken(ctx context.Context, userID string, tokenHash []byte, at, expiresAt time.Time, requestID string) error {
	if !uuidFormat.MatchString(userID) {
		return domain.ErrUserNotFound
	}
	if err := r.InvalidatePending(ctx, userID, at); err != nil {
		return err
	}
	var reqID any
	if requestID != "" {
		reqID = requestID
	}
	err := conn(ctx, r.db).Exec(`INSERT INTO password_reset_tokens (user_id, token_hash, created_at, expires_at, request_id)
		VALUES (?::uuid, ?, ?, ?, ?)`, userID, bytesParam(tokenHash), at, expiresAt, reqID).Error
	if err != nil {
		return fmt.Errorf("recuperação: criar token: %w", err)
	}
	return nil
}

// InvalidatePending closes the user's tokens that are still pending. The rows stay.
func (r *RecoveryRepository) InvalidatePending(ctx context.Context, userID string, at time.Time) error {
	if !uuidFormat.MatchString(userID) {
		return nil
	}
	err := conn(ctx, r.db).Exec("UPDATE password_reset_tokens SET used_at = ? WHERE user_id = ?::uuid AND used_at IS NULL", at, userID).Error
	if err != nil {
		return fmt.Errorf("recuperação: invalidar pendentes: %w", err)
	}
	return nil
}

// FindValid returns the user of a token that is unknown-proof valid at now,
// without consuming it. Unknown, expired and used tokens are all
// domain.ErrInvalidResetToken.
func (r *RecoveryRepository) FindValid(ctx context.Context, tokenHash []byte, now time.Time) (string, error) {
	var ids []string
	err := conn(ctx, r.db).Raw(`SELECT user_id::text FROM password_reset_tokens
		WHERE token_hash = ? AND used_at IS NULL AND expires_at > ?`, bytesParam(tokenHash), now).Scan(&ids).Error
	if err != nil {
		return "", fmt.Errorf("recuperação: consultar token: %w", err)
	}
	if len(ids) == 0 {
		return "", domain.ErrInvalidResetToken
	}
	return ids[0], nil
}

// Consume marks a valid token as used and returns its user. It is atomic: with
// simultaneous confirmations exactly one wins, because the UPDATE only matches a
// token that is still unused.
func (r *RecoveryRepository) Consume(ctx context.Context, tokenHash []byte, now time.Time) (string, error) {
	var ids []string
	err := conn(ctx, r.db).Raw(`UPDATE password_reset_tokens SET used_at = ?
		WHERE token_hash = ? AND used_at IS NULL AND expires_at > ? RETURNING user_id::text`, now, bytesParam(tokenHash), now).Scan(&ids).Error
	if err != nil {
		return "", fmt.Errorf("recuperação: consumir token: %w", err)
	}
	if len(ids) == 0 {
		return "", domain.ErrInvalidResetToken
	}
	return ids[0], nil
}

// RegisterRequest records a recovery request for the e-mail hash at the given
// instant and reports whether it still counts (at most three per hour); the
// caller answers the same either way. Old requests of that hash are dropped.
func (r *RecoveryRepository) RegisterRequest(ctx context.Context, emailHash []byte, at time.Time) (bool, error) {
	c := conn(ctx, r.db)
	if err := c.Exec("DELETE FROM password_reset_requests WHERE email_hash = ? AND requested_at < ?", bytesParam(emailHash), at.Add(-2*resetRequestWindow)).Error; err != nil {
		return false, fmt.Errorf("recuperação: limpar solicitações: %w", err)
	}
	if err := c.Exec("INSERT INTO password_reset_requests (email_hash, requested_at) VALUES (?, ?)", bytesParam(emailHash), at).Error; err != nil {
		return false, fmt.Errorf("recuperação: registrar solicitação: %w", err)
	}
	var n int64
	err := c.Raw("SELECT count(*) FROM password_reset_requests WHERE email_hash = ? AND requested_at > ? AND requested_at <= ?",
		bytesParam(emailHash), at.Add(-resetRequestWindow), at).Scan(&n).Error
	if err != nil {
		return false, fmt.Errorf("recuperação: contar solicitações: %w", err)
	}
	return n <= maxResetRequestsPerHour, nil
}

// RecordPasswordChangeFailure counts a wrong current password of the user.
func (r *RecoveryRepository) RecordPasswordChangeFailure(ctx context.Context, userID string, at time.Time) error {
	if !uuidFormat.MatchString(userID) {
		return domain.ErrUserNotFound
	}
	c := conn(ctx, r.db)
	stale := at.Add(-4 * (domain.LoginFailureWindow + domain.LoginLockDuration))
	if err := c.Exec("DELETE FROM password_change_attempts WHERE user_id = ?::uuid AND attempted_at < ?", userID, stale).Error; err != nil {
		return fmt.Errorf("troca de senha: limpar tentativas: %w", err)
	}
	if err := c.Exec("INSERT INTO password_change_attempts (user_id, attempted_at) VALUES (?::uuid, ?)", userID, at).Error; err != nil {
		return fmt.Errorf("troca de senha: registrar erro: %w", err)
	}
	return nil
}

// ClearPasswordChangeFailures resets the user's count after a success.
func (r *RecoveryRepository) ClearPasswordChangeFailures(ctx context.Context, userID string) error {
	if !uuidFormat.MatchString(userID) {
		return nil
	}
	if err := conn(ctx, r.db).Exec("DELETE FROM password_change_attempts WHERE user_id = ?::uuid", userID).Error; err != nil {
		return fmt.Errorf("troca de senha: zerar contagem: %w", err)
	}
	return nil
}

// PasswordChangeBlocked applies the same lockout rule as the login (five
// failures within fifteen minutes lock for fifteen) to the user's password
// change failures.
func (r *RecoveryRepository) PasswordChangeBlocked(ctx context.Context, userID string, now time.Time) (time.Time, bool, error) {
	if !uuidFormat.MatchString(userID) {
		return time.Time{}, false, nil
	}
	var failures []time.Time
	err := conn(ctx, r.db).Raw(`SELECT attempted_at FROM password_change_attempts
		WHERE user_id = ?::uuid ORDER BY attempted_at DESC LIMIT ?`, userID, domain.MaxLoginFailures).Scan(&failures).Error
	if err != nil {
		return time.Time{}, false, fmt.Errorf("troca de senha: consultar tentativas: %w", err)
	}
	slices.Reverse(failures)
	until, blocked := domain.Lockout(failures, now)
	return until, blocked, nil
}
