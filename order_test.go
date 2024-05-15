package ellipxobj

import (
	"log"
	"testing"
)

func TestOrder(t *testing.T) {
	a := NewOrder(Pair("BTC", "USD"), TypeBid).SetId("a9039a38-3bd4-4084-95d1-3548c1873c8b", "test")
	a.Amount = NewAmount(100000000, 8)
	a.Price, _ = NewAmountFromString("5", 5)

	if a.String() != "buy 1.00000000 BTC @ 5.00000 USD/BTC" {
		log.Printf("invalid order, expected buy 1.00000000 BTC @ 5.00000 USD/BTC, got %s", a.String())
	}
}
