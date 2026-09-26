package money

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
)

// MaxSafe is the largest magnitude, in centavos, that the API accepts: the
// largest integer a JavaScript number represents exactly (2^53 - 1).
const MaxSafe Cents = 9007199254740991

var (
	// ErrNotInteger is returned for JSON that is not an integer literal.
	ErrNotInteger = errors.New("money: o valor JSON deve ser um inteiro em centavos")
	// ErrOutOfRange is returned when the magnitude exceeds MaxSafe; the API maps
	// it to 422 with code amount_out_of_range.
	ErrOutOfRange = errors.New("money: valor fora da faixa segura")
)

var integerLiteral = regexp.MustCompile(`^-?(?:0|[1-9][0-9]*)$`)

var (
	_ json.Marshaler   = Cents(0)
	_ json.Unmarshaler = (*Cents)(nil)
)

// MarshalJSON emits an integer literal: no quotes, no decimal point, no exponent.
func (c Cents) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(nil, int64(c), 10), nil
}

// UnmarshalJSON accepts only an integer literal within [-MaxSafe, MaxSafe].
// Strings, decimals, exponent forms and null are rejected; use *Cents when a
// field may be null.
func (c *Cents) UnmarshalJSON(data []byte) error {
	if !integerLiteral.Match(data) {
		return ErrNotInteger
	}
	v, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return ErrOutOfRange
	}
	if v > int64(MaxSafe) || v < -int64(MaxSafe) {
		return ErrOutOfRange
	}
	*c = Cents(v)
	return nil
}
