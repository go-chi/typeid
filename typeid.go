package typeid

import (
	"errors"
	"fmt"
)

// Prefixer is the constraint for type-safe ID prefixes.
type Prefixer interface {
	Prefix() string
}

var (
	ErrOnlyV7         = errors.New("typeid: only UUIDv7 is supported")
	ErrNegativeInt    = errors.New("typeid: int64 must be non-negative")
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
func splitTypeid[P Prefixer](s string, suffixLen int) (suffix string, err error) {
	var p P
	want := p.Prefix()

	minLen := len(want) + 1 + suffixLen
	if len(s) < minLen {
		return "", fmt.Errorf("typeid: invalid format: %q", s)
	}

	sep := len(s) - suffixLen - 1
	if s[sep] != '_' {
		return "", fmt.Errorf("typeid: invalid format: %q", s)
	}

	prefix := s[:sep]
	if prefix != want {
		return "", fmt.Errorf("typeid: prefix mismatch: expected %q, got %q", want, prefix)
	}

	return s[sep+1:], nil
}

func appendID[P Prefixer](dst []byte) []byte {
	var p P
	dst = append(dst, p.Prefix()...)
	return append(dst, '_')
}
