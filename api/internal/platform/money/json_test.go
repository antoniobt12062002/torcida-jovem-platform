package money

import (
	"encoding/json"
	"errors"
	"testing"
)

type payload struct {
	Amount Cents `json:"amount"`
}

func TestMarshalEmitsAnIntegerLiteral(t *testing.T) {
	cases := map[Cents]string{
		123456:           `{"amount":123456}`,
		0:                `{"amount":0}`,
		-5:               `{"amount":-5}`,
		9007199254740991: `{"amount":9007199254740991}`,
	}
	for c, want := range cases {
		got, err := json.Marshal(payload{Amount: c})
		if err != nil || string(got) != want {
			t.Errorf("Marshal(%d) = %s, %v; esperado %s", int64(c), got, err, want)
		}
	}
}

func TestUnmarshalAcceptsIntegerLiterals(t *testing.T) {
	cases := map[string]Cents{
		`{"amount":123456}`:            123456,
		`{"amount":0}`:                 0,
		`{"amount":-5}`:                -5,
		`{"amount":-0}`:                0,
		`{"amount":9007199254740991}`:  9007199254740991,
		`{"amount":-9007199254740991}`: -9007199254740991,
	}
	for in, want := range cases {
		var p payload
		if err := json.Unmarshal([]byte(in), &p); err != nil || p.Amount != want {
			t.Errorf("Unmarshal(%s) = %d, %v; esperado %d", in, int64(p.Amount), err, want)
		}
	}
}

func TestUnmarshalRejectsWhatIsNotAnIntegerLiteral(t *testing.T) {
	for _, in := range []string{
		`{"amount":"123"}`,
		`{"amount":"1.234,56"}`,
		`{"amount":1.5}`,
		`{"amount":100.0}`,
		`{"amount":1e2}`,
		`{"amount":1E2}`,
		`{"amount":1e-2}`,
		`{"amount":null}`,
		`{"amount":true}`,
		`{"amount":[]}`,
		`{"amount":{}}`,
	} {
		var p payload
		err := json.Unmarshal([]byte(in), &p)
		if !errors.Is(err, ErrNotInteger) {
			t.Errorf("Unmarshal(%s): erro = %v, esperado ErrNotInteger", in, err)
		}
		if p.Amount != 0 {
			t.Errorf("Unmarshal(%s) não deveria alterar o valor, veio %d", in, int64(p.Amount))
		}
	}
}

func TestUnmarshalRejectsMagnitudeAboveTheSafeInteger(t *testing.T) {
	for _, in := range []string{
		`{"amount":9007199254740992}`,
		`{"amount":-9007199254740992}`,
		`{"amount":9223372036854775807}`,
		`{"amount":9223372036854775808}`,
		`{"amount":-9223372036854775809}`,
		`{"amount":99999999999999999999999999}`,
	} {
		var p payload
		err := json.Unmarshal([]byte(in), &p)
		if !errors.Is(err, ErrOutOfRange) {
			t.Errorf("Unmarshal(%s): erro = %v, esperado ErrOutOfRange", in, err)
		}
		if p.Amount != 0 {
			t.Errorf("Unmarshal(%s) não deveria alterar o valor, veio %d", in, int64(p.Amount))
		}
	}
}

func TestMaxSafeIsTheJavaScriptSafeInteger(t *testing.T) {
	if MaxSafe != 9007199254740991 {
		t.Errorf("MaxSafe = %d, esperado 9007199254740991", int64(MaxSafe))
	}
}

// Um campo ponteiro é a forma explícita de aceitar null.
func TestNullIsOnlyAcceptedByAnExplicitPointerField(t *testing.T) {
	var p struct {
		Amount *Cents `json:"amount"`
	}
	if err := json.Unmarshal([]byte(`{"amount":null}`), &p); err != nil || p.Amount != nil {
		t.Errorf("null em *Cents deveria virar nil sem erro: %v, %v", p.Amount, err)
	}
}
