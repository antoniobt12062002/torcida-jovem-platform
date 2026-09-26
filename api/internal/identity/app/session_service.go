package app

import (
	"context"
	"errors"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// Errors of the session service. The messages are the API error codes.
var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrSessionExpired  = errors.New("session_expired")
)

// Ports of the session service.
type (
	SessionReader interface {
		FindByToken(ctx context.Context, token string) (domain.Session, error)
		Touch(ctx context.Context, sessionID string, at time.Time, minInterval time.Duration) (bool, error)
		Revoke(ctx context.Context, sessionID string) error
	}
	UserByID interface {
		FindByID(ctx context.Context, id string) (domain.User, error)
	}
	PrincipalRoles interface {
		EffectivePermissions(ctx context.Context, userID string) ([]authz.Permission, error)
		RolesOf(ctx context.Context, userID string) ([]domain.Role, error)
	}
)

// SessionService validates sessions on every request and handles logout.
type SessionService struct {
	Sessions SessionReader
	Users    UserByID
	Roles    PrincipalRoles
	Audit    SecurityAuditor

	Idle          time.Duration // idle limit (60 minutes)
	TouchInterval time.Duration // minimum time between last_seen_at writes (1 minute)
	Now           func() time.Time
}

// Authenticated is what a valid session gives the request.
type Authenticated struct {
	Principal          authz.Principal
	User               domain.User
	SessionID          string
	CSRFToken          string
	MustChangePassword bool
}

func (s *SessionService) ready() bool {
	return s.Sessions != nil && s.Users != nil && s.Roles != nil && s.Audit != nil && s.Idle > 0 && s.TouchInterval > 0 && s.Now != nil
}

// Validate checks the session token and returns who is acting:
//
//   - an unknown, malformed, revoked or empty token, or an inactive user, is
//     ErrUnauthenticated (an inactive user invalidates every session of theirs);
//   - a session idle for more than the idle limit, or older than its absolute
//     limit, is ErrSessionExpired and is not renewed;
//   - otherwise it returns the Principal with the effective permissions read now
//     (no cache, so a role change applies on the next request), and refreshes
//     last_seen_at at most once per TouchInterval.
func (s *SessionService) Validate(ctx context.Context, token string) (Authenticated, error) {
	if !s.ready() {
		return Authenticated{}, errors.New("sessão: SessionService mal configurado")
	}
	if token == "" || len(token) > 512 {
		return Authenticated{}, ErrUnauthenticated
	}
	session, err := s.Sessions.FindByToken(ctx, token)
	if errors.Is(err, domain.ErrSessionNotFound) {
		return Authenticated{}, ErrUnauthenticated
	}
	if err != nil {
		return Authenticated{}, err
	}
	now := s.Now()
	if session.RevokedAt != nil {
		return Authenticated{}, ErrUnauthenticated
	}
	if session.Expired(now, s.Idle) {
		return Authenticated{}, ErrSessionExpired
	}
	user, err := s.Users.FindByID(ctx, session.UserID)
	if errors.Is(err, domain.ErrUserNotFound) || (err == nil && !user.Active) {
		return Authenticated{}, ErrUnauthenticated
	}
	if err != nil {
		return Authenticated{}, err
	}
	perms, err := s.Roles.EffectivePermissions(ctx, user.ID)
	if err != nil {
		return Authenticated{}, err
	}
	roles, err := s.Roles.RolesOf(ctx, user.ID)
	if err != nil {
		return Authenticated{}, err
	}
	// A failure to refresh the last-seen time must not turn a valid request into an error.
	_, _ = s.Sessions.Touch(ctx, session.ID, now, s.TouchInterval)

	principal := authz.Principal{UserID: user.ID, Roles: make([]string, len(roles)), Permissions: make(map[authz.Permission]struct{}, len(perms))}
	for i, r := range roles {
		principal.Roles[i] = string(r)
	}
	for _, p := range perms {
		principal.Permissions[p] = struct{}{}
	}
	return Authenticated{Principal: principal, User: user, SessionID: session.ID, CSRFToken: session.CSRFToken, MustChangePassword: user.MustChangePassword}, nil
}

// Logout revokes the session of the token and records `auth.logout`. It is
// idempotent: an unknown or already revoked token does nothing.
func (s *SessionService) Logout(ctx context.Context, token string) error {
	if !s.ready() {
		return errors.New("sessão: SessionService mal configurado")
	}
	if token == "" || len(token) > 512 {
		return nil
	}
	session, err := s.Sessions.FindByToken(ctx, token)
	if errors.Is(err, domain.ErrSessionNotFound) || (err == nil && session.RevokedAt != nil) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := s.Sessions.Revoke(ctx, session.ID); err != nil {
		return err
	}
	s.Audit.RecordSecurity(ctx, audit.Entry{
		ActorType: audit.ActorUser, ActorID: session.UserID, Action: audit.AuthLogout, EntityType: "user", EntityID: session.UserID,
		Outcome: audit.OutcomeSuccess,
	})
	return nil
}
