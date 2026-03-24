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
	_, err := typeid.ParseInt64[orgPrefix]("user_0h455vb4pex5v")
	fmt.Println(err)
	// Output:
	// typeid: prefix mismatch: expected "org", got "user"
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

func ExampleInt64From_rejectsNegative() {
	_, err := typeid.Int64From[orgPrefix](-1)
	fmt.Println(err)
	// Output:
	// typeid: int64 must be non-negative
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
