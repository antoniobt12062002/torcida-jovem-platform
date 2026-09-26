package domain

import "time"

// MembershipSummary is what a user listing shows of the active administrative
// membership.
type MembershipSummary struct {
	Reason    string
	GrantedAt time.Time
}

// UserSummary is a user as listed: never the password hash.
type UserSummary struct {
	ID                 string
	Email              string
	Name               string
	Active             bool
	MustChangePassword bool
	CreatedAt          time.Time
	Roles              []Role
	AdminMembership    *MembershipSummary
}

// ListFilter narrows a user listing; nil fields do not filter.
type ListFilter struct {
	Active *bool
	Role   *Role
}

// ListCursor is the position after the last item of a page: (created_at, id).
type ListCursor struct {
	CreatedAt time.Time
	ID        string
}
