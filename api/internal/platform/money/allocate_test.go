package money

import (
	"errors"
	"math"
	"math/big"
	"math/rand"
	"slices"
	"testing"
)

func sumOf(parts []Cents) *big.Int {
	total := new(big.Int)
	for _, p := range parts {
		total.Add(total, big.NewInt(int64(p)))
	}
	return total
}

func TestAllocateExamples(t *testing.T) {
	cases := []struct {
		name    string
		total   Cents
		weights []int64
		want    []Cents
	}{
		{"resto de 1 centavo vai para a primeira parcela", 10000, []int64{1, 1, 1}, []Cents{3334, 3333, 3333}},
		{"divisão exata (FIN-001: R$ 240 em 3 parcelas)", 24000, []int64{1, 1, 1}, []Cents{8000, 8000, 8000}},
		{"resto de 2 centavos vai para as duas primeiras", 6, []int64{1, 1, 1, 1}, []Cents{2, 2, 1, 1}},
		{"resto de 1 centavo com quatro partes", 5, []int64{1, 1, 1, 1}, []Cents{2, 1, 1, 1}},
		{"pesos desiguais, sem resto", 100, []int64{3, 1}, []Cents{75, 25}},
		{"pesos desiguais, resto vai para a primeira", 10, []int64{1, 2}, []Cents{4, 6}},
		{"uma única parte recebe tudo", 12345, []int64{7}, []Cents{12345}},
		{"total zero gera partes zero", 0, []int64{1, 2, 3}, []Cents{0, 0, 0}},
		{"total negativo é rateado em módulo e negado", -10000, []int64{1, 1, 1}, []Cents{-3334, -3333, -3333}},
	}
	for _, tc := range cases {
		got, err := Allocate(tc.total, tc.weights)
		if err != nil || !slices.Equal(got, tc.want) {
			t.Errorf("%s: Allocate(%d, %v) = %v, %v; esperado %v", tc.name, int64(tc.total), tc.weights, got, err, tc.want)
		}
		if err == nil && sumOf(got).Cmp(big.NewInt(int64(tc.total))) != 0 {
			t.Errorf("%s: a soma das partes %v difere do total %d", tc.name, got, int64(tc.total))
		}
	}
}

func TestAllocateInvalidWeights(t *testing.T) {
	for name, weights := range map[string][]int64{
		"nil":             nil,
		"vazio":           {},
		"peso zero":       {1, 0, 1},
		"peso negativo":   {1, -1},
		"só peso zero":    {0},
		"todos negativos": {-2, -3},
	} {
		got, err := Allocate(100, weights)
		if !errors.Is(err, ErrInvalidWeights) || got != nil {
			t.Errorf("%s: Allocate = %v, %v; esperado ErrInvalidWeights", name, got, err)
		}
	}
}

func TestAllocateWeightSumOverflow(t *testing.T) {
	if _, err := Allocate(100, []int64{math.MaxInt64, 1}); !errors.Is(err, ErrOverflow) {
		t.Errorf("soma de pesos acima de int64 deveria dar ErrOverflow, veio %v", err)
	}
}

func TestAllocateExtremeTotalsStayExact(t *testing.T) {
	for _, total := range []Cents{math.MaxInt64, math.MinInt64, math.MaxInt64 - 1, math.MinInt64 + 1} {
		got, err := Allocate(total, []int64{1, 1, 1})
		if err != nil {
			t.Fatalf("Allocate(%d): %v", int64(total), err)
		}
		if sumOf(got).Cmp(big.NewInt(int64(total))) != 0 {
			t.Errorf("Allocate(%d): soma %v difere do total", int64(total), sumOf(got))
		}
	}
	single, err := Allocate(math.MinInt64, []int64{5})
	if err != nil || len(single) != 1 || single[0] != math.MinInt64 {
		t.Errorf("Allocate(MinInt64, [5]) = %v, %v; esperado [MinInt64]", single, err)
	}
}

// MNY-03.6: as partes sempre somam o total.
func TestAllocatePartsAlwaysSumToTotal(t *testing.T) {
	rng := rand.New(rand.NewSource(20260926))
	for i := range 5000 {
		total := Cents(rng.Int63())
		if rng.Intn(2) == 0 {
			total = -total
		}
		if i%50 == 0 {
			total = Cents(rng.Int63n(1000))
		}
		weights := make([]int64, 1+rng.Intn(20))
		for j := range weights {
			weights[j] = 1 + rng.Int63n(1_000_000)
		}
		got, err := Allocate(total, weights)
		if err != nil {
			t.Fatalf("Allocate(%d, %v): %v", int64(total), weights, err)
		}
		if len(got) != len(weights) {
			t.Fatalf("esperava %d partes, veio %d", len(weights), len(got))
		}
		if sumOf(got).Cmp(big.NewInt(int64(total))) != 0 {
			t.Fatalf("Allocate(%d, %v) = %v: a soma %v difere do total", int64(total), weights, got, sumOf(got))
		}
	}
}

func TestPercentRoundsHalfUpAwayFromZero(t *testing.T) {
	cases := []struct {
		amount Cents
		bp     int64
		want   Cents
	}{
		{10000, 400, 400}, // R$ 100,00 a 4% = R$ 4,00 (exemplo de taxa do FIN-001)
		{1, 5000, 1},      // 0,5 arredonda para cima
		{-1, 5000, -1},    // simétrico para negativos
		{5, 5000, 3},      // 2,5 -> 3
		{-5, 5000, -3},    // -2,5 -> -3
		{4, 5000, 2},      // 2,0 exato
		{3, 3333, 1},      // 0,9999 -> 1
		{1, 4999, 0},      // 0,4999 -> 0
		{12345, 250, 309}, // 308,625 -> 309
		{-12345, 250, -309},
		{10000, 0, 0},
		{0, 400, 0},
		{10000, 10000, 10000},
		{10000, -400, -400}, // basis points negativos invertem o sinal
		{-10000, -400, 400},
	}
	for _, tc := range cases {
		got, err := Percent(tc.amount, tc.bp)
		if err != nil || got != tc.want {
			t.Errorf("Percent(%d, %d) = %d, %v; esperado %d", int64(tc.amount), tc.bp, int64(got), err, int64(tc.want))
		}
	}
}

func TestPercentIntermediateProductOverflow(t *testing.T) {
	exact := []struct {
		amount Cents
		bp     int64
		want   Cents
	}{
		{math.MaxInt64, 10000, math.MaxInt64}, // produto de 128 bits, resultado cabe
		{math.MinInt64, 10000, math.MinInt64},
		{math.MaxInt64, 5000, 4611686018427387904}, // 9223372036854775807 / 2, arredondado para cima
	}
	for _, tc := range exact {
		got, err := Percent(tc.amount, tc.bp)
		if err != nil || got != tc.want {
			t.Errorf("Percent(%d, %d) = %d, %v; esperado %d", int64(tc.amount), tc.bp, int64(got), err, int64(tc.want))
		}
	}
	for _, tc := range []struct {
		amount Cents
		bp     int64
	}{
		{math.MaxInt64, 10001},
		{math.MinInt64, 10001},
		{math.MinInt64, -10000}, // +2^63 não cabe em int64
		{math.MaxInt64, math.MaxInt64},
	} {
		got, err := Percent(tc.amount, tc.bp)
		if !errors.Is(err, ErrOverflow) || got != 0 {
			t.Errorf("Percent(%d, %d) = %d, %v; esperado ErrOverflow", int64(tc.amount), tc.bp, int64(got), err)
		}
	}
}
