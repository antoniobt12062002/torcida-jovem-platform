package estoquehttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

// Handler implements StrictServerInterface on top of the estoque use cases
// (added by T4). It is thin: it reads the request, calls a use case and
// turns the result or the error into the contract response. Business rules
// stay in the use cases.
type Handler struct {
	M   *estoque.Module
	Log *slog.Logger
}

// New builds the handler.
func New(m *estoque.Module, log *slog.Logger) *Handler {
	return &Handler{M: m, Log: log}
}

// Chains are the middleware chains the router builds. estoque has no public
// route (every operation requires estoque:*:read/create/adjust), so there is
// only one chain, unlike identity/http.Chains.
type Chains struct {
	Authenticated []gin.HandlerFunc
}

// ---- session (same shape as financeiro/http/handler.go and
// identity/http/handler.go's own ginContext/requestContext/sessionOf,
// duplicated here because they are package-private there and the three HTTP
// layers do not share a module, FIN-D-028)

// ginContext returns the *gin.Context the strict handler passes as context.
func ginContext(ctx context.Context) *gin.Context {
	c, _ := ctx.(*gin.Context)
	return c
}

// requestContext is the context of the request: it carries the request id
// and the audit actor, which the use cases read.
func requestContext(ctx context.Context) context.Context {
	if c := ginContext(ctx); c != nil && c.Request != nil {
		return c.Request.Context()
	}
	return ctx
}

var errNoSession = errors.New("sem sessão no contexto")

// sessionOf reads the session httpx.Authn put in the request context. Every
// resource handler calls this first: estoque never checks permissions
// itself (that is the use case's job, via Authz.Require) — it only needs the
// authenticated Principal to pass through as Actor.
func sessionOf(ctx context.Context) (httpx.SessionInfo, error) {
	info, ok := httpx.SessionFrom(requestContext(ctx))
	if !ok {
		return httpx.SessionInfo{}, errNoSession
	}
	return info, nil
}

// ---- errors (API-04, EST-D-006)

// writeError turns an error of the use cases into the problem response of
// the API. This is the single place where domain errors become HTTP.
func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errNoSession):
		httpx.WriteProblem(c, http.StatusUnauthorized, "unauthenticated", "É preciso entrar para continuar.")
	case errors.Is(err, domain.ErrCampoObrigatorio):
		// EST-D-013: a field holding only whitespace passes the contract's
		// minLength but is blank to the use case. It answers exactly what the
		// contract validator answers for an empty string — the cross-cutting
		// validation_failed, never a new code outside EST-D-006.
		httpx.WriteProblem(c, http.StatusUnprocessableEntity, "validation_failed", "Um ou mais campos são inválidos.")
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

// sentinels is the explicit status+code table of EST-D-006 — exactly its 7
// codes. Like financeiro/domain's sentinels (FIN-D-022), estoque/domain's are
// Portuguese prose, so their .Error() can never be the API code.
var sentinels = []struct {
	err    error
	status int
	code   string
}{
	{domain.ErrProdutoNaoEncontrado, http.StatusNotFound, "produto_nao_encontrado"},
	{domain.ErrCodigoDuplicado, http.StatusConflict, "codigo_duplicado"},
	{domain.ErrDevolucaoInvalida, http.StatusUnprocessableEntity, "devolucao_invalida"},
	// Conflito com o estado atual do estoque, não erro estrutural de payload
	// (EST-D-006).
	{domain.ErrSaldoInsuficiente, http.StatusConflict, "saldo_insuficiente"},
	{domain.ErrQuantidadeInvalida, http.StatusUnprocessableEntity, "quantidade_invalida"},
	{domain.ErrMotivoObrigatorio, http.StatusUnprocessableEntity, "motivo_obrigatorio"},
	{authz.ErrForbidden, http.StatusForbidden, "forbidden"},
	// domain.ErrMovimentacaoNaoEncontrada nunca chega aqui:
	// RegistrarMovimentacao.Execute já a traduz para ErrDevolucaoInvalida
	// antes de retornar (02-movimentacoes/T4).
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

// Register mounts the 6 estoque routes, all behind chains.Authenticated —
// estoque has no public route. Unlike financeiro/http.Register (which only
// returns the wrapper, FIN-D-026/027/029), this can mount routes directly
// because every method lands in a single task (T4): nothing calls Register
// with a concrete *Handler before T4 finishes, so StrictServerInterface only
// needs to be satisfied by the time T5 wires this into httpapi/router.go.
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
	auth := func(op gin.HandlerFunc) []gin.HandlerFunc {
		return append(append([]gin.HandlerFunc{}, chains.Authenticated...), op)
	}

	r.POST("/api/v1/estoque/produtos", auth(w.CreateProduto)...)
	r.GET("/api/v1/estoque/produtos", auth(w.ListProdutos)...)
	r.POST("/api/v1/estoque/movimentacoes", auth(w.CreateMovimentacao)...)
	r.GET("/api/v1/estoque/produtos/:id/movimentacoes", auth(w.ListMovimentacoes)...)
	r.POST("/api/v1/estoque/ajustes", auth(w.CreateAjuste)...)
	r.GET("/api/v1/estoque/produtos/:id/saldo", auth(w.GetSaldo)...)
}
