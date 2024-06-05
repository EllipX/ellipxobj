package ellipxobj

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
)

// Amount is a fixed point value
type Amount struct {
	value *big.Int
	exp   int
}

// NewAmount returns a new Amount object set to the specific value and decimals
func NewAmount(value int64, decimals int) *Amount {
	a := &Amount{
		value: big.NewInt(value),
		exp:   decimals,
	}
	return a
}

func (a Amount) Float() *big.Float {
	res := new(big.Float).SetInt(a.value)

	// divide by 10**exp
	return res.Quo(res, exp10f(a.exp))
}

// NewAmountFromFloat64 return a new [Amount] initialized with the value f stored with the specified number of decimals
func NewAmountFromFloat64(f float64, exp int) (*Amount, big.Accuracy) {
	return NewAmountFromFloat(big.NewFloat(f), exp)
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
		value: val,
		exp:   decimals,
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
	res = res.Mul(res, exp10f(a.exp))

	var acc big.Accuracy
	a.value, acc = res.Int(a.value)
	return a, acc
}

// Reciprocal returns 1/a in a newly allocated [Amount]
func (a *Amount) Reciprocal() (*Amount, big.Accuracy) {
	v := new(big.Float).Quo(new(big.Float).SetInt64(1), a.Float())
	return NewAmountFromFloat(v, a.exp)
}

// SetExp sets the number of decimals (exponent) of the amount, truncating or adding zeroes as needed
func (a *Amount) SetExp(e int) *Amount {
	if a.exp == e {
		// no change
		return a
	}

	if e > a.exp {
		add := e - a.exp
		a.exp = e
		a.value = a.value.Mul(a.value, exp10(add))
		return a
	}

	// e < a.exp
	// using the trick of adding 0.5 (half of exp10(sub)) to cause rounding to happen
	sub := a.exp - e
	e10 := exp10(sub)
	e10half := new(big.Int).Quo(e10, big.NewInt(2)) // 1/2
	if a.value.Sign() < 0 {
		e10half = e10half.Mul(e10half, big.NewInt(-1))
	}
	a.exp = e
	a.value = a.value.Add(a.value, e10half)
	a.value = a.value.Quo(a.value, exp10(sub))
	return a
}

func (a Amount) String() string {
	// rather than converting to float, we simply convert the int to string in base 10 and add a decimal . at the right place
	s := a.value.Text(10)

	if len(s) > a.exp {
		p := len(s) - a.exp
		return s[:p] + "." + s[p:]
	}
	if len(s) < a.exp {
		// need to add zeroes
		p := a.exp - len(s)
		return "0." + strings.Repeat("0", p) + s
	}

	// len(s) == a.exp
	return "0." + s
}

// Cmp compares two amounts, assuming both have the same exponent
func (a Amount) Cmp(b Amount) int {
	if a.exp != b.exp {
		panic("only amounts with same exponent can be compared")
	}
	return a.value.Cmp(b.value)
}

type amountJson struct {
	Value string  `json:"v"`
	Exp   int     `json:"e"`
	Float float64 `json:"f"`
}

func (a *Amount) MarshalJSON() ([]byte, error) {
	// an amount when marshalled becomes an object {"v":"123456","dec":5,"f":1.23456}
	f, _ := a.Float().Float64()
	v := &amountJson{Value: a.value.Text(10), Exp: a.exp, Float: f}
	return json.Marshal(v)
}

func (a *Amount) UnmarshalJSON(b []byte) error {
	var v any
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	err := dec.Decode(&v)
	if err != nil {
		return err
	}

	switch in := v.(type) {
	case string:
		// parse string
		na, err := NewAmountFromString(in, 0)
		if err != nil {
			return err
		}
		*a = *na
		return nil
	case json.Number:
		// parse number
		na, err := NewAmountFromString(string(in), 0)
		if err != nil {
			return err
		}
		*a = *na
		return nil
	case map[string]any:
		// we expect to find v+e or f
		// {"v":"100000000","e":8,"f":1}
		v, vok := in["v"].(string)
		e, eok := in["e"].(json.Number)
		if vok && eok {
			realV, vok := new(big.Int).SetString(v, 0)
			if !vok {
				return errors.New("failed to parse v")
			}
			realE, err := e.Int64()
			if err != nil {
				return err
			}
			a.value = realV
			a.exp = int(realE)
			return nil
		}
		// attempt f
		f, fok := in["f"].(json.Number)
		if fok {
			na, err := NewAmountFromString(string(f), 0)
			if err != nil {
				return err
			}
			*a = *na
			return nil
		}
		return fmt.Errorf("failed to parse object as Amount")
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
