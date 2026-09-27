//go:build integration

package app_test

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
)

const bootstrapPassword = "senha-do-primeiro-admin-1"

func (e *ucEnv) bootstrap() *app.BootstrapAdmin {
	return &app.BootstrapAdmin{
		Users: e.users, Members: e.members, Audit: e.rec, Hasher: e.hasher, Denylist: e.deny, Tx: e.tx, Now: e.clock,
	}
}

func bootstrapInput(role domain.Role) app.BootstrapInput {
	return app.BootstrapInput{Email: "Admin@Exemplo.com", Name: "Administrador", Password: bootstrapPassword, Role: role}
}

// IDN-01.1: em banco vazio cria o usuário ativo, com o papel, o vínculo bootstrap sem concedente e a troca de senha obrigatória.
func TestBootstrapOnAnEmptyDatabaseCreatesTheAdministrator(t *testing.T) {
	e := newUCEnv(t)

	u, err := e.bootstrap().Execute(e.ctx, bootstrapInput(""))

	if err != nil {
		t.Fatal(err)
	}
	if u.Email != "admin@exemplo.com" || !u.Active || !u.MustChangePassword {
		t.Errorf("usuário = %+v", u)
	}
	roles := e.rolesOf(t, u.ID)
	if len(roles) != 2 || !slices.Contains(roles, domain.RoleAssociado) || !slices.Contains(roles, domain.RoleAdminSistema) {
		t.Errorf("papéis = %v", roles)
	}
	m, ok, err := e.members.Active(e.ctx, u.ID)
	if err != nil || !ok || m.Reason != "bootstrap-admin" || m.GrantedBy != nil {
		t.Errorf("vínculo = %+v, ativo = %v, erro = %v", m, ok, err)
	}
	perms, _ := e.roles.EffectivePermissions(e.ctx, u.ID)
	if !slices.Contains(perms, "identity:admin:grant") {
		t.Errorf("o administrador deveria poder administrar acessos: %v", perms)
	}
}

// O papel PRESIDENTE é a única alternativa; qualquer outro é recusado sem criar nada.
func TestBootstrapAcceptsPresidenteAndRefusesAnyOtherRole(t *testing.T) {
	e := newUCEnv(t)

	if _, err := e.bootstrap().Execute(e.ctx, bootstrapInput(domain.RoleTesouraria)); !errors.Is(err, domain.ErrUnknownRole) {
		t.Errorf("TESOURARIA: err = %v", err)
	}
	if e.count(t, "users") != 0 {
		t.Fatal("um papel recusado não cria usuário")
	}
	u, err := e.bootstrap().Execute(e.ctx, bootstrapInput(domain.RolePresidente))
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(e.rolesOf(t, u.ID), domain.RolePresidente) {
		t.Errorf("papéis = %v", e.rolesOf(t, u.ID))
	}
}

// IDN-01.2: com um vínculo administrativo ativo, recusa e não cria usuário.
func TestBootstrapRefusesWhenAnActiveMembershipExists(t *testing.T) {
	e := newUCEnv(t)
	e.actor(t, domain.RoleTesouraria)
	before := e.count(t, "users")

	_, err := e.bootstrap().Execute(e.ctx, bootstrapInput(""))

	if !errors.Is(err, app.ErrBootstrapAlreadyDone) {
		t.Fatalf("err = %v", err)
	}
	if e.count(t, "users") != before || len(e.events(t, "user.bootstrap")) != 0 {
		t.Error("nada deveria ser criado nem auditado")
	}
}

// Depois do primeiro bootstrap, o segundo falha.
func TestASecondBootstrapFails(t *testing.T) {
	e := newUCEnv(t)
	if _, err := e.bootstrap().Execute(e.ctx, bootstrapInput("")); err != nil {
		t.Fatal(err)
	}

	in := bootstrapInput("")
	in.Email = "outro@exemplo.com"
	_, err := e.bootstrap().Execute(e.ctx, in)

	if !errors.Is(err, app.ErrBootstrapAlreadyDone) || e.count(t, "users") != 1 {
		t.Errorf("err = %v, usuários = %d", err, e.count(t, "users"))
	}
}

// Um vínculo encerrado não impede o bootstrap (não há administrador ativo).
func TestBootstrapIgnoresClosedMemberships(t *testing.T) {
	e := newUCEnv(t)
	u, _ := e.actor(t, domain.RoleTesouraria)
	if _, err := e.members.Revoke(e.ctx, u.ID, u.ID, "encerrado para o teste", e.now); err != nil {
		t.Fatal(err)
	}

	if _, err := e.bootstrap().Execute(e.ctx, bootstrapInput("")); err != nil {
		t.Errorf("err = %v", err)
	}
}

// IDN-01.1: o evento user.bootstrap não tem ator e traz e-mail, papel e motivo, nunca a senha.
func TestBootstrapAuditsWithoutActorAndNeverStoresThePassword(t *testing.T) {
	e := newUCEnv(t)
	u, err := e.bootstrap().Execute(e.ctx, bootstrapInput(""))
	if err != nil {
		t.Fatal(err)
	}

	evs := e.events(t, "user.bootstrap")

	if len(evs) != 1 {
		t.Fatalf("eventos = %+v", evs)
	}
	ev := evs[0]
	if ev.ActorType != "system" || ev.ActorID != nil || ev.EntityID != u.ID || ev.Outcome != "success" {
		t.Errorf("evento = %+v", ev)
	}
	after := ""
	if ev.After != nil {
		after = *ev.After
	}
	for _, want := range []string{"admin@exemplo.com", "ADMIN_SISTEMA", "bootstrap-admin"} {
		if !strings.Contains(after, want) {
			t.Errorf("after = %s, deveria conter %q", after, want)
		}
	}
	dump := e.dumpAudit(t)
	stored, _ := e.users.FindByID(e.ctx, u.ID)
	for _, secret := range []string{bootstrapPassword, stored.PasswordHash, "argon2"} {
		if strings.Contains(dump, secret) {
			t.Errorf("a auditoria não pode conter %q", secret)
		}
	}
	if strings.Contains(e.logs.String(), bootstrapPassword) {
		t.Error("o log não pode conter a senha")
	}
}

// A auditoria é da mesma transação: se falhar, nada é criado.
func TestBootstrapRollsBackWhenTheAuditFails(t *testing.T) {
	e := newUCEnv(t)
	e.breakAudit(t)

	_, err := e.bootstrap().Execute(e.ctx, bootstrapInput(""))

	if err == nil {
		t.Fatal("a falha de auditoria deveria falhar o bootstrap")
	}
	if e.count(t, "users") != 0 || e.count(t, "admin_memberships") != 0 {
		t.Errorf("nada deveria ficar: %d usuários, %d vínculos", e.count(t, "users"), e.count(t, "admin_memberships"))
	}
}

// IDN-01.3: senha ausente ou fora da política de administrador impede a criação.
func TestBootstrapRefusesAMissingOrWeakPassword(t *testing.T) {
	e := newUCEnv(t)

	for name, pw := range map[string]string{"ausente": "", "curta": "curta123", "sete": "1234567", "comum": "12345678", "só 9": "123456789"} {
		in := bootstrapInput("")
		in.Password = pw
		_, err := e.bootstrap().Execute(e.ctx, in)
		var pe *domain.PasswordError
		if !errors.As(err, &pe) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	if e.count(t, "users") != 0 {
		t.Error("nenhum usuário deveria ser criado")
	}
}

func TestBootstrapValidatesEmailAndNameAndItsOwnWiring(t *testing.T) {
	e := newUCEnv(t)

	bad := bootstrapInput("")
	bad.Email = "isto-nao-e-email"
	if _, err := e.bootstrap().Execute(e.ctx, bad); !errors.Is(err, domain.ErrInvalidEmail) {
		t.Errorf("e-mail: %v", err)
	}
	bad = bootstrapInput("")
	bad.Name = "   "
	if _, err := e.bootstrap().Execute(e.ctx, bad); !errors.Is(err, domain.ErrInvalidName) {
		t.Errorf("nome: %v", err)
	}
	if _, err := (&app.BootstrapAdmin{}).Execute(e.ctx, bootstrapInput("")); err == nil {
		t.Error("um caso de uso mal montado deveria falhar")
	}
}

// Duas execuções simultâneas nunca criam dois administradores.
func TestSimultaneousBootstrapsCreateExactlyOneAdministrator(t *testing.T) {
	e := newUCEnv(t)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i, email := range []string{"a@exemplo.com", "b@exemplo.com"} {
		wg.Go(func() {
			in := bootstrapInput("")
			in.Email = email
			_, errs[i] = e.bootstrap().Execute(e.ctx, in)
		})
	}
	wg.Wait()

	ok, refused := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, app.ErrBootstrapAlreadyDone):
			refused++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if ok != 1 || refused != 1 || e.count(t, "users") != 1 || e.count(t, "admin_memberships") != 1 {
		t.Errorf("criados = %d, recusados = %d, usuários = %d, vínculos = %d", ok, refused, e.count(t, "users"), e.count(t, "admin_memberships"))
	}
}
