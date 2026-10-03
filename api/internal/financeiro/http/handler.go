package financeirohttp

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/money"
)

// Handler implements StrictServerInterface on top of the financeiro use
// cases (added by T4-T7). It is thin: it reads the request, calls a use case
// and turns the result or the error into the contract response. Business
// rules stay in the use cases.
type Handler struct {
	M   *financeiro.Module
	Log *slog.Logger
}

// New builds the handler.
func New(m *financeiro.Module, log *slog.Logger) *Handler {
	return &Handler{M: m, Log: log}
}

// ---- errors (FIN-D-022)

// writeError turns an error of the use cases into the problem response of
// the API. This is the single place where domain errors become HTTP.
func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
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

// sentinels is the explicit status+code table of FIN-D-022. Unlike
// identity/http (whose domain sentinels are already snake_case strings),
// financeiro/domain's sentinels are Portuguese prose, so their .Error()
// can never be the API code — only platform/documents's sentinels (already
// snake_case) reuse .Error() here, matching their own package.
var sentinels = []struct {
	err    error
	status int
	code   string
}{
	{domain.ErrContaNaoEncontrada, http.StatusNotFound, "conta_nao_encontrada"},
	{domain.ErrContaTipoIncompativel, http.StatusUnprocessableEntity, "conta_tipo_incompativel"},
	{domain.ErrContaJaUtilizada, http.StatusConflict, "conta_ja_utilizada"},
	// Validação de referência de campo do corpo da requisição (conta_id
	// apontando para uma conta inexistente/incompatível), não do recurso
	// primário da URL — por isso 422, não 404 (FIN-D-022).
	{domain.ErrContaInvalida, http.StatusUnprocessableEntity, "conta_invalida"},
	{domain.ErrLancamentoTipoIncompativel, http.StatusUnprocessableEntity, "lancamento_tipo_incompativel"},
	{domain.ErrLancamentoNaoEncontrado, http.StatusNotFound, "lancamento_nao_encontrado"},
	{domain.ErrLancamentoImutavel, http.StatusConflict, "lancamento_imutavel"},
	{domain.ErrDevolucaoInvalida, http.StatusUnprocessableEntity, "devolucao_invalida"},
	{domain.ErrLancamentoNaoPodeSerRecebido, http.StatusConflict, "lancamento_nao_pode_ser_recebido"},
	{domain.ErrLancamentoNaoPodeSerPago, http.StatusConflict, "lancamento_nao_pode_ser_pago"},
	{domain.ErrLancamentoJaCancelado, http.StatusConflict, "lancamento_ja_cancelado"},
	{domain.ErrMotivoObrigatorio, http.StatusUnprocessableEntity, "motivo_obrigatorio"},
	{money.ErrOutOfRange, http.StatusUnprocessableEntity, "amount_out_of_range"},
	{authz.ErrForbidden, http.StatusForbidden, "forbidden"},
	{documents.ErrExtensionNotAllowed, http.StatusUnprocessableEntity, documents.ErrExtensionNotAllowed.Error()},
	{documents.ErrTypeNotAllowed, http.StatusUnprocessableEntity, documents.ErrTypeNotAllowed.Error()},
	{documents.ErrTypeMismatch, http.StatusUnprocessableEntity, documents.ErrTypeMismatch.Error()},
	{documents.ErrTooLarge, http.StatusRequestEntityTooLarge, documents.ErrTooLarge.Error()},
	{documents.ErrNotFound, http.StatusNotFound, documents.ErrNotFound.Error()},
	// documents.ErrOwnerTypeInvalid não está nesta tabela: não deveria
	// ocorrer para o financeiro (único owner_type usado é "lancamento"); se
	// ocorrer, cai no default (500 internal_error) como qualquer erro não
	// mapeado (FIN-D-022).
}

// mapped answers with the status and the code of a sentinel, wrapped or not.
func mapped(c *gin.Context, err error) bool {
	for _, s := range sentinels {
		if errors.Is(err, s.err) {
			httpx.WriteProblem(c, s.status, s.code, "A operação não pôde ser concluída.")
			return true
		}
	}
	return false
}

func (h *Handler) logInternal(c *gin.Context, err error) {
	if h.Log != nil {
		h.Log.Error("erro interno", "request_id", httpx.RequestIDFrom(c.Request.Context()), "error", err.Error())
	}
}

// Register builds the strict wrapper for the financeiro contract: the
// HandlerErrorFunc/ResponseErrorHandlerFunc route use-case errors through
// writeError, RequestErrorHandlerFunc/ErrorHandler cover malformed JSON and
// parameter binding. No route is mounted here yet (T4-T7 add the 15
// StrictServerInterface methods on *Handler; T8 mounts them in
// httpapi/router.go once Handler satisfies the interface in full).
func Register(server StrictServerInterface, h *Handler) ServerInterfaceWrapper {
	return ServerInterfaceWrapper{
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
}
