package identityhttp

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// A resposta lista papéis e permissões em ordem estável, seja qual for a do banco.
func TestRolesAndPermissionsAreListedInAStableOrder(t *testing.T) {
	roles := rolesOut([]domain.Role{domain.RoleTesouraria, domain.RoleAssociado, domain.RolePresidente})
	perms := permissionsOut([]authz.Permission{"identity:user:update", "audit:log:read", "identity:user:read"})

	if !slices.Equal(roles, []Role{"ASSOCIADO", "PRESIDENTE", "TESOURARIA"}) {
		t.Errorf("roles = %v", roles)
	}
	if !slices.Equal(perms, []string{"audit:log:read", "identity:user:read", "identity:user:update"}) {
		t.Errorf("permissions = %v", perms)
	}
}

// Retry-After nunca é zero: menos de um segundo restante vira 1.
func TestRetryAfterIsAtLeastOneSecond(t *testing.T) {
	for in, want := range map[time.Duration]string{0: "1", 400 * time.Millisecond: "1", -time.Second: "1", 90 * time.Second: "90", 15 * time.Minute: "900"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		retryAfter(c, in)

		if got := w.Header().Get("Retry-After"); got != want {
			t.Errorf("retryAfter(%v) = %q, esperado %q", in, got, want)
		}
	}
}

// A tabela única de erros: cada erro de domínio vira o status e o código do contrato.
func TestEveryDomainErrorMapsToItsStatusAndCode(t *testing.T) {
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{app.ErrInvalidCredentials, http.StatusUnauthorized, "invalid_credentials"},
		{app.ErrUnauthenticated, http.StatusUnauthorized, "unauthenticated"},
		{authz.ErrForbidden, http.StatusForbidden, "forbidden"},
		{domain.ErrPrivilegeEscalation, http.StatusForbidden, "privilege_escalation"},
		{domain.ErrSelfChangeForbidden, http.StatusForbidden, "self_change_forbidden"},
		{domain.ErrInvalidCurrentPassword, http.StatusForbidden, "invalid_current_password"},
		{domain.ErrUserNotFound, http.StatusNotFound, "not_found"},
		{domain.ErrEmailTaken, http.StatusConflict, "email_taken"},
		{domain.ErrLastAdmin, http.StatusConflict, "last_admin"},
		{domain.ErrAlreadyAdmin, http.StatusConflict, "already_admin"},
		{domain.ErrNotAdmin, http.StatusConflict, "not_admin"},
		{domain.ErrUserInactive, http.StatusConflict, "user_inactive"},
		{domain.ErrAdminMembershipRequired, http.StatusConflict, "admin_membership_required"},
		{domain.ErrReasonRequired, http.StatusUnprocessableEntity, "reason_required"},
		{domain.ErrAdminRoleRequired, http.StatusUnprocessableEntity, "admin_role_required"},
		{domain.ErrPasswordUnchanged, http.StatusUnprocessableEntity, "password_unchanged"},
		{domain.ErrInvalidLimit, http.StatusUnprocessableEntity, "invalid_limit"},
		{domain.ErrInvalidCursor, http.StatusUnprocessableEntity, "invalid_cursor"},
		{domain.ErrInvalidName, http.StatusUnprocessableEntity, "invalid_name"},
		{domain.ErrInvalidEmail, http.StatusUnprocessableEntity, "invalid_email"},
		{domain.ErrUnknownRole, http.StatusUnprocessableEntity, "unknown_role"},
		{&domain.PasswordError{Violation: domain.PasswordTooShort, MinLength: 10}, http.StatusUnprocessableEntity, "password_too_short"},
		{&domain.PasswordError{Violation: domain.PasswordCompromised}, http.StatusUnprocessableEntity, "password_compromised"},
		{domain.ErrInvalidResetToken, http.StatusBadRequest, "invalid_reset_token"},
		{&app.LockedError{RetryAfter: time.Minute}, http.StatusTooManyRequests, "login_blocked"},
		{&app.PasswordChangeBlockedError{RetryAfter: time.Minute}, http.StatusTooManyRequests, "password_change_blocked"},
		{fmt.Errorf("%w: disco cheio", audit.ErrWrite), http.StatusInternalServerError, "audit_failed"},
		{database.ErrUnavailable, http.StatusServiceUnavailable, "service_unavailable"},
		{fmt.Errorf("abrir conexão dsn=postgres://tj_app:segredo-de-conexao@host: %w", driver.ErrBadConn), http.StatusServiceUnavailable, "service_unavailable"},
		{errors.New("segredo-interno: falha do banco"), http.StatusInternalServerError, "internal_error"},
	}
	h := &Handler{}
	for _, tc := range cases {
		// Erros embrulhados chegam do mesmo jeito que os puros.
		for _, err := range []error{tc.err, fmt.Errorf("caso de uso: %w", tc.err)} {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			h.writeError(c, err)

			var p struct{ Code string }
			_ = json.Unmarshal(w.Body.Bytes(), &p)
			if w.Code != tc.status || p.Code != tc.code || w.Header().Get("Content-Type") != "application/problem+json" {
				t.Errorf("%v: status = %d, code = %q", err, w.Code, p.Code)
			}
			if (tc.status == http.StatusInternalServerError || tc.status == http.StatusServiceUnavailable) &&
				(strings.Contains(w.Body.String(), "segredo-interno") || strings.Contains(w.Body.String(), "disco cheio") || strings.Contains(w.Body.String(), "segredo-de-conexao")) {
				t.Errorf("o corpo não pode vazar detalhe interno: %s", w.Body.String())
			}
		}
	}
}
