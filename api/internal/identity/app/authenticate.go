package app

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
)

// Errors of the login use case. The messages are the API error codes.
var ErrInvalidCredentials = errors.New("invalid_credentials")

// LockedError is returned while the lockout is in force; RetryAfter feeds the
// Retry-After header of the 429 response.
type LockedError struct{ RetryAfter time.Duration }

func (e *LockedError) Error() string { return "login_blocked" }

// Ports of the login use case, satisfied by the infra repositories, the audit
// recorder and the password hasher.
type (
	UserFinder interface {
		FindByEmail(ctx context.Context, email string) (domain.User, error)
	}
	SessionStore interface {
		Create(ctx context.Context, userID string, tokenHash []byte, csrfToken string, at, expiresAt time.Time) (domain.Session, error)
		FindByToken(ctx context.Context, token string) (domain.Session, error)
		Revoke(ctx context.Context, sessionID string) error
	}
	AttemptStore interface {
		RecordFailure(ctx context.Context, emailHash []byte, at time.Time) error
		RecordSuccess(ctx context.Context, emailHash []byte) error
		Blocked(ctx context.Context, emailHash []byte, now time.Time) (until time.Time, blocked bool, err error)
	}
	SecurityAuditor interface {
		RecordSecurity(ctx context.Context, e audit.Entry)
	}
	PasswordChecker interface {
		Verify(plain, encoded string) (bool, error)
		BurnCycles(plain string)
	}
)

// Authenticator is the login use case.
type Authenticator struct {
	Users     UserFinder
	Sessions  SessionStore
	Attempts  AttemptStore
	Audit     SecurityAuditor
	Passwords PasswordChecker

	HashKey         []byte
	SessionAbsolute time.Duration
	// Tx runs fn in one unit of work (database.WithTx).
	Tx  func(ctx context.Context, fn func(ctx context.Context) error) error
	Now func() time.Time
}

// LoginInput is what the client sent. PresentedToken is the session cookie, if any.
type LoginInput struct {
	Email          string
	Password       string
	PresentedToken string
}

// LoginResult is what the handler needs to set the cookie and answer.
type LoginResult struct {
	Token              string
	Session            domain.Session
	UserID             string
	MustChangePassword bool
}

func (a *Authenticator) ready() bool {
	return a.Users != nil && a.Sessions != nil && a.Attempts != nil && a.Audit != nil && a.Passwords != nil &&
		len(a.HashKey) > 0 && a.SessionAbsolute > 0 && a.Tx != nil && a.Now != nil
}

// Login authenticates by e-mail and password.
//
//   - Unknown e-mail, wrong password and inactive user all return the same
//     ErrInvalidCredentials, and do the same amount of work.
//   - Five consecutive failures for the same e-mail within 15 minutes lock it for
//     15 minutes (LockedError). Attempts during the lock are not counted, so the
//     lock never extends itself and is never permanent.
//   - A success clears the failure count, revokes the session presented in the
//     cookie (if any) and issues a new token; other sessions stay valid.
//   - Events go to the audit trail as security events: a failed write is an
//     incident in the log and never changes the response.
func (a *Authenticator) Login(ctx context.Context, in LoginInput) (LoginResult, error) {
	if !a.ready() {
		return LoginResult{}, errors.New("login: Authenticator mal configurado")
	}
	now := a.Now()
	emailHash, err := domain.HashIdentifier(a.HashKey, in.Email)
	if err != nil {
		return LoginResult{}, err
	}
	hashHex := hex.EncodeToString(emailHash)

	until, blocked, err := a.Attempts.Blocked(ctx, emailHash, now)
	if err != nil {
		return LoginResult{}, err
	}
	if blocked {
		a.Audit.RecordSecurity(ctx, audit.Entry{
			ActorType: audit.ActorAnonymous, Action: audit.AuthLoginBlocked, EntityType: "login", EntityID: hashHex,
			Outcome: audit.OutcomeDenied, Context: map[string]any{"email_hash": hashHex},
		})
		return LoginResult{}, &LockedError{RetryAfter: until.Sub(now)}
	}

	user, category, err := a.check(ctx, in)
	if err != nil {
		return LoginResult{}, err
	}
	if category != "" {
		return LoginResult{}, a.fail(ctx, emailHash, hashHex, user, category, now)
	}
	return a.succeed(ctx, in, emailHash, user, now)
}

// check verifies the credentials. It returns the failure category, or "" on
// success. The password is always verified against a real or a dummy hash.
func (a *Authenticator) check(ctx context.Context, in LoginInput) (user domain.User, category string, err error) {
	email, nerr := domain.NormalizeEmail(in.Email)
	if nerr != nil {
		a.Passwords.BurnCycles(in.Password)
		return domain.User{}, "unknown_user", nil
	}
	user, err = a.Users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		a.Passwords.BurnCycles(in.Password)
		return domain.User{}, "unknown_user", nil
	}
	if err != nil {
		return domain.User{}, "", err
	}
	ok, verr := a.Passwords.Verify(in.Password, user.PasswordHash)
	switch {
	case verr != nil || !ok:
		return user, "wrong_password", nil
	case !user.Active:
		return user, "inactive", nil
	}
	return user, "", nil
}

func (a *Authenticator) fail(ctx context.Context, emailHash []byte, hashHex string, user domain.User, category string, now time.Time) error {
	if err := a.Attempts.RecordFailure(ctx, emailHash, now); err != nil {
		return err
	}
	entityType, entityID := "login", hashHex
	if user.ID != "" {
		entityType, entityID = "user", user.ID
	}
	a.Audit.RecordSecurity(ctx, audit.Entry{
		ActorType: audit.ActorAnonymous, Action: audit.AuthLoginFailed, EntityType: entityType, EntityID: entityID,
		Outcome: audit.OutcomeFailure, Context: map[string]any{"category": category, "email_hash": hashHex},
	})
	return ErrInvalidCredentials
}

func (a *Authenticator) succeed(ctx context.Context, in LoginInput, emailHash []byte, user domain.User, now time.Time) (LoginResult, error) {
	token, tokenHash, err := domain.NewSessionToken()
	if err != nil {
		return LoginResult{}, err
	}
	csrf, err := domain.NewCSRFToken()
	if err != nil {
		return LoginResult{}, err
	}
	var session domain.Session
	err = a.Tx(ctx, func(ctx context.Context) error {
		if err := a.Attempts.RecordSuccess(ctx, emailHash); err != nil {
			return err
		}
		if in.PresentedToken != "" {
			old, ferr := a.Sessions.FindByToken(ctx, in.PresentedToken)
			switch {
			case ferr == nil && old.RevokedAt == nil:
				if err := a.Sessions.Revoke(ctx, old.ID); err != nil {
					return err
				}
			case ferr != nil && !errors.Is(ferr, domain.ErrSessionNotFound):
				return ferr
			}
		}
		session, err = a.Sessions.Create(ctx, user.ID, tokenHash, csrf, now, now.Add(a.SessionAbsolute))
		return err
	})
	if err != nil {
		return LoginResult{}, fmt.Errorf("login: %w", err)
	}
	a.Audit.RecordSecurity(ctx, audit.Entry{
		ActorType: audit.ActorUser, ActorID: user.ID, Action: audit.AuthLogin, EntityType: "user", EntityID: user.ID,
		Outcome: audit.OutcomeSuccess,
	})
	return LoginResult{Token: token, Session: session, UserID: user.ID, MustChangePassword: user.MustChangePassword}, nil
}
