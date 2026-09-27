package identityhttp

import (
	"context"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
)

func (h *Handler) GrantAdminMembership(ctx context.Context, req GrantAdminMembershipRequestObject) (GrantAdminMembershipResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	err = h.M.Promote.Execute(requestContext(ctx), app.PromoteInput{
		Actor: info.Principal, TargetID: req.Id.String(), Roles: toDomainRoles(req.Body.Roles), Reason: req.Body.Reason,
	})
	if err != nil {
		return nil, err
	}
	return GrantAdminMembership204Response{}, nil
}

// RevokeAdminMembership closes the membership; it never deletes it (the history
// stays), which is why the contract has a POST with the reason in the body.
func (h *Handler) RevokeAdminMembership(ctx context.Context, req RevokeAdminMembershipRequestObject) (RevokeAdminMembershipResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	err = h.M.Revoke.Execute(requestContext(ctx), app.RevokeInput{Actor: info.Principal, TargetID: req.Id.String(), Reason: req.Body.Reason})
	if err != nil {
		return nil, err
	}
	return RevokeAdminMembership204Response{}, nil
}

// ResetUserPassword answers with the temporary password once, never cached. It
// is not logged, not audited and not sent by e-mail.
func (h *Handler) ResetUserPassword(ctx context.Context, req ResetUserPasswordRequestObject) (ResetUserPasswordResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	res, err := h.M.AdminReset.Execute(requestContext(ctx), app.AdminResetInput{Actor: info.Principal, TargetID: req.Id.String(), Reason: req.Body.Reason})
	if err != nil {
		return nil, err
	}
	noStore := "no-store"
	return ResetUserPassword200JSONResponse{
		Body:    TemporaryPassword{TemporaryPassword: res.TemporaryPassword},
		Headers: ResetUserPassword200ResponseHeaders{CacheControl: &noStore},
	}, nil
}
