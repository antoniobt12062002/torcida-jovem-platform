package app

import (
	"context"
	"errors"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

// ResetTokens is what ResetPasswordWithToken needs of the recovery repository.
type ResetTokens interface {
	FindValid(ctx context.Context, tokenHash []byte, now time.Time) (string, error)
	Consume(ctx context.Context, tokenHash []byte, now time.Time) (string, error)
	InvalidatePending(ctx context.Context, userID string, at time.Time) error
}

const maxTokenLen = 512

// ResetPasswordWithToken sets a new password with a recovery token.
//
// An unknown, expired, used or inactive-user token is always the same
// domain.ErrInvalidResetToken and an `auth.password_reset_failed` security event.
// A password that breaks the policy keeps the token valid. On success, in one
// transaction, the token is consumed (single use), the hash replaced,
// must_change_password cleared, the user's other pending tokens invalidated and
// every session revoked, and `auth.password_reset_completed` is audited. Neither
// the token nor the password reaches the audit trail.
type ResetPasswordWithToken struct {
	Recovery  ResetTokens
	Users     PasswordUsers
	Roles     RoleReader
	Sessions  SessionRevoker
	Passwords PasswordHasher
	Denylist  *password.Denylist
	Audit     Auditor
	Tx        TxFunc
	Now       Clock
}

type ResetInput struct {
	Token       string
	NewPassword string
}

func (uc *ResetPasswordWithToken) Execute(ctx context.Context, in ResetInput) error {
	if uc.Recovery == nil || uc.Users == nil || uc.Roles == nil || uc.Sessions == nil || uc.Passwords == nil || uc.Denylist == nil || uc.Audit == nil || uc.Tx == nil || uc.Now == nil {
		return errNotConfigured
	}
	now := uc.Now()
	if in.Token == "" || len(in.Token) > maxTokenLen {
		return uc.invalid(ctx, "malformed")
	}
	tokenHash := domain.HashSessionToken(in.Token)
	userID, err := uc.Recovery.FindValid(ctx, tokenHash, now)
	if errors.Is(err, domain.ErrInvalidResetToken) {
		return uc.invalid(ctx, "unknown_expired_or_used")
	}
	if err != nil {
		return err
	}
	user, err := uc.Users.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if !user.Active {
		return uc.invalid(ctx, "inactive_user")
	}
	roles, err := uc.Roles.RolesOf(ctx, user.ID)
	if err != nil {
		return err
	}
	if err := domain.ValidatePassword(in.NewPassword, roles, uc.Denylist); err != nil {
		return err
	}
	hash, err := uc.Passwords.Hash(in.NewPassword)
	if err != nil {
		return err
	}

	err = uc.Tx(ctx, func(ctx context.Context) error {
		if _, err := uc.Recovery.Consume(ctx, tokenHash, now); err != nil {
			return err
		}
		if err := uc.Users.SetPassword(ctx, user.ID, hash, false); err != nil {
			return err
		}
		if err := uc.Recovery.InvalidatePending(ctx, user.ID, now); err != nil {
			return err
		}
		if _, err := uc.Sessions.RevokeAllForUser(ctx, user.ID); err != nil {
			return err
		}
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: user.ID, Action: audit.AuthPasswordResetCompleted,
			EntityType: "user", EntityID: user.ID, Outcome: audit.OutcomeSuccess,
			Before: map[string]any{"must_change_password": user.MustChangePassword}, After: map[string]any{"must_change_password": false},
		})
	})
	if errors.Is(err, domain.ErrInvalidResetToken) { // another confirmation used it first
		return uc.invalid(ctx, "used_concurrently")
	}
	return err
}

// invalid audits the failure as a security event (no token, no address) and
// returns the one error every invalid token gets.
func (uc *ResetPasswordWithToken) invalid(ctx context.Context, reason string) error {
	uc.Audit.RecordSecurity(ctx, audit.Entry{
		ActorType: audit.ActorAnonymous, Action: audit.AuthPasswordResetFailed, EntityType: "password_reset", EntityID: "token",
		Outcome: audit.OutcomeFailure, Context: map[string]any{"reason": reason},
	})
	return domain.ErrInvalidResetToken
}
