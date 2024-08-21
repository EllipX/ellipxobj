package ellipxobj

import (
	"encoding/hex"
	"encoding/json"
	"testing"
)

type structWithAmount struct {
	A Amount
}

func TestAmount(t *testing.T) {
	a := NewAmount(42000, 3)

	if a.String() != "42.000" {
		t.Errorf("expected 42.000 but got %s", a.String())
	}

	a, _ = NewAmountFromFloat64(0.42, 5)
	if a.String() != "0.42000" {
		t.Errorf("expected 0.42000 but got %s", a.String())
	}

	a = NewAmount(42000, 10)
	if a.String() != "0.0000042000" {
		t.Errorf("expected 0.0000042000 but got %s", a.String())
	}

	a, _ = NewAmountFromFloat64(123.456, 0)
	if a.String() != "123.45600" {
		t.Errorf("expected 123.45600 but got %s", a.String())
	}

	a, _ = NewAmountFromFloat64(123.456789123456, 0)
	if a.String() != "123.456789123456" {
		t.Errorf("expected 123.456789123456 but got %s", a.String())
	}
	a.SetExp(5)
	if a.String() != "123.45679" {
		t.Errorf("expected 123.45679 but got %s", a.String())
	}
	a.SetExp(6)
	if a.String() != "123.456790" {
		t.Errorf("expected 123.456790 but got %s", a.String())
	}

	a = NewAmount(42000, 3)
	b := NewAmount(500000, 5)

	c, _ := NewAmount(0, 5).Mul(a, b)

	if c.String() != "210.00000" {
		t.Errorf("expected 210.00000 but got %s", c.String())
	}

	c = NewAmount(0, 10).Div(a, b)

	if c.String() != "8.4000000000" {
		t.Errorf("expected 8.400 but got %s / %s = %s", a, b, c.String())
	}

	v := c.Bytes()
	if hex.EncodeToString(v) != "0014138eca4800" {
		t.Errorf("expected 8.4 to become , got %s", hex.EncodeToString(v))
	}

	d := new(Amount)
	err := d.UnmarshalBinary(v)
	if err != nil {
		t.Errorf("error unmarsha binary: %s", err)
	}

	if d.String() != c.String() {
		t.Errorf("invalid unmarshal binary, value differs %s != %s", c, d)
	}

	o := &structWithAmount{}

	js, _ := json.Marshal(o)
	if string(js) != `{"A":{"v":"0","e":0,"f":0}}` {
		t.Errorf("unexpected value for invalid Amount: %s", js)
	}
}
