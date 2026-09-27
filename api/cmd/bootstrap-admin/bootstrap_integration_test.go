//go:build integration

package main

import (
	"context"
	"errors"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

// IDN-01.1: num banco recém-migrado, em que a API ainda não subiu (os papéis não foram
// sincronizados), o comando cria o administrador, com o vínculo e os eventos.
func TestBootstrapAdminWorksOnAFreshDatabaseBeforeTheAPIEverStarted(t *testing.T) {
	db := testutil.NewTestDB(t)
	owner := testutil.OwnerDBFor(t, db)
	dsn := testutil.AppDSN(t, db)
	in := app.BootstrapInput{Email: "admin@exemplo.com", Name: "Administrador", Password: "senha-do-primeiro-admin-1", Role: domain.RoleAdminSistema}

	id, err := bootstrapAdmin(context.Background(), dsn, in)

	if err != nil || id == "" {
		t.Fatalf("id = %q, erro = %v", id, err)
	}
	var roles []string
	owner.Raw(`SELECT r.name FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id = ?::uuid ORDER BY r.name`, id).Scan(&roles)
	if len(roles) != 2 || roles[0] != "ADMIN_SISTEMA" || roles[1] != "ASSOCIADO" {
		t.Errorf("papéis = %v", roles)
	}
	for _, action := range []string{"user.bootstrap", "rbac.sync"} {
		var n int64
		owner.Raw("SELECT count(*) FROM audit_log WHERE action = ?", action).Scan(&n)
		if n != 1 {
			t.Errorf("%s: %d eventos, esperado 1", action, n)
		}
	}
}

// IDN-01.2: o segundo bootstrap é recusado e nada é criado.
func TestBootstrapAdminRefusesASecondRun(t *testing.T) {
	db := testutil.NewTestDB(t)
	dsn := testutil.AppDSN(t, db)
	in := app.BootstrapInput{Email: "admin@exemplo.com", Name: "Administrador", Password: "senha-do-primeiro-admin-1"}
	if _, err := bootstrapAdmin(context.Background(), dsn, in); err != nil {
		t.Fatal(err)
	}

	in.Email = "outro@exemplo.com"
	_, err := bootstrapAdmin(context.Background(), dsn, in)

	if !errors.Is(err, app.ErrBootstrapAlreadyDone) {
		t.Errorf("err = %v", err)
	}
}
