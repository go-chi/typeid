package timerange

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ID is a time-based identifier value with boundary operations.
type ID[T int64 | uuid.UUID] struct{ val T }

// UUID is an ID holding a uuid.UUID value.
type UUID = ID[uuid.UUID]

// Int64 is an ID holding an int64 value.
type Int64 = ID[int64]

const randomBits = 15

func (id ID[T]) Get() T    { return id.val }
func (id ID[T]) IsZero() bool {
	switch v := any(id.val).(type) {
	case int64:
		return v == 0
	case uuid.UUID:
		return v == uuid.UUID{}
	}
	return false
}

func (id ID[T]) Value() (driver.Value, error) {
	switch v := any(id.val).(type) {
	case int64:
		return v, nil
	case uuid.UUID:
		return v.String(), nil
	}
	return nil, nil
}

func (id ID[T]) GetTime() time.Time {
	switch v := any(id.val).(type) {
	case int64:
		return time.UnixMilli(v >> randomBits)
	case uuid.UUID:
		return time.UnixMilli(uuidTimestamp(v))
	}
	return time.Time{}
}

func (id ID[T]) Floor() ID[T] {
	t := id.GetTime()
	switch any(id.val).(type) {
	case int64:
		return ID[T]{val: any(t.UnixMilli() << randomBits).(T)}
	case uuid.UUID:
		return ID[T]{val: any(floorUUIDBytes(t)).(T)}
	}
	return ID[T]{}
}

func (id ID[T]) Ceil() ID[T] {
	t := id.GetTime()
	switch any(id.val).(type) {
	case int64:
		return ID[T]{val: any(t.UnixMilli()<<randomBits | 0x7FFF).(T)}
	case uuid.UUID:
		return ID[T]{val: any(ceilUUIDBytes(t)).(T)}
	}
	return ID[T]{}
}

func uuidTimestamp(u uuid.UUID) int64 {
	return int64(u[0])<<40 | int64(u[1])<<32 | int64(u[2])<<24 | int64(u[3])<<16 | int64(u[4])<<8 | int64(u[5])
}

func setTimestamp(u *uuid.UUID, t time.Time) {
	ms := uint64(t.UnixMilli())
	u[0] = byte(ms >> 40)
	u[1] = byte(ms >> 32)
	u[2] = byte(ms >> 24)
	u[3] = byte(ms >> 16)
	u[4] = byte(ms >> 8)
	u[5] = byte(ms)
}

func floorUUIDBytes(t time.Time) uuid.UUID {
	var u uuid.UUID
	setTimestamp(&u, t)
	u[6] = 0x70 // version 7
	u[8] = 0x80 // variant 10xxxxxx
	return u
}

func ceilUUIDBytes(t time.Time) uuid.UUID {
	var u uuid.UUID
	setTimestamp(&u, t)
	u[6] = 0x7f // version 7 + rand_a high nibble all 1s
	u[7] = 0xff // rand_a low byte all 1s
	u[8] = 0xbf // variant 10 + 6 bits all 1s
	for i := 9; i < 16; i++ {
		u[i] = 0xff
	}
	return u
}

// FloorUUID returns the lowest valid UUIDv7 for timestamp t.
// Any UUIDv7 generated at or after t will be >= FloorUUID(t).
func FloorUUID(t time.Time) UUID { return UUID{val: floorUUIDBytes(t)} }

// CeilUUID returns the highest valid UUIDv7 for timestamp t.
// Any UUIDv7 generated at or before t will be <= CeilUUID(t).
func CeilUUID(t time.Time) UUID { return UUID{val: ceilUUIDBytes(t)} }

// FloorInt64 returns the lowest valid int64 ID for timestamp t.
func FloorInt64(t time.Time) Int64 { return Int64{val: t.UnixMilli() << randomBits} }

// CeilInt64 returns the highest valid int64 ID for timestamp t.
func CeilInt64(t time.Time) Int64 { return Int64{val: t.UnixMilli()<<randomBits | 0x7FFF} }

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
