package app

import (
	"context"
	"slices"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

const adminGrant authz.Permission = "identity:admin:grant"

// Ports of the activation use cases.
type (
	UserAdmin interface {
		FindByID(ctx context.Context, id string) (domain.User, error)
		SetActive(ctx context.Context, id string, active bool) error
		LockAdminSet(ctx context.Context) error
		ActiveHoldersOf(ctx context.Context, perm authz.Permission, exclude string) (int, error)
	}
	PermissionReader interface {
		EffectivePermissions(ctx context.Context, userID string) ([]authz.Permission, error)
	}
	SessionRevoker interface {
		RevokeAllForUser(ctx context.Context, userID string) (int64, error)
	}
	TokenInvalidator interface {
		InvalidatePending(ctx context.Context, userID string, at time.Time) error
	}
)

// UserActivation deactivates and reactivates users.
//
// Both need identity:user:update and follow the no-escalation rule: the actor
// must hold every effective permission of the target, and nobody changes their
// own account. Deactivating revokes the user's sessions and pending recovery
// tokens, and never leaves the system without an active holder of
// identity:admin:grant. Reactivating restores nothing but the active flag: the
// administrative roles and any closed membership stay as they are.
type UserActivation struct {
	Authz    Authorizer
	Users    UserAdmin
	Perms    PermissionReader
	Sessions SessionRevoker
	Recovery TokenInvalidator
	Audit    Auditor
	Tx       TxFunc
	Now      Clock
}

type ActivationInput struct {
	Actor    authz.Principal
	TargetID string
}

func (uc *UserActivation) Deactivate(ctx context.Context, in ActivationInput) error {
	return uc.change(ctx, in, false)
}

func (uc *UserActivation) Reactivate(ctx context.Context, in ActivationInput) error {
	return uc.change(ctx, in, true)
}

func (uc *UserActivation) change(ctx context.Context, in ActivationInput, active bool) error {
	if uc.Authz == nil || uc.Users == nil || uc.Perms == nil || uc.Sessions == nil || uc.Recovery == nil || uc.Audit == nil || uc.Tx == nil || uc.Now == nil {
		return errNotConfigured
	}
	if err := uc.Authz.Require(ctx, in.Actor, "identity:user:update"); err != nil {
		return err
	}
	action := audit.UserDeactivate
	if active {
		action = audit.UserReactivate
	}
	if in.TargetID == in.Actor.UserID {
		return denyChange(ctx, uc.Audit, in.Actor, action, in.TargetID, domain.ErrSelfChangeForbidden, nil)
	}
	target, err := uc.Users.FindByID(ctx, in.TargetID)
	if err != nil {
		return err
	}
	perms, err := uc.Perms.EffectivePermissions(ctx, target.ID)
	if err != nil {
		return err
	}
	if missing := authz.Covers(in.Actor, perms); len(missing) > 0 {
		return denyChange(ctx, uc.Audit, in.Actor, action, target.ID, domain.ErrPrivilegeEscalation, missing)
	}
	if target.Active == active {
		return nil
	}
	return uc.Tx(ctx, func(ctx context.Context) error {
		if !active {
			if err := ensureAdminRemains(ctx, uc.Users, uc.Perms, target.ID, true); err != nil {
				return err
			}
		}
		if err := uc.Users.SetActive(ctx, target.ID, active); err != nil {
			return err
		}
		if !active {
			if _, err := uc.Sessions.RevokeAllForUser(ctx, target.ID); err != nil {
				return err
			}
			if err := uc.Recovery.InvalidatePending(ctx, target.ID, uc.Now()); err != nil {
				return err
			}
		}
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: action,
			EntityType: "user", EntityID: target.ID, Outcome: audit.OutcomeSuccess,
			Before: map[string]any{"active": !active}, After: map[string]any{"active": active},
		})
	})
}

// ensureAdminRemains takes the admin-set lock and, when the target currently
// holds identity:admin:grant and is about to lose it (losesGrant), refuses the
// change if no other active user holds it (domain.ErrLastAdmin). The permissions
// are read after the lock, so concurrent changes see each other.
func ensureAdminRemains(ctx context.Context, users UserAdmin, perms PermissionReader, targetID string, losesGrant bool) error {
	if err := users.LockAdminSet(ctx); err != nil {
		return err
	}
	if !losesGrant {
		return nil
	}
	held, err := perms.EffectivePermissions(ctx, targetID)
	if err != nil {
		return err
	}
	if !slices.Contains(held, adminGrant) {
		return nil
	}
	others, err := users.ActiveHoldersOf(ctx, adminGrant, targetID)
	if err != nil {
		return err
	}
	if others == 0 {
		return domain.ErrLastAdmin
	}
	return nil
}

// denyChange audits a refused administrative action as a security event (its own
// transaction; a write failure is only logged) and returns the refusal.
func denyChange(ctx context.Context, a Auditor, actor authz.Principal, action audit.Action, targetID string, refusal error, missing []authz.Permission) error {
	c := map[string]any{"denial": refusal.Error()}
	if len(missing) > 0 {
		names := make([]any, len(missing))
		for i, p := range missing {
			names[i] = string(p)
		}
		c["missing_permissions"] = names
	}
	a.RecordSecurity(ctx, audit.Entry{
		ActorType: audit.ActorUser, ActorID: actor.UserID, Action: action,
		EntityType: "user", EntityID: targetID, Outcome: audit.OutcomeDenied, Context: c,
	})
	return refusal
}
