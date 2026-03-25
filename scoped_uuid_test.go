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

func ExampleNewScopedUUID() {
	id, err := typeid.NewScopedUUID[apiKeyScope](0) // dev
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	s := id.String()
	fmt.Println(strings.HasPrefix(s, "sk_dev_"))
	fmt.Println(id.Scope())
	// Output:
	// true
	// 0
}

func ExampleNewScopedUUID_withScope() {
	id, err := typeid.NewScopedUUID[apiKeyScope](2) // prod
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(strings.HasPrefix(id.String(), "sk_prod_"))
	fmt.Println(id.Scope())
	// Output:
	// true
	// 2
}

func ExampleParseScopedUUID() {
	original, _ := typeid.NewScopedUUID[apiKeyScope](1)
	parsed, err := typeid.ParseScopedUUID[apiKeyScope](original.String())
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

func ExampleParseScopedUUID_detectsScope() {
	id, _ := typeid.NewScopedUUID[apiKeyScope](0)
	parsed, _ := typeid.ParseScopedUUID[apiKeyScope](id.String())
	fmt.Println(parsed.Scope()) // dev

	id, _ = typeid.NewScopedUUID[apiKeyScope](2)
	parsed, _ = typeid.ParseScopedUUID[apiKeyScope](id.String())
	fmt.Println(parsed.Scope()) // prod
	// Output:
	// 0
	// 2
}

func ExampleParseScopedUUID_wrongPrefix() {
	_, err := typeid.ParseScopedUUID[apiKeyScope]("sk_unknown_01h455vb4pex5vsknk084sn02q")
	fmt.Println(err)
	// Output:
	// typeid: prefix mismatch: expected one of [sk_dev sk_stg sk_prod], got "sk_unknown"
}

func TestScopedUUID_KnownVector(t *testing.T) {
	raw := uuid.Must(uuid.FromBytes([]byte{
		0x01, 0x93, 0x2c, 0x1c, 0xe4, 0x00,
		0x73, 0x60,
		0x81, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef,
	}))

	for scope, wantPrefix := range []string{"sk_dev", "sk_stg", "sk_prod"} {
		id, err := typeid.ScopedUUIDFrom[apiKeyScope](raw, uint8(scope))
		if err != nil {
			t.Fatal(err)
		}
		want := wantPrefix + "_01jcp1ss00edg828t5cy4tqkff"
		if got := id.String(); got != want {
			t.Errorf("scope %d: String() = %q, want %q", scope, got, want)
		}
		parsed, err := typeid.ParseScopedUUID[apiKeyScope](want)
		if err != nil {
			t.Fatal(err)
		}
		if parsed.UUID() != raw {
			t.Errorf("scope %d: roundtrip UUID mismatch", scope)
		}
		if parsed.Scope() != uint8(scope) {
			t.Errorf("scope %d: Scope() = %d", scope, parsed.Scope())
		}
	}
}

func TestScopedUUID_JSON(t *testing.T) {
	type Key struct {
		ID APIKeyID `json:"id"`
	}

	id, _ := typeid.NewScopedUUID[apiKeyScope](1)
	original := Key{ID: id}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id":"sk_stg_`) {
		t.Errorf("expected sk_stg_ prefix in JSON: %s", data)
	}

	var decoded Key
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if original.ID != decoded.ID {
		t.Errorf("JSON roundtrip mismatch: %v != %v", original.ID, decoded.ID)
	}
}

func TestParseScopedUUID_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no underscore", "abc"},
		{"suffix too short", "sk_dev_abc"},
		{"invalid base32 char", "sk_dev_01h455vb4pex5vsknk084sn0!q"},
		{"overflow first char", "sk_dev_81h455vb4pex5vsknk084sn02q"},
		{"unknown prefix", "sk_unknown_01h455vb4pex5vsknk084sn02q"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := typeid.ParseScopedUUID[apiKeyScope](tt.input); err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
		})
	}
}

func TestScopedUUID_RejectsV4(t *testing.T) {
	v4 := uuid.New()
	if _, err := typeid.ScopedUUIDFrom[apiKeyScope](v4, 0); err == nil {
		t.Error("ScopedUUIDFrom should reject non-v7")
	}
}

func TestScopedUUID_ValueScan(t *testing.T) {
	id, _ := typeid.NewScopedUUID[apiKeyScope](1)
	val, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := val.(string)
	if !ok {
		t.Fatal("Value() should return string")
	}
	if !strings.HasPrefix(s, "sk_stg_") {
		t.Errorf("Value() = %q, expected sk_stg_ prefix", s)
	}

	var scanned APIKeyID
	if err := scanned.Scan(s); err != nil {
		t.Fatal(err)
	}
	if scanned != id {
		t.Errorf("Scan roundtrip: got %v, want %v", scanned, id)
	}
}

func TestScopedUUID_ScanBytes(t *testing.T) {
	id, _ := typeid.NewScopedUUID[apiKeyScope](2)
	s := id.String()

	var scanned APIKeyID
	if err := scanned.Scan([]byte(s)); err != nil {
		t.Fatal(err)
	}
	if scanned != id {
		t.Errorf("Scan []byte roundtrip: got %v, want %v", scanned, id)
	}
}

func TestScopedUUID_ScanInvalid(t *testing.T) {
	var id APIKeyID
	if err := id.Scan(123); err == nil {
		t.Error("Scan should reject int")
	}
	if err := id.Scan(true); err == nil {
		t.Error("Scan should reject bool")
	}
}

func TestScopedUUID_Sortable(t *testing.T) {
	a, _ := typeid.NewScopedUUID[apiKeyScope](0)
	time.Sleep(2 * time.Millisecond)
	b, _ := typeid.NewScopedUUID[apiKeyScope](0)
	if a.String() >= b.String() {
		t.Errorf("expected a < b (IDs must sort by time)\n  a = %s\n  b = %s", a, b)
	}
}

func TestScopedUUID_DifferentScopesNotEqual(t *testing.T) {
	raw := uuid.Must(uuid.NewV7())
	a, _ := typeid.ScopedUUIDFrom[apiKeyScope](raw, 0)
	b, _ := typeid.ScopedUUIDFrom[apiKeyScope](raw, 1)
	if a == b {
		t.Error("same UUID with different scopes should not be equal")
	}
}
