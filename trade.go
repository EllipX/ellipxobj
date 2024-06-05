package ellipxobj

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

// Spent returns the amount spent in that trade
func (t *Trade) Spent() *Amount {
	c, _ := NewAmount(0, t.Price.exp).Mul(t.Amount, t.Price)
	return c
}
