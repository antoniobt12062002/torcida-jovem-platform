//go:build integration

package audit_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func hookAuthorizer(t *testing.T, e env) *authz.Authorizer {
	t.Helper()
	a, err := authz.NewAuthorizer([]authz.Definition{
		{Permission: "identity:user:create"},
		{Permission: "associados:cadastro:read", CommonRead: true},
	}, audit.DeniedHook(e.rec))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func principal(perms ...authz.Permission) authz.Principal {
	p := authz.Principal{UserID: actorID, Permissions: map[authz.Permission]struct{}{}}
	for _, perm := range perms {
		p.Permissions[perm] = struct{}{}
	}
	return p
}

// RBAC-02.7 e AUD-04.6: a negação vira evento de segurança.
func TestDeniedHookRecordsAuthzDeniedAsASecurityEvent(t *testing.T) {
	e := newEnv(t)

	err := hookAuthorizer(t, e).Require(ctxWith("req-12345678", ""), principal(), "identity:user:create")

	if !errors.Is(err, authz.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
	rows := e.rows(t)
	if len(rows) != 1 {
		t.Fatalf("linhas = %+v", rows)
	}
	r := rows[0]
	if r.Action != "authz.denied" || r.Outcome != "denied" || r.ActorType != "user" || r.ActorUserID == nil || *r.ActorUserID != actorID ||
		r.EntityType != "permission" || r.EntityID != "identity:user:create" || r.RequestID == nil || *r.RequestID != "req-12345678" {
		t.Errorf("registro = %+v", r)
	}
}

func TestAllowedChecksAndCommonReadDenialsAreNotAudited(t *testing.T) {
	e := newEnv(t)
	a := hookAuthorizer(t, e)

	if err := a.Require(context.Background(), principal("identity:user:create"), "identity:user:create"); err != nil {
		t.Fatal(err)
	}
	if err := a.Require(context.Background(), principal(), "associados:cadastro:read"); !errors.Is(err, authz.ErrForbidden) {
		t.Fatal(err)
	}
	if n := len(e.rows(t)); n != 0 {
		t.Errorf("nada deveria ser auditado, linhas = %d", n)
	}
}

// A falha ao gravar a negação não altera a resposta: continua sendo 403.
func TestDeniedHookFailureDoesNotChangeTheResponseAndIsLogged(t *testing.T) {
	e := newEnv(t)
	if err := e.owner.Exec("ALTER TABLE audit_log RENAME TO audit_log_off").Error; err != nil {
		t.Fatal(err)
	}

	err := hookAuthorizer(t, e).Require(ctxWith("req-12345678", ""), principal(), "identity:user:create")

	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, a resposta deveria continuar sendo a negação", err)
	}
	if out := e.logs.String(); !strings.Contains(out, "audit_security_write_failed") || !strings.Contains(out, "authz.denied") {
		t.Errorf("o incidente deveria estar no log: %s", out)
	}
}
