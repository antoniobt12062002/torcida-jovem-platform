package identityhttp

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

func toDomainRoles(roles []Role) []domain.Role {
	out := make([]domain.Role, len(roles))
	for i, r := range roles {
		out[i] = domain.Role(r)
	}
	return out
}

// userOut is a listed user: never the password hash or any token. Every
// timestamp is normalized to UTC (AUD-01.3), independent of the server's
// local time zone: PostgreSQL's timestamptz decodes with the Go runtime's
// zone, not necessarily UTC.
func userOut(s domain.UserSummary) User {
	u := User{
		Id: parseUUID(s.ID), Email: s.Email, Name: s.Name, Active: s.Active, MustChangePassword: s.MustChangePassword,
		CreatedAt: s.CreatedAt.UTC(), Roles: rolesOut(s.Roles),
	}
	if s.AdminMembership != nil {
		u.AdminMembership = &AdminMembershipSummary{Reason: s.AdminMembership.Reason, GrantedAt: s.AdminMembership.GrantedAt.UTC()}
	}
	return u
}

func (h *Handler) ListUsers(ctx context.Context, req ListUsersRequestObject) (ListUsersResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	in := app.ListInput{Actor: info.Principal, Active: req.Params.Active}
	if req.Params.Limit != nil {
		in.Limit = *req.Params.Limit
	}
	if req.Params.Cursor != nil {
		in.Cursor = string(*req.Params.Cursor)
	}
	if req.Params.Role != nil {
		role := domain.Role(*req.Params.Role)
		in.Role = &role
	}
	page, err := h.M.ListUsers.Execute(requestContext(ctx), in)
	if err != nil {
		return nil, err
	}
	out := UserPage{Items: make([]User, len(page.Items))}
	for i, s := range page.Items {
		out.Items[i] = userOut(s)
	}
	if page.NextCursor != "" {
		next := page.NextCursor
		out.NextCursor = &next
	}
	return ListUsers200JSONResponse(out), nil
}

func (h *Handler) CreateUser(ctx context.Context, req CreateUserRequestObject) (CreateUserResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	u, err := h.M.CreateUser.Execute(requestContext(ctx), app.CreateUserInput{
		Actor: info.Principal, Email: req.Body.Email, Name: req.Body.Name, Password: req.Body.Password,
	})
	if err != nil {
		return nil, err
	}
	return CreateUser201JSONResponse(userOut(domain.UserSummary{
		ID: u.ID, Email: u.Email, Name: u.Name, Active: u.Active, MustChangePassword: u.MustChangePassword,
		CreatedAt: u.CreatedAt, Roles: []domain.Role{domain.RoleAssociado},
	})), nil
}

func (h *Handler) DeactivateUser(ctx context.Context, req DeactivateUserRequestObject) (DeactivateUserResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.M.Activation.Deactivate(requestContext(ctx), app.ActivationInput{Actor: info.Principal, TargetID: req.Id.String()}); err != nil {
		return nil, err
	}
	return DeactivateUser204Response{}, nil
}

func (h *Handler) ReactivateUser(ctx context.Context, req ReactivateUserRequestObject) (ReactivateUserResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.M.Activation.Reactivate(requestContext(ctx), app.ActivationInput{Actor: info.Principal, TargetID: req.Id.String()}); err != nil {
		return nil, err
	}
	return ReactivateUser204Response{}, nil
}

func (h *Handler) SetUserRoles(ctx context.Context, req SetUserRolesRequestObject) (SetUserRolesResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	err = h.M.AssignRoles.Execute(requestContext(ctx), app.AssignInput{
		Actor: info.Principal, TargetID: req.Id.String(), Roles: toDomainRoles(req.Body.Roles),
	})
	if err != nil {
		return nil, err
	}
	return SetUserRoles204Response{}, nil
}
