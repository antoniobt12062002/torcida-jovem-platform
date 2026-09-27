// Command bootstrap-admin creates the first administrator of the platform.
//
// The password comes only from the environment variable BOOTSTRAP_ADMIN_PASSWORD:
// there is no flag for it, so it never shows up in the process list or in the
// shell history. The command refuses to run while an active administrative
// membership exists.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

const passwordVariable = "BOOTSTRAP_ADMIN_PASSWORD"

// deps are the command's edges, so the tests need no database.
type deps struct {
	getenv    func(string) string
	stdout    io.Writer
	stderr    io.Writer
	bootstrap func(ctx context.Context, dsn string, in app.BootstrapInput) (userID string, err error)
}

func main() {
	os.Exit(run(os.Args[1:], deps{getenv: os.Getenv, stdout: os.Stdout, stderr: os.Stderr, bootstrap: bootstrapAdmin}))
}

func run(args []string, d deps) int {
	fs := flag.NewFlagSet("bootstrap-admin", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // the usage text is ours; flag errors never echo argument values
	email := fs.String("email", "", "e-mail do administrador (obrigatório)")
	name := fs.String("name", "", "nome do administrador (obrigatório)")
	role := fs.String("role", string(domain.RoleAdminSistema), "papel: ADMIN_SISTEMA (padrão) ou PRESIDENTE")
	if err := fs.Parse(args); err != nil {
		_, _ = fmt.Fprintln(d.stderr, "opções inválidas; use --email, --name e --role (a senha vai só em "+passwordVariable+")")
		return 2
	}

	switch {
	case strings.TrimSpace(*email) == "":
		return fail(d, "informe --email")
	case strings.TrimSpace(*name) == "":
		return fail(d, "informe --name")
	case *role != string(domain.RoleAdminSistema) && *role != string(domain.RolePresidente):
		return fail(d, "--role aceita apenas ADMIN_SISTEMA ou PRESIDENTE")
	}
	pw := d.getenv(passwordVariable)
	if pw == "" {
		return fail(d, "defina a variável de ambiente "+passwordVariable+" (a senha nunca é aceita na linha de comando)")
	}
	dsn := d.getenv("DATABASE_URL")
	if dsn == "" {
		return fail(d, "defina a variável de ambiente DATABASE_URL")
	}

	in := app.BootstrapInput{Email: *email, Name: *name, Password: pw, Role: domain.Role(*role)}
	_, err := d.bootstrap(context.Background(), dsn, in)
	if err != nil {
		return fail(d, explain(err, pw))
	}
	_, _ = fmt.Fprintf(d.stdout, "Administrador criado: %s (%s). A troca de senha é obrigatória no primeiro acesso.\n", strings.ToLower(strings.TrimSpace(*email)), *role)
	return 0
}

func fail(d deps, msg string) int {
	_, _ = fmt.Fprintln(d.stderr, "erro: "+msg)
	return 1
}

// explain turns the failure into a message for the operator, never with the password.
func explain(err error, pw string) string {
	var pwErr *domain.PasswordError
	switch {
	case errors.Is(err, app.ErrBootstrapAlreadyDone):
		return "já existe um administrador ativo; nada foi criado"
	case errors.As(err, &pwErr):
		return "a senha não cumpre a política de administrador (" + string(pwErr.Violation) + "); nada foi criado"
	case errors.Is(err, domain.ErrEmailTaken):
		return "já existe um usuário com este e-mail; nada foi criado"
	case errors.Is(err, domain.ErrInvalidEmail):
		return "e-mail inválido"
	case errors.Is(err, domain.ErrInvalidName):
		return "nome inválido"
	}
	return "falha ao criar o administrador: " + strings.ReplaceAll(err.Error(), pw, "[redacted]")
}

// bootstrapAdmin opens the database as the application role and runs the use case.
func bootstrapAdmin(ctx context.Context, dsn string, in app.BootstrapInput) (string, error) {
	logger := logx.New("warn", os.Stderr)
	db, err := database.Open(dsn, logger)
	if err != nil {
		return "", err
	}
	hasher, err := password.NewHasher(password.DefaultParams())
	if err != nil {
		return "", err
	}
	mod := identity.New(identity.Deps{DB: db, Recorder: audit.NewRecorder(db, logger), Hasher: hasher, Denylist: password.DefaultDenylist(), Log: logger})
	u, err := mod.Bootstrap.Execute(ctx, in)
	if err != nil {
		return "", err
	}
	return u.ID, nil
}
