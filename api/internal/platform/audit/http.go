package audit

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	openapi_types "github.com/oapi-codegen/runtime/types"

	platformapi "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/api"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

// Handler serves GET /api/v1/audit-logs. The router mounts it behind the
// authenticated chain; the permission is checked by the query.
type Handler struct {
	Query *Query
}

// GetAuditLogs answers the audit query: newest first, filters combined by AND,
// keyset cursor, 50 per page by default.
func (h Handler) GetAuditLogs(c *gin.Context, params platformapi.GetAuditLogsParams) {
	principal, ok := httpx.PrincipalFrom(c.Request.Context())
	if !ok {
		httpx.WriteProblem(c, http.StatusUnauthorized, "unauthenticated", "É preciso entrar para continuar.")
		return
	}
	f := Filter{EntityType: params.EntityType, EntityID: params.EntityId, Action: params.Action, From: params.From, To: params.To}
	if params.ActorUserId != nil {
		id := params.ActorUserId.String()
		f.ActorUserID = &id
	}
	if params.Outcome != nil {
		outcome := string(*params.Outcome)
		f.Outcome = &outcome
	}
	limit, cursor := 0, ""
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Cursor != nil {
		cursor = string(*params.Cursor)
	}

	page, err := h.Query.Search(c.Request.Context(), principal, f, limit, cursor)
	switch {
	case errors.Is(err, authz.ErrForbidden):
		httpx.WriteProblem(c, http.StatusForbidden, "forbidden", "Você não tem permissão para esta ação.")
	case errors.Is(err, ErrInvalidLimit):
		httpx.WriteProblem(c, http.StatusUnprocessableEntity, "invalid_limit", "O limite deve estar entre 1 e 100.")
	case errors.Is(err, ErrInvalidCursor):
		httpx.WriteProblem(c, http.StatusUnprocessableEntity, "invalid_cursor", "Cursor inválido.")
	case database.Unavailable(err):
		httpx.WriteProblem(c, http.StatusServiceUnavailable, "service_unavailable", "Serviço temporariamente indisponível. Tente novamente em instantes.")
	case err != nil:
		httpx.WriteProblem(c, http.StatusInternalServerError, "internal_error", "Erro interno. Informe o request_id ao suporte.")
	default:
		c.JSON(http.StatusOK, toPage(page))
	}
}

func toPage(p Page) platformapi.AuditLogPage {
	out := platformapi.AuditLogPage{Items: make([]platformapi.AuditLog, len(p.Items))}
	for i, r := range p.Items {
		item := platformapi.AuditLog{
			Action: r.Action, ActorType: platformapi.AuditLogActorType(r.ActorType), Context: r.Context, EntityId: r.EntityID,
			EntityType: r.EntityType, Id: uuidOf(r.ID), OccurredAt: r.OccurredAt, Outcome: platformapi.AuditLogOutcome(r.Outcome),
			Reason: r.Reason, RequestId: r.RequestID,
		}
		if r.ActorUserID != nil {
			id := uuidOf(*r.ActorUserID)
			item.ActorUserId = &id
		}
		if r.Before != nil {
			before := r.Before
			item.Before = &before
		}
		if r.After != nil {
			after := r.After
			item.After = &after
		}
		out.Items[i] = item
	}
	if p.NextCursor != "" {
		next := p.NextCursor
		out.NextCursor = &next
	}
	return out
}

func uuidOf(s string) openapi_types.UUID {
	var u openapi_types.UUID
	_ = u.UnmarshalText([]byte(s))
	return u
}
