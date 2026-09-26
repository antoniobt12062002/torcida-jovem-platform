package authz

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func principalWith(perms ...Permission) Principal {
	p := Principal{UserID: "u-1", Roles: []string{"QUALQUER"}, Permissions: map[Permission]struct{}{}}
	for _, perm := range perms {
		p.Permissions[perm] = struct{}{}
	}
	return p
}

// RBAC-02.2: sem a permissão, 403 forbidden; o erro não revela nada sobre o recurso.
func TestRequireAllowsAHeldPermissionAndDeniesTheOthers(t *testing.T) {
	p := principalWith("identity:user:read")

	if err := Require(p, "identity:user:read"); err != nil {
		t.Errorf("permissão presente foi negada: %v", err)
	}
	err := Require(p, "identity:user:create")
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("err = %v, esperava ErrForbidden", err)
	}
}

func TestForbiddenErrorIsGenericAndDoesNotNameThePermissionOrResource(t *testing.T) {
	err := Require(principalWith(), "financeiro:lancamento:read")

	if msg := err.Error(); strings.Contains(msg, "financeiro") || strings.Contains(msg, "lancamento") {
		t.Errorf("a mensagem revela a permissão ou o recurso: %q", msg)
	}
}

func TestAPrincipalWithNoPermissionsIsDeniedEverything(t *testing.T) {
	for _, p := range []Principal{{}, {UserID: "u-1"}, principalWith()} {
		if err := Require(p, "identity:user:read"); !errors.Is(err, ErrForbidden) {
			t.Errorf("Require(%+v) = %v", p, err)
		}
	}
}

// RBAC-02.8: a decisão usa permissões efetivas, nunca o nome do papel.
func TestRequireIgnoresRoleNames(t *testing.T) {
	p := Principal{UserID: "u-1", Roles: []string{"PRESIDENTE", "ADMIN_SISTEMA"}}

	if err := Require(p, "identity:user:read"); !errors.Is(err, ErrForbidden) {
		t.Errorf("um papel poderoso sem permissões efetivas deve ser negado: %v", err)
	}
}

type denial struct {
	principal  Principal
	permission Permission
}

func newAuthorizer(t *testing.T, defs []Definition, calls *[]denial) *Authorizer {
	t.Helper()
	a, err := NewAuthorizer(defs, func(_ context.Context, p Principal, perm Permission) {
		*calls = append(*calls, denial{p, perm})
	})
	if err != nil {
		t.Fatalf("NewAuthorizer: %v", err)
	}
	return a
}

var testDefs = []Definition{
	{Permission: "identity:user:create"},
	{Permission: "associados:cadastro:read", CommonRead: true},
}

// RBAC-02.7: toda negação chama o gancho, exceto leitura comum e permitido.
func TestAuthorizerCallsTheDenialHookForDeniedNonCommonReads(t *testing.T) {
	var calls []denial
	a := newAuthorizer(t, testDefs, &calls)
	p := principalWith()

	err := a.Require(context.Background(), p, "identity:user:create")

	if !errors.Is(err, ErrForbidden) || len(calls) != 1 || calls[0].permission != "identity:user:create" || calls[0].principal.UserID != "u-1" {
		t.Errorf("err = %v, chamadas = %+v", err, calls)
	}
}

func TestAuthorizerDoesNotCallTheHookWhenAllowed(t *testing.T) {
	var calls []denial
	a := newAuthorizer(t, testDefs, &calls)

	if err := a.Require(context.Background(), principalWith("identity:user:create"), "identity:user:create"); err != nil {
		t.Errorf("err = %v", err)
	}
	if len(calls) != 0 {
		t.Errorf("o gancho não deveria ser chamado: %+v", calls)
	}
}

func TestAuthorizerDeniesCommonReadsWithoutCallingTheHook(t *testing.T) {
	var calls []denial
	a := newAuthorizer(t, testDefs, &calls)

	err := a.Require(context.Background(), principalWith(), "associados:cadastro:read")

	if !errors.Is(err, ErrForbidden) || len(calls) != 0 {
		t.Errorf("err = %v, chamadas = %+v (leitura comum fica só no log de acesso)", err, calls)
	}
}

func TestAuthorizerDeniesAnUnknownPermissionAndAuditsIt(t *testing.T) {
	var calls []denial
	a := newAuthorizer(t, testDefs, &calls)

	err := a.Require(context.Background(), principalWith("financeiro:x:create"), "financeiro:x:create")

	if !errors.Is(err, ErrForbidden) || len(calls) != 1 {
		t.Errorf("uma permissão fora do catálogo deve ser negada e auditada: err = %v, chamadas = %d", err, len(calls))
	}
}

func TestAuthorizerWorksWithoutAHook(t *testing.T) {
	a, err := NewAuthorizer(testDefs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Require(context.Background(), principalWith(), "identity:user:create"); !errors.Is(err, ErrForbidden) {
		t.Errorf("err = %v", err)
	}
}

func TestNewAuthorizerRejectsInvalidAndConflictingDefinitions(t *testing.T) {
	if _, err := NewAuthorizer([]Definition{{Permission: "invalida"}}, nil); err == nil {
		t.Error("definição inválida deveria ser recusada")
	}
	if _, err := NewAuthorizer([]Definition{{Permission: "a:b:read"}, {Permission: "a:b:read", CommonRead: true}}, nil); err == nil {
		t.Error("a mesma permissão com definições diferentes deveria ser recusada")
	}
	if _, err := NewAuthorizer([]Definition{{Permission: "a:b:read"}, {Permission: "a:b:read"}}, nil); err != nil {
		t.Errorf("a mesma definição repetida é inofensiva: %v", err)
	}
}
