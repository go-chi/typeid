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
type AnyInt64 struct {
	val    int64
	prefix string
}

func NewAnyInt64(prefix string) (AnyInt64, error) {
	ms := time.Now().UnixMilli()

	var rb [2]byte
	if _, err := rand.Read(rb[:]); err != nil {
		return AnyInt64{}, fmt.Errorf("typeid: crypto/rand: %w", err)
	}
	r := int64(binary.BigEndian.Uint16(rb[:]) & 0x7FFF)

	return AnyInt64{val: (ms << randomBits) | r, prefix: prefix}, nil
}

func AnyInt64From(prefix string, v int64) (AnyInt64, error) {
	if v <= 0 {
		return AnyInt64{}, ErrNonPositiveInt
	}
	return AnyInt64{val: v, prefix: prefix}, nil
}

func ParseAnyInt64(s string) (AnyInt64, error) {
	j := strings.LastIndex(s, "_") + 1
	pref, suffix := s[:max(0, j-1)], s[j:]
	if len(suffix) != int64SuffixLen {
		return AnyInt64{}, fmt.Errorf("typeid: invalid format: %q", s)
	}
	v, err := decodeBase32Int64(suffix)
	if err != nil {
		return AnyInt64{}, err
	}
	if v <= 0 {
		return AnyInt64{}, ErrNonPositiveInt
	}
	return AnyInt64{val: v, prefix: pref}, nil
}

func (id AnyInt64) Int64() int64   { return id.val }
func (id AnyInt64) Prefix() string { return id.prefix }
func (id *AnyInt64) SetPrefix(s string) {
	id.prefix = s
}

func (id AnyInt64) appendText(dst []byte) []byte {
	return appendBase32Int64(dst, id.prefix, id.val)
}

func (id AnyInt64) String() string {
	var buf [64]byte
	return string(id.appendText(buf[:0]))
}

func (id AnyInt64) IsZero() bool { return id.val == 0 }

func (id AnyInt64) MarshalText() ([]byte, error) {
	if id.val <= 0 {
		return nil, ErrNonPositiveInt
	}
	return id.appendText(nil), nil
}

func (id *AnyInt64) UnmarshalText(data []byte) error {
	parsed, err := ParseAnyInt64(string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id AnyInt64) Value() (driver.Value, error) {
	if id.val <= 0 {
		return nil, ErrNonPositiveInt
	}
	return id.val, nil
}

func (id *AnyInt64) Scan(src any) error {
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
