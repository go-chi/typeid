package typeid_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/typeid"
)

func ExampleNewScopedInt64() {
	id, err := typeid.NewScopedInt64[statusScope](0)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	s := id.String()
	fmt.Println(strings.HasPrefix(s, "status_pending_"))
	fmt.Println(id.Scope())
	fmt.Println(id.Int64() > 0)
	// Output:
	// true
	// 0
	// true
}

func ExampleParseScopedInt64() {
	original, _ := typeid.NewScopedInt64[statusScope](1)
	parsed, err := typeid.ParseScopedInt64[statusScope](original.String())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(original == parsed)
	fmt.Println(parsed.Scope())
	// Output:
	// true
	// 1
}

func ExampleParseScopedInt64_wrongPrefix() {
	_, err := typeid.ParseScopedInt64[statusScope]("status_unknown_0h455vb4pex5v")
	fmt.Println(err)
	// Output:
	// typeid: prefix mismatch: expected one of [status_pending status_active], got "status_unknown"
}

func TestScopedInt64_KnownVector(t *testing.T) {
	raw := int64(1700000000000<<15) | 12345

	for scope, wantPrefix := range []string{"status_pending", "status_active"} {
		id, err := typeid.ScopedInt64From[statusScope](raw, uint8(scope))
		if err != nil {
			t.Fatal(err)
		}
		want := wantPrefix + "_01hf7yat00c1s"
		if got := id.String(); got != want {
			t.Errorf("scope %d: String() = %q, want %q", scope, got, want)
		}
		parsed, err := typeid.ParseScopedInt64[statusScope](want)
		if err != nil {
			t.Fatal(err)
		}
		if parsed.Int64() != raw {
			t.Errorf("scope %d: roundtrip Int64 mismatch", scope)
		}
		if parsed.Scope() != uint8(scope) {
			t.Errorf("scope %d: Scope() = %d", scope, parsed.Scope())
		}
	}
}

func TestScopedInt64_JSON(t *testing.T) {
	type Status struct {
		ID StatusID `json:"id"`
	}

	id, _ := typeid.NewScopedInt64[statusScope](1)
	original := Status{ID: id}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id":"status_active_`) {
		t.Errorf("expected status_active_ prefix in JSON: %s", data)
	}

	var decoded Status
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if original.ID != decoded.ID {
		t.Errorf("JSON roundtrip mismatch")
	}
}

func TestParseScopedInt64_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no underscore", "abc"},
		{"suffix too short", "status_pending_abc"},
		{"invalid base32 char", "status_pending_0h455vb4pex!v"},
		{"overflow first char", "status_pending_8h455vb4pex5v"},
		{"unknown prefix", "status_unknown_0h455vb4pex5v"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := typeid.ParseScopedInt64[statusScope](tt.input); err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
		})
	}
}

func TestScopedInt64_RejectsNegative(t *testing.T) {
	if _, err := typeid.ScopedInt64From[statusScope](-1, 0); err == nil {
		t.Error("ScopedInt64From should reject negative")
	}
}

func TestScopedInt64_ValueScan(t *testing.T) {
	id, _ := typeid.NewScopedInt64[statusScope](1)
	val, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := val.(string)
	if !ok {
		t.Fatal("Value() should return string")
	}
	if !strings.HasPrefix(s, "status_active_") {
		t.Errorf("Value() = %q, expected status_active_ prefix", s)
	}

	var scanned StatusID
	if err := scanned.Scan(s); err != nil {
		t.Fatal(err)
	}
	if scanned != id {
		t.Errorf("Scan roundtrip: got %v, want %v", scanned, id)
	}
}

func TestScopedInt64_ScanInvalid(t *testing.T) {
	var id StatusID
	if err := id.Scan(123); err == nil {
		t.Error("Scan should reject int")
	}
	if err := id.Scan(true); err == nil {
		t.Error("Scan should reject bool")
	}
}

func TestScopedInt64_Sortable(t *testing.T) {
	a, _ := typeid.NewScopedInt64[statusScope](0)
	time.Sleep(2 * time.Millisecond)
	b, _ := typeid.NewScopedInt64[statusScope](0)
	if a.String() >= b.String() {
		t.Errorf("expected a < b\n  a = %s\n  b = %s", a, b)
	}
}

func TestScopedInt64_DifferentScopesNotEqual(t *testing.T) {
	raw := int64(1700000000000<<15) | 12345
	a, _ := typeid.ScopedInt64From[statusScope](raw, 0)
	b, _ := typeid.ScopedInt64From[statusScope](raw, 1)
	if a == b {
		t.Error("same Int64 with different scopes should not be equal")
	}
}
