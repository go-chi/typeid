package typeid

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"time"
)

const randomBits = 15

// Int64 is a type-safe compact identifier. Maps to Postgres BIGINT.
//
// # Bit layout
//
//	[48-bit unix ms timestamp][15-bit crypto/rand] = 63 bits, always positive
//
// # Timestamp range
//
// 48-bit millisecond timestamp (same as UUIDv7) covers Unix epoch through
// year 10889. No action needed in our lifetimes.
//
// # Collision resistance
//
// 15 random bits = 32,768 values per millisecond. Collision probability follows
// the birthday problem: ~R²/65,536,000 expected collisions per second for R
// total IDs/sec across all servers.
//
//	    10 IDs/sec → ~1 collision per 7,500 days
//	   100 IDs/sec → ~1 collision per 1.8 hours
//	 1,000 IDs/sec → ~1 collision per 65 seconds
//
// Protect with a UNIQUE constraint and retry on conflict. For high-throughput
// resources use [UUID] instead.
//
// # Ordering (k-sortable)
//
// IDs are k-sortable: the 48-bit timestamp in the high bits dominates sort
// order, so IDs sort by creation time at millisecond granularity. Two IDs
// generated in the exact same millisecond are not ordered relative to each
// other, but they cluster on the same B-tree leaf pages — no impact on
// Postgres insert locality. Clock skew between servers may produce
// out-of-order IDs within that skew window.
type Int64[P Prefixer] struct {
	val int64
}

func NewInt64[P Prefixer]() (Int64[P], error) {
	ms := time.Now().UnixMilli()

	var rb [2]byte
	if _, err := rand.Read(rb[:]); err != nil {
		return Int64[P]{}, fmt.Errorf("typeid: crypto/rand: %w", err)
	}
	r := int64(binary.BigEndian.Uint16(rb[:]) & 0x7FFF)

	return Int64[P]{val: (ms << randomBits) | r}, nil
}

func Int64From[P Prefixer](v int64) (Int64[P], error) {
	if v < 0 {
		return Int64[P]{}, ErrNegativeInt
	}
	return Int64[P]{val: v}, nil
}

func ParseInt64[P Prefixer](s string) (Int64[P], error) {
	suffix, err := splitTypeid[P](s, int64SuffixLen)
	if err != nil {
		return Int64[P]{}, err
	}

	v, err := decodeBase32Int64(suffix)
	if err != nil {
		return Int64[P]{}, err
	}
	return Int64[P]{val: v}, nil
}

func (id Int64[P]) String() string         { return formatID[P](encodeBase32Int64(id.val)) }
func (id Int64[P]) Int64() int64           { return id.val }
func (id Int64[P]) IsZero() bool           { return id.val == 0 }
func (id Int64[P]) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

func (id *Int64[P]) UnmarshalText(data []byte) error {
	parsed, err := ParseInt64[P](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id Int64[P]) Value() (driver.Value, error) { return id.val, nil }

func (id *Int64[P]) Scan(src any) error {
	switch v := src.(type) {
	case int64:
		id.val = v
	case int:
		id.val = int64(v)
	default:
		return fmt.Errorf("typeid: cannot scan %T into Int64", src)
	}
	return nil
}
