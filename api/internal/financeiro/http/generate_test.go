package financeirohttp

import "testing"

// 06-api-http/T2: the contract embedded by oapi-codegen must parse and must
// declare exactly the 15 operations of the 15 real use cases — no handler
// exists yet (T3-T7), this only proves the contract/codegen step itself.
func TestGetSpecParsesWithFifteenOperations(t *testing.T) {
	spec, err := GetSpec()
	if err != nil {
		t.Fatalf("GetSpec: %v", err)
	}
	count := 0
	for _, item := range spec.Paths.Map() {
		count += len(item.Operations())
	}
	if count != 15 {
		t.Errorf("operações no contrato = %d, esperado 15", count)
	}
}
