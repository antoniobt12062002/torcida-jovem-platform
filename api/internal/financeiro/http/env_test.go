//go:build integration

// env_test.go is the shared integration harness for this package's handler
// tests (T4-T7 of 06-api-http). It cannot use financeirohttp.Register or
// httpapi.NewRouter: both require *Handler to satisfy StrictServerInterface
// in full, which only happens once T4-T7 together add all 15 methods (Go
// cannot satisfy an interface partially) — see FIN-D-028 in
// .specs/features/financeiro/STATE.md. Instead, newEngine wires each
// operation under test by hand, behind the same real middleware chain
// httpapi/router.go uses (Authn, CSRF, BodyLimit, Origin, contract
// validation) and a real identity module for login — only the route
// registration itself (normally generated code, ServerInterfaceWrapper) is
// reimplemented here, grown task by task as T5-T7 add their own operations.
// T8 deletes this file's route-wiring in favor of the real Register call.
package financeirohttp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	identityapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	identityhttp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/http"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email/emailtest"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/storage"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

const appOrigin = "https://app.tj.example"
const goodPassword = "uma-senha-correta-123"

// env is identity (for real login/session) and financeiro (under test)
// behind the real middleware chains, over a private database.
type env struct {
	t        *testing.T
	db       *gorm.DB
	idMod    *identity.Module
	finMod   *financeiro.Module
	h        *Handler
	engine   *gin.Engine
	contract *testutil.Contract
	hasher   *password.Hasher
	seq      int
}

func newEnv(t *testing.T) *env {
	t.Helper()
	db := testutil.NewTestDB(t)
	log := logx.New("error", io.Discard)
	rec := audit.NewRecorder(db, log)
	matrix, err := identityapp.BuildMatrix(append(identityapp.FoundationContributions(), financeiro.Contribution())...)
	if err != nil {
		t.Fatal(err)
	}
	hasher, err := password.NewHasher(password.Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1})
	if err != nil {
		t.Fatal(err)
	}
	authorizer, err := authz.NewAuthorizer(matrix.Definitions, audit.DeniedHook(rec))
	if err != nil {
		t.Fatal(err)
	}
	idMod := identity.New(identity.Deps{
		DB: db, Recorder: rec, Authorizer: authorizer, Matrix: matrix, Hasher: hasher,
		Denylist: password.NewDenylist("12345678\nsenhacomum1234\n"), Sender: &emailtest.Recorder{}, Log: log,
		HashKey: []byte("0123456789abcdef0123456789abcdef"), SessionIdle: 60 * time.Minute, SessionAbsolute: 8 * time.Hour,
		ResetTTL: 30 * time.Minute, BaseURL: appOrigin, Now: time.Now,
	})
	if _, err := idMod.Roles.Sync(context.Background(), matrix); err != nil {
		t.Fatal(err)
	}
	s3cfg := testutil.SharedS3(t)
	s3, err := storage.NewS3(context.Background(), storage.Config{
		Endpoint: s3cfg.Endpoint, Region: s3cfg.Region, Bucket: s3cfg.Bucket,
		AccessKey: s3cfg.AccessKey, SecretKey: s3cfg.SecretKey, UsePathStyle: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	docs := &documents.Service{DB: db, Storage: s3, Authz: authorizer, Audit: rec, URLTTL: 5 * time.Minute}
	finMod := financeiro.New(financeiro.Deps{DB: db, Recorder: rec, Authorizer: authorizer, Documents: docs})
	e := &env{
		t: t, db: db, idMod: idMod, finMod: finMod, hasher: hasher,
		h: New(finMod, log), contract: testutil.LoadContracts(t),
	}
	e.engine = e.newEngine(log)
	return e
}

// newEngine wires the real chains (same shape as httpapi/router.go's
// "authenticated"/"public") over identity's real routes (for login) plus
// financeiro's routes under test, registered by hand (see the file comment).
func (e *env) newEngine(log *slog.Logger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	idSpec, err := identityhttp.GetSpec()
	if err != nil {
		e.t.Fatal(err)
	}
	finSpec, err := GetSpec()
	if err != nil {
		e.t.Fatal(err)
	}
	validate, err := httpx.NewContractValidator(idSpec, finSpec)
	if err != nil {
		e.t.Fatal(err)
	}
	cookie := httpx.SessionCookie{Secure: true}
	idH := identityhttp.New(e.idMod, cookie, log)
	origin := httpx.Origin(httpx.OriginConfig{Allowed: []string{appOrigin}, Cookie: cookie})
	authn := httpx.Authn(httpx.AuthnConfig{
		Validator: idH.SessionValidator(), Cookie: cookie,
		AllowedWhilePasswordChangeIsRequired: identityhttp.AllowedWhilePasswordChangeIsRequired,
		OnAuthenticated:                      identityhttp.WithActor,
	})
	authenticated := []gin.HandlerFunc{httpx.BodyLimit(1 << 20), origin, authn, httpx.CSRF(), validate}

	r := gin.New()
	r.Use(httpx.RequestID())
	identityhttp.Register(r, idH, idH, identityhttp.Chains{
		Public:        []gin.HandlerFunc{origin, httpx.BodyLimit(64 << 10), validate},
		Authenticated: authenticated,
	})
	e.registerContaRoutes(r, authenticated)
	e.registerLancamentoRoutes(r, authenticated)
	e.registerSaldoRoute(r, authenticated)
	// financeiroUpload mirrors httpapi.FinanceiroUploadBodyLimit (FIN-D-024,
	// FIN-D-031): the one multipart route needs a body limit larger than
	// the 1 MiB `authenticated` chain allows.
	financeiroUpload := []gin.HandlerFunc{httpx.BodyLimit(11 << 20), origin, authn, httpx.CSRF(), validate}
	e.registerComprovanteRoutes(r, authenticated, financeiroUpload)
	return r
}

// registerContaRoutes mounts the 4 conta operations (T4/API-01) by hand.
func (e *env) registerContaRoutes(r gin.IRoutes, chain []gin.HandlerFunc) {
	h := e.h
	route := func(fn gin.HandlerFunc) []gin.HandlerFunc { return append(append([]gin.HandlerFunc{}, chain...), fn) }

	r.GET("/api/v1/financeiro/contas", route(func(c *gin.Context) {
		resp, err := h.ListContas(c, ListContasRequestObject{})
		writeStrictResponse(c, err, h, resp, (ListContasResponseObject).VisitListContasResponse)
	})...)
	r.POST("/api/v1/financeiro/contas", route(func(c *gin.Context) {
		var body CreateContaJSONRequestBody
		if !bindJSON(c, &body) {
			return
		}
		resp, err := h.CreateConta(c, CreateContaRequestObject{Body: &body})
		writeStrictResponse(c, err, h, resp, (CreateContaResponseObject).VisitCreateContaResponse)
	})...)
	r.PATCH("/api/v1/financeiro/contas/:id", route(func(c *gin.Context) {
		id, ok := pathUUID(c, "id")
		if !ok {
			return
		}
		var body RenameContaJSONRequestBody
		if !bindJSON(c, &body) {
			return
		}
		resp, err := h.RenameConta(c, RenameContaRequestObject{Id: ContaId(id), Body: &body})
		writeStrictResponse(c, err, h, resp, (RenameContaResponseObject).VisitRenameContaResponse)
	})...)
	r.POST("/api/v1/financeiro/contas/:id/deactivate", route(func(c *gin.Context) {
		id, ok := pathUUID(c, "id")
		if !ok {
			return
		}
		resp, err := h.DeactivateConta(c, DeactivateContaRequestObject{Id: ContaId(id)})
		writeStrictResponse(c, err, h, resp, (DeactivateContaResponseObject).VisitDeactivateContaResponse)
	})...)
}

// registerLancamentoRoutes mounts the 7 lançamento+workflow operations
// (T5/API-02, API-03) by hand.
func (e *env) registerLancamentoRoutes(r gin.IRoutes, chain []gin.HandlerFunc) {
	h := e.h
	route := func(fn gin.HandlerFunc) []gin.HandlerFunc { return append(append([]gin.HandlerFunc{}, chain...), fn) }

	r.GET("/api/v1/financeiro/lancamentos", route(func(c *gin.Context) {
		resp, err := h.ListLancamentos(c, ListLancamentosRequestObject{})
		writeStrictResponse(c, err, h, resp, (ListLancamentosResponseObject).VisitListLancamentosResponse)
	})...)
	r.POST("/api/v1/financeiro/lancamentos", route(func(c *gin.Context) {
		var body CreateLancamentoJSONRequestBody
		if !bindJSON(c, &body) {
			return
		}
		resp, err := h.CreateLancamento(c, CreateLancamentoRequestObject{Body: &body})
		writeStrictResponse(c, err, h, resp, (CreateLancamentoResponseObject).VisitCreateLancamentoResponse)
	})...)
	r.POST("/api/v1/financeiro/lancamentos/devolucoes", route(func(c *gin.Context) {
		var body CreateDevolucaoJSONRequestBody
		if !bindJSON(c, &body) {
			return
		}
		resp, err := h.CreateDevolucao(c, CreateDevolucaoRequestObject{Body: &body})
		writeStrictResponse(c, err, h, resp, (CreateDevolucaoResponseObject).VisitCreateDevolucaoResponse)
	})...)
	r.PUT("/api/v1/financeiro/lancamentos/:id", route(func(c *gin.Context) {
		id, ok := pathUUID(c, "id")
		if !ok {
			return
		}
		var body UpdateLancamentoJSONRequestBody
		if !bindJSON(c, &body) {
			return
		}
		resp, err := h.UpdateLancamento(c, UpdateLancamentoRequestObject{Id: LancamentoId(id), Body: &body})
		writeStrictResponse(c, err, h, resp, (UpdateLancamentoResponseObject).VisitUpdateLancamentoResponse)
	})...)
	r.POST("/api/v1/financeiro/lancamentos/:id/receive", route(func(c *gin.Context) {
		id, ok := pathUUID(c, "id")
		if !ok {
			return
		}
		resp, err := h.ReceiveLancamento(c, ReceiveLancamentoRequestObject{Id: LancamentoId(id)})
		writeStrictResponse(c, err, h, resp, (ReceiveLancamentoResponseObject).VisitReceiveLancamentoResponse)
	})...)
	r.POST("/api/v1/financeiro/lancamentos/:id/pay", route(func(c *gin.Context) {
		id, ok := pathUUID(c, "id")
		if !ok {
			return
		}
		resp, err := h.PayLancamento(c, PayLancamentoRequestObject{Id: LancamentoId(id)})
		writeStrictResponse(c, err, h, resp, (PayLancamentoResponseObject).VisitPayLancamentoResponse)
	})...)
	r.POST("/api/v1/financeiro/lancamentos/:id/cancel", route(func(c *gin.Context) {
		id, ok := pathUUID(c, "id")
		if !ok {
			return
		}
		var body CancelLancamentoJSONRequestBody
		if !bindJSON(c, &body) {
			return
		}
		resp, err := h.CancelLancamento(c, CancelLancamentoRequestObject{Id: LancamentoId(id), Body: &body})
		writeStrictResponse(c, err, h, resp, (CancelLancamentoResponseObject).VisitCancelLancamentoResponse)
	})...)
}

// registerSaldoRoute mounts the single saldo operation (T6/API-04) by hand.
func (e *env) registerSaldoRoute(r gin.IRoutes, chain []gin.HandlerFunc) {
	h := e.h
	route := append(append([]gin.HandlerFunc{}, chain...), func(c *gin.Context) {
		resp, err := h.GetSaldo(c, GetSaldoRequestObject{})
		writeStrictResponse(c, err, h, resp, (GetSaldoResponseObject).VisitGetSaldoResponse)
	})
	r.GET("/api/v1/financeiro/saldo", route...)
}

// registerComprovanteRoutes mounts the 3 comprovante operations (T7/API-05)
// by hand. CreateComprovante (multipart) is the one route behind
// uploadChain; the other two are plain JSON/no-body and use chain.
func (e *env) registerComprovanteRoutes(r gin.IRoutes, chain, uploadChain []gin.HandlerFunc) {
	h := e.h
	route := func(c []gin.HandlerFunc, fn gin.HandlerFunc) []gin.HandlerFunc { return append(append([]gin.HandlerFunc{}, c...), fn) }

	r.POST("/api/v1/financeiro/lancamentos/:id/comprovantes", route(uploadChain, func(c *gin.Context) {
		id, ok := pathUUID(c, "id")
		if !ok {
			return
		}
		mr, err := c.Request.MultipartReader()
		if err != nil {
			httpx.WriteProblem(c, http.StatusBadRequest, "invalid_json", "O corpo da requisição não é multipart/form-data válido.")
			return
		}
		resp, err := h.CreateComprovante(c, CreateComprovanteRequestObject{Id: LancamentoId(id), Body: mr})
		writeStrictResponse(c, err, h, resp, (CreateComprovanteResponseObject).VisitCreateComprovanteResponse)
	})...)
	r.GET("/api/v1/financeiro/lancamentos/:id/comprovantes", route(chain, func(c *gin.Context) {
		id, ok := pathUUID(c, "id")
		if !ok {
			return
		}
		resp, err := h.ListComprovantes(c, ListComprovantesRequestObject{Id: LancamentoId(id)})
		writeStrictResponse(c, err, h, resp, (ListComprovantesResponseObject).VisitListComprovantesResponse)
	})...)
	r.GET("/api/v1/financeiro/comprovantes/:documentId/url", route(chain, func(c *gin.Context) {
		id, ok := pathUUID(c, "documentId")
		if !ok {
			return
		}
		resp, err := h.GetComprovanteUrl(c, GetComprovanteUrlRequestObject{DocumentId: DocumentId(id)})
		writeStrictResponse(c, err, h, resp, (GetComprovanteUrlResponseObject).VisitGetComprovanteUrlResponse)
	})...)
}

// bindJSON reports whether body parsed as JSON, writing invalid_json
// (the same answer Register's RequestErrorHandlerFunc gives) otherwise.
func bindJSON(c *gin.Context, body any) bool {
	if err := c.ShouldBindJSON(body); err != nil {
		httpx.WriteProblem(c, http.StatusBadRequest, "invalid_json", "O corpo da requisição não é um JSON válido.")
		return false
	}
	return true
}

// pathUUID reports whether the named path param parses as a UUID, writing
// validation_failed (the same answer Register's ErrorHandler gives for a
// malformed parameter) otherwise.
func pathUUID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		httpx.WriteProblem(c, http.StatusUnprocessableEntity, "validation_failed", "Parâmetros inválidos.")
		return uuid.UUID{}, false
	}
	return id, true
}

// writeStrictResponse is what Register's generated wrapper does for every
// operation: an error goes through writeError, a response through its own
// Visit method.
func writeStrictResponse[T any](c *gin.Context, err error, h *Handler, resp T, visit func(T, http.ResponseWriter) error) {
	if err != nil {
		h.writeError(c, err)
		return
	}
	if err := visit(resp, c.Writer); err != nil {
		h.writeError(c, err)
	}
}

type opt func(*http.Request)

func cookieOpt(token string) opt {
	return func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "tj_session", Value: token}) }
}
func csrfOpt(token string) opt { return func(r *http.Request) { r.Header.Set("X-CSRF-Token", token) } }

type session struct{ token, csrf string }

func (s session) opts() []opt { return []opt{cookieOpt(s.token), csrfOpt(s.csrf)} }

func (e *env) do(method, path string, body any, opts ...opt) *httptest.ResponseRecorder {
	e.t.Helper()
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			e.t.Fatal(err)
		}
		rd = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rd)
	req.Header.Set("Origin", appOrigin)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, o := range opts {
		o(req)
	}
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)
	e.contract.ValidateResponse(e.t, req, w)
	return w
}

func (e *env) as(s session, method, path string, body any, more ...opt) *httptest.ResponseRecorder {
	e.t.Helper()
	return e.do(method, path, body, append(s.opts(), more...)...)
}

// uploadFile sends a real multipart/form-data request (field "file"),
// through the real chain and contract validation, same as do/as.
func (e *env) uploadFile(s session, path, filename, content string) *httptest.ResponseRecorder {
	e.t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		e.t.Fatal(err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		e.t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		e.t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Origin", appOrigin)
	for _, o := range s.opts() {
		o(req)
	}
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)
	e.contract.ValidateResponse(e.t, req, w)
	return w
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("corpo não é JSON: %v: %s", err, w.Body.String())
	}
	return m
}

func codeOf(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	c, _ := decode(t, w)["code"].(string)
	return c
}

// newUser creates an active user with the given roles (ASSOCIADO if none),
// mirroring identity/http/env_test.go's own fixture.
func (e *env) newUser(roles ...domain.Role) domain.User {
	e.t.Helper()
	e.seq++
	hash, err := e.hasher.Hash(goodPassword)
	if err != nil {
		e.t.Fatal(err)
	}
	email := "user" + uuid.NewString() + "@exemplo.com"
	u, err := e.idMod.Users.Create(context.Background(), domain.User{Email: email, Name: "Usuário de Teste", PasswordHash: hash, Active: true})
	if err != nil {
		e.t.Fatal(err)
	}
	if len(roles) == 0 {
		roles = []domain.Role{domain.RoleAssociado}
	}
	if err := e.idMod.Users.SetRoles(context.Background(), u.ID, roles); err != nil {
		e.t.Fatal(err)
	}
	return u
}

func (e *env) login(u domain.User) session {
	e.t.Helper()
	w := e.do(http.MethodPost, "/api/v1/auth/login", map[string]string{"email": u.Email, "password": goodPassword})
	if w.Code != http.StatusOK {
		e.t.Fatalf("login de %s: %d %s", u.Email, w.Code, w.Body.String())
	}
	var token string
	for _, c := range w.Result().Cookies() {
		if c.Name == "tj_session" {
			token = c.Value
		}
	}
	return session{token: token, csrf: decode(e.t, w)["csrf_token"].(string)}
}

// signedIn creates a user with the roles and signs them in.
func (e *env) signedIn(roles ...domain.Role) (domain.User, session) {
	e.t.Helper()
	u := e.newUser(roles...)
	return u, e.login(u)
}
