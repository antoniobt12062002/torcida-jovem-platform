// Package identity composes the identity module: the repositories, the use
// cases and their shared dependencies. The HTTP layer (identity/http) and the
// bootstrap command consume the composed Module; nothing outside identity
// reaches into its repositories.
package identity

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

// Deps are the platform pieces the module is built on.
type Deps struct {
	DB         *gorm.DB
	Recorder   *audit.Recorder
	Authorizer *authz.Authorizer
	Matrix     domain.Matrix
	Hasher     *password.Hasher
	Denylist   *password.Denylist
	Sender     email.Sender
	Log        *slog.Logger

	HashKey         []byte
	SessionIdle     time.Duration
	SessionAbsolute time.Duration
	ResetTTL        time.Duration
	BaseURL         string // base of the links sent by e-mail, without trailing slash

	// Now defaults to the wall clock; tests inject a fixed one.
	Now func() time.Time
}

// Module is the composed identity module.
type Module struct {
	Users    *infra.UserRepository
	Roles    *infra.RoleRepository
	Members  *infra.AdminMembershipRepository
	Sessions *infra.SessionRepository

	Login          *app.Authenticator
	Session        *app.SessionService
	CreateUser     *app.CreateUser
	ListUsers      *app.ListUsers
	Promote        *app.PromoteToAdmin
	Revoke         *app.RevokeAdmin
	AssignRoles    *app.AssignRoles
	Activation     *app.UserActivation
	ChangePassword *app.ChangePassword
	RequestReset   *app.RequestPasswordReset
	ResetPassword  *app.ResetPasswordWithToken
	AdminReset     *app.AdminResetPassword

	SessionAbsolute time.Duration
}

// New wires the module. It does no I/O.
func New(d Deps) *Module {
	now := d.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	tx := func(ctx context.Context, fn func(context.Context) error) error { return database.WithTx(ctx, d.DB, fn) }

	users := infra.NewUserRepository(d.DB)
	sessions := infra.NewSessionRepository(d.DB)
	members := infra.NewAdminMembershipRepository(d.DB)
	recovery := infra.NewRecoveryRepository(d.DB)
	attempts := infra.NewAttemptRepository(d.DB)
	roles := infra.NewRoleRepository(d.DB, d.Recorder)

	return &Module{
		Users: users, Roles: roles, Members: members, Sessions: sessions,
		SessionAbsolute: d.SessionAbsolute,

		Login: &app.Authenticator{
			Users: users, Sessions: sessions, Attempts: attempts, Audit: d.Recorder, Passwords: d.Hasher,
			HashKey: d.HashKey, SessionAbsolute: d.SessionAbsolute, Tx: tx, Now: now,
		},
		Session: &app.SessionService{
			Sessions: sessions, Users: users, Roles: roles, Audit: d.Recorder,
			Idle: d.SessionIdle, TouchInterval: time.Minute, Now: now,
		},
		CreateUser: &app.CreateUser{Authz: d.Authorizer, Users: users, Audit: d.Recorder, Hasher: d.Hasher, Denylist: d.Denylist, Tx: tx},
		ListUsers:  &app.ListUsers{Authz: d.Authorizer, Users: users},
		Promote: &app.PromoteToAdmin{
			Authz: d.Authorizer, Users: users, Members: members, Roles: roles, Sessions: sessions, Matrix: d.Matrix,
			Audit: d.Recorder, Tx: tx, Now: now,
		},
		Revoke: &app.RevokeAdmin{
			Authz: d.Authorizer, Users: users, Members: members, Perms: roles, Sessions: sessions, Audit: d.Recorder, Tx: tx, Now: now,
		},
		AssignRoles: &app.AssignRoles{
			Authz: d.Authorizer, Users: users, Members: members, Roles: roles, Perms: roles, Matrix: d.Matrix, Audit: d.Recorder, Tx: tx,
		},
		Activation: &app.UserActivation{
			Authz: d.Authorizer, Users: users, Perms: roles, Sessions: sessions, Recovery: recovery, Audit: d.Recorder, Tx: tx, Now: now,
		},
		ChangePassword: &app.ChangePassword{
			Users: users, Roles: roles, Throttle: recovery, Sessions: sessions, Passwords: d.Hasher, Denylist: d.Denylist,
			Audit: d.Recorder, Tx: tx, Now: now,
		},
		RequestReset: &app.RequestPasswordReset{
			Users: users, Recovery: recovery, Sender: d.Sender, Audit: d.Recorder, Tx: tx, Now: now,
			HashKey: d.HashKey, TTL: d.ResetTTL, BaseURL: d.BaseURL, Log: d.Log,
		},
		ResetPassword: &app.ResetPasswordWithToken{
			Recovery: recovery, Users: users, Roles: roles, Sessions: sessions, Passwords: d.Hasher, Denylist: d.Denylist,
			Audit: d.Recorder, Tx: tx, Now: now,
		},
		AdminReset: &app.AdminResetPassword{
			Authz: d.Authorizer, Users: users, Perms: roles, Roles: roles, Sessions: sessions, Recovery: recovery,
			Passwords: d.Hasher, Denylist: d.Denylist, Audit: d.Recorder, Tx: tx, Now: now,
		},
	}
}
