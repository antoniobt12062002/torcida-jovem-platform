package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"time"
)

// ErrSessionNotFound is returned when no session matches the token.
var ErrSessionNotFound = errors.New("session_not_found")

// Session is a server-side session. The token lives only in the cookie; the
// database keeps its SHA-256.
type Session struct {
	ID         string
	UserID     string
	CSRFToken  string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time // absolute limit: creation + 8 hours (configurable)
	RevokedAt  *time.Time
}

// Expired reports whether the session no longer counts at the given instant:
// revoked, past the absolute limit, or idle for more than the idle limit.
func (s Session) Expired(at time.Time, idle time.Duration) bool {
	return s.RevokedAt != nil || at.After(s.ExpiresAt) || at.Sub(s.LastSeenAt) > idle
}

// NewSessionToken returns a token of 256 random bits in base64url and its
// SHA-256, the only form that is stored.
func NewSessionToken() (token string, hash []byte, err error) {
	token, err = randomToken()
	if err != nil {
		return "", nil, err
	}
	return token, HashSessionToken(token), nil
}

// HashSessionToken is the SHA-256 of the token as sent by the client.
func HashSessionToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// NewCSRFToken returns a token of 256 random bits in base64url. It is stored in
// clear text with the session, because the client needs it back in `me`.
func NewCSRFToken() (string, error) { return randomToken() }

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("token aleatório: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
