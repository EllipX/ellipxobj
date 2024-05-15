package ellipxobj

import "testing"

func TestAmount(t *testing.T) {
	a := NewAmount(42000, 3)

	if a.String() != "42.000" {
		t.Errorf("expected 42.000 but got %s", a.String())
	}

	a = NewAmount(42000, 5)
	if a.String() != "0.42000" {
		t.Errorf("expected 0.42000 but got %s", a.String())
	}

	a = NewAmount(42000, 10)
	if a.String() != "0.0000042000" {
		t.Errorf("expected 0.0000042000 but got %s", a.String())
	}

	a = NewAmount(42000, 3)
	b := NewAmount(500000, 5)

	c, _ := NewAmount(0, 5).Mul(a, b)

	if c.String() != "210.00000" {
		t.Errorf("expected 210.00000 but got %s", c.String())
	}
}
