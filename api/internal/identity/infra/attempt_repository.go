package infra

import (
	"context"
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

// AttemptRepository keeps the failed login attempts of each e-mail, identified
// only by its HMAC (never the address). A success clears them.
type AttemptRepository struct{ db *gorm.DB }

func NewAttemptRepository(db *gorm.DB) *AttemptRepository { return &AttemptRepository{db: db} }

// RecordFailure stores a failed attempt at the given instant. It also drops
// that e-mail's attempts that are too old to matter, so the table does not grow
// without bound.
func (r *AttemptRepository) RecordFailure(ctx context.Context, emailHash []byte, at time.Time) error {
	c := conn(ctx, r.db)
	stale := at.Add(-4 * (domain.LoginFailureWindow + domain.LoginLockDuration))
	if err := c.Exec("DELETE FROM login_attempts WHERE email_hash = ? AND attempted_at < ?", bytesParam(emailHash), stale).Error; err != nil {
		return fmt.Errorf("tentativas: limpar antigas: %w", err)
	}
	if err := c.Exec("INSERT INTO login_attempts (email_hash, attempted_at, success) VALUES (?, ?, false)", bytesParam(emailHash), at).Error; err != nil {
		return fmt.Errorf("tentativas: registrar falha: %w", err)
	}
	return nil
}

// RecordSuccess clears the failures of the e-mail: the count starts over.
func (r *AttemptRepository) RecordSuccess(ctx context.Context, emailHash []byte) error {
	if err := conn(ctx, r.db).Exec("DELETE FROM login_attempts WHERE email_hash = ?", bytesParam(emailHash)).Error; err != nil {
		return fmt.Errorf("tentativas: zerar contagem: %w", err)
	}
	return nil
}

// Blocked reports whether attempts for the e-mail are blocked at now, and until
// when, by the rule of domain.Lockout applied to the latest failures.
func (r *AttemptRepository) Blocked(ctx context.Context, emailHash []byte, now time.Time) (until time.Time, blocked bool, err error) {
	var failures []time.Time
	err = conn(ctx, r.db).Raw(`SELECT attempted_at FROM login_attempts
		WHERE email_hash = ? AND NOT success ORDER BY attempted_at DESC LIMIT ?`, bytesParam(emailHash), domain.MaxLoginFailures).Scan(&failures).Error
	if err != nil {
		return time.Time{}, false, fmt.Errorf("tentativas: consultar: %w", err)
	}
	slices.Reverse(failures) // oldest first
	until, blocked = domain.Lockout(failures, now)
	return until, blocked, nil
}
