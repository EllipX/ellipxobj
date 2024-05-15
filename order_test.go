package ellipxobj

import (
	"encoding/json"
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

func TestMarshalOrder(t *testing.T) {
	a := NewOrder(Pair("BTC", "USD"), TypeBid).SetId("a9039a38-3bd4-4084-95d1-3548c1873c8b", "test")
	a.RequestTime = 1715773941
	//a.Unique = &TimeId{Unix: 1715773941, Nano: 987654321, Index: 42}
	a.Amount, _ = NewAmountFromFloat64(1, 8)
	a.Price, _ = NewAmountFromString("5", 5)

	data, err := json.Marshal(a)
	if err != nil {
		t.Errorf("failed to marshal json: %s", err)
		return
	}

	if string(data) != `{"id":"a9039a38-3bd4-4084-95d1-3548c1873c8b","iss":"test","iat":1715773941,"pair":["BTC","USD"],"type":"bid","status":"pending","amount":{"v":"100000000","e":8,"f":1},"price":{"v":"500000","e":5,"f":5}}` {
		t.Errorf("unexpected format for marshalled order: %s", data)
	}

	var b *Order
	err = json.Unmarshal(data, &b)
	if err != nil {
		t.Errorf("failed to unmarshal json: %s", err)
		return
	}

	if a.String() != b.String() {
		t.Errorf("failed: expected same object but got a=%s b=%s", a, b)
	}
}
