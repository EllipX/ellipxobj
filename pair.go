package ellipxobj

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"
	"sync"
)

// PairName represents a trading pair with base and quote currency.
// For example, in BTC_USD, BTC is the base currency (index 0) and
// USD is the quote currency (index 1).
type PairName [2]string

// Thread-safe cache for pair hash calculations.
// This improves performance by avoiding recalculation of hashes
// for frequently used currency pairs.
var (
	pairHashCache   = make(map[PairName][]byte) // Cache mapping pairs to their hash values
	pairHashCacheLk sync.RWMutex                // Mutex to protect concurrent access to the cache
)

// Pair creates a new PairName from base and quote currency strings.
// For example, Pair("BTC", "USD") creates the BTC_USD trading pair.
func Pair(a, b string) PairName {
	return PairName{a, b}
}

// ParsePairName parses a string in the format "BASE_QUOTE" into a PairName.
// Returns an error if the string doesn't contain the underscore separator.
// For example, "BTC_USD" would be parsed into PairName{"BTC", "USD"}.
func ParsePairName(s string) (PairName, error) {
	pos := strings.IndexByte(s, '_')
	if pos == -1 {
		return PairName{"", ""}, errors.New("malformed pair")
	}
	return PairName{s[:pos], s[pos+1:]}, nil
}

// String returns the string representation of a PairName in the format "BASE_QUOTE".
// For example, PairName{"BTC", "USD"}.String() returns "BTC_USD".
func (p PairName) String() string {
	return p[0] + "_" + p[1]
}

// UnmarshalJSON implements the json.Unmarshaler interface for PairName.
// Supports two JSON formats:
// 1. String format: "BTC_USD"
// 2. Array format: ["BTC", "USD"]
// Returns an error if the format is invalid or if the string format
// doesn't contain the required underscore separator.
func (p *PairName) UnmarshalJSON(v []byte) error {
	// Handle null value
	if string(v) == "null" {
		// do nothing
		return nil
	}

	switch v[0] {
	case '[':
		// Array format: ["BTC", "USD"]
		var t [2]string
		err := json.Unmarshal(v, &t)
		if err != nil {
			return err
		}
		*p = PairName(t)
		return nil
	case '"':
		// String format: "BTC_USD"
		var t string
		err := json.Unmarshal(v, &t)
		if err != nil {
			return err
		}
		// must contain a _
		pos := strings.IndexByte(t, '_')
		if pos == -1 {
			return errors.New("malformed pair")
		}
		(*p)[0] = t[:pos]
		(*p)[1] = t[pos+1:]
		return nil
	default:
		return errors.New("cannot unmarshal json into pair")
	}
}

// Hash returns a 32-byte SHA-256 hash representing the pair name.
// The hash is calculated by concatenating the base currency, a nil character,
// and the quote currency, then computing the SHA-256 hash of this string.
//
// Results are cached in a thread-safe map for performance optimization,
// making repeated calls with the same pair very efficient.
func (p PairName) Hash() []byte {
	// First check cache with read lock
	pairHashCacheLk.RLock()
	v, ok := pairHashCache[p]
	pairHashCacheLk.RUnlock()

	if ok {
		return v
	}

	// Cache miss - calculate hash with write lock
	pairHashCacheLk.Lock()
	defer pairHashCacheLk.Unlock()

	// Double-check in case another goroutine calculated it while we were waiting
	if v, ok := pairHashCache[p]; ok {
		return v
	}

	// Create byte slice with base, nil byte, and quote currency
	b := append([]byte(p[0]), append([]byte{0}, p[1]...)...)
	h := sha256.Sum256(b)

	// Store in cache and return
	pairHashCache[p] = h[:]
	return h[:]
}
