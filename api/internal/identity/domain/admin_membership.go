package domain

import (
	"errors"
	"strings"
	"time"
)

// Domain errors of the administrative membership; their messages are the API
// error codes.
var (
	ErrReasonRequired          = errors.New("reason_required")
	ErrAlreadyAdmin            = errors.New("already_admin")
	ErrNotAdmin                = errors.New("not_admin")
	ErrAdminMembershipRequired = errors.New("admin_membership_required")
)

const (
	minReasonLen = 10

	// BootstrapReason is the reason of the membership created by the
	// bootstrap-admin command, which has no grantor.
	BootstrapReason = "bootstrap"
)

// AdminMembership says why a user has administrative access; the roles say what
// the user can do. It is never deleted: closing it keeps the row.
type AdminMembership struct {
	ID           string
	UserID       string
	Reason       string
	GrantedBy    *string // nil only for the bootstrap
	GrantedAt    time.Time
	RevokedBy    *string
	RevokeReason *string
	RevokedAt    *time.Time
}

// NewAdminMembership opens a membership. The reason must have at least 10
// characters after trimming; the bootstrap membership, which has no grantor and
// uses BootstrapReason, is the only exception.
func NewAdminMembership(userID, reason string, grantedBy *string, at time.Time) (AdminMembership, error) {
	reason = strings.TrimSpace(reason)
	bootstrap := grantedBy == nil && reason == BootstrapReason
	if len([]rune(reason)) < minReasonLen && !bootstrap {
		return AdminMembership{}, ErrReasonRequired
	}
	return AdminMembership{UserID: userID, Reason: reason, GrantedBy: grantedBy, GrantedAt: at}, nil
}

// Active reports whether the membership is still open.
func (m AdminMembership) Active() bool { return m.RevokedAt == nil }

// Revoke returns the closed membership, recording who, when and why. The
// receiver is not changed.
func (m AdminMembership) Revoke(by, reason string, at time.Time) (AdminMembership, error) {
	if !m.Active() {
		return AdminMembership{}, ErrNotAdmin
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return AdminMembership{}, ErrReasonRequired
	}
	if by == "" {
		return AdminMembership{}, errors.New("é preciso dizer quem encerra o vínculo")
	}
	m.RevokedBy, m.RevokeReason, m.RevokedAt = &by, &reason, &at
	return m, nil
}

// CheckCanGrant refuses a new membership while the user has an active one.
func CheckCanGrant(history []AdminMembership) error {
	for _, m := range history {
		if m.Active() {
			return ErrAlreadyAdmin
		}
	}
	return nil
}

// CheckRolesHaveMembership: a role other than ASSOCIADO requires an active
// administrative membership.
func CheckRolesHaveMembership(roles []Role, hasActiveMembership bool) error {
	if hasActiveMembership {
		return nil
	}
	for _, r := range roles {
		if r != RoleAssociado {
			return ErrAdminMembershipRequired
		}
	}
	return nil
}
