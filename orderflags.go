package ellipxobj

import (
	"encoding/json"
	"fmt"
)

type OrderFlags int

const (
	FlagImmediateOrCancel OrderFlags = 1 << iota // do not create an open order after execution
	FlagFillOrKill                               // if order can't be fully executed, cancel
	FlagStop
)

func (f *OrderFlags) UnmarshalJSON(j []byte) error {
	var flags []string
	var res OrderFlags

	for _, s := range flags {
		switch s {
		case "ioc":
			res |= FlagImmediateOrCancel
		case "fok":
			res |= FlagFillOrKill
		case "stop":
			res |= FlagStop
		default:
			return fmt.Errorf("unsupported flag %s", s)
		}
	}

	*f = res
	return nil
}

func (f OrderFlags) MarshalJSON() ([]byte, error) {
	var flags []string

	if f&FlagImmediateOrCancel == FlagImmediateOrCancel {
		flags = append(flags, "ioc")
	}
	if f&FlagFillOrKill == FlagFillOrKill {
		flags = append(flags, "fok")
	}
	if f&FlagStop == FlagStop {
		flags = append(flags, "stop")
	}

	return json.Marshal(flags)
}
