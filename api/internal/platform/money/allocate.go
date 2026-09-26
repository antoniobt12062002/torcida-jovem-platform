package money

import (
	"errors"
	"math"
	"math/bits"
)

// ErrInvalidWeights is returned by Allocate for an empty weight list or a
// weight that is not positive.
var ErrInvalidWeights = errors.New("money: pesos inválidos")

const minInt64Magnitude = uint64(1) << 63

// Allocate splits total in proportion to weights. The parts always sum to
// total: each part is the floor of its share and the leftover centavos go, one
// each, to the first parts in order. A negative total is split by magnitude and
// every part is negated.
func Allocate(total Cents, weights []int64) ([]Cents, error) {
	if len(weights) == 0 {
		return nil, ErrInvalidWeights
	}
	var sum uint64
	for _, w := range weights {
		if w <= 0 {
			return nil, ErrInvalidWeights
		}
		if uint64(w) > math.MaxInt64-sum {
			return nil, ErrOverflow
		}
		sum += uint64(w)
	}

	magnitude := absCents(total)
	parts := make([]uint64, len(weights))
	var assigned uint64
	for i, w := range weights {
		hi, lo := bits.Mul64(magnitude, uint64(w))
		parts[i], _ = bits.Div64(hi, lo, sum)
		assigned += parts[i]
	}
	for i, leftover := 0, magnitude-assigned; leftover > 0; i, leftover = i+1, leftover-1 {
		parts[i]++
	}

	out := make([]Cents, len(parts))
	for i, p := range parts {
		out[i] = signed(p, total < 0)
	}
	return out, nil
}

// Percent returns amount times basisPoints (1 bp = 0,01%), rounded half up
// with negative results rounded symmetrically away from zero. The intermediate
// product uses 128 bits, so the result is exact whenever it fits in int64;
// otherwise it returns ErrOverflow.
func Percent(amount Cents, basisPoints int64) (Cents, error) {
	negative := (amount < 0) != (basisPoints < 0)

	hi, lo := bits.Mul64(absCents(amount), absInt64(basisPoints))
	lo, carry := bits.Add64(lo, 5000, 0)
	hi += carry
	if hi >= 10000 {
		return 0, ErrOverflow
	}
	q, _ := bits.Div64(hi, lo, 10000)

	if negative && q > minInt64Magnitude || !negative && q > math.MaxInt64 {
		return 0, ErrOverflow
	}
	return signed(q, negative), nil
}

func absCents(c Cents) uint64 {
	if c < 0 {
		return uint64(-(c + 1)) + 1
	}
	return uint64(c)
}

func absInt64(v int64) uint64 {
	if v < 0 {
		return uint64(-(v + 1)) + 1
	}
	return uint64(v)
}

// signed converts a magnitude that is known to fit into a Cents value.
func signed(magnitude uint64, negative bool) Cents {
	if !negative {
		return Cents(magnitude)
	}
	if magnitude == minInt64Magnitude {
		return math.MinInt64
	}
	return -Cents(magnitude)
}
