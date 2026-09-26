package domain

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/password"
)

func TestNormalizeEmailTrimsLowercasesAndAcceptsBasicAddresses(t *testing.T) {
	cases := map[string]string{
		"Ana@Exemplo.com":          "ana@exemplo.com",
		"  ana@exemplo.com  ":      "ana@exemplo.com",
		"ANA.SILVA+tj@EXEMPLO.COM": "ana.silva+tj@exemplo.com",
		"josé@exemplo.com.br":      "josé@exemplo.com.br",
	}
	for in, want := range cases {
		got, err := NormalizeEmail(in)
		if err != nil || got != want {
			t.Errorf("NormalizeEmail(%q) = %q, %v; esperado %q", in, got, err, want)
		}
	}
}

func TestNormalizeEmailRejectsInvalidAddresses(t *testing.T) {
	long := strings.Repeat("a", 250) + "@x.com"
	for _, in := range []string{
		"", "   ", "sem-arroba", "@exemplo.com", "ana@", "a@b@c.com", "ana silva@exemplo.com",
		"Ana <ana@exemplo.com>", "ana@exemplo.com\nBcc: x@y.com", "ana@exemplo.com, bia@exemplo.com", long, "ana@exe mplo.com",
		"ana\x00@exemplo.com", "ana@exemplo",
	} {
		if got, err := NormalizeEmail(in); err == nil {
			t.Errorf("NormalizeEmail(%q) deveria falhar, veio %q", in, got)
		}
	}
}

var hashKey = []byte("0123456789abcdef0123456789abcdef")

// A chave impede reverter o e-mail por dicionário a partir do banco ou da auditoria.
func TestHashEmailIsAKeyedDeterministicHMACOfTheNormalizedEmail(t *testing.T) {
	a, err := HashEmail(hashKey, "ana@exemplo.com")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := HashEmail(hashKey, "  ANA@Exemplo.COM ")
	other, _ := HashEmail([]byte("ffffffffffffffffffffffffffffffff"), "ana@exemplo.com")
	plain := sha256.Sum256([]byte("ana@exemplo.com"))

	if len(a) != 32 || !bytes.Equal(a, b) {
		t.Errorf("deveria ser determinístico e ignorar caixa e espaços: %x vs %x", a, b)
	}
	if bytes.Equal(a, other) {
		t.Error("chaves diferentes deveriam dar resultados diferentes")
	}
	if bytes.Equal(a, plain[:]) {
		t.Error("não pode ser o SHA-256 simples do e-mail")
	}
}

func TestHashEmailRejectsAnEmptyKeyAndInvalidEmails(t *testing.T) {
	if _, err := HashEmail(nil, "ana@exemplo.com"); err == nil {
		t.Error("chave vazia deveria ser recusada")
	}
	if _, err := HashEmail(hashKey, "inválido"); err == nil {
		t.Error("e-mail inválido deveria ser recusado")
	}
}

func violation(err error) (PasswordViolation, int) {
	if pe, ok := errors.AsType[*PasswordError](err); ok {
		return pe.Violation, pe.MinLength
	}
	return "", 0
}

// IDN-05.1 e IDN-05.2: 10 para qualquer papel diferente de ASSOCIADO, 8 só para ASSOCIADO.
func TestMinPasswordLengthDependsOnWhetherTheUserHoldsAnyNonAssociadoRole(t *testing.T) {
	cases := []struct {
		roles []Role
		want  int
	}{
		{[]Role{RoleAssociado}, 8},
		{[]Role{RoleTesouraria}, 10},
		{[]Role{RoleAssociado, RoleTesouraria}, 10},
		{[]Role{RolePresidente, RoleConselhoFiscal}, 10},
		{[]Role{RoleAdminSistema}, 10},
	}
	for _, tc := range cases {
		got, err := MinPasswordLength(tc.roles)
		if err != nil || got != tc.want {
			t.Errorf("MinPasswordLength(%v) = %d, %v; esperado %d", tc.roles, got, err, tc.want)
		}
	}
}

// Conjunto vazio ou papel desconhecido é inválido.
func TestMinPasswordLengthRejectsEmptyAndUnknownRoleSets(t *testing.T) {
	if _, err := MinPasswordLength(nil); err == nil {
		t.Error("conjunto vazio deveria ser inválido")
	}
	if _, err := MinPasswordLength([]Role{RoleAssociado, "CHEFE"}); err == nil {
		t.Error("papel desconhecido deveria ser inválido")
	}
}

func TestValidatePasswordEnforcesTheMinimumByRole(t *testing.T) {
	deny := password.NewDenylist("")
	nine := "abcdefgh1"

	if err := ValidatePassword(nine, []Role{RoleAssociado}, deny); err != nil {
		t.Errorf("9 caracteres bastam para ASSOCIADO: %v", err)
	}
	v, min := violation(ValidatePassword(nine, []Role{RoleAssociado, RoleEventos}, deny))
	if v != PasswordTooShort || min != 10 {
		t.Errorf("9 caracteres não bastam para papel administrativo: %q (min %d)", v, min)
	}
	v, min = violation(ValidatePassword("abcdefg", []Role{RoleAssociado}, deny))
	if v != PasswordTooShort || min != 8 {
		t.Errorf("7 caracteres não bastam para ASSOCIADO: %q (min %d)", v, min)
	}
}

// IDN-05.3 e IDN-05.7: máximo de 128 e contagem em pontos de código, sem regras de composição.
func TestValidatePasswordCountsUnicodeCodePointsAndHasNoCompositionRules(t *testing.T) {
	deny := password.NewDenylist("")
	admin := []Role{RoleTesouraria}

	if err := ValidatePassword(strings.Repeat("ñ", 10), admin, deny); err != nil {
		t.Errorf("10 pontos de código (20 bytes) deveriam bastar: %v", err)
	}
	if v, _ := violation(ValidatePassword(strings.Repeat("ñ", 9), admin, deny)); v != PasswordTooShort {
		t.Errorf("9 pontos de código não bastam: %q", v)
	}
	if err := ValidatePassword(strings.Repeat("🙂", 128), admin, deny); err != nil {
		t.Errorf("128 pontos de código (512 bytes) deveriam ser aceitos: %v", err)
	}
	if v, _ := violation(ValidatePassword(strings.Repeat("a", 129), admin, deny)); v != PasswordTooLong {
		t.Errorf("129 caracteres deveriam ser longos demais: %q", v)
	}
	if err := ValidatePassword("aaaaaaaaaaaa", admin, deny); err != nil {
		t.Errorf("sem regras de composição, só minúsculas repetidas é aceito: %v", err)
	}
}

// IDN-05.4: senha comum ou comprometida, sem diferenciar maiúsculas.
func TestValidatePasswordRejectsCompromisedPasswordsIgnoringCase(t *testing.T) {
	deny := password.NewDenylist("senhacomum123\n")

	for _, p := range []string{"senhacomum123", "SenhaComum123", "SENHACOMUM123"} {
		if v, _ := violation(ValidatePassword(p, []Role{RoleTesouraria}, deny)); v != PasswordCompromised {
			t.Errorf("%q deveria ser recusada como comprometida: %q", p, v)
		}
	}
}

// A ordem é: longa demais, curta demais, comprometida.
func TestValidatePasswordReportsTheTooShortBeforeCompromised(t *testing.T) {
	deny := password.NewDenylist("1234567\n")

	if v, _ := violation(ValidatePassword("1234567", []Role{RoleAssociado}, deny)); v != PasswordTooShort {
		t.Errorf("violação = %q, esperava too_short", v)
	}
}

func TestPasswordViolationCodesMatchTheAPI(t *testing.T) {
	if PasswordTooShort != "password_too_short" || PasswordTooLong != "password_too_long" || PasswordCompromised != "password_compromised" {
		t.Error("os códigos da API são password_too_short, password_too_long e password_compromised")
	}
	if got := (&PasswordError{Violation: PasswordTooShort, MinLength: 8}).Error(); got != "password_too_short" {
		t.Errorf("Error() = %q", got)
	}
}

func TestValidatePasswordPropagatesARoleSetError(t *testing.T) {
	if err := ValidatePassword("uma-senha-longa-1", nil, password.NewDenylist("")); err == nil {
		t.Error("sem papéis não há como saber o mínimo: deveria falhar")
	}
}

func TestHashIdentifierEqualsHashEmailForValidAddressesAndAcceptsInvalidOnes(t *testing.T) {
	a, _ := HashEmail(hashKey, "ana@exemplo.com")
	b, err := HashIdentifier(hashKey, "  ANA@exemplo.com ")
	c, err2 := HashIdentifier(hashKey, "isto não é um e-mail")

	if err != nil || err2 != nil || !bytes.Equal(a, b) || len(c) != 32 {
		t.Errorf("a = %x, b = %x, c = %x, err = %v/%v", a, b, c, err, err2)
	}
	if _, err := HashIdentifier(nil, "x"); err == nil {
		t.Error("chave vazia deveria ser recusada")
	}
}
