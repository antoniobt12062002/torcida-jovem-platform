package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/config"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/httpapi"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logx.New(cfg.LogLevel, os.Stdout)
	db, err := database.Open(cfg.DatabaseURL, logger)
	if err != nil {
		log.Fatal(err)
	}

	recorder := audit.NewRecorder(db, logger)
	matrix, err := app.BuildMatrix(app.FoundationContributions()...)
	if err != nil {
		log.Fatalf("matriz de papéis inválida: %v", err)
	}
	authorizer, err := authz.NewAuthorizer(matrix.Definitions, audit.DeniedHook(recorder))
	if err != nil {
		log.Fatalf("autorizador: %v", err)
	}
	hasher, err := password.NewHasher(password.Params{MemoryKiB: cfg.Argon2MemoryKiB, Iterations: cfg.Argon2Iterations, Parallelism: cfg.Argon2Parallelism})
	if err != nil {
		log.Fatalf("parâmetros do Argon2: %v", err)
	}
	sender, err := email.NewSender(cfg.EmailProvider, logger)
	if err != nil {
		log.Fatal(err)
	}
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
		log.Fatalf("sincronização de papéis e permissões: %v", err)
	}
	logger.Info("papéis e permissões sincronizados", "changed", synced.Changed)

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: httpapi.NewRouter(httpapi.Deps{
			Ping: database.Ping(db), Identity: mod, AuditQuery: &audit.Query{Authz: authorizer, DB: db}, Log: logger,
			AllowedOrigins: cfg.AllowedOrigins, Cookie: httpx.SessionCookie{Domain: cfg.CookieDomain, Secure: cfg.CookieSecure},
		}),
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
