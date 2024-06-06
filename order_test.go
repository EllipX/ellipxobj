package ellipxobj

import (
	"encoding/json"
	"log"
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

func TestMatchOrder(t *testing.T) {
	a := NewOrder(Pair("BTC", "USD"), TypeBid).SetId("a9039a38-3bd4-4084-95d1-3548c1873c8b", "test")
	a.Unique = &TimeId{Unix: 1715773941, Nano: 987654321, Index: 42}
	a.RequestTime = 1715773941
	a.Amount, _ = NewAmountFromFloat64(1, 8)
	a.Price, _ = NewAmountFromString("5", 5)

	b := NewOrder(Pair("BTC", "USD"), TypeAsk).SetId("1c3f54ff-1c8e-44ac-a067-c0e0ac7b944c", "test")
	b.Unique = &TimeId{Unix: 1715773941, Nano: 987654321, Index: 43}
	b.RequestTime = 1715773941
	b.Amount, _ = NewAmountFromFloat64(0.5, 8)
	b.Price, _ = NewAmountFromString("5", 5)

	tradeAmt := a.TradeAmount(b)
	if tradeAmt.String() != "0.50000000" {
		t.Errorf("invalid trade amount %s, expected 0.50000000", tradeAmt)
	}

	trade := a.Matches(b)
	if trade == nil {
		t.Errorf("no trade from [%s] vs [%s]", a, b)
	}
	log.Printf("trade = %+v", trade)
}
