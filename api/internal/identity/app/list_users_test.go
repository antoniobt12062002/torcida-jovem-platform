//go:build integration

package app_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func (e *ucEnv) list() *app.ListUsers { return &app.ListUsers{Authz: e.authorizer, Users: e.users} }

func TestListUsersRequiresIdentityUserRead(t *testing.T) {
	e := newUCEnv(t)
	_, plain := e.actor(t)

	_, err := e.list().Execute(e.ctx, app.ListInput{Actor: plain})

	if !errors.Is(err, authz.ErrForbidden) || len(e.events(t, "authz.denied")) != 1 {
		t.Errorf("err = %v", err)
	}
}

// IDN-04.8: cursor, 50 por padrão, no máximo 100.
func TestListUsersPagesWithCursorAndDefaultsTo50(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RoleDiretoria)
	for range 6 {
		e.actor(t)
	}

	first, err := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Limit: 3})
	if err != nil || len(first.Items) != 3 || first.NextCursor == "" {
		t.Fatalf("primeira página = %d itens, cursor %q, err %v", len(first.Items), first.NextCursor, err)
	}
	second, _ := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Limit: 3, Cursor: first.NextCursor})
	third, _ := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Limit: 3, Cursor: second.NextCursor})

	var ids []string
	for _, p := range []app.UserPage{first, second, third} {
		for _, it := range p.Items {
			ids = append(ids, it.ID)
		}
	}
	if len(ids) != 7 || len(slices.Compact(slices.Sorted(slices.Values(ids)))) != 7 {
		t.Errorf("deveria listar os 7 usuários sem repetição: %d", len(ids))
	}
	if third.NextCursor != "" {
		t.Errorf("a última página não tem próximo cursor: %q", third.NextCursor)
	}
	all, err := e.list().Execute(e.ctx, app.ListInput{Actor: admin})
	if err != nil || len(all.Items) != 7 || all.NextCursor != "" {
		t.Errorf("limite padrão: %d itens, cursor %q, err %v", len(all.Items), all.NextCursor, err)
	}
}

func TestListUsersRejectsInvalidLimitsCursorsAndRoles(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RoleDiretoria)

	for _, in := range []app.ListInput{
		{Actor: admin, Limit: 101}, {Actor: admin, Limit: -1},
	} {
		if _, err := e.list().Execute(e.ctx, in); !errors.Is(err, domain.ErrInvalidLimit) {
			t.Errorf("limit %d: %v", in.Limit, err)
		}
	}
	if _, err := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Limit: 100}); err != nil {
		t.Errorf("100 é o máximo aceito: %v", err)
	}
	if _, err := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Cursor: "lixo"}); !errors.Is(err, domain.ErrInvalidCursor) {
		t.Errorf("cursor inválido: %v", err)
	}
	bad := domain.Role("CHEFE")
	if _, err := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Role: &bad}); !errors.Is(err, domain.ErrUnknownRole) {
		t.Errorf("papel desconhecido: %v", err)
	}
}

func TestListUsersFiltersAndShowsRolesAndMembershipButNeverSecrets(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RoleDiretoria)
	tes, _ := e.actor(t, domain.RoleTesouraria)
	off, _ := e.actor(t)
	_ = e.users.SetActive(context.Background(), off.ID, false)
	no := false
	role := domain.RoleTesouraria

	byRole, err := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Role: &role})
	if err != nil || len(byRole.Items) != 1 || byRole.Items[0].ID != tes.ID || byRole.Items[0].AdminMembership == nil {
		t.Fatalf("por papel = %+v, %v", byRole, err)
	}
	inactive, _ := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Active: &no})
	if len(inactive.Items) != 1 || inactive.Items[0].ID != off.ID {
		t.Errorf("inativos = %+v", inactive.Items)
	}
	typ := reflect.TypeOf(domain.UserSummary{})
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		if name == "PasswordHash" || name == "Token" || name == "Hash" {
			t.Errorf("UserSummary não pode expor %s", name)
		}
	}
}

func TestListUsersRequiresAllDependencies(t *testing.T) {
	if _, err := (&app.ListUsers{}).Execute(context.Background(), app.ListInput{}); err == nil || errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v", err)
	}
}

func TestListUsersDefaultPageHas50ItemsAndAPageThatEndsExactlyHasNoNextCursor(t *testing.T) {
	e := newUCEnv(t)
	_, admin := e.actor(t, domain.RoleDiretoria)
	for i := range 54 {
		if _, err := e.users.Create(context.Background(), domain.User{Email: "massa" + itoa(i) + "@exemplo.com", Name: "M", PasswordHash: "h", Active: true}); err != nil {
			t.Fatal(err)
		}
	}

	page, err := e.list().Execute(e.ctx, app.ListInput{Actor: admin})
	if err != nil || len(page.Items) != 50 || page.NextCursor == "" {
		t.Fatalf("o padrão são 50 por página: %d itens, cursor %q, err %v", len(page.Items), page.NextCursor, err)
	}
	rest, _ := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Cursor: page.NextCursor})
	if len(rest.Items) != 5 || rest.NextCursor != "" {
		t.Errorf("resto = %d itens, cursor %q", len(rest.Items), rest.NextCursor)
	}

	// 55 usuários no total: com 55 por página o fim coincide exatamente com o limite.
	exact, _ := e.list().Execute(e.ctx, app.ListInput{Actor: admin, Limit: 55})
	if len(exact.Items) != 55 || exact.NextCursor != "" {
		t.Errorf("página exata = %d itens, cursor %q", len(exact.Items), exact.NextCursor)
	}
}
