package ellipxobj

// Trade represents a trade that happened, where two orders matched
type Trade struct {
	Id     *TimeId // trade id
	Pair   PairName
	Bid    *OrderMeta
	Ask    *OrderMeta
	Type   OrderType // taker's order type
	Amount *Amount
	Price  *Amount
}
