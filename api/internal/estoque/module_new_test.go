package estoque

import (
	"reflect"
	"testing"
)

// New does no I/O (same contract as financeiro.New) — a zero-value Deps is
// enough to prove every field of Module gets wired, without touching a
// database or any platform dependency.
func TestNewWiresEveryFieldOfTheModule(t *testing.T) {
	m := New(Deps{})
	if m == nil {
		t.Fatal("New(Deps{}) devolveu nil")
	}

	v := reflect.ValueOf(*m)
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		if v.Field(i).IsNil() {
			t.Errorf("Module.%s ficou nil depois de New", field.Name)
		}
	}
}

// Confirms the 6 real use cases (not counting the 2 repositories) are all
// present, by name, so a future field rename/removal is caught explicitly —
// not just "something is non-nil".
func TestNewWiresAllSixUseCases(t *testing.T) {
	m := New(Deps{})
	want := []string{
		"CriarProduto", "ListarProdutos",
		"RegistrarMovimentacao", "ListarMovimentacoes",
		"AjustarEstoque", "ConsultarSaldo",
	}
	v := reflect.ValueOf(*m)
	for _, name := range want {
		f := v.FieldByName(name)
		if !f.IsValid() {
			t.Errorf("Module não tem o campo %s", name)
			continue
		}
		if f.IsNil() {
			t.Errorf("Module.%s ficou nil", name)
		}
	}
	if len(want) != 6 {
		t.Fatalf("lista de verificação tem %d nomes, esperado 6", len(want))
	}
}
