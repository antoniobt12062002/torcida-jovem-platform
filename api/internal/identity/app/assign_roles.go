package app

import (
	"context"
	"slices"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Ports of AssignRoles.
type (
	AssignUsers interface {
		UserAdmin
		SetRoles(ctx context.Context, userID string, roles []domain.Role) error
	}
	MembershipReader interface {
		Active(ctx context.Context, userID string) (domain.AdminMembership, bool, error)
	}
)

// AssignRoles replaces the administrative roles of a user that already has an
// active administrative membership. ASSOCIADO is the base role and always stays;
// an empty list leaves the membership dormant with only ASSOCIADO.
//
// The actor needs identity:role:assign, cannot change their own roles, and can
// only add or remove roles whose permissions they all hold. It never leaves the
// system without an active holder of identity:admin:grant. Sessions are not
// revoked: permissions are read on every request, so the change applies at once.
type AssignRoles struct {
	Authz   Authorizer
	Users   AssignUsers
	Members MembershipReader
	Roles   RoleReader
	Perms   PermissionReader
	Matrix  domain.Matrix
	Audit   Auditor
	Tx      TxFunc
}

type AssignInput struct {
	Actor    authz.Principal
	TargetID string
	Roles    []domain.Role // administrative roles; ASSOCIADO is implicit
}

func (uc *AssignRoles) Execute(ctx context.Context, in AssignInput) error {
	if uc.Authz == nil || uc.Users == nil || uc.Members == nil || uc.Roles == nil || uc.Perms == nil || uc.Audit == nil || uc.Tx == nil {
		return errNotConfigured
	}
	if err := uc.Authz.Require(ctx, in.Actor, "identity:role:assign"); err != nil {
		return err
	}
	if in.TargetID == in.Actor.UserID {
		return denyRoleChange(ctx, uc.Audit, in.Actor, in.TargetID, domain.ErrSelfChangeForbidden, in.Roles, nil)
	}
	desired := []domain.Role{domain.RoleAssociado}
	for _, r := range in.Roles {
		if !r.Valid() {
			return domain.ErrUnknownRole
		}
		if !slices.Contains(desired, r) {
			desired = append(desired, r)
		}
	}
	target, err := uc.Users.FindByID(ctx, in.TargetID)
	if err != nil {
		return err
	}
	if !target.Active {
		return domain.ErrUserInactive
	}
	if _, has, err := uc.Members.Active(ctx, target.ID); err != nil {
		return err
	} else if !has {
		return domain.ErrAdminMembershipRequired
	}
	current, err := uc.Roles.RolesOf(ctx, target.ID)
	if err != nil {
		return err
	}
	changed := symmetricDifference(current, desired)
	if len(changed) == 0 {
		return nil
	}
	if missing := authz.Covers(in.Actor, uc.Matrix.PermissionsOf(changed)); len(missing) > 0 {
		return denyRoleChange(ctx, uc.Audit, in.Actor, target.ID, domain.ErrPrivilegeEscalation, changed, missing)
	}

	losesGrant := !slices.Contains(uc.Matrix.PermissionsOf(desired), adminGrant)
	return uc.Tx(ctx, func(ctx context.Context) error {
		if err := ensureAdminRemains(ctx, uc.Users, uc.Perms, target.ID, losesGrant); err != nil {
			return err
		}
		before, err := uc.Roles.RolesOf(ctx, target.ID)
		if err != nil {
			return err
		}
		if err := uc.Users.SetRoles(ctx, target.ID, desired); err != nil {
			return err
		}
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: audit.UserRolesSet,
			EntityType: "user", EntityID: target.ID, Outcome: audit.OutcomeSuccess,
			Before: map[string]any{"roles": roleNames(before)}, After: map[string]any{"roles": roleNames(desired)},
		})
	})
}

func symmetricDifference(a, b []domain.Role) []domain.Role {
	var out []domain.Role
	for _, r := range a {
		if !slices.Contains(b, r) {
			out = append(out, r)
		}
	}
	for _, r := range b {
		if !slices.Contains(a, r) {
			out = append(out, r)
		}
	}
	return out
}
