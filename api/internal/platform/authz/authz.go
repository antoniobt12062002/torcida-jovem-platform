package authz

import (
	"context"
	"errors"
	"fmt"
)

// ErrForbidden is what a failed check returns; the API answers 403 forbidden.
// Its message is generic on purpose: it never names the resource or the
// permission, so a denial reveals nothing about whether the resource exists.
var ErrForbidden = errors.New("forbidden")

// Principal is the authenticated user with the effective permissions of the
// current request. Roles are informational (for example the `me` response);
// authorization never reads them.
type Principal struct {
	UserID      string
	Roles       []string
	Permissions map[Permission]struct{}
}

// Has reports whether the principal holds the permission.
func (p Principal) Has(perm Permission) bool {
	_, ok := p.Permissions[perm]
	return ok
}

// Require returns ErrForbidden unless the principal holds the permission.
func Require(p Principal, perm Permission) error {
	if !p.Has(perm) {
		return ErrForbidden
	}
	return nil
}

// DeniedFunc is called when a check fails and the denial must be audited.
type DeniedFunc func(ctx context.Context, p Principal, perm Permission)

// Authorizer is Require plus the denial hook. The hook is injected at wiring
// time, so this package does not depend on audit or identity.
type Authorizer struct {
	defs     map[Permission]Definition
	onDenied DeniedFunc
}

// NewAuthorizer builds an Authorizer from the catalog definitions. onDenied may
// be nil.
func NewAuthorizer(defs []Definition, onDenied DeniedFunc) (*Authorizer, error) {
	byPerm := make(map[Permission]Definition, len(defs))
	for _, d := range defs {
		if err := d.Validate(); err != nil {
			return nil, err
		}
		if prev, ok := byPerm[d.Permission]; ok && prev != d {
			return nil, fmt.Errorf("authz: permissão %s declarada com definições diferentes", d.Permission)
		}
		byPerm[d.Permission] = d
	}
	return &Authorizer{defs: byPerm, onDenied: onDenied}, nil
}

// Require checks the permission. On denial it calls the hook, unless the
// permission is a declared common read (denied reads stay in the access log).
// A permission missing from the catalog is always denied and audited, even if
// the principal somehow holds it: what is not declared is not granted.
func (a *Authorizer) Require(ctx context.Context, p Principal, perm Permission) error {
	def, known := a.defs[perm]
	if known && p.Has(perm) {
		return nil
	}
	if known && def.CommonRead {
		return ErrForbidden
	}
	if a.onDenied != nil {
		a.onDenied(ctx, p, perm)
	}
	return ErrForbidden
}
