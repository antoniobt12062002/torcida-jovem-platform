//go:build integration

package app_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func (e *ucEnv) assign() *app.AssignRoles {
	return &app.AssignRoles{
		Authz: e.authorizer, Users: e.users, Members: e.members, Roles: e.roles, Perms: e.roles, Matrix: e.matrix, Audit: e.rec, Tx: e.tx,
	}
}

func assignIn(actor authz.Principal, target string, roles ...domain.Role) app.AssignInput {
	return app.AssignInput{Actor: actor, TargetID: target, Roles: roles}
}

func sortedRoles(r []domain.Role) []domain.Role {
	out := slices.Clone(r)
	slices.Sort(out)
	return out
}

// IDN-04.7: substitui o conjunto, audita e vale na requisição seguinte.
func TestAssignRolesReplacesTheSetKeepsAssociadoAndAudits(t *testing.T) {
	e := newUCEnv(t)
	presidenteU, presidente := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)
	e.sessionFor(t, target.ID)

	err := e.assign().Execute(e.ctx, assignIn(presidente, target.ID, domain.RoleEventos, domain.RoleDiretoria, domain.RoleEventos))

	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Role{domain.RoleAssociado, domain.RoleDiretoria, domain.RoleEventos}
	if got := sortedRoles(e.rolesOf(t, target.ID)); !slices.Equal(got, want) {
		t.Errorf("papéis = %v, esperado %v (ASSOCIADO mantido, repetições ignoradas, vários papéis)", got, want)
	}
	if !e.principal(t, target.ID).Has("identity:user:read") {
		t.Error("as novas permissões valem na requisição seguinte")
	}
	if e.activeSessions(t, target.ID) != 1 {
		t.Error("trocar papéis não revoga sessões: a permissão é lida a cada requisição")
	}
	evs := e.events(t, "user.roles_set")
	if len(evs) != 1 || evs[0].Outcome != "success" || evs[0].EntityID != target.ID || evs[0].ActorID == nil || *evs[0].ActorID != presidenteU.ID ||
		evs[0].Before == nil || !strings.Contains(*evs[0].Before, "TESOURARIA") || evs[0].After == nil || !strings.Contains(*evs[0].After, "EVENTOS") {
		t.Errorf("eventos = %+v", evs)
	}
}

func TestAnEmptyListLeavesTheMembershipDormantWithOnlyAssociado(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)

	if err := e.assign().Execute(e.ctx, assignIn(presidente, target.ID)); err != nil {
		t.Fatal(err)
	}

	if got := e.rolesOf(t, target.ID); len(got) != 1 || got[0] != domain.RoleAssociado {
		t.Errorf("papéis = %v", got)
	}
	if _, ok, _ := e.members.Active(context.Background(), target.ID); !ok {
		t.Error("o vínculo continua ativo (dormente): quem o encerra é RevokeAdmin")
	}
}

func TestAssignRolesChecksThePermissionFirst(t *testing.T) {
	e := newUCEnv(t)
	_, diretoria := e.actor(t, domain.RoleDiretoria)
	target, _ := e.actor(t, domain.RoleTesouraria)

	err := e.assign().Execute(e.ctx, assignIn(diretoria, target.ID, domain.RoleEventos))

	if !errors.Is(err, authz.ErrForbidden) || len(e.events(t, "authz.denied")) != 1 {
		t.Errorf("err = %v", err)
	}
}

// RBAC-03.3
func TestNobodyChangesTheirOwnRoles(t *testing.T) {
	e := newUCEnv(t)
	u, presidente := e.actor(t, domain.RolePresidente)

	err := e.assign().Execute(e.ctx, assignIn(presidente, u.ID, domain.RoleEventos))

	if !errors.Is(err, domain.ErrSelfChangeForbidden) {
		t.Fatalf("err = %v", err)
	}
	if evs := e.events(t, "role.change_denied"); len(evs) != 1 || evs[0].Outcome != "denied" {
		t.Errorf("eventos = %+v", evs)
	}
}

// RBAC-03.1 e .2: adicionar ou remover um papel com permissões que o ator não tem.
func TestAssignRolesRefusesAddingOrRemovingRolesBeyondTheActorsPermissions(t *testing.T) {
	e := newUCEnv(t)
	_, adminSistema := e.actor(t, domain.RoleAdminSistema)
	comEventos, _ := e.actor(t, domain.RoleEventos)
	conselheiro, _ := e.actor(t, domain.RoleConselhoFiscal, domain.RoleEventos)

	err := e.assign().Execute(e.ctx, assignIn(adminSistema, comEventos.ID, domain.RoleEventos, domain.RoleConselhoFiscal))
	if !errors.Is(err, domain.ErrPrivilegeEscalation) {
		t.Errorf("adicionar CONSELHO_FISCAL: %v", err)
	}
	err = e.assign().Execute(e.ctx, assignIn(adminSistema, conselheiro.ID, domain.RoleEventos))
	if !errors.Is(err, domain.ErrPrivilegeEscalation) {
		t.Errorf("remover CONSELHO_FISCAL: %v", err)
	}
	if got := sortedRoles(e.rolesOf(t, conselheiro.ID)); !slices.Contains(got, domain.RoleConselhoFiscal) {
		t.Errorf("os papéis não podem ter mudado: %v", got)
	}
	if evs := e.events(t, "role.change_denied"); len(evs) != 2 || evs[0].Outcome != "denied" {
		t.Errorf("eventos = %+v", evs)
	}
}

// IDN-06.9
func TestAssignRolesRequiresAnActiveMembershipAndValidRoles(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	plain, _ := e.actor(t) // ASSOCIADO, sem vínculo
	admin, _ := e.actor(t, domain.RoleTesouraria)

	if err := e.assign().Execute(e.ctx, assignIn(presidente, plain.ID, domain.RoleEventos)); !errors.Is(err, domain.ErrAdminMembershipRequired) {
		t.Errorf("sem vínculo: %v", err)
	}
	if err := e.assign().Execute(e.ctx, assignIn(presidente, admin.ID, "CHEFE")); !errors.Is(err, domain.ErrUnknownRole) {
		t.Errorf("papel desconhecido: %v", err)
	}
	if err := e.assign().Execute(e.ctx, assignIn(presidente, "0f8fad5b-d9cb-469f-a165-70867728950e", domain.RoleEventos)); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("alvo inexistente: %v", err)
	}
	inactive, _ := e.actor(t, domain.RoleTesouraria)
	_ = e.users.SetActive(context.Background(), inactive.ID, false)
	if err := e.assign().Execute(e.ctx, assignIn(presidente, inactive.ID, domain.RoleEventos)); !errors.Is(err, domain.ErrUserInactive) {
		t.Errorf("alvo inativo: %v", err)
	}
	if got := e.rolesOf(t, plain.ID); len(got) != 1 {
		t.Errorf("papéis = %v", got)
	}
}

func TestAssignRolesWithNoChangeIsANoOp(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)

	if err := e.assign().Execute(e.ctx, assignIn(presidente, target.ID, domain.RoleTesouraria, domain.RoleAssociado)); err != nil {
		t.Errorf("err = %v", err)
	}
	if len(e.events(t, "user.roles_set")) != 0 {
		t.Error("um no-op não audita")
	}
}

func TestAssignRolesNeverRemovesTheLastAdminAccessHolder(t *testing.T) {
	e := newUCEnv(t)
	only, _ := e.actor(t, domain.RoleAdminSistema)
	actor := e.withPermsOf(t, domain.RolePresidente)

	err := e.assign().Execute(e.ctx, assignIn(actor, only.ID, domain.RoleTesouraria))

	if !errors.Is(err, domain.ErrLastAdmin) {
		t.Fatalf("err = %v", err)
	}
	if got := e.rolesOf(t, only.ID); !slices.Contains(got, domain.RoleAdminSistema) {
		t.Errorf("papéis = %v", got)
	}
	if err := e.assign().Execute(e.ctx, assignIn(actor, only.ID, domain.RoleAdminSistema, domain.RoleTesouraria)); err != nil {
		t.Errorf("manter identity:admin:grant é permitido: %v", err)
	}
}

func TestSimultaneousRoleChangesNeverLeaveTheSystemWithoutAnAdministrator(t *testing.T) {
	e := newUCEnv(t)
	a, _ := e.actor(t, domain.RoleAdminSistema)
	b, _ := e.actor(t, domain.RoleAdminSistema)
	actor := e.withPermsOf(t, domain.RolePresidente)
	errs := make([]error, 2)

	var wg sync.WaitGroup
	for i, id := range []string{a.ID, b.ID} {
		wg.Go(func() { errs[i] = e.assign().Execute(e.ctx, assignIn(actor, id, domain.RoleTesouraria)) })
	}
	wg.Wait()

	ok, last := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, domain.ErrLastAdmin):
			last++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if ok != 1 || last != 1 {
		t.Errorf("sucessos = %d, last_admin = %d", ok, last)
	}
}

// AUD-01.2
func TestAssignRolesIsRolledBackWhenTheAuditFails(t *testing.T) {
	e := newUCEnv(t)
	_, presidente := e.actor(t, domain.RolePresidente)
	target, _ := e.actor(t, domain.RoleTesouraria)
	e.breakAudit(t)

	err := e.assign().Execute(e.ctx, assignIn(presidente, target.ID, domain.RoleEventos))

	if !errors.Is(err, audit.ErrWrite) {
		t.Errorf("err = %v", err)
	}
	if got := sortedRoles(e.rolesOf(t, target.ID)); !slices.Contains(got, domain.RoleTesouraria) || slices.Contains(got, domain.RoleEventos) {
		t.Errorf("os papéis deveriam continuar como estavam: %v", got)
	}
}

func TestAssignRolesRequiresAllDependencies(t *testing.T) {
	if err := (&app.AssignRoles{}).Execute(context.Background(), app.AssignInput{}); err == nil || errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v", err)
	}
}
