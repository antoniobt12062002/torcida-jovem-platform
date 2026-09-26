package money

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// ErrInvalidFormat is returned by Parse for text that is not a valid pt-BR amount.
var ErrInvalidFormat = errors.New("money: formato inválido")

// Digits, dots as thousands separators (groups of three) and at most one comma
// followed by one or two decimal digits. Signs and the "R$" prefix are not accepted.
var amountPattern = regexp.MustCompile(`^(?:\d{1,3}(?:\.\d{3})+|\d+)(?:,\d{1,2})?$`)

// Format renders the amount in pt-BR, for example "R$ 1.234,56" or "-R$ 1.234,56",
// separated by an ordinary space.
func (c Cents) Format() string {
	var magnitude uint64
	if c < 0 {
		magnitude = uint64(-(c + 1)) + 1
	} else {
		magnitude = uint64(c)
	}

	reais := groupThousands(strconv.FormatUint(magnitude/100, 10))
	centavos := strconv.FormatUint(magnitude%100, 10)
	if len(centavos) == 1 {
		centavos = "0" + centavos
	}

	out := "R$ " + reais + "," + centavos
	if c < 0 {
		out = "-" + out
	}
	return out
}

func groupThousands(digits string) string {
	if len(digits) <= 3 {
		return digits
	}
	var b strings.Builder
	first := len(digits) % 3
	if first > 0 {
		b.WriteString(digits[:first])
	}
	for i := first; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(digits[i : i+3])
	}
	return b.String()
}

// Parse reads a non-negative pt-BR amount such as "1.234,56" into centavos.
// Text with more than two decimal places, characters other than digits, dots
// and one comma, or misplaced thousands separators returns ErrInvalidFormat.
// A value that does not fit in int64 returns ErrOverflow.
func Parse(text string) (Cents, error) {
	if !amountPattern.MatchString(text) {
		return 0, ErrInvalidFormat
	}
	whole, fraction, _ := strings.Cut(strings.ReplaceAll(text, ".", ""), ",")

	var centavos uint64
	switch len(fraction) {
	case 1:
		centavos = uint64(fraction[0]-'0') * 10
	case 2:
		centavos = uint64(fraction[0]-'0')*10 + uint64(fraction[1]-'0')
	}

	reais, err := strconv.ParseUint(whole, 10, 64)
	if err != nil || reais > (math.MaxInt64-centavos)/100 {
		return 0, ErrOverflow
	}
	return Cents(reais*100 + centavos), nil
}
