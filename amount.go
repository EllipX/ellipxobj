package ellipxobj

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
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

// NewAmountFromFloat64 return a new [Amount] initialized with the value f stored with the specified number of decimals
func NewAmountFromFloat64(f float64, decimals int) (*Amount, big.Accuracy) {
	return NewAmountFromFloat(big.NewFloat(f), decimals)
}

// NewAmountFromFloat return a new [Amount] initialized with the value f stored with the specified number of decimals
func NewAmountFromFloat(f *big.Float, decimals int) (*Amount, big.Accuracy) {
	if decimals <= 0 {
		// let's attempt to guess a good decimal value
		s := f.Text('f', -1)
		pos := strings.IndexByte(s, '.')
		if pos == -1 {
			// no decimals at all?
			decimals = 5
		} else {
			// 123.456 pos=3 len(s)=7
			decimals = len(s) - pos - 1
		}
	}
	if decimals < 5 {
		decimals = 5
	}

	// multiply f by 10**decimals
	f = new(big.Float).Mul(f, exp10f(decimals))

	// add 0.5 so that f.Int returns a rounded value
	f = f.Add(f, big.NewFloat(0.5*float64(f.Sign())))
	val, acc := f.Int(nil)

	a := &Amount{
		Value:    val,
		Decimals: decimals,
	}

	return a, acc
}

// NewAmountFromString return a new [Amount] initialized with the passed string value
func NewAmountFromString(s string, decimals int) (*Amount, error) {
	f, _, err := big.ParseFloat(s, 0, 1024, big.ToNearestEven)
	if err != nil {
		return nil, err
	}

	a, _ := NewAmountFromFloat(f, decimals)
	return a, nil
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

// SetExp sets the number of decimals (exponent) of the amount, truncating or adding zeroes as needed
func (a *Amount) SetExp(e int) *Amount {
	if a.Decimals == e {
		// no change
		return a
	}

	if e > a.Decimals {
		add := e - a.Decimals
		a.Decimals = e
		a.Value = a.Value.Mul(a.Value, exp10(add))
		return a
	}

	// e < a.Decimals
	sub := a.Decimals - e
	a.Decimals = e
	a.Value = a.Value.Quo(a.Value, exp10(sub))
	return a
}

func (a Amount) String() string {
	// rather than converting to float, we simply convert the int to string in base 10 and add a decimal . at the right place
	s := a.Value.Text(10)

	if len(s) > a.Decimals {
		p := len(s) - a.Decimals
		return s[:p] + "." + s[p:]
	}
	if len(s) < a.Decimals {
		// need to add zeroes
		p := a.Decimals - len(s)
		return "0." + strings.Repeat("0", p) + s
	}

	// len(s) == a.Decimals
	return "0." + s
}

type amountJson struct {
	Value    string  `json:"v"`
	Decimals int     `json:"dec"`
	Float    float64 `json:"f"`
}

func (a *Amount) MarshalJSON() ([]byte, error) {
	// an amount when marshalled becomes an object {"v":"123456","dec":5,"f":1.23456}
	f, _ := a.Float().Float64()
	v := &amountJson{Value: a.Value.Text(10), Decimals: a.Decimals, Float: f}
	return json.Marshal(v)
}

func (a *Amount) UnmarshalJSON(b []byte) error {
	var v any
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	switch in := v.(type) {
	case string:
		// parse string
		na, err := NewAmountFromString(in, a.Decimals)
		if err != nil {
			return err
		}
		*a = *na
		return nil
	default:
		return fmt.Errorf("unsupported amount type %T", v)
	}
}

var (
	exp10cache  = make(map[int]*big.Int)
	exp10fcache = make(map[int]*big.Float)
	exp10lock   sync.RWMutex
)

// exp10 returns 10**v as [big.Int], caching results since it's likely we'll need the same values more than once
func exp10(v int) *big.Int {
	exp10lock.RLock()
	res, ok := exp10cache[v]
	exp10lock.RUnlock()

	if ok {
		return res
	}

	res = new(big.Int).Exp(new(big.Int).SetInt64(10), new(big.Int).SetInt64(int64(v)), nil)

	exp10lock.Lock()
	defer exp10lock.Unlock()

	exp10cache[v] = res
	return res
}

// exp10f returns 10**v as [big.Float], caching results since it's likely we'll need the same values more than once
func exp10f(v int) *big.Float {
	exp10lock.RLock()
	res, ok := exp10fcache[v]
	exp10lock.RUnlock()

	if ok {
		return res
	}

	res = new(big.Float).SetInt(exp10(v))

	exp10lock.Lock()
	defer exp10lock.Unlock()

	exp10fcache[v] = res
	return res
}
