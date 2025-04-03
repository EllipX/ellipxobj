package ellipxobj

import (
	"fmt"
	"time"
)

// Order represents a financial trading order with all parameters needed to execute it.
// An order can be either a buy (bid) or sell (ask) on a specific trading pair.
//
// Orders can be specified in two ways:
// 1. Fixed Amount: Specifies the exact quantity of the base asset to trade
// 2. SpendLimit: Specifies the maximum amount of the quote asset to spend/receive
//
// At least one of Amount or SpendLimit must be set for a valid order.
//
// Market orders have nil Price, while limit orders specify the desired price.
// Orders can have various flags that modify their behavior (see OrderFlags).
type Order struct {
	OrderId     string      `json:"id"`                    // Unique order ID assigned by the broker
	BrokerId    string      `json:"iss"`                   // ID of the broker that issued this order
	UserId      string      `json:"usr,omitempty"`         // Optional ID or hash of the user owner of the order
	RequestTime uint64      `json:"iat"`                   // Unix timestamp when the order was placed
	Unique      *TimeId     `json:"uniq,omitempty"`        // Unique ID allocated on order ingress for strict ordering
	Target      *TimeId     `json:"target,omitempty"`      // Target order to be updated (for order modifications)
	Version     uint64      `json:"ver"`                   // Version counter, incremented each time order is modified
	Pair        PairName    `json:"pair"`                  // Trading pair (e.g., BTC_USD)
	Type        OrderType   `json:"type"`                  // Type of order (BID/ASK, Buy/Sell)
	Status      OrderStatus `json:"status"`                // Current status of the order (Pending, Open, Filled, etc.)
	Flags       OrderFlags  `json:"flags,omitempty"`       // Special behavior flags (IOC, FOK, etc.)
	Amount      *Amount     `json:"amount,omitempty"`      // Quantity of base asset to trade (if nil, SpendLimit must be set)
	Price       *Amount     `json:"price,omitempty"`       // Limit price (if nil, this is a market order)
	SpendLimit  *Amount     `json:"spend_limit,omitempty"` // Maximum amount of quote asset to spend/receive (if nil, Amount must be set)
	StopPrice   *Amount     `json:"stop_price,omitempty"`  // Trigger price for stop orders (ignored if Stop flag not set)
}

type OrderMeta struct {
	OrderId  string  `json:"id"`
	BrokerId string  `json:"iss"`
	Unique   *TimeId `json:"uniq,omitempty"`
}

// NewOrder creates a new order with the specified pair and type.
// It initializes the order with the current time and sets the status to Pending.
// After creation, the order should have OrderId and BrokerId set with SetId(),
// and at least one of Amount or SpendLimit must be set for the order to be valid.
//
// Example:
//
//	order := NewOrder(Pair("BTC", "USD"), OrderBid).
//	    SetId("order123", "broker1").
//	    SetAmount(NewAmount(10000, 8)). // 0.1 BTC
//	    SetPrice(NewAmount(2000000, 2)) // $20,000.00
func NewOrder(pair PairName, typ OrderType) *Order {
	res := &Order{
		RequestTime: uint64(time.Now().Unix()),
		Pair:        pair,
		Type:        typ,
		Status:      OrderPending,
	}

	return res
}

// SetId sets the order and broker identifiers on this order.
// Every order must have these IDs set before it can be considered valid.
// Returns the order itself for method chaining.
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

	// reverse amount and spend limit, duplicate
	res.Amount, res.SpendLimit = o.SpendLimit.Dup(), o.Amount.Dup()

	// if we have a target price, set it to 1/price
	if o.Price != nil {
		res.Price, _ = o.Price.Reciprocal()
	}

	return res
}

// Dup returns a copy of Order including objects such as Amount duplicated
func (o *Order) Dup() *Order {
	if o == nil {
		return nil
	}

	res := &Order{}
	*res = *o

	if o.Unique != nil {
		res.Unique = &TimeId{}
		*res.Unique = *o.Unique
	}

	res.Amount = o.Amount.Dup()
	res.Price = o.Price.Dup()
	res.SpendLimit = o.SpendLimit.Dup()
	res.StopPrice = o.StopPrice.Dup()

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

// NominalAmount calculates and returns the effective quantity of the base asset
// that would be traded, considering both Amount and SpendLimit constraints.
//
// This method handles different order configurations:
// - For market orders (nil Price) with Amount: Returns the Amount directly
// - For market orders with only SpendLimit: Returns nil (cannot calculate without price)
// - For limit orders with only SpendLimit: Calculates Amount = SpendLimit / Price
// - For limit orders with both Amount and SpendLimit: Returns the smaller of:
//   - The specified Amount
//   - The calculated amount based on SpendLimit/Price
//
// The amountExp parameter specifies the decimal precision for the returned Amount.
func (a *Order) NominalAmount(amountExp int) *Amount {
	if a.Price == nil {
		// For market orders (nil Price), we can't calculate based on SpendLimit
		// even if we have a SpendLimit, it can't be used without a price
		return a.Amount
	}

	if a.Amount == nil {
		// We have only Price+SpendLimit (SpendLimit-based limit order)
		// Calculate Amount = SpendLimit / Price
		return NewAmount(0, amountExp).Div(a.SpendLimit, a.Price)
	}

	// We have an Amount, possibly with a SpendLimit too
	amt := a.Amount
	if a.SpendLimit != nil {
		// We have both Amount & SpendLimit - check which is more restrictive
		amt2 := NewAmount(0, amountExp).Div(a.SpendLimit, a.Price)
		if amt.Cmp(amt2) > 0 {
			// SpendLimit is more restrictive than Amount, use it instead
			return amt2
		}
	}
	return amt
}

// TradeAmount calculates the actual amount that can be traded between this order and order b.
// This is a key component of the matching engine, determining the exact quantity
// to use when creating a trade between two orders.
//
// The method calculates this by considering:
// 1. This order's Amount and/or SpendLimit (converted to an amount using b's Price)
// 2. The available amount in the counterparty order (b.Amount)
//
// For SpendLimit-based orders, the amount is calculated as SpendLimit/Price.
// When both Amount and SpendLimit are specified, the more restrictive one is used.
// The result is further limited by the available amount in order b.
//
// Returns the maximum amount that can be traded between the two orders.
func (a *Order) TradeAmount(b *Order) *Amount {
	// Calculate the maximum amount this order can trade
	amt := a.Amount
	if amt == nil {
		// No Amount means we have a SpendLimit
		// Calculate amount = SpendLimit / Price
		// Use b.Amount's precision for consistent decimal handling
		amt = NewAmount(0, b.Amount.exp).Div(a.SpendLimit, b.Price)
	} else if a.SpendLimit != nil {
		// We have both Amount and SpendLimit - check which is more restrictive
		amt2 := NewAmount(0, b.Amount.exp).Div(a.SpendLimit, b.Price)
		if amt.Cmp(amt2) > 0 {
			// SpendLimit is more restrictive than Amount
			amt = amt2
		}
	}

	// Limit by the available amount in order b
	if amt.Cmp(b.Amount) > 0 {
		return b.Amount
	}

	return amt
}

// Matches determines if this order (a) can match with the provided order (b),
// returning a Trade object if a match is possible, or nil if no match can occur.
//
// For a match to occur:
// 1. Orders must have opposite types (bid vs ask)
// 2. For limit orders, prices must be compatible:
//   - For a bid (buy), a.Price must be >= b.Price
//   - For an ask (sell), a.Price must be <= b.Price
//
// 3. There must be a non-zero amount that can be traded
//
// The order 'b' is assumed to be an open resting order with a defined Price.
// The price used for the trade will be b.Price (the resting order's price),
// providing price improvement for the incoming order when possible.
//
// The Type field in the returned Trade indicates the type of the incoming order (a).
func (a *Order) Matches(b *Order) *Trade {
	switch a.Type {
	case TypeBid: // This is a buy order
		// Check order type compatibility
		if b.Type != TypeAsk {
			return nil // Can't match buy with buy
		}

		// Check price compatibility for limit orders
		if a.Price != nil {
			if a.Price.Cmp(b.Price) < 0 {
				// Bid price lower than ask price, trade cannot happen
				return nil
			}
		}

		// Compute the tradable amount considering Amount and SpendLimit
		amt := a.TradeAmount(b)
		if amt.IsZero() {
			// Nothing to trade
			return nil
		}

		// Create the trade object
		t := &Trade{
			Pair:   a.Pair,
			Bid:    a.Meta(),  // This order is the bid
			Ask:    b.Meta(),  // The other order is the ask
			Type:   TypeBid,   // The trade is from the perspective of this order
			Amount: amt.Dup(), // Copy the amount to avoid side effects
			Price:  b.Price,   // Use the resting order's price
		}

		return t

	case TypeAsk: // This is a sell order
		// Check order type compatibility
		if b.Type != TypeBid {
			return nil // Can't match sell with sell
		}

		// Check price compatibility for limit orders
		if a.Price != nil {
			if a.Price.Cmp(b.Price) > 0 {
				// Ask price higher than bid price, trade cannot happen
				return nil
			}
		}

		// Compute the tradable amount considering Amount and SpendLimit
		amt := a.TradeAmount(b)
		if amt.IsZero() {
			// Nothing to trade
			return nil
		}

		// Create the trade object
		t := &Trade{
			Pair:   a.Pair,
			Bid:    b.Meta(),  // The other order is the bid
			Ask:    a.Meta(),  // This order is the ask
			Type:   TypeAsk,   // The trade is from the perspective of this order
			Amount: amt.Dup(), // Copy the amount to avoid side effects
			Price:  b.Price,   // Use the resting order's price
		}

		return t

	default:
		return nil // Invalid order type
	}
}

// Deduct reduces this order's Amount and/or SpendLimit by the quantities in the given trade.
// This is called after a trade is executed to update the order's remaining quantities.
//
// The method handles both Amount-based and SpendLimit-based orders:
// - For Amount-based orders: Reduces Amount by the trade's Amount
// - For SpendLimit-based orders: Reduces SpendLimit by the trade's Spent value
// - For orders with both: Reduces both values appropriately
//
// Returns true if the order is fully consumed (either Amount or SpendLimit reduced to zero).
// Note that even if this method returns false (order not fully consumed), the remaining
// quantity might be too small to execute further trades.
//
// This method modifies the order in place and should be called only once per trade.
func (o *Order) Deduct(t *Trade) bool {
	fullyConsumed := false

	// Update Amount if set
	if o.Amount != nil {
		o.Amount = o.Amount.Sub(o.Amount, t.Amount)
		if o.Amount.IsZero() {
			fullyConsumed = true
		}
	}

	// Update SpendLimit if set
	if o.SpendLimit != nil {
		o.SpendLimit = o.SpendLimit.Sub(o.SpendLimit, t.Spent())
		if o.SpendLimit.Sign() < 0 {
			// Rounding may cause a negative result, set to zero in that case
			o.SpendLimit = NewAmount(0, o.SpendLimit.exp)
			fullyConsumed = true
		}
	}

	return fullyConsumed
}
