// Package ellipxobj provides core types for a trading/exchange system.
package ellipxobj

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// TimeIdDataLen defines the number of bytes in the binary representation of a TimeId.
const TimeIdDataLen = 16

// TimeId represents a unique timestamp-based identifier with nanosecond precision.
// It's used for precisely ordering events like orders and trades, and provides
// a unique, comparable, and sortable identifier even when multiple events
// occur at the exact same time.
type TimeId struct {
	Type  string `json:"type"` // Type of object ("order" or "trade")
	Unix  uint64 `json:"unix"` // Unix timestamp in seconds
	Nano  uint32 `json:"nano"` // Nanosecond component [0, 999999999]
	Index uint32 `json:"idx"`  // Sequential index for events occurring at the same nanosecond
}

// TimeIdUnique provides a mechanism to ensure TimeIds are always unique
// and monotonically increasing within a process, even when created
// in rapid succession or with system clock changes.
type TimeIdUnique struct {
	Last TimeId // Tracks the last generated TimeId to ensure uniqueness
}

// Global instance for generating process-wide unique TimeIds
var uniqueTime TimeIdUnique

// NewTimeId returns a new TimeId initialized with the current system time.
// The Index field starts at 0 and Type is left empty.
// Note: This method does not guarantee uniqueness if called in rapid succession.
// Use NewUniqueTimeId for guaranteed uniqueness.
func NewTimeId() *TimeId {
	t := time.Now()
	res := &TimeId{
		Unix: uint64(t.Unix()),
		Nano: uint32(t.Nanosecond()),
	}
	return res
}

// NewUniqueTimeId returns a guaranteed unique TimeId within the current process.
// It uses the global uniqueTime variable to ensure that even if called multiple times
// within the same nanosecond, each TimeId will be unique by incrementing the Index field.
// This ensures strict ordering of events even at extremely high throughput.
func NewUniqueTimeId() *TimeId {
	t := NewTimeId()
	uniqueTime.Unique(t)
	return t
}

// ParseTimeId parses a string representation of a TimeId.
// The expected format is either "type:unix:nano:index" or "unix:nano:index".
// For example:
// - "order:1649134672:123456789:0" (with type)
// - "1649134672:123456789:0" (without type)
//
// Returns an error if the format is incorrect or values cannot be parsed as integers.
func ParseTimeId(s string) (*TimeId, error) {
	vA := strings.SplitN(s, ":", 4)
	if len(vA) < 3 {
		return nil, fmt.Errorf("invalid format for TimeId: %s", s)
	}

	// Handle optional type field
	typ := ""
	if len(vA) == 4 {
		typ = vA[0]
		vA = vA[1:]
	}

	// Parse the numeric components
	vN := make([]uint64, 3)
	var err error
	bits := 64 // First number (unix) is 64-bit
	for n, sub := range vA {
		vN[n], err = strconv.ParseUint(sub, 10, bits)
		if err != nil {
			return nil, fmt.Errorf("failed to parse TimeId element %s: %w", sub, err)
		}
		bits = 32 // nano and index are 32-bit
	}

	t := &TimeId{
		Type:  typ,
		Unix:  vN[0],
		Nano:  uint32(vN[1]),
		Index: uint32(vN[2]),
	}
	return t, nil
}

// Time returns the [TimeId] timestamp, which may be when the ID was generated
func (t TimeId) Time() time.Time {
	return time.Unix(int64(t.Unix), int64(t.Nano))
}

// String returns a string matching this TimeId
func (t TimeId) String() string {
	if t.Type != "" {
		return fmt.Sprintf("%s:%d:%d:%d", t.Type, t.Unix, t.Nano, t.Index)
	}
	// Type should never be empty
	return fmt.Sprintf("nil:%d:%d:%d", t.Unix, t.Nano, t.Index)
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
	bits := 64
	for n, sub := range vA {
		vN[n], err = strconv.ParseUint(sub, 10, bits)
		if err != nil {
			return fmt.Errorf("failed to parse TimeId element %s: %w", sub, err)
		}
		bits = 32
	}

	t.Type = typ
	t.Unix = vN[0]
	t.Nano = uint32(vN[1])
	t.Index = uint32(vN[2])
	return nil
}

// Bytes returns a 128bits (TimeIdDataLen bytes) bigendian sortable version of this TimeId. If buf is not nil, the data
// is appended to it.
func (t TimeId) Bytes(buf []byte) []byte {
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], t.Unix)
	buf = append(buf, tmp[:]...)
	binary.BigEndian.PutUint32(tmp[:4], t.Nano)
	binary.BigEndian.PutUint32(tmp[4:], t.Index)
	return append(buf, tmp[:]...)
}

func (t TimeId) MarshalBinary() ([]byte, error) {
	return t.Bytes(nil), nil
}

// UnmarshalBinary will convert a binary value back to TimeId. Type will not be kept
func (t *TimeId) UnmarshalBinary(v []byte) error {
	if len(v) != 16 {
		return errors.New("bad data length for timeId")
	}
	t.Unix = binary.BigEndian.Uint64(v[:8])
	t.Nano = binary.BigEndian.Uint32(v[8:12])
	t.Index = binary.BigEndian.Uint32(v[12:])
	return nil
}

// Unique ensures the provided TimeId is always higher (later) than the latest
// one processed by this TimeIdUnique instance. If the provided TimeId is
// already higher, it becomes the new "last" value. If not, the TimeId is
// modified to be one increment higher than the current "last" value.
//
// This method guarantees strict time ordering even when:
// - Multiple TimeIds are created within the same nanosecond
// - The system clock moves backward (due to NTP adjustments, etc.)
// - TimeIds are created on different systems with slightly unsynchronized clocks
//
// The comparison follows a hierarchical order: Unix seconds, then nanoseconds, then index.
func (u *TimeIdUnique) Unique(t *TimeId) {
	// If new time is after last recorded time, update last
	if t.Unix > u.Last.Unix {
		u.Last = *t
		return
	}

	if t.Unix == u.Last.Unix {
		// Same second, check nanoseconds
		if t.Nano > u.Last.Nano {
			u.Last = *t
			return
		}

		if t.Nano == u.Last.Nano {
			// Same nanosecond, check index
			if t.Index > u.Last.Index {
				u.Last.Index = t.Index
				return
			}
		}
	}

	// New time is not after last time, so increment the index of the last time
	// and use that instead (ensuring monotonically increasing sequence)
	u.Last.Index += 1
	*t = u.Last
}

// New creates and returns a new TimeId that is guaranteed to be unique
// within the scope of this TimeIdUnique instance.
// This is a convenience method that combines NewTimeId() and Unique() in one call.
func (u *TimeIdUnique) New() *TimeId {
	t := NewTimeId()
	u.Unique(t)
	return t
}

// Cmp compares two TimeId values and returns:
//
//	-1 if a < b (a is earlier than b)
//	 0 if a == b (a and b represent the same moment)
//	+1 if a > b (a is later than b)
//
// The comparison uses a hierarchical approach:
// 1. First comparing Unix seconds
// 2. Then nanoseconds if seconds are equal
// 3. Finally index if both seconds and nanoseconds are equal
//
// This provides a total ordering of TimeId values suitable for sorting.
func (a TimeId) Cmp(b TimeId) int {
	// Compare seconds first
	if a.Unix > b.Unix {
		return 1
	} else if a.Unix < b.Unix {
		return -1
	}

	// Seconds are equal, compare nanoseconds
	if a.Nano > b.Nano {
		return 1
	} else if a.Nano < b.Nano {
		return -1
	}

	// Seconds and nanoseconds are equal, compare index
	if a.Index > b.Index {
		return 1
	} else if a.Index < b.Index {
		return -1
	}

	// All fields are equal
	return 0
}
