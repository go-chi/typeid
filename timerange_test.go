package typeid_test

import (
	"testing"
	"time"

	"github.com/go-chi/typeid"
	"github.com/google/uuid"
)

func TestFloorUUID(t *testing.T) {
	now := time.Now()
	floor := typeid.FloorUUID[userPrefix](now)
	u := floor.UUID()

	if u.Version() != 7 {
		t.Fatalf("version = %d, want 7", u.Version())
	}
	if u.Variant() != uuid.RFC4122 {
		t.Fatalf("variant = %d, want RFC4122", u.Variant())
	}

	// Any UUIDv7 generated after now must be >= floor.
	for range 100 {
		id, err := typeid.NewUUID[userPrefix]()
		if err != nil {
			t.Fatal(err)
		}
		if id.UUID().String() < u.String() {
			t.Fatalf("NewUUID %s < floor %s", id.UUID(), u)
		}
	}
}

func TestCeilUUID(t *testing.T) {
	now := time.Now()
	ceil := typeid.CeilUUID[userPrefix](now)
	floor := typeid.FloorUUID[userPrefix](now)

	u := ceil.UUID()
	if u.Version() != 7 {
		t.Fatalf("version = %d, want 7", u.Version())
	}
	if u.Variant() != uuid.RFC4122 {
		t.Fatalf("variant = %d, want RFC4122", u.Variant())
	}
	if ceil.UUID().String() < floor.UUID().String() {
		t.Fatalf("ceil %s < floor %s", ceil.UUID(), floor.UUID())
	}
}

func TestFloorCeilUUID_Bracket(t *testing.T) {
	id, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	u := id.UUID()
	ts := id.GetTime()

	floor := typeid.FloorUUID[userPrefix](ts)
	ceil := typeid.CeilUUID[userPrefix](ts)

	if floor.UUID().String() > u.String() {
		t.Fatalf("floor %s > id %s", floor.UUID(), u)
	}
	if ceil.UUID().String() < u.String() {
		t.Fatalf("ceil %s < id %s", ceil.UUID(), u)
	}
}

func TestFloorUUID_TimestampRoundTrip(t *testing.T) {
	ts := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	floor := typeid.FloorUUID[userPrefix](ts)
	if got := floor.GetTime(); got.UnixMilli() != ts.UnixMilli() {
		t.Fatalf("GetTime() = %v, want %v", got, ts)
	}
}

func TestFloorInt64(t *testing.T) {
	now := time.Now()
	floor := typeid.FloorInt64[orgPrefix](now)

	for range 100 {
		id, err := typeid.NewInt64[orgPrefix]()
		if err != nil {
			t.Fatal(err)
		}
		if id.Int64() < floor.Int64() {
			t.Fatalf("NewInt64 %d < floor %d", id.Int64(), floor.Int64())
		}
	}
}

func TestCeilInt64(t *testing.T) {
	now := time.Now()
	ceil := typeid.CeilInt64[orgPrefix](now)
	floor := typeid.FloorInt64[orgPrefix](now)

	if ceil.Int64() < floor.Int64() {
		t.Fatalf("ceil %d < floor %d", ceil.Int64(), floor.Int64())
	}
}

func TestFloorCeilInt64_Bracket(t *testing.T) {
	id, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	v := id.Int64()
	ts := id.GetTime()

	floor := typeid.FloorInt64[orgPrefix](ts)
	ceil := typeid.CeilInt64[orgPrefix](ts)

	if floor.Int64() > v {
		t.Fatalf("floor %d > id %d", floor.Int64(), v)
	}
	if ceil.Int64() < v {
		t.Fatalf("ceil %d < id %d", ceil.Int64(), v)
	}
}

func TestInt64_GetTime(t *testing.T) {
	ts := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	floor := typeid.FloorInt64[orgPrefix](ts)
	if got := floor.GetTime(); got.UnixMilli() != ts.UnixMilli() {
		t.Fatalf("GetTime() = %v, want %v", got, ts)
	}
}

func TestTimeRange_ToSql_BothSet(t *testing.T) {
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC)
	r := typeid.UUIDRange[userPrefix]("id", &since, &until)

	sql, args, err := r.ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if sql != "id BETWEEN ? AND ?" {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2", len(args))
	}
}

func TestTimeRange_ToSql_SinceOnly(t *testing.T) {
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r := typeid.UUIDRange[userPrefix]("id", &since, nil)

	sql, args, err := r.ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if sql != "id >= ?" {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
}

func TestTimeRange_ToSql_UntilOnly(t *testing.T) {
	until := time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC)
	r := typeid.UUIDRange[userPrefix]("id", nil, &until)

	sql, args, err := r.ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if sql != "id <= ?" {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 1 {
		t.Fatalf("args len = %d, want 1", len(args))
	}
}

func TestTimeRange_ToSql_Neither(t *testing.T) {
	r := typeid.UUIDRange[userPrefix]("id", nil, nil)

	sql, args, err := r.ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if sql != "1=1" {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 0 {
		t.Fatalf("args len = %d, want 0", len(args))
	}
}

func TestTimeRange_ToSql_Int64(t *testing.T) {
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC)
	r := typeid.Int64Range[orgPrefix]("id", &since, &until)

	sql, args, err := r.ToSql()
	if err != nil {
		t.Fatal(err)
	}
	if sql != "id BETWEEN ? AND ?" {
		t.Fatalf("sql = %q", sql)
	}
	if len(args) != 2 {
		t.Fatalf("args len = %d, want 2", len(args))
	}
}

func TestTimeRange_FloorCeil_Accessors(t *testing.T) {
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r := typeid.UUIDRange[userPrefix]("id", &since, nil)

	if _, ok := r.Floor(); !ok {
		t.Fatal("Floor() returned false")
	}
	if _, ok := r.Ceil(); ok {
		t.Fatal("Ceil() returned true for nil until")
	}
}
