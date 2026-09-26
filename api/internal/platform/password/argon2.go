// Package password hashes and verifies passwords with argon2id.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLen = 16
	keyLen  = 32

	maxMemoryKiB   = 1 << 20 // 1 GiB: a tampered hash must not exhaust the server
	maxIterations  = 20
	maxParallelism = 255
)

// ErrInvalidHash is returned for a hash that is not a valid encoded argon2id
// hash. It never includes the hash itself.
var ErrInvalidHash = errors.New("password: hash inválido")

// Params are the argon2id cost parameters.
type Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
}

// DefaultParams follows the OWASP Password Storage Cheat Sheet minimum for
// argon2id: 19 MiB of memory, 2 iterations, parallelism 1.
func DefaultParams() Params {
	return Params{MemoryKiB: 19456, Iterations: 2, Parallelism: 1}
}

// Validate checks the parameters against sane bounds.
func (p Params) Validate() error {
	switch {
	case p.MemoryKiB == 0 || p.MemoryKiB > maxMemoryKiB:
		return fmt.Errorf("password: memória fora da faixa 1..%d KiB", maxMemoryKiB)
	case p.Iterations == 0 || p.Iterations > maxIterations:
		return fmt.Errorf("password: iterações fora da faixa 1..%d", maxIterations)
	case p.Parallelism == 0:
		return errors.New("password: paralelismo deve ser positivo")
	case p.MemoryKiB < 8*uint32(p.Parallelism):
		return errors.New("password: a memória deve ser de pelo menos 8 KiB por unidade de paralelismo")
	}
	return nil
}

// Hash returns the PHC-encoded argon2id hash of plain, with a random salt. The
// encoding carries the parameters, so Verify needs nothing else.
func Hash(plain string, p Params) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: sal aleatório: %w", err)
	}
	key := argon2.IDKey([]byte(plain), salt, p.Iterations, p.MemoryKiB, p.Parallelism, keyLen)
	enc := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.MemoryKiB, p.Iterations, p.Parallelism, enc.EncodeToString(salt), enc.EncodeToString(key)), nil
}

// Verify reports whether plain matches the encoded hash, comparing in constant
// time and using the parameters stored in the hash. A malformed hash returns
// ErrInvalidHash.
func Verify(plain, encoded string) (bool, error) {
	p, salt, want, err := decode(encoded)
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(plain), salt, p.Iterations, p.MemoryKiB, p.Parallelism, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func decode(encoded string) (p Params, salt, key []byte, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v="+strconv.Itoa(argon2.Version) {
		return p, nil, nil, ErrInvalidHash
	}
	var m, t, par uint64
	fields := strings.Split(parts[3], ",")
	if len(fields) != 3 {
		return p, nil, nil, ErrInvalidHash
	}
	for i, f := range fields {
		name := string("mtp"[i])
		v, ok := strings.CutPrefix(f, name+"=")
		if !ok {
			return p, nil, nil, ErrInvalidHash
		}
		n, perr := strconv.ParseUint(v, 10, 32)
		if perr != nil {
			return p, nil, nil, ErrInvalidHash
		}
		switch i {
		case 0:
			m = n
		case 1:
			t = n
		default:
			par = n
		}
	}
	if par > maxParallelism {
		return p, nil, nil, ErrInvalidHash
	}
	p = Params{MemoryKiB: uint32(m), Iterations: uint32(t), Parallelism: uint8(par)}
	if p.Validate() != nil {
		return p, nil, nil, ErrInvalidHash
	}
	enc := base64.RawStdEncoding
	salt, serr := enc.DecodeString(parts[4])
	key, kerr := enc.DecodeString(parts[5])
	if serr != nil || kerr != nil || len(salt) < 8 || len(key) < 16 || len(key) > 64 {
		return p, nil, nil, ErrInvalidHash
	}
	return p, salt, key, nil
}

// Hasher hashes with fixed parameters (from the configuration) and can spend
// the same effort as a verification for a user that does not exist, so the
// response time does not reveal whether an account exists.
type Hasher struct {
	params Params
	dummy  string
}

// NewHasher validates the parameters and prepares the dummy hash, which is the
// hash of a random password nobody knows.
func NewHasher(p Params) (*Hasher, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("password: segredo do hash falso: %w", err)
	}
	dummy, err := Hash(string(secret), p)
	if err != nil {
		return nil, err
	}
	return &Hasher{params: p, dummy: dummy}, nil
}

// Hash hashes plain with the configured parameters.
func (h *Hasher) Hash(plain string) (string, error) { return Hash(plain, h.params) }

// Verify checks plain against an encoded hash.
func (h *Hasher) Verify(plain, encoded string) (bool, error) { return Verify(plain, encoded) }

// BurnCycles does the work of a verification against the dummy hash and
// discards the result. Login calls it when the user does not exist.
func (h *Hasher) BurnCycles(plain string) { _, _ = Verify(plain, h.dummy) }
