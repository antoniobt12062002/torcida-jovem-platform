package money

import (
	"encoding/json"
	"errors"
	"math"
	"math/rand"
	"os"
	"strings"
	"testing"
)

type vectors struct {
	Format []struct {
		Cents int64  `json:"cents"`
		Text  string `json:"text"`
	} `json:"format"`
	ParseValid []struct {
		Text  string `json:"text"`
		Cents int64  `json:"cents"`
	} `json:"parse_valid"`
	ParseInvalid []string `json:"parse_invalid"`
}

func loadVectors(t *testing.T) vectors {
	t.Helper()
	raw, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var v vectors
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Format) == 0 || len(v.ParseValid) == 0 || len(v.ParseInvalid) == 0 {
		t.Fatal("vetores compartilhados vazios")
	}
	return v
}

func TestFormatVectors(t *testing.T) {
	for _, tc := range loadVectors(t).Format {
		got := Cents(tc.Cents).Format()
		if got != tc.Text {
			t.Errorf("Format(%d) = %q, esperado %q", tc.Cents, got, tc.Text)
		}
		if strings.Contains(got, " ") || strings.Contains(got, " ") {
			t.Errorf("Format(%d) usa espaço não separável; deveria ser um espaço comum: %q", tc.Cents, got)
		}
	}
}

func TestFormatExtremes(t *testing.T) {
	cases := map[Cents]string{
		math.MaxInt64: "R$ 92.233.720.368.547.758,07",
		math.MinInt64: "-R$ 92.233.720.368.547.758,08",
	}
	for c, want := range cases {
		if got := c.Format(); got != want {
			t.Errorf("Format(%d) = %q, esperado %q", int64(c), got, want)
		}
	}
}

func TestParseVectors(t *testing.T) {
	v := loadVectors(t)
	for _, tc := range v.ParseValid {
		got, err := Parse(tc.Text)
		if err != nil || got != Cents(tc.Cents) {
			t.Errorf("Parse(%q) = %d, %v; esperado %d", tc.Text, got, err, tc.Cents)
		}
	}
	for _, text := range v.ParseInvalid {
		got, err := Parse(text)
		if !errors.Is(err, ErrInvalidFormat) || got != 0 {
			t.Errorf("Parse(%q) = %d, %v; esperado ErrInvalidFormat", text, got, err)
		}
	}
}

func TestParseOverflow(t *testing.T) {
	if got, err := Parse("92.233.720.368.547.758,07"); err != nil || got != math.MaxInt64 {
		t.Errorf("o maior valor representável deveria ser aceito: %d, %v", got, err)
	}
	for _, text := range []string{"92.233.720.368.547.758,08", "92233720368547758,99", "999999999999999999999999"} {
		if _, err := Parse(text); !errors.Is(err, ErrOverflow) {
			t.Errorf("Parse(%q) deveria devolver ErrOverflow, veio %v", text, err)
		}
	}
}

// Para valores não negativos, o texto de Format sem o prefixo "R$ " volta ao valor original.
func TestFormatThenParseRoundTrips(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for range 2000 {
		c := Cents(rng.Int63())
		text := strings.TrimPrefix(c.Format(), "R$ ")
		got, err := Parse(text)
		if err != nil || got != c {
			t.Fatalf("ida e volta de %d falhou: %q -> %d, %v", int64(c), text, got, err)
		}
	}
}
