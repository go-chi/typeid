package typeid_test

import (
	"testing"
	"time"

	"github.com/go-chi/typeid"
)

func TestUUID_GetTime(t *testing.T) {
	id, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	got := id.GetTime()
	if got.IsZero() {
		t.Fatal("GetTime() returned zero")
	}
	if time.Since(got) > time.Second {
		t.Fatalf("GetTime() = %v, too far from now", got)
	}
}

func TestInt64_GetTime(t *testing.T) {
	id, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	got := id.GetTime()
	if got.IsZero() {
		t.Fatal("GetTime() returned zero")
	}
	if time.Since(got) > time.Second {
		t.Fatalf("GetTime() = %v, too far from now", got)
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
