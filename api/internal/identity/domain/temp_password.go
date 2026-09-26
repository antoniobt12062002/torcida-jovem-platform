package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// tempAlphabet has no lookalikes (no 0, O, 1, l, I), so it can be read out loud.
const tempAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"

// NewTemporaryPassword returns a random password of four groups of four
// characters, like "Xk3p-9Rmw-Tq7d-Fh2z" (about 92 bits), for the administrative
// reset. It is 19 characters long, so it meets the strictest policy.
func NewTemporaryPassword() (string, error) {
	max := big.NewInt(int64(len(tempAlphabet)))
	out := make([]byte, 0, 19)
	for i := range 16 {
		if i > 0 && i%4 == 0 {
			out = append(out, '-')
		}
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("senha temporária: %w", err)
		}
		out = append(out, tempAlphabet[n.Int64()])
	}
	return string(out), nil
}
