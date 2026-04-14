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

func TestAnyUUID_json(t *testing.T) {
	type Request struct {
		ID typeid.AnyUUID `json:"id"`
	}

	suffix := "01jcp1ss00edg828t5cy4tqkff"
	inputs := []string{
		`{"id":"whatever_prefix_` + suffix + `"}`,
		`{"id":"other_prefix_` + suffix + `"}`,
		`{"id":"` + suffix + `"}`,
	}
	for _, raw := range inputs {
		var req Request
		if err := json.Unmarshal([]byte(raw), &req); err != nil {
			t.Fatalf("Unmarshal %s: %v", raw, err)
		}
		if req.ID.UUID().String() == "" || req.ID.UUID().Version() != 7 {
			t.Fatalf("expected v7 UUID, got %v", req.ID.UUID())
		}
	}
}

// ExampleAnyUUID_switchToTypedUUID shows narrowing [AnyUUID] to [UUID] after inspecting [AnyUUID.Prefix].
// Use [UUIDFrom] when the prefix matches; it keeps the same UUID bytes under the typed wrapper.
func ExampleAnyUUID_switchToTypedUUID() {
	const payload = `{"id":"user_01jcp1ss00edg828t5cy4tqkff"}`
	type Request struct {
		ID typeid.AnyUUID `json:"id"`
	}
	var req Request
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		fmt.Println("unmarshal:", err)
		return
	}

	var userID UserID
	var err error
	switch req.ID.Prefix() {
	case "user":
		userID, err = typeid.UUIDFrom[userPrefix](req.ID.UUID())
	default:
		fmt.Println("unknown prefix")
		return
	}
	if err != nil {
		fmt.Println("narrow:", err)
		return
	}
	fmt.Println(userID.String())
	// Output:
	// user_01jcp1ss00edg828t5cy4tqkff
}

func TestAnyUUID_narrowToUserPrefix(t *testing.T) {
	suffix := "01jcp1ss00edg828t5cy4tqkff"
	anyID, err := typeid.ParseAnyUUID("user_" + suffix)
	if err != nil {
		t.Fatal(err)
	}
	var userID UserID
	switch anyID.Prefix() {
	case "user":
		userID, err = typeid.UUIDFrom[userPrefix](anyID.UUID())
	default:
		t.Fatalf("unexpected prefix %q", anyID.Prefix())
	}
	if err != nil {
		t.Fatal(err)
	}
	if userID.UUID() != anyID.UUID() {
		t.Errorf("UUID mismatch")
	}
	if got := userID.String(); got != "user_"+suffix {
		t.Errorf("String() = %q", got)
	}
}

func TestAnyUUID_prefixAndSetPrefix(t *testing.T) {
	suffix := "01jcp1ss00edg828t5cy4tqkff"
	id, err := typeid.ParseAnyUUID("foo_" + suffix)
	if err != nil {
		t.Fatal(err)
	}
	if got := id.Prefix(); got != "foo" {
		t.Fatalf("Prefix() = %q, want foo", got)
	}

	id.SetPrefix("bar")
	if got := id.Prefix(); got != "bar" {
		t.Fatalf("after SetPrefix, Prefix() = %q, want bar", got)
	}
	wantText := "bar_" + suffix
	if got, _ := id.MarshalText(); string(got) != wantText {
		t.Fatalf("MarshalText = %q, want %q", got, wantText)
	}
}

func TestNewAnyUUID(t *testing.T) {
	id, err := typeid.NewAnyUUID("user")
	if err != nil {
		t.Fatal(err)
	}
	if id.Prefix() != "user" {
		t.Errorf("prefix = %q, want %q", id.Prefix(), "user")
	}
	if id.IsZero() {
		t.Error("new AnyUUID should not be zero")
	}
	if id.UUID().Version() != 7 {
		t.Error("expected UUIDv7")
	}
}

func TestAnyUUIDFrom(t *testing.T) {
	raw := uuid.Must(uuid.NewV7())
	id, err := typeid.AnyUUIDFrom("team", raw)
	if err != nil {
		t.Fatal(err)
	}
	if id.UUID() != raw {
		t.Errorf("UUID mismatch: got %s, want %s", id.UUID(), raw)
	}
	if id.Prefix() != "team" {
		t.Errorf("prefix = %q, want %q", id.Prefix(), "team")
	}
}

func TestAnyUUIDFrom_RejectsV4(t *testing.T) {
	v4 := uuid.New()
	_, err := typeid.AnyUUIDFrom("user", v4)
	if err == nil {
		t.Error("expected error for non-v7 UUID")
	}
}

func TestAnyUUID_String(t *testing.T) {
	id, _ := typeid.NewAnyUUID("user")
	s := id.String()
	if !strings.HasPrefix(s, "user_") {
		t.Errorf("expected user_ prefix, got %q", s)
	}
	if len(s) != len("user")+1+26 {
		t.Errorf("unexpected length %d", len(s))
	}
}

func TestAnyUUID_SetPrefix(t *testing.T) {
	id, _ := typeid.NewAnyUUID("apiKey")
	if !strings.HasPrefix(id.String(), "apiKey_") {
		t.Fatalf("expected apiKey_ prefix, got %q", id.String())
	}

	id.SetPrefix("apiKeySandbox")
	if !strings.HasPrefix(id.String(), "apiKeySandbox_") {
		t.Errorf("expected apiKeySandbox_ prefix after SetPrefix, got %q", id.String())
	}

	// Underlying UUID unchanged
	id2, _ := typeid.NewAnyUUID("apiKey")
	raw := id2.UUID()
	id2.SetPrefix("other")
	if id2.UUID() != raw {
		t.Error("SetPrefix should not change the UUID")
	}
}

func TestAnyUUID_MarshalText_RejectsZero(t *testing.T) {
	var id typeid.AnyUUID
	_, err := id.MarshalText()
	if err == nil {
		t.Error("MarshalText should reject zero")
	}
}

func TestAnyUUID_UnmarshalText(t *testing.T) {
	original, _ := typeid.NewAnyUUID("proj")
	data, _ := original.MarshalText()

	var parsed typeid.AnyUUID
	if err := parsed.UnmarshalText(data); err != nil {
		t.Fatal(err)
	}
	if parsed.UUID() != original.UUID() {
		t.Error("UUID mismatch after unmarshal")
	}
	if parsed.Prefix() != "proj" {
		t.Errorf("prefix = %q, want %q", parsed.Prefix(), "proj")
	}
}

func TestAnyUUID_UnmarshalText_MultiWordPrefix(t *testing.T) {
	original, _ := typeid.NewAnyUUID("apiKeySandbox")
	data, _ := original.MarshalText()

	var parsed typeid.AnyUUID
	if err := parsed.UnmarshalText(data); err != nil {
		t.Fatal(err)
	}
	if parsed.Prefix() != "apiKeySandbox" {
		t.Errorf("prefix = %q, want %q", parsed.Prefix(), "apiKeySandbox")
	}
	if parsed.UUID() != original.UUID() {
		t.Error("UUID mismatch")
	}
}

func TestAnyUUID_Value(t *testing.T) {
	id, _ := typeid.NewAnyUUID("key")
	val, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := val.(string)
	if !ok {
		t.Fatal("Value should return string")
	}
	if _, err := uuid.Parse(s); err != nil {
		t.Errorf("Value should return valid UUID string: %v", err)
	}
}

func TestAnyUUID_Value_RejectsZero(t *testing.T) {
	var id typeid.AnyUUID
	_, err := id.Value()
	if err == nil {
		t.Error("Value should reject zero")
	}
}

func TestAnyUUID_Scan(t *testing.T) {
	original, _ := typeid.NewAnyUUID("user")
	raw := original.UUID().String()

	var scanned typeid.AnyUUID
	if err := scanned.Scan(raw); err != nil {
		t.Fatal(err)
	}
	if scanned.UUID() != original.UUID() {
		t.Error("UUID mismatch after scan")
	}
}

func TestAnyUUID_ScanRawBytes(t *testing.T) {
	original, _ := typeid.NewAnyUUID("user")
	raw := original.UUID()

	var scanned typeid.AnyUUID
	if err := scanned.Scan(raw[:]); err != nil {
		t.Fatal(err)
	}
	if scanned.UUID() != original.UUID() {
		t.Error("UUID mismatch after scan from bytes")
	}
}

func TestAnyUUID_ScanInvalid(t *testing.T) {
	var id typeid.AnyUUID
	if err := id.Scan(123); err == nil {
		t.Error("Scan should reject int")
	}
	v4 := uuid.New()
	if err := id.Scan(v4.String()); err == nil {
		t.Error("Scan should reject non-v7")
	}
}

func TestAnyUUID_DBRoundTrip(t *testing.T) {
	id, _ := typeid.NewAnyUUID("apiKey")

	val, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}

	var scanned typeid.AnyUUID
	if err := scanned.Scan(val); err != nil {
		t.Fatal(err)
	}

	scanned.SetPrefix("apiKeySandbox")

	if scanned.UUID() != id.UUID() {
		t.Error("UUID mismatch in round-trip")
	}
	if !strings.HasPrefix(scanned.String(), "apiKeySandbox_") {
		t.Errorf("expected apiKeySandbox_ prefix, got %q", scanned.String())
	}
}

func TestUUID_Any(t *testing.T) {
	typed, _ := typeid.NewUUID[userPrefix]()
	any := typed.Any()

	if any.UUID() != typed.UUID() {
		t.Error("UUID mismatch")
	}
	if any.Prefix() != "user" {
		t.Errorf("prefix = %q, want %q", any.Prefix(), "user")
	}
	if any.String() != typed.String() {
		t.Errorf("String mismatch: any=%q, typed=%q", any.String(), typed.String())
	}

	any.SetPrefix("admin")
	if any.UUID() != typed.UUID() {
		t.Error("UUID changed after SetPrefix")
	}
	if !strings.HasPrefix(any.String(), "admin_") {
		t.Errorf("expected admin_ prefix, got %q", any.String())
	}
}

func TestAnyUUID_GetTime(t *testing.T) {
	before := time.Now()
	id, _ := typeid.NewAnyUUID("user")
	after := time.Now()

	got := id.GetTime()
	if got.Before(before.Truncate(time.Millisecond)) {
		t.Errorf("GetTime %v before creation time %v", got, before)
	}
	if got.After(after.Add(time.Millisecond)) {
		t.Errorf("GetTime %v after creation time %v", got, after)
	}
}

func TestAnyUUID_JSON(t *testing.T) {
	type Record struct {
		ID   typeid.AnyUUID `json:"id"`
		Name string         `json:"name"`
	}

	id, _ := typeid.NewAnyUUID("apiKey")
	original := Record{ID: id, Name: "test"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id":"apiKey_`) {
		t.Errorf("JSON should contain apiKey_ prefix: %s", data)
	}

	var decoded Record
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID.UUID() != original.ID.UUID() {
		t.Error("UUID mismatch after JSON round-trip")
	}
	if decoded.ID.Prefix() != "apiKey" {
		t.Errorf("prefix = %q, want %q", decoded.ID.Prefix(), "apiKey")
	}
}

func BenchmarkAnyUUID_String(b *testing.B) {
	id, _ := typeid.NewAnyUUID("apiKeySandbox")
	b.ResetTimer()
	for b.Loop() {
		_ = id.String()
	}
}

func BenchmarkAnyUUID_Parse(b *testing.B) {
	id, _ := typeid.NewAnyUUID("apiKeySandbox")
	s := id.String()
	b.ResetTimer()
	for b.Loop() {
		typeid.ParseAnyUUID(s) //nolint:errcheck
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
