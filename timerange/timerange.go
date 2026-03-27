package timerange

import (
	"database/sql/driver"
	"fmt"
	"time"

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

type uuidValue struct{ val uuid.UUID }

func (b uuidValue) UUID() uuid.UUID              { return b.val }
func (b uuidValue) IsZero() bool                 { return b.val == uuid.UUID{} }
func (b uuidValue) Value() (driver.Value, error) { return b.val.String(), nil }

type int64Value struct{ val int64 }

func (b int64Value) Int64() int64                 { return b.val }
func (b int64Value) IsZero() bool                 { return b.val == 0 }
func (b int64Value) Value() (driver.Value, error) { return b.val, nil }

// FloorUUID returns the lowest valid UUIDv7 for timestamp t.
// Any UUIDv7 generated at or after t will be >= FloorUUID(t).
func FloorUUID(t time.Time) UUID {
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
	return uuidValue{val: u}
}

// CeilUUID returns the highest valid UUIDv7 for timestamp t.
// Any UUIDv7 generated at or before t will be <= CeilUUID(t).
func CeilUUID(t time.Time) UUID {
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
	return uuidValue{val: u}
}

const randomBits = 15

// FloorInt64 returns the lowest valid int64 ID for timestamp t.
func FloorInt64(t time.Time) Int64 {
	ms := t.UnixMilli()
	return int64Value{val: ms << randomBits}
}

// CeilInt64 returns the highest valid int64 ID for timestamp t.
func CeilInt64(t time.Time) Int64 {
	ms := t.UnixMilli()
	return int64Value{val: ms<<randomBits | 0x7FFF}
}

// TimeRange holds optional floor/ceil bounds for time-based ID range queries.
// It satisfies squirrel.Sqlizer structurally via ToSql.
type TimeRange struct {
	column string
	floor  driver.Valuer
	ceil   driver.Valuer
}

// UUIDRange returns a TimeRange over UUID IDs for the given time window.
// Nil since/until means unbounded on that side.
func UUIDRange(column string, since, until *time.Time) TimeRange {
	r := TimeRange{column: column}
	if since != nil {
		r.floor = FloorUUID(*since)
	}
	if until != nil {
		r.ceil = CeilUUID(*until)
	}
	return r
}

// Int64Range returns a TimeRange over Int64 IDs for the given time window.
// Nil since/until means unbounded on that side.
func Int64Range(column string, since, until *time.Time) TimeRange {
	r := TimeRange{column: column}
	if since != nil {
		r.floor = FloorInt64(*since)
	}
	if until != nil {
		r.ceil = CeilInt64(*until)
	}
	return r
}

func (r TimeRange) Floor() (driver.Valuer, bool) { return r.floor, r.floor != nil }
func (r TimeRange) Ceil() (driver.Valuer, bool)  { return r.ceil, r.ceil != nil }

// ToSql satisfies squirrel.Sqlizer structurally (no import needed).
func (r TimeRange) ToSql() (string, []any, error) {
	switch {
	case r.floor != nil && r.ceil != nil:
		fv, err := r.floor.Value()
		if err != nil {
			return "", nil, fmt.Errorf("typeid: floor value: %w", err)
		}
		cv, err := r.ceil.Value()
		if err != nil {
			return "", nil, fmt.Errorf("typeid: ceil value: %w", err)
		}
		return r.column + " BETWEEN ? AND ?", []any{fv, cv}, nil
	case r.floor != nil:
		fv, err := r.floor.Value()
		if err != nil {
			return "", nil, fmt.Errorf("typeid: floor value: %w", err)
		}
		return r.column + " >= ?", []any{fv}, nil
	case r.ceil != nil:
		cv, err := r.ceil.Value()
		if err != nil {
			return "", nil, fmt.Errorf("typeid: ceil value: %w", err)
		}
		return r.column + " <= ?", []any{cv}, nil
	default:
		return "1=1", nil, nil
	}
}
