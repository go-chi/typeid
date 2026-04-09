package typeid

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// AnyPrefixUUID is a UUIDv7 typeid string that accepts any prefix (or none) when parsing
// and keeps that prefix for [AnyPrefixUUID.Prefix], [AnyPrefixUUID.SetPrefix], and text marshaling.
type AnyPrefixUUID struct {
	val    uuid.UUID
	prefix string
}

func NewAnyPrefixUUID() (AnyPrefixUUID, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return AnyPrefixUUID{}, err
	}
	return AnyPrefixUUID{val: u}, nil
}

func AnyPrefixUUIDFrom(u uuid.UUID) (AnyPrefixUUID, error) {
	if u.Version() != 7 {
		return AnyPrefixUUID{}, ErrOnlyV7
	}
	return AnyPrefixUUID{val: u}, nil
}

func ParseAnyPrefixUUID(s string) (AnyPrefixUUID, error) {
	j := strings.LastIndex(s, "_") + 1
	pref, suffix := s[:max(0, j-1)], s[j:]
	if len(suffix) != uuidSuffixLen {
		return AnyPrefixUUID{}, fmt.Errorf("typeid: invalid format: %q", s)
	}
	b, err := decodeBase32UUID(suffix)
	if err != nil {
		return AnyPrefixUUID{}, err
	}
	u := uuid.UUID(b)
	if u.Version() != 7 {
		return AnyPrefixUUID{}, ErrOnlyV7
	}
	return AnyPrefixUUID{val: u, prefix: pref}, nil
}

func (id AnyPrefixUUID) UUID() uuid.UUID { return id.val }
func (id AnyPrefixUUID) Prefix() string  { return id.prefix }
func (id *AnyPrefixUUID) SetPrefix(s string) {
	id.prefix = s
}

func (id AnyPrefixUUID) appendText(dst []byte) []byte {
	return appendBase32UUID(dst, id.prefix, id.val)
}

func (id AnyPrefixUUID) String() string { return string(id.appendText(nil)) }

func (id AnyPrefixUUID) IsZero() bool { return id.val == uuid.UUID{} }

func (id AnyPrefixUUID) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.appendText(nil), nil
}

func (id *AnyPrefixUUID) UnmarshalText(data []byte) error {
	parsed, err := ParseAnyPrefixUUID(string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id AnyPrefixUUID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.val.String(), nil
}

func (id *AnyPrefixUUID) Scan(src any) (err error) {
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
		return fmt.Errorf("typeid: cannot scan %T into AnyPrefixUUID", src)
	}
	if u.Version() != 7 {
		return ErrOnlyV7
	}
	id.val = u
	id.prefix = ""
	return nil
}
