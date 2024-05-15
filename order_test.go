package ellipxobj

import (
	"testing"
)

func TestOrder(t *testing.T) {
	a := NewOrder(Pair("BTC", "USD"), TypeBid).SetId("a9039a38-3bd4-4084-95d1-3548c1873c8b", "test")
	a.Amount, _ = NewAmountFromFloat64(1, 8)
	a.Price, _ = NewAmountFromString("5", 5)

	if a.String() != "buy 1.00000000 BTC @ 5.00000 USD/BTC" {
		t.Errorf("invalid order, expected buy 1.00000000 BTC @ 5.00000 USD/BTC, got %s", a.String())
	}

	// reverse order
	b := a.Reverse()
	if b.String() != "sell 1.00000000 BTC worth @ 0.20000 BTC/USD" {
		t.Errorf("invalid order, expected, got %s", b.String())
	}
}
