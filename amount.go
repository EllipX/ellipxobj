package ellipxobj

import (
	"math/big"
	"sync"
)

// Amount is a fixed point value
type Amount struct {
	Value    *big.Int
	Decimals int
}

// NewAmount returns a new Amount object set to the specific value and decimals
func NewAmount(value int64, decimals int) *Amount {
	a := &Amount{
		Value:    big.NewInt(value),
		Decimals: decimals,
	}
	return a
}

func (a Amount) Float() *big.Float {
	res := new(big.Float).SetInt(a.Value)

	// divide by 10**Decimals
	return res.Quo(res, exp10f(a.Decimals))
}

// NewAmountFromFloat return a new [Amount] initialized with the value f stored with the specified number of decimals
func NewAmountFromFloat(f *big.Float, decimals int) (*Amount, big.Accuracy) {
	// multiply f by 10**decimals
	val, acc := new(big.Float).Mul(f, exp10f(decimals)).Int(nil)

	a := &Amount{
		Value:    val,
		Decimals: decimals,
	}

	return a, acc
}

// Mul sets a=x*y and returns a
func (a *Amount) Mul(x, y *Amount) (*Amount, big.Accuracy) {
	res := new(big.Float).Mul(x.Float(), y.Float())
	res = res.Mul(res, exp10f(a.Decimals))

	var acc big.Accuracy
	a.Value, acc = res.Int(a.Value)
	return a, acc
}

// Reciprocal returns 1/a in a newly allocated [Amount]
func (a *Amount) Reciprocal() (*Amount, big.Accuracy) {
	v := new(big.Float).Quo(new(big.Float).SetInt64(1), a.Float())
	return NewAmountFromFloat(v, a.Decimals)
}

var (
	exp10cache = make(map[int]*big.Float)
	exp10lock  sync.RWMutex
)

// exp10f returns 10**v as [big.Float], caching results since it's likely we'll need the same values more than once
func exp10f(v int) *big.Float {
	exp10lock.RLock()
	res, ok := exp10cache[v]
	exp10lock.RUnlock()

	if ok {
		return res
	}

	dec := new(big.Int).Exp(new(big.Int).SetInt64(10), new(big.Int).SetInt64(int64(v)), nil)
	res = new(big.Float).SetInt(dec)

	exp10lock.Lock()
	defer exp10lock.Unlock()

	exp10cache[v] = res
	return res
}
