package ellipxobj

import "math/big"

type Amount struct {
	Value    *big.Int
	Decimals int
}
