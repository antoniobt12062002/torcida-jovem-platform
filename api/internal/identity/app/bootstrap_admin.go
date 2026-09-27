package app

import (
	"context"
	"errors"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

// ErrBootstrapAlreadyDone is returned when an active administrative membership
// already exists: the bootstrap only creates the first administrator.
var ErrBootstrapAlreadyDone = errors.New("bootstrap_already_done")

// Ports of the bootstrap, satisfied by the infra repositories.
type (
	BootstrapUsers interface {
		Create(ctx context.Context, u domain.User) (domain.User, error)
		SetRoles(ctx context.Context, userID string, roles []domain.Role) error
		LockAdminSet(ctx context.Context) error
	}
	BootstrapMembers interface {
		Grant(ctx context.Context, m domain.AdminMembership) (domain.AdminMembership, error)
		AnyActive(ctx context.Context) (bool, error)
	}
)

// BootstrapAdmin creates the first administrator without an existing account
// to authorize it. It runs only while no administrative membership is active
// (checked under the admin-set lock, so two runs together cannot create two),
// applies the administrator password policy, and in one transaction creates
// the active user with must_change_password, its roles, the membership with the
// fixed reason `bootstrap-admin` and no grantor, and the `user.bootstrap` audit
// entry, which has no actor and never carries the password or its hash. A new
// user has no sessions, so there is none to revoke.
type BootstrapAdmin struct {
	Users    BootstrapUsers
	Members  BootstrapMembers
	Audit    Auditor
	Hasher   PasswordHasher
	Denylist *password.Denylist
	Tx       TxFunc
	Now      Clock
}

type BootstrapInput struct {
	Email    string
	Name     string
	Password string
	Role     domain.Role // empty means ADMIN_SISTEMA; PRESIDENTE is the only other
}

func (uc *BootstrapAdmin) Execute(ctx context.Context, in BootstrapInput) (domain.User, error) {
	if uc.Users == nil || uc.Members == nil || uc.Audit == nil || uc.Hasher == nil || uc.Denylist == nil || uc.Tx == nil || uc.Now == nil {
		return domain.User{}, errNotConfigured
	}
	role := in.Role
	if role == "" {
		role = domain.RoleAdminSistema
	}
	if role != domain.RoleAdminSistema && role != domain.RolePresidente {
		return domain.User{}, domain.ErrUnknownRole
	}
	email, err := domain.NormalizeEmail(in.Email)
	if err != nil {
		return domain.User{}, domain.ErrInvalidEmail
	}
	name, err := validName(in.Name)
	if err != nil {
		return domain.User{}, err
	}
	roles := []domain.Role{domain.RoleAssociado, role}
	if err := domain.ValidatePassword(in.Password, roles, uc.Denylist); err != nil {
		return domain.User{}, err
	}
	hash, err := uc.Hasher.Hash(in.Password)
	if err != nil {
		return domain.User{}, err
	}

	var created domain.User
	err = uc.Tx(ctx, func(ctx context.Context) error {
		if err := uc.Users.LockAdminSet(ctx); err != nil {
			return err
		}
		exists, err := uc.Members.AnyActive(ctx)
		if err != nil {
			return err
		}
		if exists {
			return ErrBootstrapAlreadyDone
		}
		u, err := uc.Users.Create(ctx, domain.User{Email: email, Name: name, PasswordHash: hash, Active: true, MustChangePassword: true})
		if err != nil {
			return err
		}
		if err := uc.Users.SetRoles(ctx, u.ID, roles); err != nil {
			return err
		}
		m, err := domain.NewAdminMembership(u.ID, domain.BootstrapReason, nil, uc.Now())
		if err != nil {
			return err
		}
		if _, err := uc.Members.Grant(ctx, m); err != nil {
			return err
		}
		created = u
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorSystem, Action: audit.UserBootstrap, EntityType: "user", EntityID: u.ID,
			Outcome: audit.OutcomeSuccess, Reason: domain.BootstrapReason,
			After: map[string]any{"email": u.Email, "role": string(role), "membership_reason": domain.BootstrapReason},
		})
	})
	if err != nil {
		return domain.User{}, err
	}
	return created, nil
}
