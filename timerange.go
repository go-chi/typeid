package typeid

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// TimeRange holds optional floor/ceil bounds for time-based ID range queries
// against a primary key column. It satisfies squirrel.Sqlizer structurally
// via [TimeRange.ToSql], so it can be passed directly to squirrel.Where
// without importing squirrel in this package.
//
// Construct via [UUIDRange] or [Int64Range].
type TimeRange struct {
	column string
	floor  driver.Valuer
	ceil   driver.Valuer
}

// UUIDRange builds a [TimeRange] that brackets column with [FloorUUID] / [CeilUUID].
// Nil since or until leaves that side unbounded.
func UUIDRange[P Prefixer](column string, since, until *time.Time) TimeRange {
	r := TimeRange{column: column}
	if since != nil {
		r.floor = floorUUID[P](*since)
	}
	if until != nil {
		r.ceil = ceilUUID[P](*until)
	}
	return r
}

// Int64Range builds a [TimeRange] that brackets column with [FloorInt64] / [CeilInt64].
// Nil since or until leaves that side unbounded.
func Int64Range[P Prefixer](column string, since, until *time.Time) TimeRange {
	r := TimeRange{column: column}
	if since != nil {
		r.floor = floorInt64[P](*since)
	}
	if until != nil {
		r.ceil = ceilInt64[P](*until)
	}
	return r
}

// Floor returns the lower-bound ID and true, or (nil, false) if unbounded.
func (r TimeRange) Floor() (driver.Valuer, bool) { return r.floor, r.floor != nil }

// Ceil returns the upper-bound ID and true, or (nil, false) if unbounded.
func (r TimeRange) Ceil() (driver.Valuer, bool) { return r.ceil, r.ceil != nil }

// ToSql emits a SQL predicate and bind args for the range.
// Returns "column BETWEEN ? AND ?", "column >= ?", "column <= ?",
// or "1=1" depending on which bounds are set.
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
