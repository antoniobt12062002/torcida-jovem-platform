package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

// User is an account that authenticates. It does not reference the associado:
// that link belongs to the associados module (associados.associados.user_id).
type User struct {
	ID                 string
	Email              string // normalized
	Name               string
	PasswordHash       string
	Active             bool
	MustChangePassword bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Errors of the user repository; their messages are the API error codes.
var (
	ErrEmailTaken   = errors.New("email_taken")
	ErrUserNotFound = errors.New("user_not_found")

	// ErrInvalidResetToken covers an unknown, expired and already used recovery
	// token alike, so the response never says which.
	ErrInvalidResetToken = errors.New("invalid_reset_token")
)

const maxEmailLen = 254

// NormalizeEmail trims spaces, lowercases and checks the basic form: a single
// bare address (no display name), up to 254 characters, with a domain that has
// a dot. Internationalized domains and aliases (`+tag`) are not normalized.
func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || len(email) > maxEmailLen || strings.ContainsAny(email, "\r\n\x00 \t") {
		return "", errors.New("e-mail inválido")
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", errors.New("e-mail inválido")
	}
	at := strings.LastIndexByte(email, '@')
	if !strings.Contains(email[at+1:], ".") {
		return "", errors.New("e-mail inválido")
	}
	return email, nil
}

// HashIdentifier is the HMAC-SHA256, under the secret key, of what a client
// typed as an e-mail, trimmed and lowercased but not validated. Login uses it so
// that an invalid address is counted and audited like an unknown one, without
// ever storing the text.
func HashIdentifier(key []byte, raw string) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("chave do hash de e-mail ausente")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(strings.ToLower(strings.TrimSpace(raw))))
	return mac.Sum(nil), nil
}

// HashEmail is the HMAC-SHA256 of the normalized e-mail under the secret key.
// It is what login attempts and audit entries store instead of the e-mail, so a
// leaked table cannot be reversed with a dictionary of addresses.
func HashEmail(key []byte, raw string) ([]byte, error) {
	if len(key) == 0 {
		return nil, errors.New("chave do hash de e-mail ausente")
	}
	email, err := NormalizeEmail(raw)
	if err != nil {
		return nil, err
	}
	return HashIdentifier(key, email)
}

const (
	minPasswordAssociado = 8
	minPasswordAdmin     = 10
	maxPasswordLen       = 128
)

// PasswordViolation is the API code of a password policy failure.
type PasswordViolation string

const (
	PasswordTooShort    PasswordViolation = "password_too_short"
	PasswordTooLong     PasswordViolation = "password_too_long"
	PasswordCompromised PasswordViolation = "password_compromised"
)

// PasswordError is a password policy failure.
type PasswordError struct {
	Violation PasswordViolation
	MinLength int // for password_too_short
}

func (e *PasswordError) Error() string { return string(e.Violation) }

// MinPasswordLength is 10 when the user holds any role other than ASSOCIADO and
// 8 when the only role is ASSOCIADO. An empty or unknown role set is invalid.
func MinPasswordLength(roles []Role) (int, error) {
	if len(roles) == 0 {
		return 0, errors.New("o usuário precisa de ao menos um papel")
	}
	min := minPasswordAssociado
	for _, r := range roles {
		if !r.Valid() {
			return 0, errors.New("papel desconhecido")
		}
		if r != RoleAssociado {
			min = minPasswordAdmin
		}
	}
	return min, nil
}

// ValidatePassword applies the policy: at most 128 Unicode code points, at
// least the minimum for the roles, and not in the list of common or compromised
// passwords. There are no composition rules. Length is counted in code points.
func ValidatePassword(plain string, roles []Role, deny *password.Denylist) error {
	min, err := MinPasswordLength(roles)
	if err != nil {
		return err
	}
	n := utf8.RuneCountInString(plain)
	switch {
	case n > maxPasswordLen:
		return &PasswordError{Violation: PasswordTooLong}
	case n < min:
		return &PasswordError{Violation: PasswordTooShort, MinLength: min}
	case deny.Contains(plain):
		return &PasswordError{Violation: PasswordCompromised}
	}
	return nil
}
