package ellipxobj

import (
	"fmt"
	"time"
)

type Order struct {
	OrderId     string      `json:"id"`                    // order ID assigned by the broker
	BrokerId    string      `json:"iss"`                   // id of the broker
	RequestTime uint64      `json:"iat"`                   // unix timestamp when the order was placed
	Unique      *TimeId     `json:"uniq,omitempty"`        // unique ID allocated on order igress
	Pair        PairName    `json:"pair"`                  // the name of the pair the order is on
	Type        OrderType   `json:"type"`                  // type of order (buy or sell)
	Status      OrderStatus `json:"status"`                // new orders will always be in "pending" state
	Flags       OrderFlags  `json:"flags"`                 // order flags
	Amount      *Amount     `json:"amount,omitempty"`      // optional amount, if nil SpendLimit must be set
	Price       *Amount     `json:"price,omitempty"`       // price, if nil this will be a market order
	SpendLimit  *Amount     `json:"spend_limit,omitempty"` // optional spending limit, if nil Amount must be set
	StopPrice   *Amount     `json:"stop_price,omitempty"`  // ignored if flag Stop is not set
}

type OrderMeta struct {
	OrderId  string  `json:"id"`
	BrokerId string  `json:"iss"`
	Unique   *TimeId `json:"uniq,omitempty"`
}

func NewOrder(pair PairName, typ OrderType) *Order {
	res := &Order{
		RequestTime: uint64(time.Now().Unix()),
		Pair:        pair,
		Type:        typ,
		Status:      OrderPending,
	}

	return res
}

func (o *Order) SetId(orderId, brokerId string) *Order {
	o.OrderId = orderId
	o.BrokerId = brokerId
	return o
}

func (o *Order) IsValid() error {
	if o.OrderId == "" {
		return ErrOrderIdMissing
	}
	if o.BrokerId == "" {
		return ErrBrokerIdMissing
	}
	if !o.Type.IsValid() {
		return ErrOrderTypeNotValid
	}
	if !o.Status.IsValid() {
		return ErrOrderStatusNotValid
	}
	if o.Amount == nil && o.SpendLimit == nil {
		return ErrOrderNeedsAmount
	}

	return nil
}

func (o *Order) Meta() *OrderMeta {
	res := &OrderMeta{
		OrderId:  o.OrderId,
		BrokerId: o.BrokerId,
		Unique:   o.Unique,
	}
	return res
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

func (o *Order) String() string {
	// transform order into a human friendly string, there are various cases we can process
	if err := o.IsValid(); err != nil {
		return fmt.Sprintf("(invalid order: %s)", err)
	}

	res := "buy"
	if o.Type == TypeAsk {
		res = "sell"
	}
	res += " "
	if o.Amount != nil && o.SpendLimit != nil {
		res += fmt.Sprintf("%s %s or up to %s %s worth", o.Amount, o.Pair[0], o.SpendLimit, o.Pair[1])
	} else if o.SpendLimit != nil {
		res += fmt.Sprintf("%s %s worth", o.SpendLimit, o.Pair[1])
	} else if o.Amount != nil {
		res += fmt.Sprintf("%s %s", o.Amount, o.Pair[0])
	}

	if o.Price != nil {
		res += fmt.Sprintf(" @ %s %s/%s", o.Price, o.Pair[1], o.Pair[0])
	}

	return res
}
