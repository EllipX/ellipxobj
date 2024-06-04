package ellipxobj

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const TimeIdDataLen = 16 // number of bytes in a timeid

type TimeId struct {
	Type  string `json:"type"` // order | trade
	Unix  uint64 `json:"unix"` // unix timestamp in seconds
	Nano  uint32 `json:"nano"` // [0, 999999999]
	Index uint32 `json:"idx"`  // index if multiple ids are generated with the same unix/nano values
}

type TimeIdUnique struct {
	Last TimeId
}

var uniqueTime TimeIdUnique

// NewTimeId returns a new TimeId for now
func NewTimeId() *TimeId {
	t := time.Now()
	res := &TimeId{
		Unix: uint64(t.Unix()),
		Nano: uint32(t.Nanosecond()),
	}
	return res
}

// NewUniqueTimeId returns a unique (in the local process) [TimeId]
func NewUniqueTimeId() *TimeId {
	t := NewTimeId()
	uniqueTime.Unique(t)
	return t
}

// Time returns the [TimeId] timestamp, which may be when the ID was generated
func (t TimeId) Time() time.Time {
	return time.Unix(int64(t.Unix), int64(t.Nano))
}

func (t TimeId) String() string {
	if t.Type != "" {
		return fmt.Sprintf("%s:%d:%d:%d", t.Type, t.Unix, t.Nano, t.Index)
	}
	return fmt.Sprintf("%d:%d:%d", t.Unix, t.Nano, t.Index)
}

func (t TimeId) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t *TimeId) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	vA := strings.SplitN(s, ":", 4)
	if len(vA) < 3 {
		return fmt.Errorf("invalid format for TimeId: %s", s)
	}
	typ := ""
	if len(vA) == 4 {
		typ = vA[0]
		vA = vA[1:]
	}
	vN := make([]uint64, 3)
	for n, sub := range vA {
		vN[n], err = strconv.ParseUint(sub, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse TimeId element %s: %w", sub, err)
		}
	}

	t.Type = typ
	t.Unix = vN[0]
	t.Nano = uint32(vN[1])
	t.Index = uint32(vN[2])
	return nil
}

// Bytes returns a 128bits bigendian sortable version of this TimeId
func (t TimeId) Bytes() [TimeIdDataLen]byte {
	var res [TimeIdDataLen]byte
	binary.BigEndian.PutUint64(res[:8], t.Unix)
	binary.BigEndian.PutUint32(res[8:12], t.Nano)
	binary.BigEndian.PutUint32(res[12:], t.Index)

	return res
}

// Unique ensures the provided [TimeId] is always higher than the latest
// one and will update it if not the case
func (u *TimeIdUnique) Unique(t *TimeId) {
	if t.Unix > u.Last.Unix {
		u.Last = *t
		return
	}
	if t.Unix == u.Last.Unix {
		if t.Nano > u.Last.Nano {
			u.Last = *t
			return
		}
		if t.Nano == u.Last.Nano {
			if t.Index > u.Last.Index {
				u.Last.Index = t.Index
				return
			}
		}
	}

	// re-use last
	u.Last.Index += 1
	*t = u.Last
}

// Cmp returns an integer comparing two TimeId time point. The result will be 0 if a == b, -1 if a < b, and +1 if a > b.
func (a TimeId) Cmp(b TimeId) int {
	if a.Unix > b.Unix {
		return 1
	} else if a.Unix < b.Unix {
		return -1
	}

	if a.Nano > b.Nano {
		return 1
	} else if a.Nano < b.Nano {
		return -1
	}

	if a.Index > b.Index {
		return 1
	} else if a.Index < b.Index {
		return -1
	}

	return 0
}
