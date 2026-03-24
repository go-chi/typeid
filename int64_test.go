package typeid_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/typeid"
)

func ExampleNewInt64() {
	id, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	s := id.String()

	prefix, suffix, _ := strings.Cut(s, "_")
	fmt.Println(prefix)
	fmt.Println(len(suffix))
	fmt.Println(id.Int64() > 0)
	// Output:
	// org
	// 13
	// true
}

func ExampleParseInt64() {
	original, _ := typeid.NewInt64[orgPrefix]()
	parsed, err := typeid.ParseInt64[orgPrefix](original.String())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(original == parsed)
	// Output:
	// true
}

func ExampleParseInt64_wrongPrefix() {
	_, err := typeid.ParseInt64[orgPrefix]("foo_0h455vb4pex5v")
	fmt.Println(err)
	// Output:
	// typeid: prefix mismatch: expected "org", got "foo"
}

func ExampleInt64From() {
	id, _ := typeid.NewInt64[orgPrefix]()
	raw := id.Int64()
	reconstructed, err := typeid.Int64From[orgPrefix](raw)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(id == reconstructed)
	// Output:
	// true
}

func ExampleInt64From_rejectsNonPositive() {
	_, err := typeid.Int64From[orgPrefix](-1)
	fmt.Println(err)
	_, err = typeid.Int64From[orgPrefix](0)
	fmt.Println(err)
	// Output:
	// typeid: non-positive Int64
	// typeid: non-positive Int64
}

func ExampleInt64_IsZero() {
	var id OrgID
	fmt.Println(id.IsZero())
	id, _ = typeid.NewInt64[orgPrefix]()
	fmt.Println(id.IsZero())
	// Output:
	// true
	// false
}

func ExampleInt64_json() {
	type Org struct {
		ID   OrgID  `json:"id"`
		Name string `json:"name"`
	}

	id, _ := typeid.NewInt64[orgPrefix]()
	original := Org{ID: id, Name: "Polygon"}
	data, _ := json.Marshal(original)

	var decoded Org
	_ = json.Unmarshal(data, &decoded)
	fmt.Println(original.ID == decoded.ID)
	fmt.Println(strings.Contains(string(data), `"id":"org_`))
	// Output:
	// true
	// true
}

func ExampleInt64_Value() {
	id, _ := typeid.NewInt64[orgPrefix]()
	val, _ := id.Value()
	v, ok := val.(int64)
	fmt.Println(ok)
	fmt.Println(v > 0)
	// Output:
	// true
	// true
}

func ExampleInt64_Scan() {
	id, _ := typeid.NewInt64[orgPrefix]()
	raw := id.Int64()

	var scanned OrgID
	err := scanned.Scan(raw)
	fmt.Println(err == nil)
	fmt.Println(id == scanned)
	// Output:
	// true
	// true
}

func TestInt64_RejectZeroAndNegative(t *testing.T) {
	var zero OrgID

	if _, err := zero.MarshalText(); err == nil {
		t.Error("MarshalText should reject zero")
	}
	if _, err := zero.Value(); err == nil {
		t.Error("Value should reject zero")
	}

	var scanned OrgID
	if err := scanned.Scan(int64(0)); err == nil {
		t.Error("Scan should reject zero")
	}
	if err := scanned.Scan(int64(-1)); err == nil {
		t.Error("Scan should reject negative")
	}
	if err := scanned.Scan(int(-1)); err == nil {
		t.Error("Scan should reject negative int")
	}
}

func TestParseInt64_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no underscore", "abc"},
		{"suffix too short", "org_abc"},
		{"suffix too long", "org_0h455vb4pex5vv"},
		{"invalid base32 char", "org_0h455vb4pex!v"},
		{"overflow first char", "org_8h455vb4pex5v"},
		{"zero", "org_0000000000000"},
		{"wrong prefix", "user_0h455vb4pex5v"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := typeid.ParseInt64[orgPrefix](tt.input); err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
		})
	}
}

func TestInt64_ScanInvalid(t *testing.T) {
	var id OrgID

	if err := id.Scan("hello"); err == nil {
		t.Error("Scan should reject string")
	}
	if err := id.Scan(true); err == nil {
		t.Error("Scan should reject bool")
	}
	if err := id.Scan(3.14); err == nil {
		t.Error("Scan should reject float64")
	}
}

func TestInt64_KnownVector(t *testing.T) {
	// timestamp=1700000000000ms, random=12345
	raw := int64(1700000000000<<15) | 12345
	id, err := typeid.Int64From[orgPrefix](raw)
	if err != nil {
		t.Fatal(err)
	}

	const want = "org_01hf7yat00c1s"
	if got := id.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	parsed, err := typeid.ParseInt64[orgPrefix](want)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Int64() != raw {
		t.Errorf("roundtrip Int64 mismatch: got %d, want %d", parsed.Int64(), raw)
	}
}

func BenchmarkInt64_String(b *testing.B) {
	id, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		_ = id.String()
	}
}

func BenchmarkInt64_MarshalText(b *testing.B) {
	id, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		id.MarshalText() //nolint:errcheck
	}
}

func BenchmarkInt64_Parse(b *testing.B) {
	id, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		b.Fatal(err)
	}
	s := id.String()
	b.ResetTimer()
	for b.Loop() {
		typeid.ParseInt64[orgPrefix](s) //nolint:errcheck
	}
}

func TestInt64_Sortable(t *testing.T) {
	a, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	b, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		t.Fatal(err)
	}

	if a.Int64() >= b.Int64() {
		t.Errorf("expected a < b numerically\n  a = %d\n  b = %d", a.Int64(), b.Int64())
	}
	if a.String() >= b.String() {
		t.Errorf("expected a < b lexicographically\n  a = %s\n  b = %s", a, b)
	}
}
