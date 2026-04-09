package typeid

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// AnyPrefixInt64 is a compact typeid string that accepts any prefix (or none) when parsing
// and keeps that prefix for [AnyPrefixInt64.Prefix], [AnyPrefixInt64.SetPrefix], and text marshaling.
type AnyPrefixInt64 struct {
	val    int64
	prefix string
}

func NewAnyPrefixInt64() (AnyPrefixInt64, error) {
	ms := time.Now().UnixMilli()

	var rb [2]byte
	if _, err := rand.Read(rb[:]); err != nil {
		return AnyPrefixInt64{}, fmt.Errorf("typeid: crypto/rand: %w", err)
	}
	r := int64(binary.BigEndian.Uint16(rb[:]) & 0x7FFF)

	return AnyPrefixInt64{val: (ms << randomBits) | r}, nil
}

func AnyPrefixInt64From(v int64) (AnyPrefixInt64, error) {
	if v <= 0 {
		return AnyPrefixInt64{}, ErrNonPositiveInt
	}
	return AnyPrefixInt64{val: v}, nil
}

func ParseAnyPrefixInt64(s string) (AnyPrefixInt64, error) {
	j := strings.LastIndex(s, "_") + 1
	pref, suffix := s[:max(0, j-1)], s[j:]
	if len(suffix) != int64SuffixLen {
		return AnyPrefixInt64{}, fmt.Errorf("typeid: invalid format: %q", s)
	}
	v, err := decodeBase32Int64(suffix)
	if err != nil {
		return AnyPrefixInt64{}, err
	}
	if v <= 0 {
		return AnyPrefixInt64{}, ErrNonPositiveInt
	}
	return AnyPrefixInt64{val: v, prefix: pref}, nil
}

func (id AnyPrefixInt64) Int64() int64   { return id.val }
func (id AnyPrefixInt64) Prefix() string { return id.prefix }
func (id *AnyPrefixInt64) SetPrefix(s string) {
	id.prefix = s
}

func (id AnyPrefixInt64) appendText(dst []byte) []byte {
	return appendBase32Int64(dst, id.prefix, id.val)
}

func (id AnyPrefixInt64) String() string { return string(id.appendText(nil)) }

func (id AnyPrefixInt64) IsZero() bool { return id.val == 0 }

func (id AnyPrefixInt64) MarshalText() ([]byte, error) {
	if id.val <= 0 {
		return nil, ErrNonPositiveInt
	}
	return id.appendText(nil), nil
}

func (id *AnyPrefixInt64) UnmarshalText(data []byte) error {
	parsed, err := ParseAnyPrefixInt64(string(data))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

func (id AnyPrefixInt64) Value() (driver.Value, error) {
	if id.val <= 0 {
		return nil, ErrNonPositiveInt
	}
	return id.val, nil
}

func (id *AnyPrefixInt64) Scan(src any) error {
	var v int64
	switch sv := src.(type) {
	case int64:
		v = sv
	case int:
		v = int64(sv)
	default:
		return fmt.Errorf("typeid: cannot scan %T into AnyPrefixInt64", src)
	}
	if v <= 0 {
		return ErrNonPositiveInt
	}
	id.val = v
	id.prefix = ""
	return nil
}
