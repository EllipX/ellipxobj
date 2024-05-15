package ellipxobj

type Order struct {
	OrderId     string      `json:"id"`     // order ID assigned by the broker
	BrokerId    string      `json:"iss"`    // id of the broker
	RequestTime uint64      `json:"iat"`    // unix timestamp when the order was placed
	Unique      TimeId      `json:"uniq"`   // unique ID allocated on order igress
	Pair        PairName    `json:"pair"`   // the name of the pair the order is on
	Status      OrderStatus `json:"status"` // new orders will always be in "pending" state
	Flags       OrderFlags  `json:"flags"`
	Amount      *Amount     `json:"amount"`
	Price       *Amount     `json:"price"` // price
	SpendLimit  *Amount     `json:"spend_limit,omitempty"`
	StopPrice   *Amount     `json:"stop_price,omitempty"` // ignored if flag Stop is not set
}

type SignedOrder struct {
	Order      string            `json:"order"` // json-encoded
	Signatures []*OrderSignature `json:"sigs"`
}

type OrderSignature struct {
	Issuer    string `json:"iss"` // name of signature issuer (if broker, broker id)
	Signature string `json:"sig"` // base64url encoded signature
}
