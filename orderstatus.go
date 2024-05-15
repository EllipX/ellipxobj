package ellipxobj

import (
	"encoding/json"
	"fmt"
)

type OrderStatus int

const (
	OrderInvalid OrderStatus = -1
	OrderPending OrderStatus = iota
	OrderRunning
	OrderOpen
	OrderStop // pending for a trigger
	OrderDone
	OrderCancel // cancelled or overwritten order
)

func (s OrderStatus) String() string {
	switch s {
	case OrderPending:
		return "pending"
	case OrderRunning:
		return "running"
	case OrderOpen:
		return "open"
	case OrderStop:
		return "stop"
	case OrderDone:
		return "done"
	case OrderCancel:
		return "cancel"
	default:
		return "invalid"
	}
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderPending, OrderRunning, OrderOpen, OrderStop, OrderDone, OrderCancel:
		return true
	default:
		return false
	}
}

func OrderStatusByString(s string) OrderStatus {
	switch s {
	case "pending":
		return OrderPending
	case "running":
		return OrderRunning
	case "open":
		return OrderOpen
	case "stop":
		return OrderStop
	case "done":
		return OrderDone
	case "cancel":
		return OrderCancel
	default:
		return OrderInvalid
	}
}

func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *OrderStatus) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	v := OrderStatusByString(str)
	if v == OrderInvalid {
		return fmt.Errorf("invalid order status %s", s)
	}
	*s = v
	return nil
}
