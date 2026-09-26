package app

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

// UserWriter is what CreateUser needs of the user repository.
type UserWriter interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	SetRoles(ctx context.Context, userID string, roles []domain.Role) error
}

// CreateUser creates an account. Every account is born with the role ASSOCIADO
// only, and with must_change_password: the initial password chosen by the
// creator must not stay as the definitive credential. Administrative access
// exists only through PromoteToAdmin.
type CreateUser struct {
	Authz    Authorizer
	Users    UserWriter
	Audit    Auditor
	Hasher   PasswordHasher
	Denylist *password.Denylist
	Tx       TxFunc
}

type CreateUserInput struct {
	Actor    authz.Principal
	Email    string
	Name     string
	Password string
}

const maxNameLen = 120

// Execute checks the permission first, then validates the input, hashes the
// password and, in one transaction, creates the user and audits `user.create`.
func (uc *CreateUser) Execute(ctx context.Context, in CreateUserInput) (domain.User, error) {
	if uc.Authz == nil || uc.Users == nil || uc.Audit == nil || uc.Hasher == nil || uc.Denylist == nil || uc.Tx == nil {
		return domain.User{}, errNotConfigured
	}
	if err := uc.Authz.Require(ctx, in.Actor, "identity:user:create"); err != nil {
		return domain.User{}, err
	}
	email, err := domain.NormalizeEmail(in.Email)
	if err != nil {
		return domain.User{}, domain.ErrInvalidEmail
	}
	name := strings.TrimSpace(in.Name)
	if name == "" || utf8.RuneCountInString(name) > maxNameLen || strings.ContainsFunc(name, unicode.IsControl) {
		return domain.User{}, domain.ErrInvalidName
	}
	roles := []domain.Role{domain.RoleAssociado}
	if err := domain.ValidatePassword(in.Password, roles, uc.Denylist); err != nil {
		return domain.User{}, err
	}
	hash, err := uc.Hasher.Hash(in.Password)
	if err != nil {
		return domain.User{}, err
	}

	var created domain.User
	err = uc.Tx(ctx, func(ctx context.Context) error {
		u, err := uc.Users.Create(ctx, domain.User{Email: email, Name: name, PasswordHash: hash, Active: true, MustChangePassword: true})
		if err != nil {
			return err
		}
		if err := uc.Users.SetRoles(ctx, u.ID, roles); err != nil {
			return err
		}
		created = u
		return uc.Audit.Record(ctx, audit.Entry{
			ActorType: audit.ActorUser, ActorID: in.Actor.UserID, Action: audit.UserCreate,
			EntityType: "user", EntityID: u.ID, Outcome: audit.OutcomeSuccess,
			After: map[string]any{"email": u.Email, "name": u.Name, "active": true, "must_change_password": true, "roles": []any{string(domain.RoleAssociado)}},
		})
	})
	if err != nil {
		return domain.User{}, err
	}
	return created, nil
}
