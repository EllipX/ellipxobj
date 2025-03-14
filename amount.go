// Package ellipxobj provides core types for a trading/exchange system.
package ellipxobj

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/KarpelesLab/typutil"
)

// Amount represents a fixed-point decimal value with arbitrary precision.
// It uses a big.Int for the value and an exponent to represent the decimal position.
// For example, 123.456 would be stored as value=123456 and exp=3.
// This allows for precise decimal arithmetic without floating-point errors.
type Amount struct {
	value *big.Int // The integer value (significand)
	exp   int      // The exponent (number of decimal places)
}

// NewAmount returns a new Amount object set to the specific value and decimals.
// For example, NewAmount(12345, 2) creates the value 123.45
func NewAmount(value int64, decimals int) *Amount {
	a := &Amount{
		value: big.NewInt(value),
		exp:   decimals,
	}
	return a
}

// Float converts the Amount to a big.Float representation.
// This method divides the internal integer value by 10^exp to get the
// actual decimal value. Returns zero for zero amounts.
func (a Amount) Float() *big.Float {
	if a.Sign() == 0 {
		return new(big.Float)
	}
	res := new(big.Float).SetInt(a.value)

	// divide by 10**exp
	return res.Quo(res, exp10f(a.exp))
}

// NewAmountFromFloat64 returns a new Amount initialized with the value f
// stored with the specified number of decimal places.
// Returns the Amount and the accuracy of the conversion.
func NewAmountFromFloat64(f float64, exp int) (*Amount, big.Accuracy) {
	return NewAmountFromFloat(big.NewFloat(f), exp)
}

// NewAmountFromFloat returns a new Amount initialized with the big.Float value
// stored with the specified number of decimal places.
// If decimals <= 0, it will automatically determine an appropriate precision
// based on the input value, with a minimum of 5 decimal places.
// Returns the Amount and the accuracy of the conversion.
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

// Dup returns a copy of the Amount object so that modifying one won't affect the other
func (a *Amount) Dup() *Amount {
	if a == nil {
		return nil
	}
	res := &Amount{
		value: new(big.Int).Set(a.value),
		exp:   a.exp,
	}
	return res
}

// NewAmountFromString return a new [Amount] initialized with the passed string value
func NewAmountFromString(s string, decimals int) (*Amount, error) {
	if decimals == 0 {
		pos := strings.IndexAny(s, "eE")
		extraE := 0
		if pos != -1 {
			// we have a eXX value
			v, err := strconv.ParseInt(s[pos+1:], 0, 64)
			if err != nil {
				return nil, fmt.Errorf("%w: %s", ErrAmountParseFailed, err)
			}
			extraE = 0 - int(v)
			s = s[:pos]
		}
		pos = strings.IndexByte(s, '.')
		if pos == -1 {
			v, ok := new(big.Int).SetString(s, 0)
			if !ok {
				return nil, ErrAmountParseFailed
			}
			if extraE < 0 {
				// adjust
				v = v.Mul(v, exp10(0-extraE))
				extraE = 0
			}
			return &Amount{value: v, exp: extraE}, nil
		}
		// we have a dot, assume decimal & take position of dot as value for e
		e := len(s) - pos - 1
		v, ok := new(big.Int).SetString(s[:pos]+s[pos+1:], 10)
		if !ok {
			return nil, ErrAmountParseFailed
		}
		e += extraE
		if e < 0 {
			// adjust
			v = v.Mul(v, exp10(0-e))
			e = 0
		}
		return &Amount{value: v, exp: e}, nil
	}
	f, _, err := big.ParseFloat(s, 0, 1024, big.ToNearestEven)
	if err != nil {
		return nil, err
	}

	a, _ := NewAmountFromFloat(f, decimals)
	return a, nil
}

// NewAmountRaw returns a new [Amount] initialized with the passed values as is
func NewAmountRaw(v *big.Int, decimals int) *Amount {
	return &Amount{value: v, exp: decimals}
}

// Mul sets a=x*y and returns a.
// The multiplication preserves the desired exponent in parameter 'a'
// by automatically adjusting precision after multiplying the values.
// For example, with a.exp=5, x=1.23 (exp=2), y=4.56 (exp=2),
// the result will be 5.60880 adjusted to have 5 decimal places.
func (a *Amount) Mul(x, y *Amount) *Amount {
	if a.value == nil {
		a.value = new(big.Int)
	}
	a.value.Mul(x.value, y.value)
	exp := a.exp
	a.exp = x.exp + y.exp
	return a.SetExp(exp)
}

// Reciprocal returns 1/a in a newly allocated Amount.
// Returns the Amount and the accuracy of the calculation.
// The precision is maintained at the same level as the original Amount.
func (a Amount) Reciprocal() (*Amount, big.Accuracy) {
	v := new(big.Float).Quo(new(big.Float).SetInt64(1), a.Float())
	return NewAmountFromFloat(v, a.exp)
}

// Neg returns -a (the negation) in a newly allocated Amount.
// The exponent/precision remains unchanged.
func (a Amount) Neg() *Amount {
	v := new(big.Int).Neg(a.value)
	return NewAmountRaw(v, a.exp)
}

// Value returns the amount's value *big.Int
func (a Amount) Value() *big.Int {
	return a.value
}

// Exp returns the amount's exp value
func (a Amount) Exp() int {
	return a.exp
}

// SetExp sets the number of decimals (exponent) of the amount.
// When increasing precision (e > a.exp), this adds zeros to the right.
// When decreasing precision (e < a.exp), this rounds the value to the nearest
// decimal place using banker's rounding (round half to even).
//
// Examples:
// - Setting 123.456 from exp=3 to exp=5 gives 123.45600
// - Setting 123.456 from exp=3 to exp=2 gives 123.46
//
// Returns the amount itself for method chaining.
func (a *Amount) SetExp(e int) *Amount {
	if a.exp == e {
		// no change
		return a
	}

	if e > a.exp {
		// Increasing precision (adding decimal places)
		add := e - a.exp
		a.exp = e
		a.value = a.value.Mul(a.value, exp10(add))
		return a
	}

	// Decreasing precision (removing decimal places)
	// Using the trick of adding 0.5 (half of exp10(sub)) to cause rounding to happen
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

	if a.exp == 0 {
		return s
	}

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

func (a Amount) IsZero() bool {
	return a.value.BitLen() == 0
}

func (a Amount) Sign() int {
	if a.value == nil {
		return 0
	}
	return a.value.Sign()
}

// Cmp compares two amounts, assuming both have the same exponent
func (a Amount) Cmp(b *Amount) int {
	if a.exp != b.exp {
		panic("only amounts with same exponent can be compared")
	}
	return a.value.Cmp(b.value)
}

// Div sets a=x/y and returns a.
// The division ensures appropriate precision by automatically
// adjusting x's exponent before performing the division.
// The result maintains the precision specified in parameter 'a'.
func (a *Amount) Div(x, y *Amount) *Amount {
	// When we do x/y, the resulting exponent will be x.exp-y.exp,
	// so we need to add a.exp to x.exp to achieve the desired precision
	x = x.Dup().SetExp(y.exp + a.exp)

	a.value = a.value.Quo(x.value, y.value)
	return a
}

// Add sets a=x+y and returns a.
// Before adding, both x and y are converted to match the precision of 'a'.
// This ensures that decimal places align correctly during addition.
func (a *Amount) Add(x, y *Amount) *Amount {
	if x.exp != a.exp {
		x = x.Dup().SetExp(a.exp)
	}
	if y.exp != a.exp {
		y = y.Dup().SetExp(a.exp)
	}
	a.value = a.value.Add(x.value, y.value)
	return a
}

// Sub sets a=x-y and returns a.
// Before subtracting, both x and y are converted to match the precision of 'a'.
// This ensures that decimal places align correctly during subtraction.
func (a *Amount) Sub(x, y *Amount) *Amount {
	if x.exp != a.exp {
		x = x.Dup().SetExp(a.exp)
	}
	if y.exp != a.exp {
		y = y.Dup().SetExp(a.exp)
	}
	a.value = a.value.Sub(x.value, y.value)
	return a
}

type amountJson struct {
	Value string  `json:"v"`
	Exp   int     `json:"e"`
	Float float64 `json:"f"`
}

func (a Amount) MarshalJSON() ([]byte, error) {
	if a.value == nil {
		v := &amountJson{Value: "0", Exp: a.exp}
		return json.Marshal(v)
	}
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

	return a.Scan(v)
}

func (a *Amount) Scan(v any) error {
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
		e, eok := in["e"]
		if vok && eok {
			realV, vok := new(big.Int).SetString(v, 0)
			if !vok {
				return errors.New("failed to parse v")
			}
			realE, err := typutil.As[int](e)
			if err != nil {
				return err
			}
			a.value = realV
			a.exp = realE
			return nil
		}
		// attempt f
		f, fok := in["f"]
		if fok {
			sf, err := typutil.As[string](f)
			if err != nil {
				return err
			}
			na, err := NewAmountFromString(sf, 0)
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

func (a Amount) Bytes() []byte {
	// convert amount into bytes
	// 0x00 (version) + exp (int) + val
	buf := binary.AppendVarint([]byte{0x00}, int64(a.exp))
	return append(buf, a.value.Bytes()...)
}

func (a Amount) MarshalBinary() ([]byte, error) {
	return a.Bytes(), nil
}

func (a *Amount) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return errors.New("data too short")
	}
	if data[0] != 0 {
		return errors.New("invalid version")
	}
	exp, n := binary.Varint(data[1:])
	if n <= 0 {
		return errors.New("invalid amount encoding")
	}

	// all ready
	a.exp = int(exp)
	a.value = new(big.Int).SetBytes(data[n+1:])
	return nil
}

// Thread-safe caches for powers of 10 calculations.
// These are used to optimize repeated calculations of common exponents.
var (
	exp10cache  = make(map[int]*big.Int)   // Cache for 10^n as big.Int values
	exp10fcache = make(map[int]*big.Float) // Cache for 10^n as big.Float values
	exp10lock   sync.RWMutex               // RWMutex to protect concurrent access to caches
)

// exp10 returns 10^v as a big.Int with thread-safe caching.
// This function is used extensively for exponent-related calculations,
// and caching improves performance significantly for repeated operations.
func exp10(v int) *big.Int {
	// First check the cache with a read lock
	exp10lock.RLock()
	res, ok := exp10cache[v]
	exp10lock.RUnlock()

	if ok {
		return res
	}

	// Cache miss - calculate the value
	res = new(big.Int).Exp(new(big.Int).SetInt64(10), new(big.Int).SetInt64(int64(v)), nil)

	// Store in cache with a write lock
	exp10lock.Lock()
	defer exp10lock.Unlock()

	exp10cache[v] = res
	return res
}

// exp10f returns 10^v as a big.Float with thread-safe caching.
// Like exp10, this optimizes floating point calculations involving powers of 10.
// It reuses the exp10 function's result and converts it to a big.Float.
func exp10f(v int) *big.Float {
	// First check the cache with a read lock
	exp10lock.RLock()
	res, ok := exp10fcache[v]
	exp10lock.RUnlock()

	if ok {
		return res
	}

	// Cache miss - convert from integer version
	res = new(big.Float).SetInt(exp10(v))

	// Store in cache with a write lock
	exp10lock.Lock()
	defer exp10lock.Unlock()

	exp10fcache[v] = res
	return res
}
