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

func (id UUID[P]) appendText(dst []byte) []byte {
	var p P
	return appendBase32UUID(dst, p.Prefix(), id.val)
}
func (id UUID[P]) String() string {
	var buf [64]byte
	return string(id.appendText(buf[:0]))
}
func (id UUID[P]) UUID() uuid.UUID { return id.val }
func (id UUID[P]) IsZero() bool    { return id.val == uuid.UUID{} }
func (id UUID[P]) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.appendText(nil), nil
}

func (id *UUID[P]) UnmarshalText(data []byte) error {
	parsed, err := ParseUUID[P](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id UUID[P]) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.val.String(), nil
}

// Any converts a typed UUID to an AnyUUID with the same prefix and value.
func (id UUID[P]) Any() AnyUUID[AnyPrefix] {
	var p P
	return AnyUUID[AnyPrefix]{val: id.val, prefix: AnyPrefix(p.Prefix())}
}

func (id *UUID[P]) Scan(src any) (err error) {
	var u uuid.UUID
	switch v := src.(type) {
	case string:
		if u, err = uuid.Parse(v); err != nil {
			return err
		}
	case []byte:
		switch {
		case len(v) == 16:
			copy(u[:], v)
		default:
			if u, err = uuid.ParseBytes(v); err != nil {
				return err
			}
		}
	case [16]byte:
		u = uuid.UUID(v)
	default:
		return fmt.Errorf("typeid: cannot scan %T into UUID", src)
	}
	if u.Version() != 7 {
		return ErrOnlyV7
	}
	id.val = u
	return nil
}
