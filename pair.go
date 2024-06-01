package ellipxobj

import (
	"encoding/json"
	"errors"
	"strings"
)

type PairName [2]string

func Pair(a, b string) PairName {
	return PairName{a, b}
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
