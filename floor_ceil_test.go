package typeid

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

type testUserPrefix struct{}

func (testUserPrefix) Prefix() string { return "user" }

type testOrgPrefix struct{}

func (testOrgPrefix) Prefix() string { return "org" }

func TestFloorUUID(t *testing.T) {
	now := time.Now()
	floor := floorUUID[testUserPrefix](now)
	u := floor.UUID()

	if u.Version() != 7 {
		t.Fatalf("version = %d, want 7", u.Version())
	}
	if u.Variant() != uuid.RFC4122 {
		t.Fatalf("variant = %d, want RFC4122", u.Variant())
	}

	// Any UUIDv7 generated after now must be >= floor.
	for range 100 {
		id, err := NewUUID[testUserPrefix]()
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
	ceil := ceilUUID[testUserPrefix](now)
	floor := floorUUID[testUserPrefix](now)

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
	id, err := NewUUID[testUserPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	u := id.UUID()
	ts := id.GetTime()

	floor := floorUUID[testUserPrefix](ts)
	ceil := ceilUUID[testUserPrefix](ts)

	if floor.UUID().String() > u.String() {
		t.Fatalf("floor %s > id %s", floor.UUID(), u)
	}
	if ceil.UUID().String() < u.String() {
		t.Fatalf("ceil %s < id %s", ceil.UUID(), u)
	}
}

func TestFloorUUID_TimestampRoundTrip(t *testing.T) {
	ts := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	floor := floorUUID[testUserPrefix](ts)
	if got := floor.GetTime(); got.UnixMilli() != ts.UnixMilli() {
		t.Fatalf("GetTime() = %v, want %v", got, ts)
	}
}

func TestFloorInt64(t *testing.T) {
	now := time.Now()
	floor := floorInt64[testOrgPrefix](now)

	for range 100 {
		id, err := NewInt64[testOrgPrefix]()
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
	ceil := ceilInt64[testOrgPrefix](now)
	floor := floorInt64[testOrgPrefix](now)

	if ceil.Int64() < floor.Int64() {
		t.Fatalf("ceil %d < floor %d", ceil.Int64(), floor.Int64())
	}
}

func TestFloorCeilInt64_Bracket(t *testing.T) {
	id, err := NewInt64[testOrgPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	v := id.Int64()
	ts := id.GetTime()

	floor := floorInt64[testOrgPrefix](ts)
	ceil := ceilInt64[testOrgPrefix](ts)

	if floor.Int64() > v {
		t.Fatalf("floor %d > id %d", floor.Int64(), v)
	}
	if ceil.Int64() < v {
		t.Fatalf("ceil %d < id %d", ceil.Int64(), v)
	}
}

func TestInt64_GetTime(t *testing.T) {
	ts := time.Date(2026, 3, 27, 12, 0, 0, 0, time.UTC)
	floor := floorInt64[testOrgPrefix](ts)
	if got := floor.GetTime(); got.UnixMilli() != ts.UnixMilli() {
		t.Fatalf("GetTime() = %v, want %v", got, ts)
	}
}
