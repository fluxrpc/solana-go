package solana_go

import "errors"

// The Solana wire format prefixes variable-length sequences with a
// "shortvec" (compact-u16) length: a little-endian base-128 varint capped at
// 3 bytes and a maximum value of 0xFFFF.

var (
	errShortvecTruncated    = errors.New("shortvec: truncated length")
	errShortvecTooLarge     = errors.New("shortvec: length exceeds uint16 range")
	errShortvecNonCanonical = errors.New("shortvec: non-canonical length encoding")
)

// appendShortvecLen appends the shortvec encoding of n to dst.
func appendShortvecLen(dst []byte, n int) []byte {
	for n >= 0x80 {
		dst = append(dst, byte(n)|0x80)
		n >>= 7
	}
	return append(dst, byte(n))
}

// shortvecLen returns the number of bytes appendShortvecLen writes for n.
func shortvecLen(n int) int {
	size := 1
	for n >= 0x80 {
		n >>= 7
		size++
	}
	return size
}

// decodeShortvecLen decodes a shortvec length from the start of data,
// returning the value and the number of bytes consumed.
func decodeShortvecLen(data []byte) (value int, bytesRead int, err error) {
	for i := 0; i < 3; i++ {
		if i >= len(data) {
			return 0, 0, errShortvecTruncated
		}
		b := data[i]
		// A zero continuation byte adds nothing, making the encoding
		// non-minimal. The Solana runtime rejects such aliases, and
		// accepting them would let one message have several encodings.
		if i > 0 && b == 0 {
			return 0, 0, errShortvecNonCanonical
		}
		value |= int(b&0x7F) << (7 * i)
		if b&0x80 == 0 {
			if value > 0xFFFF {
				return 0, 0, errShortvecTooLarge
			}
			return value, i + 1, nil
		}
	}
	// A third byte with the continuation bit set can only encode values
	// beyond the uint16 range.
	return 0, 0, errShortvecTooLarge
}
