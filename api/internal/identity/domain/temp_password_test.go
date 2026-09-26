package domain

import (
	"regexp"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

var tempFormat = regexp.MustCompile(`^[A-HJ-NP-Za-km-z2-9]{4}(-[A-HJ-NP-Za-km-z2-9]{4}){3}$`)

func TestTemporaryPasswordHasAReadableFormatWithoutAmbiguousCharacters(t *testing.T) {
	for range 200 {
		p, err := NewTemporaryPassword()
		if err != nil || !tempFormat.MatchString(p) {
			t.Fatalf("senha temporária %q (err %v) fora do formato xxxx-xxxx-xxxx-xxxx sem caracteres ambíguos", p, err)
		}
	}
}

func TestTemporaryPasswordsAreDifferentEveryTimeAndSatisfyEveryPolicy(t *testing.T) {
	seen := map[string]bool{}
	deny := password.DefaultDenylist()
	for range 300 {
		p, _ := NewTemporaryPassword()
		if seen[p] {
			t.Fatal("senha temporária repetida")
		}
		seen[p] = true
		if err := ValidatePassword(p, []Role{RolePresidente, RoleAssociado}, deny); err != nil {
			t.Fatalf("a senha temporária %q deveria cumprir a política mais rígida: %v", p, err)
		}
	}
}
