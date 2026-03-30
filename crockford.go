package typeid

import (
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/google/uuid"
)

// Crockford base32 alphabet (lowercase). Excludes i, l, o, u.
const alphabet = "0123456789abcdefghjkmnpqrstvwxyz"

// decode maps every ASCII byte to its 5-bit value (0–31), or 0xFF if invalid.
var decode = func() (t [256]byte) {
	for i := range t {
		t[i] = 0xFF
	}
	for i, c := range alphabet {
		t[c] = byte(i)
		if c >= 'a' && c <= 'z' {
			t[c-32] = byte(i)
		}
	}
	return
}()

func decodeChar(c byte) (byte, error) {
	v := decode[c]
	if v == 0xFF {
		return 0, fmt.Errorf("typeid: invalid base32 character %q", c)
	}
	return v, nil
}

// appendBase32UUID appends an optional prefix and underscore, then the 26-char typeid suffix for u.
func appendBase32UUID(dst []byte, prefix string, u uuid.UUID) []byte {
	dst = slices.Grow(dst, uuidSuffixLen+len(prefix)+min(1, len(prefix)))
	if len(prefix) > 0 {
		dst = append(append(dst, prefix...), '_')
	}
	hi := binary.BigEndian.Uint64(u[:8])
	lo := binary.BigEndian.Uint64(u[8:])

	var buf [uuidSuffixLen]byte

	for i := 25; i >= 14; i-- {
		buf[i] = alphabet[lo&0x1F]
		lo >>= 5
	}
	// char 13 straddles hi/lo: 4 remaining lo bits + 1 bit from hi
	buf[13] = alphabet[(lo&0x0F)|((hi&0x01)<<4)]
	hi >>= 1

	for i := 12; i >= 1; i-- {
		buf[i] = alphabet[hi&0x1F]
		hi >>= 5
	}
	buf[0] = alphabet[hi&0x07]

	return append(dst, buf[:]...)
}

// UUID decoding (26 chars -> 128 bits)
func decodeBase32UUID(s string) ([16]byte, error) {
	if len(s) != uuidSuffixLen {
		return [16]byte{}, fmt.Errorf("typeid: invalid suffix length %d", len(s))
	}

	v, err := decodeChar(s[0])
	if err != nil {
		return [16]byte{}, err
	}
	if v > 7 {
		return [16]byte{}, ErrOverflowBase32
	}
	hi := uint64(v)

	for i := 1; i <= 12; i++ {
		v, err = decodeChar(s[i])
		if err != nil {
			return [16]byte{}, err
		}
		hi = (hi << 5) | uint64(v)
	}

	// char 13 straddle: top 1 bit → hi, bottom 4 bits → lo
	v, err = decodeChar(s[13])
	if err != nil {
		return [16]byte{}, err
	}
	hi = (hi << 1) | uint64(v>>4)
	lo := uint64(v & 0x0F)

	for i := 14; i <= 25; i++ {
		v, err = decodeChar(s[i])
		if err != nil {
			return [16]byte{}, err
		}
		lo = (lo << 5) | uint64(v)
	}

	var out [16]byte
	binary.BigEndian.PutUint64(out[:8], hi)
	binary.BigEndian.PutUint64(out[8:], lo)
	return out, nil
}

// appendBase32Int64 appends an optional prefix and underscore, then the 13-char typeid suffix for n.
func appendBase32Int64(dst []byte, prefix string, n int64) []byte {
	dst = slices.Grow(dst, int64SuffixLen+len(prefix)+min(1, len(prefix)))
	if len(prefix) > 0 {
		dst = append(append(dst, prefix...), '_')
	}
	u := uint64(n)
	var buf [int64SuffixLen]byte

	for i := 12; i >= 1; i-- {
		buf[i] = alphabet[u&0x1F]
		u >>= 5
	}
	buf[0] = alphabet[u&0x07]

	return append(dst, buf[:]...)
}

// Int64 decoding (13 chars -> 63 bits)
func decodeBase32Int64(s string) (int64, error) {
	if len(s) != int64SuffixLen {
		return 0, fmt.Errorf("typeid: invalid suffix length %d", len(s))
	}

	v, err := decodeChar(s[0])
	if err != nil {
		return 0, err
	}
	if v > 7 {
		return 0, ErrOverflowBase32
	}
	val := uint64(v)

	for i := 1; i < int64SuffixLen; i++ {
		v, err = decodeChar(s[i])
		if err != nil {
			return 0, err
		}
		val = (val << 5) | uint64(v)
	}

	if val > 1<<63-1 {
		return 0, ErrOverflowInt64
	}
	return int64(val), nil
}
