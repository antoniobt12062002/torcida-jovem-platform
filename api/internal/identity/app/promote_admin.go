package app

import (
	"context"
	"slices"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Ports of the promotion and withdrawal of administrative access.
type (
	PromoteUsers interface {
		FindByID(ctx context.Context, id string) (domain.User, error)
		SetRoles(ctx context.Context, userID string, roles []domain.Role) error
		SetMustChangePassword(ctx context.Context, id string, mustChange bool) error
	}
	MembershipGranter interface {
		Grant(ctx context.Context, m domain.AdminMembership) (domain.AdminMembership, error)
	}
)

// PromoteToAdmin gives a user administrative access: the membership (why) and
// the roles (what). It is the only way to grant administrative roles.
//
// In one transaction it creates the membership with the reason, assigns the
// roles (ASSOCIADO always stays), forces a password change, revokes every
// session of the target and audits `admin.promote`. The actor needs
// identity:admin:grant, cannot promote themselves, and cannot grant roles whose
// permissions they do not hold (authz.Covers); refusals are audited as security
// events (`role.change_denied`).
type PromoteToAdmin struct {
	Authz    Authorizer
	Users    PromoteUsers
	Members  MembershipGranter
	Roles    RoleReader
	Sessions SessionRevoker
	Matrix   domain.Matrix
	Audit    Auditor
	Tx       TxFunc
	Now      Clock
}

type PromoteInput struct {
	Actor    authz.Principal
	TargetID string
	Roles    []domain.Role // administrative roles; ASSOCIADO is implicit
	Reason   string
}

func (uc *PromoteToAdmin) Execute(ctx context.Context, in PromoteInput) error {
	if uc.Authz == nil || uc.Users == nil || uc.Members == nil || uc.Roles == nil || uc.Sessions == nil || uc.Audit == nil || uc.Tx == nil || uc.Now == nil {
		return errNotConfigured
	}
	if err := uc.Authz.Require(ctx, in.Actor, "identity:admin:grant"); err != nil {
		return err
	}
	if in.TargetID == in.Actor.UserID {
		return denyRoleChange(ctx, uc.Audit, in.Actor, in.TargetID, domain.ErrSelfChangeForbidden, in.Roles, nil)
	}
	requested, err := administrativeRoles(in.Roles)
	if err != nil {
		return err
	}
	target, err := uc.Users.FindByID(ctx, in.TargetID)
	if err != nil {
		return err
	}
	if !target.Active {
		return domain.ErrUserInactive
	}
	membership, err := domain.NewAdminMembership(target.ID, in.Reason, &in.Actor.UserID, uc.Now())
	if err != nil {
		return err
	}
	grantedRoles := append([]domain.Role{domain.RoleAssociado}, requested...)
	if missing := authz.Covers(in.Actor, uc.Matrix.PermissionsOf(grantedRoles)); len(missing) > 0 {
		return denyRoleChange(ctx, uc.Audit, in.Actor, target.ID, domain.ErrPrivilegeEscalation, requested, missing)
	}

	return uc.Tx(ctx, func(ctx context.Context) error {
		before, err := uc.Roles.RolesOf(ctx, target.ID)
		if err != nil {
			return err
		}
		if _, err := uc.Members.Grant(ctx, membership); err != nil {
			return err
		}
		after := unionRoles(before, grantedRoles)
		if err := uc.Users.SetRoles(ctx, target.ID, after); err != nil {
			return err
		}
		if err := uc.Users.SetMustChangePassword(ctx, target.ID, true); err != nil {
			return err
		}
		if _, err := uc.Sessions.RevokeAllForUser(ctx, target.ID); err != nil {
			return err
		}
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: audit.AdminPromote,
			EntityType: "user", EntityID: target.ID, Outcome: audit.OutcomeSuccess, Reason: membership.Reason,
			Before: map[string]any{"roles": roleNames(before), "admin_membership": false},
			After:  map[string]any{"roles": roleNames(after), "admin_membership": true, "must_change_password": true},
		})
	})
}

// administrativeRoles validates the requested roles, drops repetitions and
// ASSOCIADO, and requires at least one administrative role.
func administrativeRoles(roles []domain.Role) ([]domain.Role, error) {
	var out []domain.Role
	for _, r := range roles {
		if !r.Valid() {
			return nil, domain.ErrUnknownRole
		}
		if r != domain.RoleAssociado && !slices.Contains(out, r) {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return nil, domain.ErrAdminRoleRequired
	}
	return out, nil
}

func unionRoles(a, b []domain.Role) []domain.Role {
	out := slices.Clone(a)
	for _, r := range b {
		if !slices.Contains(out, r) {
			out = append(out, r)
		}
	}
	slices.Sort(out)
	return out
}

func roleNames(roles []domain.Role) []any {
	names := make([]string, len(roles))
	for i, r := range roles {
		names[i] = string(r)
	}
	slices.Sort(names)
	out := make([]any, len(names))
	for i, n := range names {
		out[i] = n
	}
	return out
}

// denyRoleChange audits a refused change of roles or membership as a security
// event (`role.change_denied`) and returns the refusal.
func denyRoleChange(ctx context.Context, a Auditor, actor authz.Principal, targetID string, refusal error, requested []domain.Role, missing []authz.Permission) error {
	c := map[string]any{"denial": refusal.Error(), "requested_roles": roleNames(requested)}
	if len(missing) > 0 {
		names := make([]any, len(missing))
		for i, p := range missing {
			names[i] = string(p)
		}
		c["missing_permissions"] = names
	}
	a.RecordSecurity(ctx, audit.Entry{
		ActorType: audit.ActorUser, ActorID: actor.UserID, Action: audit.RoleChangeDenied,
		EntityType: "user", EntityID: targetID, Outcome: audit.OutcomeDenied, Context: c,
	})
	return refusal
}
