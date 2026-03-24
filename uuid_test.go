package typeid_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/go-chi/typeid"
)

func ExampleNewUUID() {
	id, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	s := id.String()

	prefix, suffix, _ := strings.Cut(s, "_")
	fmt.Println(prefix)
	fmt.Println(len(suffix))
	fmt.Println(int(id.UUID().Version()))
	// Output:
	// user
	// 26
	// 7
}

func ExampleParseUUID() {
	original, _ := typeid.NewUUID[userPrefix]()
	parsed, err := typeid.ParseUUID[userPrefix](original.String())
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(original == parsed)
	// Output:
	// true
}

func ExampleParseUUID_wrongPrefix() {
	_, err := typeid.ParseUUID[userPrefix]("team_01h455vb4pex5vsknk084sn02q")
	fmt.Println(err)
	// Output:
	// typeid: prefix mismatch: expected "user", got "team"
}

func ExampleUUIDFrom() {
	raw := uuid.Must(uuid.NewV7())
	id, err := typeid.UUIDFrom[userPrefix](raw)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(id.UUID() == raw)
	// Output:
	// true
}

func ExampleUUIDFrom_rejectsV4() {
	v4 := uuid.New()
	_, err := typeid.UUIDFrom[userPrefix](v4)
	fmt.Println(err)
	// Output:
	// typeid: only UUIDv7 is supported
}

func ExampleUUID_IsZero() {
	var id UserID
	fmt.Println(id.IsZero())
	id, _ = typeid.NewUUID[userPrefix]()
	fmt.Println(id.IsZero())
	// Output:
	// true
	// false
}

func ExampleUUID_json() {
	type User struct {
		ID   UserID `json:"id"`
		Name string `json:"name"`
	}

	id, _ := typeid.NewUUID[userPrefix]()
	original := User{ID: id, Name: "Alice"}
	data, _ := json.Marshal(original)

	var decoded User
	_ = json.Unmarshal(data, &decoded)
	fmt.Println(original.ID == decoded.ID)
	fmt.Println(strings.Contains(string(data), `"id":"user_`))
	// Output:
	// true
	// true
}

func ExampleUUID_Value() {
	id, _ := typeid.NewUUID[userPrefix]()
	val, _ := id.Value()
	s, ok := val.(string)
	fmt.Println(ok)
	_, err := uuid.Parse(s)
	fmt.Println(err == nil)
	// Output:
	// true
	// true
}

func ExampleUUID_Scan() {
	id, _ := typeid.NewUUID[userPrefix]()
	raw := id.UUID().String()

	var scanned UserID
	err := scanned.Scan(raw)
	fmt.Println(err == nil)
	fmt.Println(id == scanned)
	// Output:
	// true
	// true
}

func TestUUID_RejectZero(t *testing.T) {
	var zero UserID

	if _, err := zero.MarshalText(); err == nil {
		t.Error("MarshalText should reject zero")
	}
	if _, err := zero.Value(); err == nil {
		t.Error("Value should reject zero")
	}
}

func TestParseUUID_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no underscore", "abc"},
		{"suffix too short", "user_abc"},
		{"suffix too long", "user_01h455vb4pex5vsknk084sn02qq"},
		{"invalid base32 char", "user_01h455vb4pex5vsknk084sn0!q"},
		{"overflow first char", "user_81h455vb4pex5vsknk084sn02q"},
		{"wrong prefix", "org_01h455vb4pex5vsknk084sn02q"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := typeid.ParseUUID[userPrefix](tt.input); err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
		})
	}
}

func TestUUID_ScanRawBytes(t *testing.T) {
	id, _ := typeid.NewUUID[userPrefix]()
	raw := id.UUID()

	var scanned UserID
	if err := scanned.Scan(raw[:]); err != nil {
		t.Fatalf("Scan raw 16-byte slice: %v", err)
	}
	if scanned != id {
		t.Errorf("got %s, want %s", scanned, id)
	}
}

func TestUUID_ScanInvalid(t *testing.T) {
	var id UserID

	// wrong type
	if err := id.Scan(123); err == nil {
		t.Error("Scan should reject int")
	}
	if err := id.Scan(true); err == nil {
		t.Error("Scan should reject bool")
	}

	// non-v7 UUID (v4)
	v4 := uuid.New()
	if err := id.Scan(v4.String()); err == nil {
		t.Error("Scan should reject non-v7 UUID string")
	}
	if err := id.Scan(v4[:]); err == nil {
		t.Error("Scan should reject non-v7 UUID bytes")
	}
	if err := id.Scan([16]byte(v4)); err == nil {
		t.Error("Scan should reject non-v7 [16]byte")
	}

	// malformed string
	if err := id.Scan("not-a-uuid"); err == nil {
		t.Error("Scan should reject malformed string")
	}
}

func TestUUID_KnownVector(t *testing.T) {
	// UUIDv7: 01932c1c-e400-7360-8123-456789abcdef
	raw := uuid.Must(uuid.FromBytes([]byte{
		0x01, 0x93, 0x2c, 0x1c, 0xe4, 0x00,
		0x73, 0x60,
		0x81, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
	}))
	id, err := typeid.UUIDFrom[userPrefix](raw)
	if err != nil {
		t.Fatal(err)
	}

	const want = "user_01jcp1ss00edg828t5cy4tqkff"
	if got := id.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}

	parsed, err := typeid.ParseUUID[userPrefix](want)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.UUID() != raw {
		t.Errorf("roundtrip UUID mismatch: got %s, want %s", parsed.UUID(), raw)
	}
}

func BenchmarkUUID_String(b *testing.B) {
	id, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		_ = id.String()
	}
}

func BenchmarkUUID_MarshalText(b *testing.B) {
	id, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for b.Loop() {
		id.MarshalText() //nolint:errcheck
	}
}

func BenchmarkUUID_Parse(b *testing.B) {
	id, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		b.Fatal(err)
	}
	s := id.String()
	b.ResetTimer()
	for b.Loop() {
		typeid.ParseUUID[userPrefix](s) //nolint:errcheck
	}
}

func TestUUID_Sortable(t *testing.T) {
	a, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	b, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		t.Fatal(err)
	}
	if a.String() >= b.String() {
		t.Errorf("expected a < b (IDs must sort by time)\n  a = %s\n  b = %s", a, b)
	}
}
