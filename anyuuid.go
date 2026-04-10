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
//
// Use [AnyPrefix] as P for unconstrained prefixes, or a custom enum type
// implementing [VariablePrefixer] for a known set of variants.
type AnyUUID[P Prefixer] struct {
	prefix P
	val    uuid.UUID
}

func NewAnyUUID[P Prefixer](p P) (AnyUUID[P], error) {
	u, err := uuid.NewV7()
	if err != nil {
		return AnyUUID[P]{}, err
	}
	return AnyUUID[P]{val: u, prefix: p}, nil
}

func AnyUUIDFrom[P Prefixer](p P, u uuid.UUID) (AnyUUID[P], error) {
	if u.Version() != 7 {
		return AnyUUID[P]{}, ErrOnlyV7
	}
	return AnyUUID[P]{val: u, prefix: p}, nil
}

func ParseAnyUUID[P Prefixer](s string) (AnyUUID[P], error) {
	var p P
	j := strings.LastIndex(s, "_") + 1
	pref, suffix := s[:max(0, j-1)], s[j:]
	if len(suffix) != uuidSuffixLen {
		return AnyUUID[P]{}, fmt.Errorf("typeid: invalid format: %q", s)
	}
	if vp, ok := any(&p).(VariablePrefixer); ok {
		vp.ParsePrefix(pref)
	}
	if p.Prefix() != pref {
		return AnyUUID[P]{}, fmt.Errorf("typeid: invalid prefix: %q", pref)
	}
	b, err := decodeBase32UUID(suffix)
	if err != nil {
		return AnyUUID[P]{}, err
	}
	u := uuid.UUID(b)
	if u.Version() != 7 {
		return AnyUUID[P]{}, ErrOnlyV7
	}
	return AnyUUID[P]{val: u, prefix: p}, nil
}

func (id AnyUUID[P]) UUID() uuid.UUID { return id.val }
func (id AnyUUID[P]) Prefix() string  { return id.prefix.Prefix() }
func (id AnyUUID[P]) Variant() P      { return id.prefix }
func (id *AnyUUID[P]) SetPrefix(p P)  { id.prefix = p }

func (id AnyUUID[P]) appendText(dst []byte) []byte {
	return appendBase32UUID(dst, id.prefix.Prefix(), id.val)
}

func (id AnyUUID[P]) String() string {
	var buf [64]byte
	return string(id.appendText(buf[:0]))
}

func (id AnyUUID[P]) IsZero() bool { return id.val == uuid.UUID{} }

func (id AnyUUID[P]) MarshalText() ([]byte, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.appendText(nil), nil
}

func (id *AnyUUID[P]) UnmarshalText(data []byte) error {
	parsed, err := ParseAnyUUID[P](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id AnyUUID[P]) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, ErrZeroUUID
	}
	return id.val.String(), nil
}

func (id *AnyUUID[P]) Scan(src any) (err error) {
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
func (id AnyUUID[P]) GetTime() time.Time {
	ms := int64(binary.BigEndian.Uint16(id.val[:2]))<<32 | int64(binary.BigEndian.Uint32(id.val[2:6]))
	return time.UnixMilli(ms)
}
