package app

import (
	"context"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Ports of RevokeAdmin.
type (
	RevokeUsers interface {
		UserAdmin
		SetRoles(ctx context.Context, userID string, roles []domain.Role) error
	}
	MembershipRevoker interface {
		Active(ctx context.Context, userID string) (domain.AdminMembership, bool, error)
		Revoke(ctx context.Context, userID, by, reason string, at time.Time) (domain.AdminMembership, error)
	}
)

// RevokeAdmin withdraws a user's administrative access: it closes the
// membership (kept as history), leaves only the role ASSOCIADO, revokes the
// user's sessions and audits `admin.revoke`, in one transaction.
//
// The actor needs identity:admin:revoke, cannot act on themselves, and must hold
// every effective permission of the target (no escalation). It never leaves the
// system without an active holder of identity:admin:grant.
type RevokeAdmin struct {
	Authz    Authorizer
	Users    RevokeUsers
	Members  MembershipRevoker
	Perms    PermissionReader
	Sessions SessionRevoker
	Audit    Auditor
	Tx       TxFunc
	Now      Clock
}

type RevokeInput struct {
	Actor    authz.Principal
	TargetID string
	Reason   string
}

func (uc *RevokeAdmin) Execute(ctx context.Context, in RevokeInput) error {
	if uc.Authz == nil || uc.Users == nil || uc.Members == nil || uc.Perms == nil || uc.Sessions == nil || uc.Audit == nil || uc.Tx == nil || uc.Now == nil {
		return errNotConfigured
	}
	if err := uc.Authz.Require(ctx, in.Actor, "identity:admin:revoke"); err != nil {
		return err
	}
	if in.TargetID == in.Actor.UserID {
		return denyRoleChange(ctx, uc.Audit, in.Actor, in.TargetID, domain.ErrSelfChangeForbidden, nil, nil)
	}
	target, err := uc.Users.FindByID(ctx, in.TargetID)
	if err != nil {
		return err
	}
	if _, active, err := uc.Members.Active(ctx, target.ID); err != nil {
		return err
	} else if !active {
		return domain.ErrNotAdmin
	}
	if reasonBlank(in.Reason) {
		return domain.ErrReasonRequired
	}
	perms, err := uc.Perms.EffectivePermissions(ctx, target.ID)
	if err != nil {
		return err
	}
	if missing := authz.Covers(in.Actor, perms); len(missing) > 0 {
		return denyRoleChange(ctx, uc.Audit, in.Actor, target.ID, domain.ErrPrivilegeEscalation, nil, missing)
	}

	return uc.Tx(ctx, func(ctx context.Context) error {
		if err := ensureAdminRemains(ctx, uc.Users, uc.Perms, target.ID, true); err != nil {
			return err
		}
		closed, err := uc.Members.Revoke(ctx, target.ID, in.Actor.UserID, in.Reason, uc.Now())
		if err != nil {
			return err
		}
		if err := uc.Users.SetRoles(ctx, target.ID, []domain.Role{domain.RoleAssociado}); err != nil {
			return err
		}
		if _, err := uc.Sessions.RevokeAllForUser(ctx, target.ID); err != nil {
			return err
		}
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: audit.AdminRevoke,
			EntityType: "user", EntityID: target.ID, Outcome: audit.OutcomeSuccess, Reason: *closed.RevokeReason,
			Before: map[string]any{"admin_membership": true, "permissions": permNames(perms)},
			After:  map[string]any{"admin_membership": false, "roles": []any{string(domain.RoleAssociado)}},
		})
	})
}

func permNames(perms []authz.Permission) []any {
	out := make([]any, len(perms))
	for i, p := range perms {
		out[i] = string(p)
	}
	return out
}
