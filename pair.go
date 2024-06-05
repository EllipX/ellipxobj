package ellipxobj

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"strings"
)

type PairName [2]string

func Pair(a, b string) PairName {
	return PairName{a, b}
}

func ParsePairName(s string) (PairName, error) {
	pos := strings.IndexByte(s, '_')
	if pos == -1 {
		return PairName{"", ""}, errors.New("malformed pair")
	}
	return PairName{s[:pos], s[pos+1:]}, nil
}

func (p PairName) String() string {
	return p[0] + "_" + p[1]
}

func (p *PairName) UnmarshalJSON(v []byte) error {
	// can be either a string or an array of two strings
	if string(v) == "null" {
		// do nothing
		return nil
	}

	switch v[0] {
	case '[':
		var t [2]string
		err := json.Unmarshal(v, &t)
		if err != nil {
			return err
		}
		*p = PairName(t)
		return nil
	case '"':
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

// Hash returns a 32 bytes hash (sha256) representing the pair name with a nil
// character between the two names
func (p *PairName) Hash() [32]byte {
	b := append([]byte(p[0]), append([]byte{0}, p[1]...)...)
	return sha256.Sum256(b)
}
