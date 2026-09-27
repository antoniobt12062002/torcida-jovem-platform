package identityhttp

import (
	"context"
	"errors"
)

// The handlers of the administrative membership arrive in T76; until then they
// answer with an internal error, so the interface is complete.
var errPending = errors.New("handler ainda não implementado")

func (h *Handler) GrantAdminMembership(context.Context, GrantAdminMembershipRequestObject) (GrantAdminMembershipResponseObject, error) {
	return nil, errPending
}

func (h *Handler) RevokeAdminMembership(context.Context, RevokeAdminMembershipRequestObject) (RevokeAdminMembershipResponseObject, error) {
	return nil, errPending
}

func (h *Handler) ResetUserPassword(context.Context, ResetUserPasswordRequestObject) (ResetUserPasswordResponseObject, error) {
	return nil, errPending
}
