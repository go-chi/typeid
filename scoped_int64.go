package typeid

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"time"
)

// ScopedInt64 is a type-safe compact identifier that supports multiple prefixes.
// The scope selects which prefix is used for string encoding.
type ScopedInt64[P ScopedPrefixer] struct {
	val   int64
	scope uint8
}

func NewScopedInt64[P ScopedPrefixer](scope uint8) (ScopedInt64[P], error) {
	ms := time.Now().UnixMilli()

	var rb [2]byte
	if _, err := rand.Read(rb[:]); err != nil {
		return ScopedInt64[P]{}, fmt.Errorf("typeid: crypto/rand: %w", err)
	}
	r := int64(binary.BigEndian.Uint16(rb[:]) & 0x7FFF)

	return ScopedInt64[P]{val: (ms << randomBits) | r, scope: scope}, nil
}

func ScopedInt64From[P ScopedPrefixer](v int64, scope uint8) (ScopedInt64[P], error) {
	if v < 0 {
		return ScopedInt64[P]{}, ErrNegativeInt
	}
	return ScopedInt64[P]{val: v, scope: scope}, nil
}

func ParseScopedInt64[P ScopedPrefixer](s string) (ScopedInt64[P], error) {
	scope, suffix, err := splitScopedTypeid[P](s, int64SuffixLen)
	if err != nil {
		return ScopedInt64[P]{}, err
	}

	v, err := decodeBase32Int64(suffix)
	if err != nil {
		return ScopedInt64[P]{}, err
	}
	return ScopedInt64[P]{val: v, scope: scope}, nil
}

func (id ScopedInt64[P]) String() string { return formatScopedID[P](id.scope, encodeBase32Int64(id.val)) }
func (id ScopedInt64[P]) Int64() int64   { return id.val }
func (id ScopedInt64[P]) Scope() uint8   { return id.scope }
func (id ScopedInt64[P]) IsZero() bool   { return id.val == 0 }
func (id ScopedInt64[P]) MarshalText() ([]byte, error) { return []byte(id.String()), nil }

func (id *ScopedInt64[P]) UnmarshalText(data []byte) error {
	parsed, err := ParseScopedInt64[P](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id ScopedInt64[P]) Value() (driver.Value, error) { return id.String(), nil }

func (id *ScopedInt64[P]) Scan(src any) error {
	switch v := src.(type) {
	case string:
		parsed, err := ParseScopedInt64[P](v)
		if err != nil {
			return err
		}
		*id = parsed
	case []byte:
		parsed, err := ParseScopedInt64[P](string(v))
		if err != nil {
			return err
		}
		*id = parsed
	default:
		return fmt.Errorf("typeid: cannot scan %T into ScopedInt64", src)
	}
	return nil
}
