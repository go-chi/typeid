package typeid

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// TimeRange holds optional floor/ceil bounds for time-based ID range queries.
// It satisfies squirrel.Sqlizer structurally via ToSql.
type TimeRange struct {
	column string
	floor  driver.Valuer
	ceil   driver.Valuer
}

// UUIDRange returns a TimeRange over UUID[P] IDs for the given time window.
// Nil since/until means unbounded on that side.
func UUIDRange[P Prefixer](column string, since, until *time.Time) TimeRange {
	r := TimeRange{column: column}
	if since != nil {
		r.floor = FloorUUID[P](*since)
	}
	if until != nil {
		r.ceil = CeilUUID[P](*until)
	}
	return r
}

// Int64Range returns a TimeRange over Int64[P] IDs for the given time window.
// Nil since/until means unbounded on that side.
func Int64Range[P Prefixer](column string, since, until *time.Time) TimeRange {
	r := TimeRange{column: column}
	if since != nil {
		r.floor = FloorInt64[P](*since)
	}
	if until != nil {
		r.ceil = CeilInt64[P](*until)
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
