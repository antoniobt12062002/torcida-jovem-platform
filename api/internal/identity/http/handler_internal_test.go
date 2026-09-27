package identityhttp

import (
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
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
