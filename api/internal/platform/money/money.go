// Package money is shared infrastructure for monetary values. It knows only
// amounts in centavos and operations on them; financial entities such as
// entries, revenue, expenses or products belong to the business modules.
package money

import (
	"errors"
	"math"
)

// Cents is an amount of Brazilian reais, in centavos: 10000 is R$ 100,00.
type Cents int64

// ErrOverflow is returned when a result does not fit in int64.
var ErrOverflow = errors.New("money: overflow")

func (c Cents) Add(o Cents) (Cents, error) {
	if (o > 0 && c > math.MaxInt64-o) || (o < 0 && c < math.MinInt64-o) {
		return 0, ErrOverflow
	}
	return c + o, nil
}

func (c Cents) Sub(o Cents) (Cents, error) {
	if (o < 0 && c > math.MaxInt64+o) || (o > 0 && c < math.MinInt64+o) {
		return 0, ErrOverflow
	}
	return c - o, nil
}

func (c Cents) Neg() (Cents, error) {
	if c == math.MinInt64 {
		return 0, ErrOverflow
	}
	return -c, nil
}

// Cmp returns -1, 0 or 1 when c is less than, equal to or greater than o.
func (c Cents) Cmp(o Cents) int {
	switch {
	case c < o:
		return -1
	case c > o:
		return 1
	default:
		return 0
	}
}
