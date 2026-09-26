package password

import (
	"os"
	"strings"
	"testing"
)

// IDN-05.4: comparação sem diferenciar maiúsculas, com a lista embutida.
func TestDefaultDenylistContainsCommonPasswordsIgnoringCase(t *testing.T) {
	d := DefaultDenylist()

	for _, p := range []string{"password", "PASSWORD", "PassWord", "12345678", "iloveyou", "IloveYOU", "qwertyuiop"} {
		if !d.Contains(p) {
			t.Errorf("%q deveria estar na lista", p)
		}
	}
}

func TestDefaultDenylistDoesNotContainStrongOrEmptyPasswords(t *testing.T) {
	d := DefaultDenylist()

	for _, p := range []string{"", "correct horse battery staple", "k9#Lm2$vQx7!zPa4", "senha-longa-e-única-2026"} {
		if d.Contains(p) {
			t.Errorf("%q não deveria estar na lista", p)
		}
	}
}

func TestDefaultDenylistIsEmbeddedAndSubstantial(t *testing.T) {
	if n := DefaultDenylist().Len(); n < 30000 {
		t.Errorf("a lista embutida tem só %d entradas", n)
	}
	first, second := DefaultDenylist(), DefaultDenylist()
	if first != second {
		t.Error("a lista deveria ser carregada uma vez só")
	}
}

// O arquivo está no formato que Contains espera: minúsculas, 8+ caracteres, sem repetição.
func TestDenylistFileIsNormalized(t *testing.T) {
	raw, err := os.ReadFile("denylist.txt")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for i, line := range strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n") {
		switch {
		case line != strings.ToLower(line):
			t.Fatalf("linha %d não está em minúsculas: %q", i+1, line)
		case len([]rune(line)) < 8:
			t.Fatalf("linha %d tem menos de 8 caracteres: %q", i+1, line)
		case strings.ContainsAny(line, " \t\r"):
			t.Fatalf("linha %d tem espaço ou CR: %q", i+1, line)
		case seen[line]:
			t.Fatalf("linha %d repetida: %q", i+1, line)
		}
		seen[line] = true
	}
}

// Fonte e licença ficam registradas junto do arquivo.
func TestDenylistSourceAndLicenseAreRecorded(t *testing.T) {
	raw, err := os.ReadFile("denylist.SOURCE.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"SecLists", "MIT License", "Daniel Miessler", "xato-net-10-million-passwords-100000.txt"} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("denylist.SOURCE.txt deveria citar %q", want)
		}
	}
}

func TestNewDenylistNormalizesAndIgnoresBlankLines(t *testing.T) {
	d := NewDenylist("Segredo123\n\n  \nOUTRA-Senha\r\n")

	if !d.Contains("segredo123") || !d.Contains("OUTRA-SENHA") || d.Contains("") || d.Len() != 2 {
		t.Errorf("Len = %d", d.Len())
	}
}
