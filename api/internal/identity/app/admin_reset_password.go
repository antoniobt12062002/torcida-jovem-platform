package app

import (
	"context"
	"errors"
	"unicode/utf8"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

const minResetReasonLen = 10

// AdminResetPassword is the administrative alternative to the e-mail recovery.
//
// The actor needs identity:user:reset_password, gives a reason of at least ten
// characters, cannot reset their own password through this path and must hold
// every effective permission of the target (no escalation). In one transaction
// the system generates a random temporary password, replaces the hash, sets
// must_change_password, revokes the target's sessions and pending recovery
// tokens and audits `user.password_reset` with the reason. The temporary
// password is returned once to the actor; it is never stored in clear, logged,
// audited or e-mailed, and neither are the old password or any hash.
type AdminResetPassword struct {
	Authz     Authorizer
	Users     PasswordUsers
	Perms     PermissionReader
	Roles     RoleReader
	Sessions  SessionRevoker
	Recovery  TokenInvalidator
	Passwords PasswordHasher
	Denylist  *password.Denylist
	Audit     Auditor
	Tx        TxFunc
	Now       Clock
}

type AdminResetInput struct {
	Actor    authz.Principal
	TargetID string
	Reason   string
}

// AdminResetResult carries the temporary password, shown exactly once.
type AdminResetResult struct{ TemporaryPassword string }

func (uc *AdminResetPassword) Execute(ctx context.Context, in AdminResetInput) (AdminResetResult, error) {
	if uc.Authz == nil || uc.Users == nil || uc.Perms == nil || uc.Roles == nil || uc.Sessions == nil || uc.Recovery == nil ||
		uc.Passwords == nil || uc.Denylist == nil || uc.Audit == nil || uc.Tx == nil || uc.Now == nil {
		return AdminResetResult{}, errNotConfigured
	}
	if err := uc.Authz.Require(ctx, in.Actor, "identity:user:reset_password"); err != nil {
		return AdminResetResult{}, err
	}
	if in.TargetID == in.Actor.UserID {
		return AdminResetResult{}, denyChange(ctx, uc.Audit, in.Actor, audit.UserPasswordReset, in.TargetID, domain.ErrSelfChangeForbidden, nil)
	}
	reason := trimmed(in.Reason)
	if utf8.RuneCountInString(reason) < minResetReasonLen {
		return AdminResetResult{}, domain.ErrReasonRequired
	}
	target, err := uc.Users.FindByID(ctx, in.TargetID)
	if err != nil {
		return AdminResetResult{}, err
	}
	if !target.Active {
		return AdminResetResult{}, domain.ErrUserInactive
	}
	perms, err := uc.Perms.EffectivePermissions(ctx, target.ID)
	if err != nil {
		return AdminResetResult{}, err
	}
	if missing := authz.Covers(in.Actor, perms); len(missing) > 0 {
		return AdminResetResult{}, denyChange(ctx, uc.Audit, in.Actor, audit.UserPasswordReset, target.ID, domain.ErrPrivilegeEscalation, missing)
	}
	roles, err := uc.Roles.RolesOf(ctx, target.ID)
	if err != nil {
		return AdminResetResult{}, err
	}
	temporary, err := uc.temporaryPassword(roles)
	if err != nil {
		return AdminResetResult{}, err
	}
	hash, err := uc.Passwords.Hash(temporary)
	if err != nil {
		return AdminResetResult{}, err
	}

	err = uc.Tx(ctx, func(ctx context.Context) error {
		if err := uc.Users.SetPassword(ctx, target.ID, hash, true); err != nil {
			return err
		}
		if _, err := uc.Sessions.RevokeAllForUser(ctx, target.ID); err != nil {
			return err
		}
		if err := uc.Recovery.InvalidatePending(ctx, target.ID, uc.Now()); err != nil {
			return err
		}
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: audit.UserPasswordReset,
			EntityType: "user", EntityID: target.ID, Outcome: audit.OutcomeSuccess, Reason: reason,
			Before: map[string]any{"must_change_password": target.MustChangePassword}, After: map[string]any{"must_change_password": true},
		})
	})
	if err != nil {
		return AdminResetResult{}, err
	}
	return AdminResetResult{TemporaryPassword: temporary}, nil
}

// temporaryPassword draws random passwords until one meets the policy of the
// target's roles (they all do, in practice: it is 19 characters and random).
func (uc *AdminResetPassword) temporaryPassword(roles []domain.Role) (string, error) {
	for range 5 {
		p, err := domain.NewTemporaryPassword()
		if err != nil {
			return "", err
		}
		if domain.ValidatePassword(p, roles, uc.Denylist) == nil {
			return p, nil
		}
	}
	return "", errors.New("senha temporária: nenhuma senha gerada cumpriu a política")
}
