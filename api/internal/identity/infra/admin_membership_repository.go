package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// AdminMembershipRepository persists the administrative memberships. It has no
// delete operation: closing a membership keeps the row, and the history stays.
type AdminMembershipRepository struct{ db *gorm.DB }

func NewAdminMembershipRepository(db *gorm.DB) *AdminMembershipRepository {
	return &AdminMembershipRepository{db: db}
}

type membershipRow struct {
	ID           string
	UserID       string
	Reason       string
	GrantedBy    *string
	GrantedAt    time.Time
	RevokedBy    *string
	RevokeReason *string
	RevokedAt    *time.Time
}

func (r membershipRow) membership() domain.AdminMembership {
	return domain.AdminMembership{
		ID: r.ID, UserID: r.UserID, Reason: r.Reason, GrantedBy: r.GrantedBy, GrantedAt: r.GrantedAt,
		RevokedBy: r.RevokedBy, RevokeReason: r.RevokeReason, RevokedAt: r.RevokedAt,
	}
}

const membershipColumns = `id::text AS id, user_id::text AS user_id, reason, granted_by::text AS granted_by, granted_at,
	revoked_by::text AS revoked_by, revoke_reason, revoked_at`

// Grant stores a new membership. If the user already has an active one it
// returns domain.ErrAlreadyAdmin; with simultaneous grants exactly one wins,
// because a partial unique index in the database decides.
func (r *AdminMembershipRepository) Grant(ctx context.Context, m domain.AdminMembership) (domain.AdminMembership, error) {
	var row membershipRow
	err := conn(ctx, r.db).Raw(`INSERT INTO admin_memberships (user_id, reason, granted_by, granted_at)
		VALUES (?::uuid, ?, ?::uuid, ?) RETURNING `+membershipColumns,
		m.UserID, m.Reason, m.GrantedBy, m.GrantedAt).Scan(&row).Error
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "admin_memberships_active_uq" {
			return domain.AdminMembership{}, domain.ErrAlreadyAdmin
		}
		return domain.AdminMembership{}, fmt.Errorf("vínculo administrativo: conceder: %w", err)
	}
	return row.membership(), nil
}

// Active returns the user's active membership, if any.
func (r *AdminMembershipRepository) Active(ctx context.Context, userID string) (domain.AdminMembership, bool, error) {
	if !uuidFormat.MatchString(userID) {
		return domain.AdminMembership{}, false, nil
	}
	var rows []membershipRow
	err := conn(ctx, r.db).Raw("SELECT "+membershipColumns+" FROM admin_memberships WHERE user_id = ?::uuid AND revoked_at IS NULL", userID).Scan(&rows).Error
	if err != nil {
		return domain.AdminMembership{}, false, fmt.Errorf("vínculo administrativo: consultar: %w", err)
	}
	if len(rows) == 0 {
		return domain.AdminMembership{}, false, nil
	}
	return rows[0].membership(), true, nil
}

// History lists every membership of the user, newest first.
func (r *AdminMembershipRepository) History(ctx context.Context, userID string) ([]domain.AdminMembership, error) {
	if !uuidFormat.MatchString(userID) {
		return nil, nil
	}
	var rows []membershipRow
	err := conn(ctx, r.db).Raw("SELECT "+membershipColumns+" FROM admin_memberships WHERE user_id = ?::uuid ORDER BY granted_at DESC, id DESC", userID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("vínculo administrativo: histórico: %w", err)
	}
	out := make([]domain.AdminMembership, len(rows))
	for i, row := range rows {
		out[i] = row.membership()
	}
	return out, nil
}

// Revoke closes the user's active membership, recording who, when and why, and
// returns it. Without an active membership it returns domain.ErrNotAdmin.
func (r *AdminMembershipRepository) Revoke(ctx context.Context, userID, by, reason string, at time.Time) (domain.AdminMembership, error) {
	var closed domain.AdminMembership
	err := database.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx, _ := database.TxFrom(ctx)
		if !uuidFormat.MatchString(userID) {
			return domain.ErrNotAdmin
		}
		var rows []membershipRow
		err := tx.Raw("SELECT "+membershipColumns+" FROM admin_memberships WHERE user_id = ?::uuid AND revoked_at IS NULL FOR UPDATE", userID).Scan(&rows).Error
		if err != nil {
			return fmt.Errorf("vínculo administrativo: consultar: %w", err)
		}
		if len(rows) == 0 {
			return domain.ErrNotAdmin
		}
		next, err := rows[0].membership().Revoke(by, reason, at)
		if err != nil {
			return err
		}
		res := tx.Exec(`UPDATE admin_memberships SET revoked_by = ?::uuid, revoke_reason = ?, revoked_at = ?
			WHERE id = ?::uuid AND revoked_at IS NULL`,
			next.RevokedBy, next.RevokeReason, next.RevokedAt, next.ID)
		if res.Error != nil {
			return fmt.Errorf("vínculo administrativo: encerrar: %w", res.Error)
		}
		if res.RowsAffected == 0 { // someone else closed it first
			return domain.ErrNotAdmin
		}
		closed = next
		return nil
	})
	return closed, err
}
