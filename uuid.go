package typeid

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UUID is a type-safe UUIDv7 identifier with a compile-time prefix.
// Maps to Postgres uuid.
type UUID[P Prefixer] struct {
	val uuid.UUID
}

// NewUUID generates a fresh UUIDv7-backed identifier with the given prefix.
func NewUUID[P Prefixer]() (UUID[P], error) {
	u, err := uuid.NewV7()
	if err != nil {
		return UUID[P]{}, err
	}
	return UUID[P]{val: u}, nil
}

// UUIDFrom wraps an existing uuid.UUID, returning an error if it is not v7.
func UUIDFrom[P Prefixer](u uuid.UUID) (UUID[P], error) {
	if u.Version() != 7 {
		return UUID[P]{}, ErrOnlyV7
	}
	return UUID[P]{val: u}, nil
}

// ParseUUID decodes a "prefix_<base32>" string into a typed UUID.
func ParseUUID[P Prefixer](s string) (UUID[P], error) {
	suffix, err := splitTypeid[P](s, uuidSuffixLen)
	if err != nil {
		return UUID[P]{}, err
	}
	b, err := decodeBase32UUID(suffix)
	if err != nil {
		return UUID[P]{}, err
	}
	u := uuid.UUID(b)
	if u.Version() != 7 {
		return UUID[P]{}, ErrOnlyV7
	}
	return UUID[P]{val: u}, nil
}

func (id UUID[P]) appendText(dst []byte) []byte {
	var p P
	dst = growSlice(dst, len(p.Prefix())+1+uuidSuffixLen)
	return appendBase32UUID(appendID[P](dst), id.val)
}
func (id UUID[P]) String() string { return string(id.appendText(nil)) }
func (id UUID[P]) UUID() uuid.UUID              { return id.val }
func (id UUID[P]) IsZero() bool                 { return id.val == uuid.UUID{} }
func (id UUID[P]) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.appendText(nil), nil
}

func (id *UUID[P]) UnmarshalText(data []byte) error {
	parsed, err := ParseUUID[P](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id UUID[P]) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.val.String(), nil
}

// GetTime extracts the millisecond-precision creation timestamp from the UUIDv7.
func (id UUID[P]) GetTime() time.Time {
	return time.UnixMilli(uuidTimestamp(id.val))
}

func uuidTimestamp(u uuid.UUID) int64 {
	return int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 | int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
}

func setUUIDTimestamp(u *uuid.UUID, t time.Time) {
	ms := uint64(t.UnixMilli())
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)
}

// FloorUUID returns the lowest possible UUID[P] for timestamp t.
// All random and variant bits are minimized. Useful as the lower bound
// in range queries: WHERE id >= FloorUUID(since).
func FloorUUID[P Prefixer](t time.Time) UUID[P] {
	var u uuid.UUID
	setUUIDTimestamp(&u, t)
	u[6] = 0x70 // version 7
	u[8] = 0x80 // variant 10xxxxxx
	return UUID[P]{val: u}
}

// CeilUUID returns the highest possible UUID[P] for timestamp t.
// All random and variant bits are maximized. Useful as the upper bound
// in range queries: WHERE id <= CeilUUID(until).
func CeilUUID[P Prefixer](t time.Time) UUID[P] {
	var u uuid.UUID
	setUUIDTimestamp(&u, t)
	u[6] = 0x7f // version 7 + rand_a high nibble all 1s
	u[7] = 0xff // rand_a low byte all 1s
	u[8] = 0xbf // variant 10 + 6 bits all 1s
	for i := 9; i < 16; i++ {
		u[i] = 0xff
	}
	return UUID[P]{val: u}
}

func (id *UUID[P]) Scan(src any) (err error) {
	var u uuid.UUID
	switch v := src.(type) {
	case string:
		if u, err = uuid.Parse(v); err != nil {
			return err
		}
	case []byte:
		switch {
		case len(v) == 16:
			copy(u[:], v)
		default:
			if u, err = uuid.ParseBytes(v); err != nil {
				return err
			}
		}
	case [16]byte:
		u = uuid.UUID(v)
	default:
		return fmt.Errorf("typeid: cannot scan %T into UUID", src)
	}
	if u.Version() != 7 {
		return ErrOnlyV7
	}
	id.val = u
	return nil
}
