package app

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

// User listing limits (IDN-04.8).
const (
	defaultListLimit = 50
	maxListLimit     = 100
)

// UserLister is what ListUsers needs of the user repository.
type UserLister interface {
	List(ctx context.Context, filter domain.ListFilter, after *domain.ListCursor, limit int) ([]domain.UserSummary, error)
}

// ListUsers lists users, newest first, with cursor pagination over
// (created_at, id), 50 per page by default and at most 100, filterable by active
// and role. Items never carry the password hash or any token.
type ListUsers struct {
	Authz Authorizer
	Users UserLister
}

type ListInput struct {
	Actor  authz.Principal
	Limit  int // 0 means the default
	Cursor string
	Active *bool
	Role   *domain.Role
}

// UserPage is one page; NextCursor is empty on the last one.
type UserPage struct {
	Items      []domain.UserSummary
	NextCursor string
}

func (uc *ListUsers) Execute(ctx context.Context, in ListInput) (UserPage, error) {
	if uc.Authz == nil || uc.Users == nil {
		return UserPage{}, errNotConfigured
	}
	if err := uc.Authz.Require(ctx, in.Actor, "identity:user:read"); err != nil {
		return UserPage{}, err
	}
	limit := in.Limit
	switch {
	case limit == 0:
		limit = defaultListLimit
	case limit < 0 || limit > maxListLimit:
		return UserPage{}, domain.ErrInvalidLimit
	}
	var after *domain.ListCursor
	if in.Cursor != "" {
		c, err := domain.DecodeCursor(in.Cursor)
		if err != nil {
			return UserPage{}, err
		}
		after = &c
	}
	if in.Role != nil && !in.Role.Valid() {
		return UserPage{}, domain.ErrUnknownRole
	}

	items, err := uc.Users.List(ctx, domain.ListFilter{Active: in.Active, Role: in.Role}, after, limit+1)
	if err != nil {
		return UserPage{}, err
	}
	page := UserPage{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		last := page.Items[limit-1]
		page.NextCursor = domain.EncodeCursor(domain.ListCursor{CreatedAt: last.CreatedAt, ID: last.ID})
	}
	return page, nil
}
