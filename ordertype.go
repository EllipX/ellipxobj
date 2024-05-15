package ellipxobj

type OrderType int

const (
	TypeInvalid OrderType = -1
	TypeBid     OrderType = iota // buy
	TypeAsk                      // sell
)

func (t OrderType) String() string {
	switch t {
	case TypeBid:
		return "bid"
	case TypeAsk:
		return "ask"
	default:
		return "invalid"
	}
}

func (t OrderType) Reverse() OrderType {
	switch t {
	case TypeBid:
		return TypeAsk
	case TypeAsk:
		return TypeBid
	default:
		return TypeInvalid
	}
}

func OrderTypeByString(v string) OrderType {
	switch v {
	case "bid":
		return TypeBid
	case "ask":
		return TypeAsk
	default:
		return TypeInvalid
	}
}
