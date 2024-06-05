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
	Flags       OrderFlags  `json:"flags,omitempty"`       // order flags
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

// NominalAmount returns the order's maximum amount if one can be computed, or nil
func (a *Order) NominalAmount(amountExp int) *Amount {
	if a.Price == nil {
		// even if we have a SpendLimit, it can't be used without a price
		return a.Amount
	}
	if a.Amount == nil {
		// we have only Price+SpendLimit
		return NewAmount(0, amountExp).Div(a.SpendLimit, a.Price)
	}

	amt := a.Amount
	if a.SpendLimit != nil {
		// we have both amount & spend limit
		amt2 := NewAmount(0, amountExp).Div(a.SpendLimit, a.Price)
		if amt.Cmp(amt2) > 0 {
			// spend limit is lower than amount, return the spend limit amount
			return amt2
		}
	}
	return amt
}

// TradeAmount returns the order's trade amount in case it matches against b
func (a *Order) TradeAmount(b *Order) *Amount {
	amt := a.Amount
	if amt == nil {
		// no amount means we have a spend limit
		// open orders always have an amount, use that to put the right exp
		// amount = a.SpendLimit / b.Price
		amt = NewAmount(0, b.Amount.exp).Div(a.SpendLimit, b.Price)
	} else if a.SpendLimit != nil {
		amt2 := NewAmount(0, b.Amount.exp).Div(a.SpendLimit, b.Price)
		if amt.Cmp(amt2) > 0 {
			amt = amt2
		}
	}

	// if amt > b.Amount, return b.Amount
	if amt.Cmp(b.Amount) > 0 {
		return b.Amount
	}

	return amt
}

// Matches returns a Trade if a can consume b
// Because b is assumed to be an open order, it must have a Price
func (a *Order) Matches(b *Order) *Trade {
	switch a.Type {
	case TypeBid:
		if b.Type != TypeAsk {
			return nil
		}
		if a.Price != nil {
			if a.Price.Cmp(b.Price) < 0 {
				// bid price lower than ask, trade cannot happen
				return nil
			}
		}
		// compute the traded amount
		amt := a.TradeAmount(b)
		if amt.IsZero() {
			// nothing to trade
			return nil
		}

		t := &Trade{
			Pair:   a.Pair,
			Bid:    a.Meta(),
			Ask:    b.Meta(),
			Type:   TypeBid,
			Amount: amt,
			Price:  b.Price,
		}

		return t
	case TypeAsk:
		if b.Type != TypeBid {
			return nil
		}
		if a.Price != nil {
			if a.Price.Cmp(b.Price) > 0 {
				// ask price higher than bid, trade cannot happen
				return nil
			}
		}
		// compute the traded amount
		amt := a.TradeAmount(b)
		if amt.IsZero() {
			// nothing to trade
			return nil
		}

		t := &Trade{
			Pair:   a.Pair,
			Bid:    a.Meta(),
			Ask:    b.Meta(),
			Type:   TypeAsk,
			Amount: amt,
			Price:  b.Price,
		}

		return t
	default:
		return nil
	}
}

// Deduct deducts the trade's value from the order, and return true if this order
// is fully consumed. Note that even if an order isn't fully consumed, it might be
// too low to be executed.
func (o *Order) Deduct(t *Trade) bool {
	fullyConsumed := false
	if o.Amount != nil {
		o.Amount = o.Amount.Sub(o.Amount, t.Amount)
		if o.Amount.IsZero() {
			fullyConsumed = true
		}
	}
	if o.SpendLimit != nil {
		o.SpendLimit = o.SpendLimit.Sub(o.SpendLimit, t.Spent())
		if o.SpendLimit.Sign() < 0 {
			// rounding may cause a <0 result
			o.SpendLimit = NewAmount(0, o.SpendLimit.exp)
			fullyConsumed = true
		}
	}

	return fullyConsumed
}
