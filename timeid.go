package ellipxobj

import (
	"encoding/binary"
	"time"
)

type TimeId struct {
	Unix  uint64 `json:"unix"`
	Nano  uint32 `json:"nano"` // [0, 999999999]
	Index uint32 `json:"idx"`  // index if multiple ids are generated with the same unix/nano values
}

func (t TimeId) Time() time.Time {
	return time.Unix(int64(t.Unix), int64(t.Nano))
}

// Bytes returns a 128bits bigendian sortable version of this TimeId
func (t TimeId) Bytes() []byte {
	res := make([]byte, 16)
	binary.BigEndian.PutUint64(res[:8], t.Unix)
	binary.BigEndian.PutUint32(res[8:12], t.Nano)
	binary.BigEndian.PutUint32(res[12:], t.Index)

	return res
}
