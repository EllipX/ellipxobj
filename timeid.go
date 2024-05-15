package ellipxobj

import (
	"encoding/binary"
	"fmt"
	"time"
)

type TimeId struct {
	Unix  uint64 `json:"unix"`
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

func (t TimeId) Time() time.Time {
	return time.Unix(int64(t.Unix), int64(t.Nano))
}

func (t TimeId) String() string {
	return fmt.Sprintf("%d:%d:%d", t.Unix, t.Nano, t.Index)
}

// Bytes returns a 128bits bigendian sortable version of this TimeId
func (t TimeId) Bytes() []byte {
	res := make([]byte, 16)
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
