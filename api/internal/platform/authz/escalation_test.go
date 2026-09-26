package authz

import (
	"slices"
	"testing"
)

// RBAC-03.1 e RBAC-03.2: só se concede ou retira o que se possui.
func TestCoversReturnsNothingWhenThePrincipalHoldsEveryNeededPermission(t *testing.T) {
	p := principalWith("a:b:read", "a:b:create", "c:d:approve")

	if missing := Covers(p, []Permission{"a:b:read", "c:d:approve"}); len(missing) != 0 {
		t.Errorf("missing = %v", missing)
	}
}

func TestCoversReturnsExactlyTheMissingPermissionsInOrder(t *testing.T) {
	p := principalWith("a:b:read")

	got := Covers(p, []Permission{"z:z:delete", "a:b:read", "c:d:approve", "b:b:update"})

	want := []Permission{"b:b:update", "c:d:approve", "z:z:delete"}
	if !slices.Equal(got, want) {
		t.Errorf("missing = %v, esperado %v", got, want)
	}
}

func TestCoversWithNoNeededPermissionsAlwaysPasses(t *testing.T) {
	for _, p := range []Principal{{}, principalWith(), principalWith("a:b:read")} {
		if missing := Covers(p, nil); len(missing) != 0 {
			t.Errorf("Covers(%+v, nil) = %v", p, missing)
		}
		if missing := Covers(p, []Permission{}); len(missing) != 0 {
			t.Errorf("Covers(%+v, vazio) = %v", p, missing)
		}
	}
}

func TestCoversListsARepeatedMissingPermissionOnce(t *testing.T) {
	got := Covers(principalWith(), []Permission{"a:b:read", "a:b:read"})

	if !slices.Equal(got, []Permission{"a:b:read"}) {
		t.Errorf("missing = %v", got)
	}
}

// Só as permissões do Principal recebido contam: nome de papel não vale.
func TestCoversIgnoresRoleNames(t *testing.T) {
	p := Principal{UserID: "u-1", Roles: []string{"PRESIDENTE"}}

	if got := Covers(p, []Permission{"a:b:read"}); len(got) != 1 {
		t.Errorf("missing = %v", got)
	}
}
