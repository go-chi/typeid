package typeid

import (
	"errors"
	"fmt"
	"strings"
)

// Prefixer is the constraint for type-safe ID prefixes.
type Prefixer interface {
	Prefix() string
}

// VariablePrefixer is optionally implemented by prefix types that accept
// multiple string representations (e.g. "api_key" and "api_key_sandbox").
// ParsePrefix sets the receiver to the variant matching s and reports success.
type VariablePrefixer interface {
	ParsePrefix(s string) bool
}

// AnyPrefix accepts any prefix string. Use it as the type parameter for
// [AnyUUID] or [AnyInt64] when the set of valid prefixes is not known at
// compile time:
//
//	type FlexID = typeid.AnyUUID[typeid.AnyPrefix]
type AnyPrefix string

func (p AnyPrefix) Prefix() string             { return string(p) }
func (p *AnyPrefix) ParsePrefix(s string) bool { *p = AnyPrefix(s); return true }

var (
	ErrOnlyV7         = errors.New("typeid: only UUIDv7 is supported")
	ErrZeroUUID       = errors.New("typeid: zero UUID")
	ErrNonPositiveInt = errors.New("typeid: non-positive Int64")
	ErrOverflowBase32 = errors.New("typeid: base32 overflow at pos 0")
	ErrOverflowInt64  = errors.New("typeid: value overflows int64")
)

const (
	uuidSuffixLen  = 26 // 130-bit capacity, 128 used
	int64SuffixLen = 13 // 65-bit capacity, 63 used
)

// splitTypeid splits "prefix_<suffix>" from the right using known suffix length.
// Supports underscores in the prefix (e.g. "project_env_<suffix>").
func splitTypeid[P Prefixer](s string, suffixLen int) (string, error) {
	var p P
	want := p.Prefix()
	j := strings.LastIndex(s, "_") + 1 // 0 = bare suffix; else first byte after last '_'
	prefix, suffix := s[:max(0, j-1)], s[j:]
	if len(suffix) != suffixLen {
		return "", fmt.Errorf("typeid: invalid format: %q", s)
	}
	if prefix != want {
		return "", fmt.Errorf("typeid: prefix mismatch: expected %q, got %q", want, prefix)
	}
	return suffix, nil
}
