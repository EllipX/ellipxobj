package ellipxobj

import (
	"encoding/json"
	"fmt"
)

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

func (t OrderType) IsValid() bool {
	switch t {
	case TypeBid, TypeAsk:
		return true
	default:
		return false
	}
}

func (t OrderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *OrderType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	v := OrderTypeByString(s)
	if v == TypeInvalid {
		return fmt.Errorf("invalid order type %q", s)
	}
	*t = v
	return nil
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
