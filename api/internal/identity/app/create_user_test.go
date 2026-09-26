//go:build integration

package app_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func (e *ucEnv) createUser() *app.CreateUser {
	return &app.CreateUser{Authz: e.authorizer, Users: e.users, Audit: e.rec, Hasher: e.hasher, Denylist: e.deny, Tx: e.tx}
}

func input(actor authz.Principal, email, name, pw string) app.CreateUserInput {
	return app.CreateUserInput{Actor: actor, Email: email, Name: name, Password: pw}
}

// IDN-04.1 e IDN-05.8: conta só com ASSOCIADO e troca de senha obrigatória.
func TestCreateUserMakesAnAssociadoWithMustChangePasswordAndAudits(t *testing.T) {
	e := newUCEnv(t)
	admin, actor := e.actor(t, domain.RoleAdminSistema)

	u, err := e.createUser().Execute(e.ctx, input(actor, "  Nova@Exemplo.com ", "  Nova Pessoa ", "uma-senha-inicial-1"))

	if err != nil {
		t.Fatal(err)
	}
	if u.ID == "" || u.Email != "nova@exemplo.com" || u.Name != "Nova Pessoa" || !u.Active || !u.MustChangePassword {
		t.Errorf("usuário = %+v", u)
	}
	if got := e.rolesOf(t, u.ID); len(got) != 1 || got[0] != domain.RoleAssociado {
		t.Errorf("papéis = %v, esperado só ASSOCIADO", got)
	}
	stored, _ := e.users.FindByID(e.ctx, u.ID)
	if ok, _ := e.hasher.Verify("uma-senha-inicial-1", stored.PasswordHash); !ok {
		t.Error("a senha inicial deveria estar gravada como hash argon2id")
	}
	evs := e.events(t, "user.create")
	if len(evs) != 1 || evs[0].ActorID == nil || *evs[0].ActorID != admin.ID || evs[0].EntityID != u.ID || evs[0].Outcome != "success" {
		t.Fatalf("eventos = %+v", evs)
	}
	if evs[0].After == nil || strings.Contains(*evs[0].After, "argon2") || strings.Contains(*evs[0].After, "uma-senha") {
		t.Errorf("a auditoria não pode conter senha nem hash: %v", evs[0].After)
	}
}

// RBAC-02.3: a permissão é verificada antes de qualquer leitura ou escrita.
func TestCreateUserChecksThePermissionBeforeAnythingElse(t *testing.T) {
	e := newUCEnv(t)
	_, noPerm := e.actor(t, domain.RoleDiretoria) // só identity:user:read

	_, err := e.createUser().Execute(e.ctx, input(noPerm, "não-é-um-email", "", "x"))

	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperava ErrForbidden antes de validar a entrada", err)
	}
	if e.count(t, "users") != 1 { // só o ator
		t.Error("nada pode ter sido escrito")
	}
	if len(e.events(t, "authz.denied")) != 1 {
		t.Error("a negação deveria ser auditada")
	}
}

// IDN-04.2 e a borda: e-mail repetido.
func TestCreateUserRejectsARepeatedEmailAndLeavesNoAudit(t *testing.T) {
	e := newUCEnv(t)
	_, actor := e.actor(t, domain.RoleAdminSistema)
	uc := e.createUser()
	if _, err := uc.Execute(e.ctx, input(actor, "nova@exemplo.com", "Nova", "uma-senha-inicial-1")); err != nil {
		t.Fatal(err)
	}

	_, err := uc.Execute(e.ctx, input(actor, "NOVA@exemplo.com", "Outra", "uma-senha-inicial-1"))

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("err = %v", err)
	}
	if n := len(e.events(t, "user.create")); n != 1 {
		t.Errorf("eventos user.create = %d, a tentativa repetida não pode auditar", n)
	}
}

// IDN-05.2, .3 e .4: política de senha do ASSOCIADO.
func TestCreateUserAppliesTheAssociadoPasswordPolicy(t *testing.T) {
	e := newUCEnv(t)
	_, actor := e.actor(t, domain.RoleAdminSistema)

	cases := map[string]domain.PasswordViolation{
		"curta1":                 domain.PasswordTooShort,
		strings.Repeat("a", 129): domain.PasswordTooLong,
		"12345678":               domain.PasswordCompromised,
		"SenhaComum1234":         domain.PasswordCompromised,
	}
	for pw, want := range cases {
		_, err := e.createUser().Execute(e.ctx, input(actor, "x"+itoa(len(pw))+"@exemplo.com", "Nome", pw))
		pe, ok := errors.AsType[*domain.PasswordError](err)
		if !ok || pe.Violation != want {
			t.Errorf("senha %.12q: err = %v, esperava %s", pw, err, want)
		}
	}
	if _, err := e.createUser().Execute(e.ctx, input(actor, "ok8@exemplo.com", "Nome", "8chars!!")); err != nil {
		t.Errorf("8 caracteres bastam para ASSOCIADO: %v", err)
	}
}

func TestCreateUserValidatesEmailAndName(t *testing.T) {
	e := newUCEnv(t)
	_, actor := e.actor(t, domain.RoleAdminSistema)

	for _, tc := range []struct {
		email, name string
		want        error
	}{
		{"isto não é e-mail", "Nome", domain.ErrInvalidEmail},
		{"a@exemplo.com", "   ", domain.ErrInvalidName},
		{"b@exemplo.com", strings.Repeat("n", 121), domain.ErrInvalidName},
		{"c@exemplo.com", "Nome\nQuebra", domain.ErrInvalidName},
	} {
		if _, err := e.createUser().Execute(e.ctx, input(actor, tc.email, tc.name, "uma-senha-inicial-1")); !errors.Is(err, tc.want) {
			t.Errorf("%q/%.10q: err = %v, esperava %v", tc.email, tc.name, err, tc.want)
		}
	}
}

// AUD-01.2: se a auditoria falha, o usuário não é criado.
func TestCreateUserIsRolledBackWhenTheAuditFails(t *testing.T) {
	e := newUCEnv(t)
	_, actor := e.actor(t, domain.RoleAdminSistema)
	e.breakAudit(t)

	_, err := e.createUser().Execute(e.ctx, input(actor, "nova@exemplo.com", "Nova", "uma-senha-inicial-1"))

	if !errors.Is(err, audit.ErrWrite) {
		t.Errorf("err = %v, esperava ErrWrite (audit_failed)", err)
	}
	if e.count(t, "users") != 1 {
		t.Errorf("o usuário deveria ter sido revertido, usuários = %d", e.count(t, "users"))
	}
}

func TestCreateUserRequiresAllDependencies(t *testing.T) {
	if _, err := (&app.CreateUser{}).Execute(t.Context(), app.CreateUserInput{}); err == nil || errors.Is(err, authz.ErrForbidden) {
		t.Errorf("um caso de uso mal montado deveria falhar com erro de configuração, veio %v", err)
	}
}
