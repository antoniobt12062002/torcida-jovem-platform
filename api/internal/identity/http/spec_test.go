package identityhttp

import (
	"slices"
	"testing"
)

// API-01.2: o contrato embutido no código gerado é o de api/openapi/identity.yaml.
func TestEmbeddedSpecHasTheIdentityOperations(t *testing.T) {
	spec, err := GetSpec()
	if err != nil {
		t.Fatalf("contrato embutido: %v", err)
	}

	var ops []string
	for path, item := range spec.Paths.Map() {
		for method := range item.Operations() {
			ops = append(ops, method+" "+path)
		}
	}
	if len(ops) != 14 {
		t.Errorf("operações = %d (%v), esperado 14", len(ops), ops)
	}
	for _, want := range []string{"POST /api/v1/users/{id}/admin-membership/revoke", "POST /api/v1/auth/login"} {
		if !slices.Contains(ops, want) {
			t.Errorf("faltou %s", want)
		}
	}
}
