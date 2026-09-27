package identityhttp

import (
	"context"
	"errors"
)

// The handlers of the user and administrative operations arrive in T75 and T76;
// until then they answer with an internal error, so the interface is complete.
var errPending = errors.New("handler ainda não implementado")

func (h *Handler) ListUsers(context.Context, ListUsersRequestObject) (ListUsersResponseObject, error) {
	return nil, errPending
}

func (h *Handler) CreateUser(context.Context, CreateUserRequestObject) (CreateUserResponseObject, error) {
	return nil, errPending
}

func (h *Handler) GrantAdminMembership(context.Context, GrantAdminMembershipRequestObject) (GrantAdminMembershipResponseObject, error) {
	return nil, errPending
}

func (h *Handler) RevokeAdminMembership(context.Context, RevokeAdminMembershipRequestObject) (RevokeAdminMembershipResponseObject, error) {
	return nil, errPending
}

func (h *Handler) DeactivateUser(context.Context, DeactivateUserRequestObject) (DeactivateUserResponseObject, error) {
	return nil, errPending
}

func (h *Handler) ResetUserPassword(context.Context, ResetUserPasswordRequestObject) (ResetUserPasswordResponseObject, error) {
	return nil, errPending
}

func (h *Handler) ReactivateUser(context.Context, ReactivateUserRequestObject) (ReactivateUserResponseObject, error) {
	return nil, errPending
}

func (h *Handler) SetUserRoles(context.Context, SetUserRolesRequestObject) (SetUserRolesResponseObject, error) {
	return nil, errPending
}
