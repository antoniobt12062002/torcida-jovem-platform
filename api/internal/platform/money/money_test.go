package money

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"strings"
	"testing"
)

func TestAddAndSubAreExact(t *testing.T) {
	cases := []struct {
		name string
		got  func() (Cents, error)
		want Cents
	}{
		{"soma simples", func() (Cents, error) { return Cents(10000).Add(2550) }, 12550},
		{"soma com negativo", func() (Cents, error) { return Cents(500).Add(-700) }, -200},
		{"subtração simples", func() (Cents, error) { return Cents(10000).Sub(2550) }, 7450},
		{"subtração resulta negativa", func() (Cents, error) { return Cents(100).Sub(250) }, -150},
		{"soma no limite superior", func() (Cents, error) { return Cents(math.MaxInt64 - 1).Add(1) }, math.MaxInt64},
		{"subtração no limite inferior", func() (Cents, error) { return Cents(math.MinInt64 + 1).Sub(1) }, math.MinInt64},
	}
	for _, tc := range cases {
		got, err := tc.got()
		if err != nil || got != tc.want {
			t.Errorf("%s: %d, %v; esperado %d", tc.name, got, err, tc.want)
		}
	}
}

func TestOverflowReturnsErrOverflowWithoutWrapping(t *testing.T) {
	cases := []struct {
		name string
		op   func() (Cents, error)
	}{
		{"soma acima do máximo", func() (Cents, error) { return Cents(math.MaxInt64).Add(1) }},
		{"soma abaixo do mínimo", func() (Cents, error) { return Cents(math.MinInt64).Add(-1) }},
		{"subtração acima do máximo", func() (Cents, error) { return Cents(math.MaxInt64).Sub(-1) }},
		{"subtração abaixo do mínimo", func() (Cents, error) { return Cents(math.MinInt64).Sub(1) }},
		{"negação do mínimo", func() (Cents, error) { return Cents(math.MinInt64).Neg() }},
	}
	for _, tc := range cases {
		got, err := tc.op()
		if !errors.Is(err, ErrOverflow) {
			t.Errorf("%s: erro = %v, esperado ErrOverflow", tc.name, err)
		}
		if got != 0 {
			t.Errorf("%s: valor = %d, esperado 0 quando há erro (sem dar a volta)", tc.name, got)
		}
	}
}

func TestNeg(t *testing.T) {
	for in, want := range map[Cents]Cents{100: -100, -100: 100, 0: 0, math.MaxInt64: -math.MaxInt64} {
		if got, err := in.Neg(); err != nil || got != want {
			t.Errorf("Neg(%d) = %d, %v; esperado %d", in, got, err, want)
		}
	}
}

func TestCmp(t *testing.T) {
	cases := []struct {
		a, b Cents
		want int
	}{{1, 2, -1}, {2, 1, 1}, {5, 5, 0}, {-3, 2, -1}, {0, 0, 0}, {math.MinInt64, math.MaxInt64, -1}}
	for _, tc := range cases {
		if got := tc.a.Cmp(tc.b); got != tc.want {
			t.Errorf("Cmp(%d, %d) = %d, esperado %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCentsIsAnInt64CountOfCentavos(t *testing.T) {
	var c Cents = 10000 // R$ 100,00
	if int64(c) != 10000 {
		t.Errorf("Cents deveria ser int64 em centavos, veio %d", int64(c))
	}
}

// MNY-01.5: o pacote não expõe função que aceite ou devolva float32 ou float64.
func TestPackageDoesNotExposeFloats(t *testing.T) {
	var isFloat func(ast.Expr) bool
	isFloat = func(e ast.Expr) bool {
		switch v := e.(type) {
		case *ast.Ident:
			return v.Name == "float32" || v.Name == "float64"
		case *ast.ArrayType:
			return isFloat(v.Elt)
		case *ast.StarExpr:
			return isFloat(v.X)
		case *ast.Ellipsis:
			return isFloat(v.Elt)
		}
		return false
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(token.NewFileSet(), e.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		checked++
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !fn.Name.IsExported() {
				continue
			}
			for _, list := range []*ast.FieldList{fn.Type.Params, fn.Type.Results} {
				if list == nil {
					continue
				}
				for _, field := range list.List {
					if isFloat(field.Type) {
						t.Errorf("%s (%s) expõe float na assinatura", fn.Name.Name, e.Name())
					}
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("nenhum arquivo do pacote foi inspecionado")
	}
}
