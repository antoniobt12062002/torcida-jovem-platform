package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/config"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/httpapi"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/documents"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/storage"
)

// errStorageRequired: financeiro exposes comprovante upload/download via
// platform/documents (06-api-http) — STORAGE_ENABLED=true is required to
// compose it, never a per-request fallback (FIN-D-023). AD-012 requires
// every route in the contract to always have a registered operation, so the
// dependency can't be conditional at request time.
var errStorageRequired = errors.New("STORAGE_ENABLED=true é obrigatório: financeiro expõe anexo de comprovantes via platform/documents")

// boot wires every dependency and returns the HTTP handler, without starting
// a server — the split lets tests exercise the whole composition (06-api-http
// /T1) without binding a port. main is otherwise unchanged: load config, boot,
// serve, handle shutdown.
func boot(cfg config.Config, logger *slog.Logger) (http.Handler, error) {
	db, err := database.Open(cfg.DatabaseURL, logger)
	if err != nil {
		return nil, err
	}

	recorder := audit.NewRecorder(db, logger)
	matrix, err := app.BuildMatrix(append(app.FoundationContributions(), financeiro.Contribution(), estoque.Contribution())...)
	if err != nil {
		return nil, fmt.Errorf("matriz de papéis inválida: %w", err)
	}
	authorizer, err := authz.NewAuthorizer(matrix.Definitions, audit.DeniedHook(recorder))
	if err != nil {
		return nil, fmt.Errorf("autorizador: %w", err)
	}
	hasher, err := password.NewHasher(password.Params{MemoryKiB: cfg.Argon2MemoryKiB, Iterations: cfg.Argon2Iterations, Parallelism: cfg.Argon2Parallelism})
	if err != nil {
		return nil, fmt.Errorf("parâmetros do Argon2: %w", err)
	}
	sender, err := email.NewSender(cfg.EmailProvider, logger)
	if err != nil {
		return nil, err
	}

	if !cfg.StorageEnabled {
		return nil, errStorageRequired
	}
	s3, err := storage.NewS3(context.Background(), storage.Config{
		Endpoint: cfg.S3Endpoint, Region: cfg.S3Region, Bucket: cfg.S3Bucket,
		AccessKey: cfg.S3AccessKey, SecretKey: cfg.S3SecretKey, UsePathStyle: cfg.S3UsePathStyle,
	})
	if err != nil {
		return nil, fmt.Errorf("armazenamento s3: %w", err)
	}
	docs := &documents.Service{DB: db, Storage: s3, Authz: authorizer, Audit: recorder, URLTTL: cfg.DocumentURLTTL}
	fin := financeiro.New(financeiro.Deps{DB: db, Recorder: recorder, Authorizer: authorizer, Documents: docs})
	est := estoque.New(estoque.Deps{DB: db, Recorder: recorder, Authorizer: authorizer})

	mod := identity.New(identity.Deps{
		DB: db, Recorder: recorder, Authorizer: authorizer, Matrix: matrix, Hasher: hasher, Denylist: password.DefaultDenylist(),
		Sender: sender, Log: logger, HashKey: cfg.AuthHashKey, SessionIdle: cfg.SessionIdle, SessionAbsolute: cfg.SessionAbsolute,
		ResetTTL: cfg.PasswordResetTTL, BaseURL: cfg.AppBaseURL,
	})

	// The roles and permissions in the database follow the code at every start
	// (idempotent, audited as rbac.sync when something changed). Migrations do
	// not run here: they are applied by tj_owner, apart from the API.
	syncCtx, cancelSync := context.WithTimeout(context.Background(), 30*time.Second)
	synced, err := mod.Roles.Sync(syncCtx, matrix)
	cancelSync()
	if err != nil {
		return nil, fmt.Errorf("sincronização de papéis e permissões: %w", err)
	}
	logger.Info("papéis e permissões sincronizados", "changed", synced.Changed)

	return httpapi.NewRouter(httpapi.Deps{
		Ping: database.Ping(db), Identity: mod, Financeiro: fin, Estoque: est, AuditQuery: &audit.Query{Authz: authorizer, DB: db}, Log: logger,
		AllowedOrigins: cfg.AllowedOrigins, Cookie: httpx.SessionCookie{Domain: cfg.CookieDomain, Secure: cfg.CookieSecure},
	}), nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logx.New(cfg.LogLevel, os.Stdout)
	handler, err := boot(cfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
