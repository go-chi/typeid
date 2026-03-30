package typeid_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"

	"github.com/go-chi/typeid"
)

// Prefix definitions — in practice these live next to each domain entity.

type userPrefix struct{}

func (userPrefix) Prefix() string { return "user" }

type orgPrefix struct{}

func (orgPrefix) Prefix() string { return "org" }

// Type aliases give readable names.
type (
	UserID = typeid.UUID[userPrefix]
	OrgID  = typeid.Int64[orgPrefix]
)

// Compile-time interface checks.
var (
	_ fmt.Stringer             = UserID{}
	_ fmt.Stringer             = OrgID{}
	_ fmt.Stringer             = typeid.AnyUUID{}
	_ fmt.Stringer             = typeid.AnyInt64{}
	_ encoding.TextMarshaler   = UserID{}
	_ encoding.TextMarshaler   = OrgID{}
	_ encoding.TextMarshaler   = typeid.AnyUUID{}
	_ encoding.TextMarshaler   = typeid.AnyInt64{}
	_ encoding.TextUnmarshaler = (*UserID)(nil)
	_ encoding.TextUnmarshaler = (*OrgID)(nil)
	_ encoding.TextUnmarshaler = (*typeid.AnyUUID)(nil)
	_ encoding.TextUnmarshaler = (*typeid.AnyInt64)(nil)
	_ driver.Valuer            = UserID{}
	_ driver.Valuer            = OrgID{}
	_ driver.Valuer            = typeid.AnyUUID{}
	_ driver.Valuer            = typeid.AnyInt64{}
	_ sql.Scanner              = (*UserID)(nil)
	_ sql.Scanner              = (*OrgID)(nil)
	_ sql.Scanner              = (*typeid.AnyUUID)(nil)
	_ sql.Scanner              = (*typeid.AnyInt64)(nil)
)

func Example() {
	orgID, err := typeid.NewInt64[orgPrefix]()
	if err != nil {
		panic(err)
	}

	userID, err := typeid.NewUUID[userPrefix]()
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.HasPrefix(orgID.String(), "org_"))
	fmt.Println(strings.HasPrefix(userID.String(), "user_"))
	// Output:
	// true
	// true
}
