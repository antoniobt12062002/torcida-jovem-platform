package identityhttp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

// recoveryAccepted is the fixed answer of the recovery request: the same whether
// or not the account exists.
const recoveryAccepted = "Se existir uma conta vinculada ao e-mail informado, enviaremos instruções."

func parseUUID(id string) openapi_types.UUID {
	u, _ := uuid.Parse(id)
	return openapi_types.UUID(u)
}

func (h *Handler) Login(ctx context.Context, req LoginRequestObject) (LoginResponseObject, error) {
	c, rc := ginContext(ctx), requestContext(ctx)
	presented, _ := h.Cookie.Read(c)
	res, err := h.M.Login.Login(rc, app.LoginInput{Email: req.Body.Email, Password: req.Body.Password, PresentedToken: presented})
	if err != nil {
		return nil, err
	}
	body, err := h.authContext(rc, res.UserID, res.Session.CSRFToken)
	if err != nil {
		return nil, err
	}
	h.Cookie.Set(c, res.Token, h.M.SessionAbsolute)
	return Login200JSONResponse{Body: body}, nil
}

func (h *Handler) Logout(ctx context.Context, _ LogoutRequestObject) (LogoutResponseObject, error) {
	c := ginContext(ctx)
	token, _ := h.Cookie.Read(c)
	if err := h.M.Session.Logout(requestContext(ctx), token); err != nil {
		return nil, err
	}
	h.Cookie.Clear(c)
	return Logout204Response{}, nil
}

func (h *Handler) GetMe(ctx context.Context, _ GetMeRequestObject) (GetMeResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	body, err := h.authContext(requestContext(ctx), info.Principal.UserID, info.CSRFToken)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, app.ErrUnauthenticated
		}
		return nil, err
	}
	return GetMe200JSONResponse(body), nil
}

func (h *Handler) ChangePassword(ctx context.Context, req ChangePasswordRequestObject) (ChangePasswordResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	err = h.M.ChangePassword.Execute(requestContext(ctx), app.ChangePasswordInput{
		UserID: info.Principal.UserID, CurrentSessionID: info.SessionID, Current: req.Body.CurrentPassword, New: req.Body.NewPassword,
	})
	if err != nil {
		return nil, err
	}
	return ChangePassword204Response{}, nil
}

func (h *Handler) RequestPasswordReset(ctx context.Context, req RequestPasswordResetRequestObject) (RequestPasswordResetResponseObject, error) {
	if err := h.M.RequestReset.Execute(requestContext(ctx), app.RequestResetInput{Email: req.Body.Email}); err != nil {
		return nil, err
	}
	return RequestPasswordReset202JSONResponse{Message: recoveryAccepted}, nil
}

func (h *Handler) ConfirmPasswordReset(ctx context.Context, req ConfirmPasswordResetRequestObject) (ConfirmPasswordResetResponseObject, error) {
	if err := h.M.ResetPassword.Execute(requestContext(ctx), app.ResetInput{Token: req.Body.Token, NewPassword: req.Body.NewPassword}); err != nil {
		return nil, err
	}
	return ConfirmPasswordReset204Response{}, nil
}
