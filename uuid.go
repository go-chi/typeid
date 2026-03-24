package typeid

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// UUID is a type-safe UUIDv7 identifier with a compile-time prefix.
// Maps to Postgres uuid.
type UUID[P Prefixer] struct {
	val uuid.UUID
}

func NewUUID[P Prefixer]() (UUID[P], error) {
	u, err := uuid.NewV7()
	if err != nil {
		return UUID[P]{}, err
	}
	return UUID[P]{val: u}, nil
}

func UUIDFrom[P Prefixer](u uuid.UUID) (UUID[P], error) {
	if u.Version() != 7 {
		return UUID[P]{}, ErrOnlyV7
	}
	return UUID[P]{val: u}, nil
}

func ParseUUID[P Prefixer](s string) (UUID[P], error) {
	suffix, err := splitTypeid[P](s, uuidSuffixLen)
	if err != nil {
		return UUID[P]{}, err
	}

	b, err := decodeBase32UUID(suffix)
	if err != nil {
		return UUID[P]{}, err
	}
	u := uuid.UUID(b)
	if u.Version() != 7 {
		return UUID[P]{}, ErrOnlyV7
	}
	return UUID[P]{val: u}, nil
}

func (id UUID[P]) String() string         { return formatID[P](encodeBase32UUID(id.val)) }
func (id UUID[P]) UUID() uuid.UUID        { return id.val }
func (id UUID[P]) IsZero() bool           { return id.val == uuid.UUID{} }
func (id UUID[P]) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

func (id *UUID[P]) UnmarshalText(data []byte) error {
	parsed, err := ParseUUID[P](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id UUID[P]) Value() (driver.Value, error) { return id.val.String(), nil }

func (id *UUID[P]) Scan(src any) error {
	switch v := src.(type) {
	case string:
		u, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		id.val = u
	case []byte:
		u, err := uuid.ParseBytes(v)
		if err != nil {
			return err
		}
		id.val = u
	case [16]byte:
		id.val = uuid.UUID(v)
	default:
		return fmt.Errorf("typeid: cannot scan %T into UUID", src)
	}
	return nil
}
