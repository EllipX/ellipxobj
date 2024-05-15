package ellipxobj

import "errors"

var (
	ErrOrderIdMissing      = errors.New("order id is required")
	ErrBrokerIdMissing     = errors.New("broker id is required")
	ErrOrderTypeNotValid   = errors.New("order type is not valid")
	ErrOrderStatusNotValid = errors.New("order status is not valid")
	ErrOrderNeedsAmount    = errors.New("order amount or spend limit is required")
)
