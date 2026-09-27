package identityhttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

// Handler implements StrictServerInterface on top of the identity use cases. It
// is thin: it reads the principal and the request, calls a use case and turns
// the result or the error into the contract responses. Business rules stay in
// the use cases.
type Handler struct {
	M      *identity.Module
	Cookie httpx.SessionCookie
	Log    *slog.Logger
}

// New builds the handler.
func New(m *identity.Module, cookie httpx.SessionCookie, log *slog.Logger) *Handler {
	return &Handler{M: m, Cookie: cookie, Log: log}
}

// AllowedWhilePasswordChangeIsRequired are the routes a user with
// must_change_password may still call: sign out, see who they are and change the
// password. Every other authenticated route answers 403 password_change_required.
var AllowedWhilePasswordChangeIsRequired = map[string]bool{
	"POST /api/v1/auth/logout":   true,
	"GET /api/v1/auth/me":        true,
	"POST /api/v1/auth/password": true,
}

// Chains are the middleware chains the router builds; each route is registered
// behind the one that fits it.
type Chains struct {
	Public        []gin.HandlerFunc
	Authenticated []gin.HandlerFunc
}

// ginContext returns the *gin.Context the strict handler passes as context.
func ginContext(ctx context.Context) *gin.Context {
	c, _ := ctx.(*gin.Context)
	return c
}

// requestContext is the context of the request: it carries the request id and
// the audit actor, which the use cases read.
func requestContext(ctx context.Context) context.Context {
	if c := ginContext(ctx); c != nil && c.Request != nil {
		return c.Request.Context()
	}
	return ctx
}

var errNoSession = errors.New("sem sessão no contexto")

func sessionOf(ctx context.Context) (httpx.SessionInfo, error) {
	info, ok := httpx.SessionFrom(requestContext(ctx))
	if !ok {
		return httpx.SessionInfo{}, errNoSession
	}
	return info, nil
}

// SessionValidator adapts the identity session service to the HTTP layer.
func (h *Handler) SessionValidator() httpx.SessionValidator { return sessionAdapter{h.M.Session} }

type sessionAdapter struct{ svc *app.SessionService }

func (a sessionAdapter) Validate(ctx context.Context, token string) (httpx.SessionInfo, error) {
	got, err := a.svc.Validate(ctx, token)
	switch {
	case errors.Is(err, app.ErrUnauthenticated):
		return httpx.SessionInfo{}, httpx.ErrUnauthenticated
	case errors.Is(err, app.ErrSessionExpired):
		return httpx.SessionInfo{}, httpx.ErrSessionExpired
	case err != nil:
		return httpx.SessionInfo{}, err
	}
	return httpx.SessionInfo{Principal: got.Principal, SessionID: got.SessionID, CSRFToken: got.CSRFToken, MustChangePassword: got.MustChangePassword}, nil
}

// WithActor puts the authenticated user in the context for the audit trail.
func WithActor(ctx context.Context, info httpx.SessionInfo) context.Context {
	return audit.WithActor(ctx, info.Principal.UserID)
}

// authContext builds what login and `me` answer, from fresh data.
func (h *Handler) authContext(ctx context.Context, userID, csrf string) (AuthContext, error) {
	user, err := h.M.Users.FindByID(ctx, userID)
	if err != nil {
		return AuthContext{}, err
	}
	roles, err := h.M.Roles.RolesOf(ctx, userID)
	if err != nil {
		return AuthContext{}, err
	}
	perms, err := h.M.Roles.EffectivePermissions(ctx, userID)
	if err != nil {
		return AuthContext{}, err
	}
	membership, ok, err := h.M.Members.Active(ctx, userID)
	if err != nil {
		return AuthContext{}, err
	}
	out := AuthContext{
		User:               UserRef{Id: parseUUID(user.ID), Email: user.Email, Name: user.Name},
		Roles:              rolesOut(roles),
		Permissions:        permissionsOut(perms),
		CsrfToken:          csrf,
		MustChangePassword: user.MustChangePassword,
	}
	if ok {
		out.AdminMembership = &AdminMembershipSummary{Reason: membership.Reason, GrantedAt: membership.GrantedAt}
	}
	return out, nil
}

func rolesOut(roles []domain.Role) []Role {
	out := make([]Role, len(roles))
	for i, r := range roles {
		out[i] = Role(r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func permissionsOut(perms []authz.Permission) []string {
	out := make([]string, len(perms))
	for i, p := range perms {
		out[i] = string(p)
	}
	sort.Strings(out)
	return out
}

// ---- errors

// writeError turns an error of the use cases into the problem response of the
// API. This is the single place where domain errors become HTTP.
func (h *Handler) writeError(c *gin.Context, err error) {
	var locked *app.LockedError
	var pwBlocked *app.PasswordChangeBlockedError
	var pwErr *domain.PasswordError
	switch {
	case errors.As(err, &locked):
		retryAfter(c, locked.RetryAfter)
		httpx.WriteProblem(c, http.StatusTooManyRequests, "login_blocked", "Muitas tentativas. Tente novamente mais tarde.")
	case errors.As(err, &pwBlocked):
		retryAfter(c, pwBlocked.RetryAfter)
		httpx.WriteProblem(c, http.StatusTooManyRequests, "password_change_blocked", "Muitas tentativas. Tente novamente mais tarde.")
	case errors.As(err, &pwErr):
		httpx.WriteProblem(c, http.StatusUnprocessableEntity, string(pwErr.Violation), "A senha não cumpre a política.")
	case errors.Is(err, app.ErrInvalidCredentials):
		httpx.WriteProblem(c, http.StatusUnauthorized, "invalid_credentials", "E-mail ou senha inválidos.")
	case errors.Is(err, app.ErrUnauthenticated), errors.Is(err, errNoSession):
		httpx.WriteProblem(c, http.StatusUnauthorized, "unauthenticated", "É preciso entrar para continuar.")
	case errors.Is(err, authz.ErrForbidden):
		httpx.WriteProblem(c, http.StatusForbidden, "forbidden", "Você não tem permissão para esta ação.")
	case errors.Is(err, domain.ErrUserNotFound):
		httpx.WriteProblem(c, http.StatusNotFound, "not_found", "Usuário não encontrado.")
	case errors.Is(err, domain.ErrInvalidResetToken):
		httpx.WriteProblem(c, http.StatusBadRequest, "invalid_reset_token", "O link de recuperação é inválido ou expirou.")
	case mapped(c, err):
	case database.Unavailable(err):
		h.logInternal(c, err)
		httpx.WriteProblem(c, http.StatusServiceUnavailable, "service_unavailable", "Serviço temporariamente indisponível. Tente novamente em instantes.")
	case errors.Is(err, audit.ErrWrite):
		h.logInternal(c, err)
		httpx.WriteProblem(c, http.StatusInternalServerError, "audit_failed", "Não foi possível registrar a auditoria; nada foi alterado.")
	default:
		h.logInternal(c, err)
		httpx.WriteProblem(c, http.StatusInternalServerError, "internal_error", "Erro interno. Informe o request_id ao suporte.")
	}
}

// sentinels are the domain errors whose message is the API code, by status.
var sentinels = []struct {
	status int
	errs   []error
}{
	{http.StatusForbidden, []error{domain.ErrPrivilegeEscalation, domain.ErrSelfChangeForbidden, domain.ErrInvalidCurrentPassword}},
	{http.StatusConflict, []error{domain.ErrEmailTaken, domain.ErrLastAdmin, domain.ErrAlreadyAdmin, domain.ErrNotAdmin, domain.ErrUserInactive, domain.ErrAdminMembershipRequired}},
	{http.StatusUnprocessableEntity, []error{
		domain.ErrPasswordUnchanged, domain.ErrReasonRequired, domain.ErrAdminRoleRequired, domain.ErrInvalidLimit, domain.ErrInvalidCursor,
		domain.ErrInvalidName, domain.ErrInvalidEmail, domain.ErrUnknownRole,
	}},
}

// mapped answers with the status and the code of a sentinel, wrapped or not.
func mapped(c *gin.Context, err error) bool {
	for _, group := range sentinels {
		for _, sentinel := range group.errs {
			if errors.Is(err, sentinel) {
				httpx.WriteProblem(c, group.status, sentinel.Error(), "A operação não pôde ser concluída.")
				return true
			}
		}
	}
	return false
}

func (h *Handler) logInternal(c *gin.Context, err error) {
	if h.Log != nil {
		h.Log.Error("erro interno", "request_id", httpx.RequestIDFrom(c.Request.Context()), "error", err.Error())
	}
}

func retryAfter(c *gin.Context, d time.Duration) {
	secs := int(d.Seconds())
	if secs < 1 {
		secs = 1
	}
	c.Header("Retry-After", strconv.Itoa(secs))
}

// Register mounts the identity routes, each behind its chain. The public ones
// are login and the two password recovery routes; every other route is behind
// the authenticated chain.
func Register(r gin.IRoutes, server StrictServerInterface, h *Handler, chains Chains) {
	w := ServerInterfaceWrapper{
		Handler: NewStrictHandlerWithOptions(server, nil, StrictGinServerOptions{
			RequestErrorHandlerFunc: func(c *gin.Context, _ error) {
				httpx.WriteProblem(c, http.StatusBadRequest, "invalid_json", "O corpo da requisição não é um JSON válido.")
			},
			HandlerErrorFunc:         h.writeError,
			ResponseErrorHandlerFunc: h.writeError,
		}),
		ErrorHandler: func(c *gin.Context, _ error, _ int) {
			httpx.WriteProblem(c, http.StatusUnprocessableEntity, "validation_failed", "Parâmetros inválidos.")
		},
	}
	pub := func(op gin.HandlerFunc) []gin.HandlerFunc { return append(append([]gin.HandlerFunc{}, chains.Public...), op) }
	auth := func(op gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, chains.Authenticated...), op)
	}

	r.POST("/api/v1/auth/login", pub(w.Login)...)
	r.POST("/api/v1/auth/password-reset/request", pub(w.RequestPasswordReset)...)
	r.POST("/api/v1/auth/password-reset/confirm", pub(w.ConfirmPasswordReset)...)

	r.POST("/api/v1/auth/logout", auth(w.Logout)...)
	r.GET("/api/v1/auth/me", auth(w.GetMe)...)
	r.POST("/api/v1/auth/password", auth(w.ChangePassword)...)
	r.GET("/api/v1/users", auth(w.ListUsers)...)
	r.POST("/api/v1/users", auth(w.CreateUser)...)
	r.POST("/api/v1/users/:id/deactivate", auth(w.DeactivateUser)...)
	r.POST("/api/v1/users/:id/reactivate", auth(w.ReactivateUser)...)
	r.PUT("/api/v1/users/:id/roles", auth(w.SetUserRoles)...)
	r.POST("/api/v1/users/:id/admin-membership", auth(w.GrantAdminMembership)...)
	r.POST("/api/v1/users/:id/admin-membership/revoke", auth(w.RevokeAdminMembership)...)
	r.POST("/api/v1/users/:id/password-reset", auth(w.ResetUserPassword)...)
}
