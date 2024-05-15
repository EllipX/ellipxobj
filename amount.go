package ellipxobj

import "math/big"

type Amount struct {
	Value    *big.Int
	Decimals int
}

// NewAmount returns a new Amount object set to the specific value and decimals
func NewAmount(value int64, decimals int) *Amount {
	a := &Amount{
		Value:    new(big.Int).SetInt64(value),
		Decimals: decimals,
	}
	return a
}

func (a Amount) Float() *big.Float {
	res := new(big.Float).SetInt(a.Value)

	// divide by 10**Decimals
	dec := new(big.Int).Exp(new(big.Int).SetInt64(10), new(big.Int).SetInt64(int64(a.Decimals)), nil)

	return res.Quo(res, new(big.Float).SetInt(dec))
}

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

// exp10f returns 10**v as [big.Float]
func exp10f(v int) *big.Float {
	res := new(big.Int).Exp(new(big.Int).SetInt64(10), new(big.Int).SetInt64(int64(v)), nil)
	return new(big.Float).SetInt(res)
}
