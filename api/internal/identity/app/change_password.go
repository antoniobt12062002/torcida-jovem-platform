package app

import (
	"context"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

// PasswordChangeBlockedError is returned while the password change is locked
// after repeated wrong current passwords; RetryAfter feeds Retry-After (429).
type PasswordChangeBlockedError struct{ RetryAfter time.Duration }

func (e *PasswordChangeBlockedError) Error() string { return "password_change_blocked" }

// Ports of ChangePassword.
type (
	PasswordUsers interface {
		FindByID(ctx context.Context, id string) (domain.User, error)
		SetPassword(ctx context.Context, id, hash string, mustChange bool) error
	}
	RoleReader interface {
		RolesOf(ctx context.Context, userID string) ([]domain.Role, error)
	}
	ChangeThrottle interface {
		PasswordChangeBlocked(ctx context.Context, userID string, now time.Time) (until time.Time, blocked bool, err error)
		RecordPasswordChangeFailure(ctx context.Context, userID string, at time.Time) error
		ClearPasswordChangeFailures(ctx context.Context, userID string) error
	}
	SessionsExcept interface {
		RevokeAllForUser(ctx context.Context, userID string) (int64, error)
		RevokeAllForUserExcept(ctx context.Context, userID, keepSessionID string) (int64, error)
	}
	PasswordService interface {
		Hash(plain string) (string, error)
		Verify(plain, encoded string) (bool, error)
	}
)

// ChangePassword lets an authenticated user change their own password. It needs
// the current one; the wrong ones are counted apart from the login attempts (five
// in fifteen minutes lock the change for fifteen minutes); the new password
// must follow the policy for the user's roles and differ from the current one.
// Success clears must_change_password, revokes the user's other sessions and is
// audited, in one transaction.
type ChangePassword struct {
	Users     PasswordUsers
	Roles     RoleReader
	Throttle  ChangeThrottle
	Sessions  SessionsExcept
	Passwords PasswordService
	Denylist  *password.Denylist
	Audit     Auditor
	Tx        TxFunc
	Now       Clock
}

type ChangePasswordInput struct {
	UserID           string
	CurrentSessionID string // kept; empty revokes every session
	Current          string
	New              string
}

func (uc *ChangePassword) Execute(ctx context.Context, in ChangePasswordInput) error {
	if uc.Users == nil || uc.Roles == nil || uc.Throttle == nil || uc.Sessions == nil || uc.Passwords == nil || uc.Denylist == nil || uc.Audit == nil || uc.Tx == nil || uc.Now == nil {
		return errNotConfigured
	}
	now := uc.Now()
	if until, blocked, err := uc.Throttle.PasswordChangeBlocked(ctx, in.UserID, now); err != nil {
		return err
	} else if blocked {
		return &PasswordChangeBlockedError{RetryAfter: until.Sub(now)}
	}
	user, err := uc.Users.FindByID(ctx, in.UserID)
	if err != nil {
		return err
	}
	if ok, _ := uc.Passwords.Verify(in.Current, user.PasswordHash); !ok {
		if err := uc.Throttle.RecordPasswordChangeFailure(ctx, user.ID, now); err != nil {
			return err
		}
		uc.Audit.RecordSecurity(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: user.ID, Action: audit.UserPasswordChange,
			EntityType: "user", EntityID: user.ID, Outcome: audit.OutcomeFailure,
			Context: map[string]any{"failure": "invalid_current_password"},
		})
		return domain.ErrInvalidCurrentPassword
	}
	if same, _ := uc.Passwords.Verify(in.New, user.PasswordHash); same {
		return domain.ErrPasswordUnchanged
	}
	roles, err := uc.Roles.RolesOf(ctx, user.ID)
	if err != nil {
		return err
	}
	if err := domain.ValidatePassword(in.New, roles, uc.Denylist); err != nil {
		return err
	}
	hash, err := uc.Passwords.Hash(in.New)
	if err != nil {
		return err
	}
	return uc.Tx(ctx, func(ctx context.Context) error {
		if err := uc.Users.SetPassword(ctx, user.ID, hash, false); err != nil {
			return err
		}
		var revokeErr error
		if in.CurrentSessionID == "" {
			_, revokeErr = uc.Sessions.RevokeAllForUser(ctx, user.ID)
		} else {
			_, revokeErr = uc.Sessions.RevokeAllForUserExcept(ctx, user.ID, in.CurrentSessionID)
		}
		if revokeErr != nil {
			return revokeErr
		}
		if err := uc.Throttle.ClearPasswordChangeFailures(ctx, user.ID); err != nil {
			return err
		}
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: user.ID, Action: audit.UserPasswordChange,
			EntityType: "user", EntityID: user.ID, Outcome: audit.OutcomeSuccess,
			Before: map[string]any{"must_change_password": user.MustChangePassword}, After: map[string]any{"must_change_password": false},
		})
	})
}
