package typeid

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AnyUUID is a UUIDv7 identifier with a runtime-configurable prefix.
// Unlike [UUID], the prefix is not fixed at compile time.
type AnyUUID struct {
	val    uuid.UUID
	prefix string
}

func NewAnyUUID(prefix string) (AnyUUID, error) {
	u, err := uuid.NewV7()
	if err != nil {
		return AnyUUID{}, err
	}
	return AnyUUID{val: u, prefix: prefix}, nil
}

func AnyUUIDFrom(prefix string, u uuid.UUID) (AnyUUID, error) {
	if u.Version() != 7 {
		return AnyUUID{}, ErrOnlyV7
	}
	return AnyUUID{val: u, prefix: prefix}, nil
}

func ParseAnyUUID(s string) (AnyUUID, error) {
	j := strings.LastIndex(s, "_") + 1
	pref, suffix := s[:max(0, j-1)], s[j:]
	if len(suffix) != uuidSuffixLen {
		return AnyUUID{}, fmt.Errorf("typeid: invalid format: %q", s)
	}
	b, err := decodeBase32UUID(suffix)
	if err != nil {
		return AnyUUID{}, err
	}
	u := uuid.UUID(b)
	if u.Version() != 7 {
		return AnyUUID{}, ErrOnlyV7
	}
	return AnyUUID{val: u, prefix: pref}, nil
}

func (id AnyUUID) UUID() uuid.UUID { return id.val }
func (id AnyUUID) Prefix() string  { return id.prefix }
func (id *AnyUUID) SetPrefix(s string) {
	id.prefix = s
}

func (id AnyUUID) appendText(dst []byte) []byte {
	return appendBase32UUID(dst, id.prefix, id.val)
}

func (id AnyUUID) String() string {
	var buf [64]byte
	return string(id.appendText(buf[:0]))
}

func (id AnyUUID) IsZero() bool { return id.val == uuid.UUID{} }

func (id AnyUUID) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.appendText(nil), nil
}

func (id *AnyUUID) UnmarshalText(data []byte) error {
	parsed, err := ParseAnyUUID(string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id AnyUUID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.val.String(), nil
}

func (id *AnyUUID) Scan(src any) (err error) {
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
		return fmt.Errorf("typeid: cannot scan %T into AnyUUID", src)
	}
	if u.Version() != 7 {
		return ErrOnlyV7
	}
	id.val = u
	return nil
}

// GetTime extracts the millisecond-precision creation timestamp from the UUIDv7.
func (id AnyUUID) GetTime() time.Time {
	ms := int64(binary.BigEndian.Uint16(id.val[:2]))<<32 | int64(binary.BigEndian.Uint32(id.val[2:6]))
	return time.UnixMilli(ms)
}
