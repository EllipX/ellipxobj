package ellipxobj

import (
	"encoding/json"
	"fmt"
	"time"
)

// Trade represents a trade that happened, where two orders matched
type Trade struct {
	Id     *TimeId    `json:"id"` // trade id
	Pair   PairName   `json:"pair"`
	Bid    *OrderMeta `json:"bid"`
	Ask    *OrderMeta `json:"ask"`
	Type   OrderType  `json:"type"` // taker's order type
	Amount *Amount    `json:"amount"`
	Price  *Amount    `json:"price"`
}

// tradeMarshalled is the representation of Trade when marshalled as JSON
type tradeMarshalled struct {
	Id     *TimeId    `json:"id"` // trade id
	Pair   PairName   `json:"pair"`
	Bid    *OrderMeta `json:"bid"`
	Ask    *OrderMeta `json:"ask"`
	Type   OrderType  `json:"type"` // taker's order type
	Amount *Amount    `json:"amount"`
	Price  *Amount    `json:"price"`
	Date   time.Time  `json:"date"`
}

func (t *Trade) MarshalJSON() ([]byte, error) {
	obj := &tradeMarshalled{
		Id:     t.Id,
		Pair:   t.Pair,
		Bid:    t.Bid,
		Ask:    t.Ask,
		Type:   t.Type,
		Amount: t.Amount,
		Price:  t.Price,
		Date:   t.Id.Time(),
	}

	return json.Marshal(obj)
}

// Spent returns the amount spent in that trade
func (t *Trade) Spent() *Amount {
	return NewAmount(0, t.Price.exp).Mul(t.Amount, t.Price)
}

func (t *Trade) String() string {
	return fmt.Sprintf("trade %s %s @ %s %s each (spent %s %s)", t.Amount, t.Pair[0], t.Price, t.Pair[1], t.Spent(), t.Pair[1])
}
