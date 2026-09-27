package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

const secret = "senha-secreta-do-admin-1"

type harness struct {
	out, errOut bytes.Buffer
	calls       []app.BootstrapInput
	dsn         string
	result      error
	env         map[string]string
}

func newHarness() *harness {
	return &harness{env: map[string]string{"BOOTSTRAP_ADMIN_PASSWORD": secret, "DATABASE_URL": "postgres://x"}}
}

func (h *harness) run(args ...string) int {
	return run(args, deps{
		getenv: func(k string) string { return h.env[k] },
		stdout: &h.out, stderr: &h.errOut,
		bootstrap: func(_ context.Context, dsn string, in app.BootstrapInput) (string, error) {
			h.calls = append(h.calls, in)
			h.dsn = dsn
			return "0f8fad5b-d9cb-469f-a165-70867728950e", h.result
		},
	})
}

func (h *harness) everything() string { return h.out.String() + h.errOut.String() }

// IDN-01.1: e-mail, nome e o papel padrão ADMIN_SISTEMA; a senha vem do ambiente.
func TestRunCreatesTheAdministratorWithTheDefaultRoleAndThePasswordFromTheEnvironment(t *testing.T) {
	h := newHarness()

	code := h.run("--email", "admin@exemplo.com", "--name", "Administrador")

	if code != 0 || len(h.calls) != 1 {
		t.Fatalf("código = %d, chamadas = %d, saída = %s", code, len(h.calls), h.everything())
	}
	got := h.calls[0]
	if got.Email != "admin@exemplo.com" || got.Name != "Administrador" || got.Role != domain.RoleAdminSistema || got.Password != secret {
		t.Errorf("entrada = %+v", got)
	}
	if h.dsn != "postgres://x" {
		t.Errorf("dsn = %q", h.dsn)
	}
	if !strings.Contains(h.out.String(), "admin@exemplo.com") || !strings.Contains(h.out.String(), "ADMIN_SISTEMA") {
		t.Errorf("a saída deveria confirmar o e-mail e o papel: %s", h.out.String())
	}
}

// IDN-01.1: --role aceita PRESIDENTE.
func TestRunAcceptsThePresidenteRole(t *testing.T) {
	h := newHarness()

	code := h.run("--email", "a@exemplo.com", "--name", "Ana", "--role", "PRESIDENTE")

	if code != 0 || len(h.calls) != 1 || h.calls[0].Role != domain.RolePresidente {
		t.Errorf("código = %d, chamadas = %+v", code, h.calls)
	}
}

// IDN-01.1: qualquer outro papel é recusado antes de tocar o banco.
func TestRunRefusesAnyOtherRole(t *testing.T) {
	for _, role := range []string{"TESOURARIA", "ASSOCIADO", "admin_sistema", "", "QUALQUER"} {
		h := newHarness()

		code := h.run("--email", "a@exemplo.com", "--name", "Ana", "--role", role)

		if code == 0 || len(h.calls) != 0 {
			t.Errorf("papel %q: código = %d, chamadas = %d", role, code, len(h.calls))
		}
	}
}

// IDN-01.4: não existe opção de senha na linha de comando, e o valor nunca é ecoado.
func TestRunHasNoPasswordFlagAndNeverEchoesTheValue(t *testing.T) {
	h := newHarness()
	h.env["BOOTSTRAP_ADMIN_PASSWORD"] = ""

	for _, flag := range []string{"--password", "-password", "--senha", "--pass"} {
		code := h.run("--email", "a@exemplo.com", "--name", "Ana", flag, secret)
		if code == 0 || len(h.calls) != 0 {
			t.Errorf("%s: código = %d, chamadas = %d", flag, code, len(h.calls))
		}
	}
	if strings.Contains(h.everything(), secret) {
		t.Errorf("a senha passada na linha de comando não pode ser ecoada: %s", h.everything())
	}
}

// IDN-01.1 e .3: sem BOOTSTRAP_ADMIN_PASSWORD sai com código diferente de zero, sem chamar o caso de uso.
func TestRunFailsWithoutThePasswordVariable(t *testing.T) {
	h := newHarness()
	h.env["BOOTSTRAP_ADMIN_PASSWORD"] = ""

	code := h.run("--email", "a@exemplo.com", "--name", "Ana")

	if code == 0 || len(h.calls) != 0 || !strings.Contains(h.errOut.String(), "BOOTSTRAP_ADMIN_PASSWORD") {
		t.Errorf("código = %d, chamadas = %d, erro = %s", code, len(h.calls), h.errOut.String())
	}
}

func TestRunFailsWithoutEmailNameOrDatabase(t *testing.T) {
	cases := map[string]struct {
		args []string
		env  map[string]string
		want string
	}{
		"sem e-mail":   {[]string{"--name", "Ana"}, nil, "--email"},
		"sem nome":     {[]string{"--email", "a@exemplo.com"}, nil, "--name"},
		"sem database": {[]string{"--email", "a@exemplo.com", "--name", "Ana"}, map[string]string{"DATABASE_URL": ""}, "DATABASE_URL"},
	}
	for name, tc := range cases {
		h := newHarness()
		for k, v := range tc.env {
			h.env[k] = v
		}

		code := h.run(tc.args...)

		if code == 0 || len(h.calls) != 0 || !strings.Contains(h.errOut.String(), tc.want) {
			t.Errorf("%s: código = %d, chamadas = %d, erro = %s", name, code, len(h.calls), h.errOut.String())
		}
	}
}

// IDN-01.2 e .3: a recusa do caso de uso vira código diferente de zero, com mensagem, sem a senha.
func TestRunTurnsAUseCaseRefusalIntoANonZeroExitWithoutLeakingThePassword(t *testing.T) {
	for name, err := range map[string]error{
		"já existe admin": app.ErrBootstrapAlreadyDone,
		"senha fraca":     &domain.PasswordError{Violation: domain.PasswordTooShort, MinLength: 10},
		"erro interno":    errors.New("falha com " + secret),
	} {
		h := newHarness()
		h.result = err

		code := h.run("--email", "a@exemplo.com", "--name", "Ana")

		if code == 0 {
			t.Errorf("%s: deveria sair com código diferente de zero", name)
		}
		if strings.Contains(h.everything(), secret) {
			t.Errorf("%s: a saída não pode conter a senha: %s", name, h.everything())
		}
		if h.errOut.Len() == 0 {
			t.Errorf("%s: deveria explicar a recusa", name)
		}
	}
}

func TestRunExplainsWhyItRefused(t *testing.T) {
	h := newHarness()
	h.result = app.ErrBootstrapAlreadyDone

	h.run("--email", "a@exemplo.com", "--name", "Ana")

	if !strings.Contains(h.errOut.String(), "já existe um administrador ativo") {
		t.Errorf("erro = %s", h.errOut.String())
	}
}
