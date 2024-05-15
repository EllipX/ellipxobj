package ellipxobj

type Order struct {
	OrderId     string      `json:"id"`     // order ID assigned by the broker
	BrokerId    string      `json:"iss"`    // id of the broker
	RequestTime uint64      `json:"iat"`    // unix timestamp when the order was placed
	Unique      TimeId      `json:"uniq"`   // unique ID allocated on order igress
	Pair        PairName    `json:"pair"`   // the name of the pair the order is on
	Type        OrderType   `json:"type"`   // type of order (buy or sell)
	Status      OrderStatus `json:"status"` // new orders will always be in "pending" state
	Flags       OrderFlags  `json:"flags"`
	Amount      *Amount     `json:"amount,omitempty"`      // optional amount, if nil SpendLimit must be set
	Price       *Amount     `json:"price,omitempty"`       // price, if nil this will be a market order
	SpendLimit  *Amount     `json:"spend_limit,omitempty"` // optional spending limit, if nil Amount must be set
	StopPrice   *Amount     `json:"stop_price,omitempty"`  // ignored if flag Stop is not set
}

type SignedOrder struct {
	Order      string            `json:"order"` // json-encoded
	Signatures []*OrderSignature `json:"sigs"`
}

type OrderSignature struct {
	Issuer    string `json:"iss"` // name of signature issuer (if broker, broker id)
	Signature string `json:"sig"` // base64url encoded signature
}

// Reverse reverses an order's pair, updating Amount and Price accordingly
func (o *Order) Reverse() *Order {
	res := &Order{}
	*res = *o // copy all values

	// reverse pair
	res.Pair = PairName{o.Pair[1], o.Pair[0]}
	res.Type = o.Type.Reverse()

	// reverse amount and spend limit
	res.Amount, res.SpendLimit = o.SpendLimit, o.Amount

	// if we have a target price, set it to 1/price
	if o.Price != nil {
		res.Price, _ = o.Price.Reciprocal()
	}

	return res
}
