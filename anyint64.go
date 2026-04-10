package typeid

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// AnyInt64 is a compact identifier with a runtime-configurable prefix.
// Unlike [Int64], the prefix is not fixed at compile time.
//
// Use [AnyPrefix] as P for unconstrained prefixes, or a custom enum type
// implementing [VariablePrefixer] for a known set of variants.
type AnyInt64[P Prefixer] struct {
	prefix P
	val    int64
}

func NewAnyInt64[P Prefixer](p P) (AnyInt64[P], error) {
	ms := time.Now().UnixMilli()

	var rb [2]byte
	if _, err := rand.Read(rb[:]); err != nil {
		return AnyInt64[P]{}, fmt.Errorf("typeid: crypto/rand: %w", err)
	}
	r := int64(binary.BigEndian.Uint16(rb[:]) & 0x7FFF)

	return AnyInt64[P]{val: (ms << randomBits) | r, prefix: p}, nil
}

func AnyInt64From[P Prefixer](p P, v int64) (AnyInt64[P], error) {
	if v <= 0 {
		return AnyInt64[P]{}, ErrNonPositiveInt
	}
	return AnyInt64[P]{val: v, prefix: p}, nil
}

func ParseAnyInt64[P Prefixer](s string) (AnyInt64[P], error) {
	var p P
	j := strings.LastIndex(s, "_") + 1
	pref, suffix := s[:max(0, j-1)], s[j:]
	if len(suffix) != int64SuffixLen {
		return AnyInt64[P]{}, fmt.Errorf("typeid: invalid format: %q", s)
	}
	if vp, ok := any(&p).(VariablePrefixer); ok {
		vp.ParsePrefix(pref)
	}
	if p.Prefix() != pref {
		return AnyInt64[P]{}, fmt.Errorf("typeid: invalid prefix: %q", pref)
	}
	v, err := decodeBase32Int64(suffix)
	if err != nil {
		return AnyInt64[P]{}, err
	}
	if v <= 0 {
		return AnyInt64[P]{}, ErrNonPositiveInt
	}
	return AnyInt64[P]{val: v, prefix: p}, nil
}

func (id AnyInt64[P]) Int64() int64   { return id.val }
func (id AnyInt64[P]) Prefix() string { return id.prefix.Prefix() }
func (id AnyInt64[P]) Variant() P     { return id.prefix }
func (id *AnyInt64[P]) SetPrefix(p P) { id.prefix = p }

func (id AnyInt64[P]) appendText(dst []byte) []byte {
	return appendBase32Int64(dst, id.prefix.Prefix(), id.val)
}

func (id AnyInt64[P]) String() string {
	var buf [64]byte
	return string(id.appendText(buf[:0]))
}

func (id AnyInt64[P]) IsZero() bool { return id.val == 0 }

func (id AnyInt64[P]) MarshalText() ([]byte, error) {
	if id.val <= 0 {
		return nil, ErrNonPositiveInt
	}
	return id.appendText(nil), nil
}

func (id *AnyInt64[P]) UnmarshalText(data []byte) error {
	parsed, err := ParseAnyInt64[P](string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id AnyInt64[P]) Value() (driver.Value, error) {
	if id.val <= 0 {
		return nil, ErrNonPositiveInt
	}
	return id.val, nil
}

func (id *AnyInt64[P]) Scan(src any) error {
	var v int64
	switch sv := src.(type) {
	case int64:
		v = sv
	case int:
		v = int64(sv)
	default:
		return fmt.Errorf("typeid: cannot scan %T into AnyInt64", src)
	}
	if v <= 0 {
		return ErrNonPositiveInt
	}
	id.val = v
	return nil
}
