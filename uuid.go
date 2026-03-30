package typeid

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

// UUID is a type-safe UUIDv7 identifier with a compile-time prefix.
// Maps to Postgres uuid.
type UUID[P Prefixer] struct {
	val    uuid.UUID
	prefix string // only used when P is [AnyPrefix]; holds runtime prefix for parse/marshal
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
	var p P
	var suffix string
	var dynPref string
	switch any(p).(type) {
	case AnyPrefix:
		var err error
		dynPref, suffix, err = splitTypeidAny(s, uuidSuffixLen)
		if err != nil {
			return UUID[P]{}, err
		}
	default:
		var err error
		suffix, err = splitTypeid[P](s, uuidSuffixLen)
		if err != nil {
			return UUID[P]{}, err
		}
	}
	b, err := decodeBase32UUID(suffix)
	if err != nil {
		return UUID[P]{}, err
	}
	u := uuid.UUID(b)
	if u.Version() != 7 {
		return UUID[P]{}, ErrOnlyV7
	}
	switch any(p).(type) {
	case AnyPrefix:
		return UUID[P]{val: u, prefix: dynPref}, nil
	default:
		return UUID[P]{val: u}, nil
	}
}

// Prefix returns the type's fixed prefix, or the runtime prefix for [UUID[AnyPrefix]].
func (id UUID[P]) Prefix() string {
	var p P
	switch any(p).(type) {
	case AnyPrefix:
		return id.prefix
	default:
		return p.Prefix()
	}
}

// SetPrefix updates the stored prefix for [UUID[AnyPrefix]] only; it is a no-op for other P.
func (id *UUID[P]) SetPrefix(s string) {
	var p P
	if _, ok := any(p).(AnyPrefix); ok {
		id.prefix = s
	}
}

func (id UUID[P]) appendText(dst []byte) []byte {
	pref := id.Prefix()
	n := uuidSuffixLen
	if pref != "" {
		n += len(pref) + 1
	}
	dst = growSlice(dst, n)
	if pref != "" {
		dst = append(dst, pref...)
		dst = append(dst, '_')
	}
	return appendBase32UUID(dst, id.val)
}
func (id UUID[P]) String() string { return string(id.appendText(nil)) }
func (id UUID[P]) UUID() uuid.UUID              { return id.val }
func (id UUID[P]) IsZero() bool                 { return id.val == uuid.UUID{} }
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
	var p P
	if _, ok := any(p).(AnyPrefix); ok {
		id.prefix = ""
	}
	return nil
}
