package domain

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"time"
)

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

// EncodeCursor turns the position into the opaque string clients send back.
func EncodeCursor(c ListCursor) string {
	b, _ := json.Marshal(cursorWire{T: c.CreatedAt.UTC().Format(time.RFC3339Nano), ID: c.ID})
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodeCursor parses a cursor; anything that is not one is ErrInvalidCursor.
func DecodeCursor(s string) (ListCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return ListCursor{}, ErrInvalidCursor
	}
	var w cursorWire
	if err := json.Unmarshal(raw, &w); err != nil {
		return ListCursor{}, ErrInvalidCursor
	}
	t, err := time.Parse(time.RFC3339Nano, w.T)
	if err != nil || !uuidCursorID.MatchString(w.ID) {
		return ListCursor{}, ErrInvalidCursor
	}
	return ListCursor{CreatedAt: t, ID: w.ID}, nil
}

type cursorWire struct {
	T  string `json:"t"`
	ID string `json:"id"`
}

var uuidCursorID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
