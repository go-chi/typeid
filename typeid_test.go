package typeid_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"log"
	"strings"

	"github.com/go-chi/typeid"
)

// Prefix definitions — in practice these live next to each domain entity.

type userPrefix struct{}

func (userPrefix) Prefix() string { return "user" }

type orgPrefix struct{}

func (orgPrefix) Prefix() string { return "org" }

// Scoped prefix definitions — IDs with multiple prefix variants.

type apiKeyScope struct{}

func (apiKeyScope) Prefixes() []string { return []string{"sk_dev", "sk_stg", "sk_prod"} }

type statusScope struct{}

func (statusScope) Prefixes() []string { return []string{"status_pending", "status_active"} }

// Type aliases give readable names.
type (
	UserID   = typeid.UUID[userPrefix]
	OrgID    = typeid.Int64[orgPrefix]
	APIKeyID = typeid.ScopedUUID[apiKeyScope]
	StatusID = typeid.ScopedInt64[statusScope]
)

// Compile-time interface checks.
var (
	_ fmt.Stringer             = UserID{}
	_ fmt.Stringer             = OrgID{}
	_ fmt.Stringer             = APIKeyID{}
	_ fmt.Stringer             = StatusID{}
	_ encoding.TextMarshaler   = UserID{}
	_ encoding.TextMarshaler   = OrgID{}
	_ encoding.TextMarshaler   = APIKeyID{}
	_ encoding.TextMarshaler   = StatusID{}
	_ encoding.TextUnmarshaler = (*UserID)(nil)
	_ encoding.TextUnmarshaler = (*OrgID)(nil)
	_ encoding.TextUnmarshaler = (*APIKeyID)(nil)
	_ encoding.TextUnmarshaler = (*StatusID)(nil)
	_ driver.Valuer            = UserID{}
	_ driver.Valuer            = OrgID{}
	_ driver.Valuer            = APIKeyID{}
	_ driver.Valuer            = StatusID{}
	_ sql.Scanner              = (*UserID)(nil)
	_ sql.Scanner              = (*OrgID)(nil)
	_ sql.Scanner              = (*APIKeyID)(nil)
	_ sql.Scanner              = (*StatusID)(nil)
)

// Desired usage (from README).
func Example() {
	type Org struct {
		ID   OrgID  `json:"id"`
		Name string `json:"name"`
	}

	type User struct {
		ID    UserID `json:"id"`
		OrgID OrgID  `json:"org_id"`
		Name  string `json:"name"`
	}

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

	log.Println(orgID.String())
	log.Println(userID.String())
}
