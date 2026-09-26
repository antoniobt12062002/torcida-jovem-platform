package password

import (
	"strings"
	"testing"
	"time"
)

// Parâmetros baixos só para os testes serem rápidos; o padrão real é DefaultParams.
var fast = Params{MemoryKiB: 64, Iterations: 1, Parallelism: 1}

func TestDefaultParamsFollowTheOWASPMinimum(t *testing.T) {
	d := DefaultParams()

	if d.MemoryKiB != 19456 || d.Iterations != 2 || d.Parallelism != 1 {
		t.Errorf("padrão = %+v, esperado 19456 KiB, 2 iterações, paralelismo 1 (OWASP: m=19 MiB, t=2, p=1)", d)
	}
	if err := d.Validate(); err != nil {
		t.Errorf("o padrão deveria ser válido: %v", err)
	}
}

func TestHashThenVerifyRoundTrips(t *testing.T) {
	h, err := Hash("senha-longa-e-correta", fast)
	if err != nil {
		t.Fatal(err)
	}

	ok, err := Verify("senha-longa-e-correta", h)
	if err != nil || !ok {
		t.Errorf("Verify = %v, %v", ok, err)
	}
	if ok, err := Verify("outra-senha", h); err != nil || ok {
		t.Errorf("senha errada: %v, %v", ok, err)
	}
	if ok, _ := Verify("", h); ok {
		t.Error("senha vazia não pode validar")
	}
}

// O hash carrega parâmetros e sal (formato PHC) e não repete entre chamadas.
func TestHashEmbedsParametersAndARandomSalt(t *testing.T) {
	a, _ := Hash("mesma-senha", fast)
	b, _ := Hash("mesma-senha", fast)

	if !strings.HasPrefix(a, "$argon2id$v=19$m=64,t=1,p=1$") {
		t.Errorf("formato = %q", a)
	}
	if a == b {
		t.Error("o sal deveria ser aleatório: dois hashes da mesma senha não podem ser iguais")
	}
	parts := strings.Split(a, "$")
	if len(parts) != 6 || len(parts[4]) < 20 || len(parts[5]) < 40 {
		t.Errorf("partes = %v", parts)
	}
}

func TestVerifyUsesTheParametersStoredInTheHash(t *testing.T) {
	old, _ := Hash("senha-antiga", Params{MemoryKiB: 32, Iterations: 2, Parallelism: 2})

	if ok, err := Verify("senha-antiga", old); err != nil || !ok {
		t.Errorf("um hash com parâmetros antigos deve continuar válido: %v, %v", ok, err)
	}
}

const (
	okSalt = "c2FsdHNhbHRzYWx0c2FsdA"                                 // 16 bytes
	okKey  = "aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaA" // 32+ bytes
)

func encoded(version, params, salt, key string) string {
	return "$argon2id$" + version + "$" + params + "$" + salt + "$" + key
}

func TestVerifyRejectsMalformedHashesWithoutEchoingThem(t *testing.T) {
	good, _ := Hash("senha", fast)
	ok := "v=19"
	cases := map[string]string{
		"vazio":                    "",
		"texto qualquer":           "texto-qualquer",
		"outro algoritmo":          "$argon2i$v=19$m=64,t=1,p=1$" + okSalt + "$" + okKey,
		"versão diferente":         encoded("v=18", "m=64,t=1,p=1", okSalt, okKey),
		"memória não numérica":     encoded(ok, "m=x,t=1,p=1", okSalt, okKey),
		"parâmetros faltando":      encoded(ok, "m=64,t=1", okSalt, okKey),
		"ordem dos parâmetros":     encoded(ok, "t=1,m=64,p=1", okSalt, okKey),
		"sal inválido":             encoded(ok, "m=64,t=1,p=1", "!!!", okKey),
		"chave inválida":           encoded(ok, "m=64,t=1,p=1", okSalt, "!!!"),
		"sal curto demais":         encoded(ok, "m=64,t=1,p=1", "c2FsdA", okKey),
		"chave curta demais":       encoded(ok, "m=64,t=1,p=1", okSalt, "aGFzaA"),
		"chave longa demais":       encoded(ok, "m=64,t=1,p=1", okSalt, strings.Repeat("QUJD", 30)),
		"memória zero":             encoded(ok, "m=0,t=1,p=1", okSalt, okKey),
		"paralelismo acima de 255": encoded(ok, "m=4096,t=1,p=257", okSalt, okKey), // 257 viraria 1 num uint8
		"parte extra":              strings.Replace(good, "$argon2id$", "$argon2id$x$", 1),
	}
	for name, c := range cases {
		okVerify, err := Verify("senha", c)
		if okVerify || err == nil {
			t.Errorf("%s: esperava erro e ok=false, veio %v, %v", name, okVerify, err)
			continue
		}
		if c != "" && strings.Contains(err.Error(), c) {
			t.Errorf("%s: o erro não pode repetir o hash: %v", name, err)
		}
	}
	if okVerify, err := Verify("senha", encoded(ok, "m=64,t=1,p=1", okSalt, okKey)); err != nil || okVerify {
		t.Errorf("um hash bem formado de outra senha deveria só não conferir: %v, %v", okVerify, err)
	}
}

func TestVerifyRejectsAbsurdParametersInAHash(t *testing.T) {
	// Um hash adulterado com memória gigante não pode esgotar o servidor.
	h := "$argon2id$v=19$m=4194304,t=1,p=1$c2FsdHNhbHRzYWx0c2FsdA$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNoaGFzaGhhc2g"

	if ok, err := Verify("senha", h); ok || err == nil {
		t.Errorf("ok = %v, err = %v", ok, err)
	}
}

func TestParamsValidation(t *testing.T) {
	bad := []Params{
		{MemoryKiB: 0, Iterations: 1, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 0, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 1, Parallelism: 0},
		{MemoryKiB: 8, Iterations: 1, Parallelism: 2}, // memória menor que 8*p
		{MemoryKiB: 1<<20 + 1, Iterations: 1, Parallelism: 1},
		{MemoryKiB: 64, Iterations: 21, Parallelism: 1},
	}
	for _, p := range bad {
		if err := p.Validate(); err == nil {
			t.Errorf("%+v deveria ser inválido", p)
		}
		if _, err := Hash("x", p); err == nil {
			t.Errorf("Hash com %+v deveria falhar", p)
		}
	}
}

func TestHasherHashesVerifiesAndBurnsCyclesForUnknownUsers(t *testing.T) {
	h, err := NewHasher(fast)
	if err != nil {
		t.Fatal(err)
	}

	enc, err := h.Hash("senha-do-usuario")
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := h.Verify("senha-do-usuario", enc); err != nil || !ok {
		t.Errorf("Verify = %v, %v", ok, err)
	}
	h.BurnCycles("qualquer-senha") // não entra em pânico nem devolve nada
	if _, err := NewHasher(Params{}); err == nil {
		t.Error("NewHasher com parâmetros inválidos deveria falhar")
	}
}

func TestBurnCyclesDoesTheSameWorkAsAVerification(t *testing.T) {
	h, _ := NewHasher(fast)

	if ok, _ := Verify("qualquer-senha", h.dummy); ok {
		t.Error("o hash falso não pode aceitar senha alguma que um usuário possa escolher")
	}
	if !strings.HasPrefix(h.dummy, "$argon2id$v=19$m=64,t=1,p=1$") {
		t.Errorf("o hash falso deveria usar os mesmos parâmetros: %q", h.dummy)
	}
}

// BurnCycles precisa gastar o mesmo esforço de uma verificação, senão o tempo de
// resposta revelaria que o usuário não existe.
func TestBurnCyclesCostsAboutAsMuchAsAVerification(t *testing.T) {
	heavy := Params{MemoryKiB: 32768, Iterations: 2, Parallelism: 1}
	h, err := NewHasher(heavy)
	if err != nil {
		t.Fatal(err)
	}
	enc, _ := h.Hash("senha")

	start := time.Now()
	_, _ = h.Verify("senha", enc)
	verify := time.Since(start)
	start = time.Now()
	h.BurnCycles("senha")
	burn := time.Since(start)

	if burn < verify/4 {
		t.Errorf("BurnCycles levou %v, uma verificação leva %v: o esforço deveria ser parecido", burn, verify)
	}
}
