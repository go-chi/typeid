package typeid

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// ScopedUUID is a type-safe UUIDv7 identifier that supports multiple prefixes.
// The scope selects which prefix is used for string encoding.
type ScopedUUID[P ScopedPrefixer] struct {
	val   uuid.UUID
	scope uint8
}

func NewScopedUUID[P ScopedPrefixer](scope uint8) (ScopedUUID[P], error) {
	u, err := uuid.NewV7()
	if err != nil {
		return ScopedUUID[P]{}, err
	}
	return ScopedUUID[P]{val: u, scope: scope}, nil
}

func ScopedUUIDFrom[P ScopedPrefixer](u uuid.UUID, scope uint8) (ScopedUUID[P], error) {
	if u.Version() != 7 {
		return ScopedUUID[P]{}, ErrOnlyV7
	}
	return ScopedUUID[P]{val: u, scope: scope}, nil
}

func ParseScopedUUID[P ScopedPrefixer](s string) (ScopedUUID[P], error) {
	scope, suffix, err := splitScopedTypeid[P](s, uuidSuffixLen)
	if err != nil {
		return ScopedUUID[P]{}, err
	}

	b, err := decodeBase32UUID(suffix)
	if err != nil {
		return ScopedUUID[P]{}, err
	}
	u := uuid.UUID(b)
	if u.Version() != 7 {
		return ScopedUUID[P]{}, ErrOnlyV7
	}
	return ScopedUUID[P]{val: u, scope: scope}, nil
}

func (id ScopedUUID[P]) String() string  { return formatScopedID[P](id.scope, encodeBase32UUID(id.val)) }
func (id ScopedUUID[P]) UUID() uuid.UUID { return id.val }
func (id ScopedUUID[P]) Scope() uint8    { return id.scope }
func (id ScopedUUID[P]) IsZero() bool    { return id.val == uuid.UUID{} }
func (id ScopedUUID[P]) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

func (id *ScopedUUID[P]) UnmarshalText(data []byte) error {
	parsed, err := ParseScopedUUID[P](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id ScopedUUID[P]) Value() (driver.Value, error) { return id.String(), nil }

func (id *ScopedUUID[P]) Scan(src any) error {
	switch v := src.(type) {
	case string:
		parsed, err := ParseScopedUUID[P](v)
		if err != nil {
			return err
		}
		*id = parsed
	case []byte:
		parsed, err := ParseScopedUUID[P](string(v))
		if err != nil {
			return err
		}
		*id = parsed
	default:
		return fmt.Errorf("typeid: cannot scan %T into ScopedUUID", src)
	}
	return nil
}
