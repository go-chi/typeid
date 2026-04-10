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

// Variable prefix test types.

type apiKeyMode uint8

const (
	apiKeyLive apiKeyMode = iota
	apiKeySandbox
)

func (p apiKeyMode) Prefix() string {
	switch p {
	case apiKeySandbox:
		return "api_key_sandbox"
	default:
		return "api_key"
	}
}

func (p *apiKeyMode) ParsePrefix(s string) bool {
	switch s {
	case "api_key":
		*p = apiKeyLive
		return true
	case "api_key_sandbox":
		*p = apiKeySandbox
		return true
	}
	return false
}

// Compile-time interface checks.
var (
	_ fmt.Stringer             = UserID{}
	_ fmt.Stringer             = OrgID{}
	_ fmt.Stringer             = typeid.AnyUUID[typeid.AnyPrefix]{}
	_ fmt.Stringer             = typeid.AnyInt64[typeid.AnyPrefix]{}
	_ encoding.TextMarshaler   = UserID{}
	_ encoding.TextMarshaler   = OrgID{}
	_ encoding.TextMarshaler   = typeid.AnyUUID[typeid.AnyPrefix]{}
	_ encoding.TextMarshaler   = typeid.AnyInt64[typeid.AnyPrefix]{}
	_ encoding.TextUnmarshaler = (*UserID)(nil)
	_ encoding.TextUnmarshaler = (*OrgID)(nil)
	_ encoding.TextUnmarshaler = (*typeid.AnyUUID[typeid.AnyPrefix])(nil)
	_ encoding.TextUnmarshaler = (*typeid.AnyInt64[typeid.AnyPrefix])(nil)
	_ driver.Valuer            = UserID{}
	_ driver.Valuer            = OrgID{}
	_ driver.Valuer            = typeid.AnyUUID[typeid.AnyPrefix]{}
	_ driver.Valuer            = typeid.AnyInt64[typeid.AnyPrefix]{}
	_ sql.Scanner              = (*UserID)(nil)
	_ sql.Scanner              = (*OrgID)(nil)
	_ sql.Scanner              = (*typeid.AnyUUID[typeid.AnyPrefix])(nil)
	_ sql.Scanner              = (*typeid.AnyInt64[typeid.AnyPrefix])(nil)
	_ typeid.Prefixer          = typeid.AnyPrefix("")
	_ typeid.VariablePrefixer  = (*typeid.AnyPrefix)(nil)
	_ typeid.Prefixer          = apiKeyMode(0)
	_ typeid.VariablePrefixer  = (*apiKeyMode)(nil)
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
