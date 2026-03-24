package typeid_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/typeid"
	"github.com/google/uuid"
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
	_, err := typeid.ParseUUID[userPrefix]("org_01h455vb4pex5vsknk084sn02q")
	fmt.Println(err)
	// Output:
	// typeid: prefix mismatch: expected "user", got "org"
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
