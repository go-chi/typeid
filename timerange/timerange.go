package timerange

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/go-chi/typeid"
	"github.com/google/uuid"
)

// UUID is satisfied by any typeid.UUID[P] regardless of prefix.
type UUID interface {
	UUID() uuid.UUID
	IsZero() bool
	Value() (driver.Value, error)
}

// Int64 is satisfied by any typeid.Int64[P] regardless of prefix.
type Int64 interface {
	Int64() int64
	IsZero() bool
	Value() (driver.Value, error)
}

// FloorUUID returns the lowest valid UUID[P] for timestamp t.
// Any UUIDv7 generated at or after t will be >= FloorUUID(t).
func FloorUUID[P typeid.Prefixer](t time.Time) typeid.UUID[P] {
	ms := uint64(t.UnixMilli())
	var u uuid.UUID
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)
	u[6] = 0x70 // version 7
	u[8] = 0x80 // variant 10xxxxxx
	id, _ := typeid.UUIDFrom[P](u)
	return id
}

// CeilUUID returns the highest valid UUID[P] for timestamp t.
// Any UUIDv7 generated at or before t will be <= CeilUUID(t).
func CeilUUID[P typeid.Prefixer](t time.Time) typeid.UUID[P] {
	ms := uint64(t.UnixMilli())
	var u uuid.UUID
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)
	u[6] = 0x7f // version 7 + rand_a high nibble all 1s
	u[7] = 0xff // rand_a low byte all 1s
	u[8] = 0xbf // variant 10 + 6 bits all 1s
	for i := 9; i < 16; i++ {
		u[i] = 0xff
	}
	id, _ := typeid.UUIDFrom[P](u)
	return id
}

const randomBits = 15

// FloorInt64 returns the lowest valid Int64[P] for timestamp t.
func FloorInt64[P typeid.Prefixer](t time.Time) typeid.Int64[P] {
	ms := t.UnixMilli()
	id, _ := typeid.Int64From[P](ms << randomBits)
	return id
}

// CeilInt64 returns the highest valid Int64[P] for timestamp t.
func CeilInt64[P typeid.Prefixer](t time.Time) typeid.Int64[P] {
	ms := t.UnixMilli()
	id, _ := typeid.Int64From[P](ms<<randomBits | 0x7FFF)
	return id
}

type valuer interface{ Value() (driver.Value, error) }

// TimeRange holds optional floor/ceil bounds for time-based ID range queries.
// It satisfies squirrel.Sqlizer structurally via ToSql.
type TimeRange[T valuer] struct {
	column string
	floor  *T
	ceil   *T
}

// UUIDRange returns a TimeRange for UUID[P] IDs over the given time window.
// Nil since/until means unbounded on that side.
func UUIDRange[P typeid.Prefixer](column string, since, until *time.Time) TimeRange[typeid.UUID[P]] {
	var r TimeRange[typeid.UUID[P]]
	r.column = column
	if since != nil {
		f := FloorUUID[P](*since)
		r.floor = &f
	}
	if until != nil {
		c := CeilUUID[P](*until)
		r.ceil = &c
	}
	return r
}

// Int64Range returns a TimeRange for Int64[P] IDs over the given time window.
// Nil since/until means unbounded on that side.
func Int64Range[P typeid.Prefixer](column string, since, until *time.Time) TimeRange[typeid.Int64[P]] {
	var r TimeRange[typeid.Int64[P]]
	r.column = column
	if since != nil {
		f := FloorInt64[P](*since)
		r.floor = &f
	}
	if until != nil {
		c := CeilInt64[P](*until)
		r.ceil = &c
	}
	return r
}

func (r TimeRange[T]) Floor() (T, bool) {
	if r.floor == nil {
		var zero T
		return zero, false
	}
	return *r.floor, true
}

func (r TimeRange[T]) Ceil() (T, bool) {
	if r.ceil == nil {
		var zero T
		return zero, false
	}
	return *r.ceil, true
}

// ToSql satisfies squirrel.Sqlizer structurally (no import needed).
func (r TimeRange[T]) ToSql() (string, []any, error) {
	switch {
	case r.floor != nil && r.ceil != nil:
		fv, err := (*r.floor).Value()
		if err != nil {
			return "", nil, fmt.Errorf("typeid: floor value: %w", err)
		}
		cv, err := (*r.ceil).Value()
		if err != nil {
			return "", nil, fmt.Errorf("typeid: ceil value: %w", err)
		}
		return r.column + " BETWEEN ? AND ?", []any{fv, cv}, nil
	case r.floor != nil:
		fv, err := (*r.floor).Value()
		if err != nil {
			return "", nil, fmt.Errorf("typeid: floor value: %w", err)
		}
		return r.column + " >= ?", []any{fv}, nil
	case r.ceil != nil:
		cv, err := (*r.ceil).Value()
		if err != nil {
			return "", nil, fmt.Errorf("typeid: ceil value: %w", err)
		}
		return r.column + " <= ?", []any{cv}, nil
	default:
		return "1=1", nil, nil
	}
}
