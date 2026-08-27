// Package binary provides a small, allocation-free encoder and decoder for the
// little-endian ("bincode") and Borsh binary layouts used by Solana account
// state and instruction data.
//
// The package is deliberately minimal: there is no reflection or struct-tag
// machinery. Callers read and write fields explicitly in layout order, which
// keeps the hot paths inlineable and makes buffer ownership explicit.
//
// Errors are sticky: the first failed read records the error and every
// subsequent read returns the zero value without advancing. Decode a struct
// with straight-line reads and check Err once at the end:
//
//	dec := bin.NewDecoder(data)
//	mint := dec.ReadPublicKey()
//	owner := dec.ReadPublicKey()
//	amount := dec.ReadUint64()
//	if err := dec.Err(); err != nil {
//		return err
//	}
//
// ReadBytes, ReadBorshString, and ReadBincodeString return views that alias the
// input buffer rather than copies — the same convention as
// TransactionFromBytes in the root package. Use the Copy variants when the
// result must outlive the buffer.
package binary

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf8"
	"unsafe"

	solana "github.com/fluxrpc/solana-go"
)

// Sentinel errors, matchable with errors.Is against Decoder.Err results.
var (
	// ErrUnexpectedEOF is recorded when a read needs more bytes than remain.
	ErrUnexpectedEOF = errors.New("binary: unexpected EOF")

	// ErrInvalidTag is recorded when an Option, COption or Bool tag holds a
	// value other than 0 or 1.
	ErrInvalidTag = errors.New("binary: invalid tag")

	// ErrNonCanonical is recorded when a compact-u16 length uses a
	// non-minimal encoding. The Solana runtime rejects such aliases, and
	// accepting them would let one payload have several encodings.
	ErrNonCanonical = errors.New("binary: non-canonical compact-u16")

	// ErrOverflow is recorded when a length cannot be represented by its wire
	// format or a requested byte count is negative.
	ErrOverflow = errors.New("binary: length overflow")

	// ErrInvalidUTF8 is recorded when a string contains invalid UTF-8 bytes.
	ErrInvalidUTF8 = errors.New("binary: invalid UTF-8")
)

// Decoder reads little-endian and Borsh values from a byte slice.
//
// The zero value is an empty decoder; use NewDecoder. A Decoder must not be
// used from multiple goroutines concurrently.
type Decoder struct {
	data []byte
	pos  int

	// Failure state. The read methods record only the bare sentinel plus its
	// two operands (below) so their cold paths are plain assignments — a call
	// or fmt.Errorf in the body would push them past the inlining budget. Err
	// wraps the sentinel with this context once, lazily.
	err     error
	failPos int // offset at the point of failure
	failArg int // bytes needed (EOF), tag value (tag), or count (overflow)
}

// NewDecoder returns a Decoder reading from data. The Decoder aliases data;
// the caller must not mutate it while decoding.
func NewDecoder(data []byte) *Decoder {
	return &Decoder{data: data}
}

// Err returns the first error encountered by any read, or nil. The returned
// error wraps one of the package sentinels and adds offset context.
func (d *Decoder) Err() error {
	if d.err == nil {
		return nil
	}
	return d.wrapErr()
}

// wrapErr builds the offset context for the recorded sentinel. Split from
// Err so Err's no-error fast path stays inlineable.
func (d *Decoder) wrapErr() error {
	// Identity comparisons: the read methods store the bare sentinels, and
	// once wrapped the identity no longer matches, so context is built once.
	switch d.err {
	case ErrUnexpectedEOF:
		d.err = fmt.Errorf("need %d bytes at offset %d, have %d: %w",
			d.failArg, d.failPos, len(d.data)-d.failPos, ErrUnexpectedEOF)
	case ErrInvalidTag:
		d.err = fmt.Errorf("tag %d at offset %d: %w", d.failArg, d.failPos, ErrInvalidTag)
	case ErrNonCanonical:
		d.err = fmt.Errorf("zero continuation byte at offset %d: %w", d.failPos, ErrNonCanonical)
	case ErrOverflow:
		d.err = fmt.Errorf("length %d at offset %d: %w", d.failArg, d.failPos, ErrOverflow)
	case ErrInvalidUTF8:
		d.err = fmt.Errorf("string at offset %d: %w", d.failPos, ErrInvalidUTF8)
	}
	return d.err
}

// Pos returns the current read offset.
func (d *Decoder) Pos() int { return d.pos }

// Len returns the total length of the underlying buffer.
func (d *Decoder) Len() int { return len(d.data) }

// Remaining returns the number of unread bytes.
func (d *Decoder) Remaining() int { return len(d.data) - d.pos }

// fail records the first failure. It is written to stay call-free so the
// read methods that contain it remain inlineable.
func (d *Decoder) fail(sentinel error, arg int) {
	if d.err == nil {
		d.err = sentinel
		d.failPos = d.pos
		d.failArg = arg
	}
}

// Skip advances the read position by n bytes without interpreting them.
func (d *Decoder) Skip(n int) {
	if d.err != nil {
		return
	}
	if n < 0 {
		d.fail(ErrOverflow, n)
		return
	}
	if len(d.data)-d.pos < n {
		d.fail(ErrUnexpectedEOF, n)
		return
	}
	d.pos += n
}

// ReadUint8 reads one byte.
func (d *Decoder) ReadUint8() uint8 {
	p := d.pos
	if d.err != nil || len(d.data)-p < 1 {
		d.fail(ErrUnexpectedEOF, 1)
		return 0
	}
	d.pos = p + 1
	return d.data[p]
}

// ReadUint16 reads a little-endian uint16.
func (d *Decoder) ReadUint16() uint16 {
	p := d.pos
	if d.err != nil || len(d.data)-p < 2 {
		d.fail(ErrUnexpectedEOF, 2)
		return 0
	}
	d.pos = p + 2
	return binary.LittleEndian.Uint16(d.data[p:])
}

// ReadUint32 reads a little-endian uint32.
func (d *Decoder) ReadUint32() uint32 {
	p := d.pos
	if d.err != nil || len(d.data)-p < 4 {
		d.fail(ErrUnexpectedEOF, 4)
		return 0
	}
	d.pos = p + 4
	return binary.LittleEndian.Uint32(d.data[p:])
}

// ReadUint64 reads a little-endian uint64.
func (d *Decoder) ReadUint64() uint64 {
	p := d.pos
	if d.err != nil || len(d.data)-p < 8 {
		d.fail(ErrUnexpectedEOF, 8)
		return 0
	}
	d.pos = p + 8
	return binary.LittleEndian.Uint64(d.data[p:])
}

// ReadInt64 reads a little-endian int64 (Borsh i64, e.g. Unix timestamps).
func (d *Decoder) ReadInt64() int64 {
	return int64(d.ReadUint64())
}

// ReadBool reads a Borsh bool: a single byte that must be 0 or 1.
func (d *Decoder) ReadBool() bool {
	p := d.pos
	if d.err != nil || len(d.data)-p < 1 {
		d.fail(ErrUnexpectedEOF, 1)
		return false
	}
	v := d.data[p]
	if v > 1 {
		d.fail(ErrInvalidTag, int(v))
		return false
	}
	d.pos = p + 1
	return v == 1
}

// ReadBytes returns the next n bytes as a view aliasing the input buffer.
// The view is only valid while the buffer is; use ReadBytesCopy otherwise.
func (d *Decoder) ReadBytes(n int) []byte {
	p := d.pos
	if d.err != nil || n < 0 || len(d.data)-p < n {
		if n < 0 {
			d.fail(ErrOverflow, n)
		} else {
			d.fail(ErrUnexpectedEOF, n)
		}
		return nil
	}
	d.pos = p + n
	return d.data[p : p+n : p+n]
}

// ReadBytesCopy returns the next n bytes as a freshly allocated copy.
func (d *Decoder) ReadBytesCopy(n int) []byte {
	v := d.ReadBytes(n)
	if v == nil {
		return nil
	}
	out := make([]byte, n)
	copy(out, v)
	return out
}

// ReadPublicKey reads a 32-byte public key.
func (d *Decoder) ReadPublicKey() solana.PublicKey {
	p := d.pos
	if d.err != nil || len(d.data)-p < solana.PublicKeyLength {
		d.fail(ErrUnexpectedEOF, solana.PublicKeyLength)
		return solana.PublicKey{}
	}
	d.pos = p + solana.PublicKeyLength
	return solana.PublicKey(d.data[p : p+solana.PublicKeyLength])
}

// ReadSignature reads a 64-byte signature.
func (d *Decoder) ReadSignature() solana.Signature {
	p := d.pos
	if d.err != nil || len(d.data)-p < solana.SignatureLength {
		d.fail(ErrUnexpectedEOF, solana.SignatureLength)
		return solana.Signature{}
	}
	d.pos = p + solana.SignatureLength
	return solana.Signature(d.data[p : p+solana.SignatureLength])
}

// ReadHash reads a 32-byte hash.
func (d *Decoder) ReadHash() solana.Hash {
	return solana.Hash(d.ReadPublicKey())
}

// ReadBorshString reads a Borsh string: a uint32 byte length followed by the
// bytes. The returned string aliases the input buffer via unsafe.String; it
// is only valid while the buffer is neither mutated nor reused. Use
// ReadBorshStringCopy when the string must outlive the buffer.
func (d *Decoder) ReadBorshString() string {
	n := int(d.ReadUint32())
	b := d.ReadBytes(n)
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// ReadBorshStringCopy reads a Borsh string as a freshly allocated string.
func (d *Decoder) ReadBorshStringCopy() string {
	return string(d.ReadBytes(int(d.ReadUint32())))
}

// ReadBincodeString reads the bincode String layout used by Solana's native
// programs: a little-endian uint64 byte length followed by UTF-8 bytes. The
// returned string aliases the input; use ReadBincodeStringCopy to take ownership.
func (d *Decoder) ReadBincodeString() string {
	b := d.readBincodeStringBytes()
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// ReadBincodeStringCopy reads a bincode String as an owned Go string.
func (d *Decoder) ReadBincodeStringCopy() string {
	return string(d.readBincodeStringBytes())
}

func (d *Decoder) readBincodeStringBytes() []byte {
	n := d.ReadUint64()
	if d.err != nil {
		return nil
	}
	maxInt := uint64(^uint(0) >> 1)
	if n > maxInt {
		d.fail(ErrOverflow, int(maxInt))
		return nil
	}
	b := d.ReadBytes(int(n))
	if d.err == nil && !utf8.Valid(b) {
		d.pos -= len(b)
		d.fail(ErrInvalidUTF8, len(b))
		return nil
	}
	return b
}

// ReadOption reads a Borsh Option tag: a single byte that must be 0 (None)
// or 1 (Some). It returns whether a value follows.
func (d *Decoder) ReadOption() bool {
	return d.ReadBool()
}

// ReadCOption reads an SPL Token COption tag: a little-endian uint32 that
// must be 0 (None) or 1 (Some). It returns whether a value follows.
func (d *Decoder) ReadCOption() bool {
	p := d.pos
	if d.err != nil || len(d.data)-p < 4 {
		d.fail(ErrUnexpectedEOF, 4)
		return false
	}
	v := binary.LittleEndian.Uint32(d.data[p:])
	if v > 1 {
		d.fail(ErrInvalidTag, int(v))
		return false
	}
	d.pos = p + 4
	return v == 1
}

// ReadCompactU16 reads a Solana compact-u16 ("shortvec") length: a
// little-endian base-128 varint capped at 3 bytes and a maximum value of
// 0xFFFF. Non-minimal encodings are rejected as ErrNonCanonical.
func (d *Decoder) ReadCompactU16() int {
	// Single-byte fast path: almost all real-world lengths are < 0x80.
	p := d.pos
	if d.err == nil && len(d.data)-p >= 1 {
		if b0 := d.data[p]; b0 < 0x80 {
			d.pos = p + 1
			return int(b0)
		}
	}
	return d.readCompactU16Slow()
}

// readCompactU16Slow handles truncation, prior errors and the two- and
// three-byte encodings. Split out so the single-byte fast path in
// ReadCompactU16 stays inlineable.
func (d *Decoder) readCompactU16Slow() int {
	p := d.pos
	if d.err != nil || len(d.data)-p < 1 {
		d.fail(ErrUnexpectedEOF, 1)
		return 0
	}
	b0 := d.data[p]
	if b0 < 0x80 {
		d.pos = p + 1
		return int(b0)
	}
	value := int(b0 & 0x7F)
	rest := d.data[p+1:]
	if len(rest) < 1 {
		d.fail(ErrUnexpectedEOF, 2)
		return 0
	}
	// A zero continuation byte adds nothing, making the encoding
	// non-minimal. The Solana runtime rejects such aliases, and accepting
	// them would let one payload have several encodings.
	b1 := rest[0]
	if b1 == 0 {
		d.fail(ErrNonCanonical, int(b1))
		return 0
	}
	value |= int(b1&0x7F) << 7
	if b1 < 0x80 {
		d.pos = p + 2
		return value
	}
	if len(rest) < 2 {
		d.fail(ErrUnexpectedEOF, 3)
		return 0
	}
	b2 := rest[1]
	if b2 == 0 {
		d.fail(ErrNonCanonical, int(b2))
		return 0
	}
	value |= int(b2&0x7F) << 14
	// A third byte with the continuation bit set, or a value beyond the
	// uint16 range, can only encode out-of-range lengths.
	if b2 >= 0x80 || value > 0xFFFF {
		d.fail(ErrOverflow, value)
		return 0
	}
	d.pos = p + 3
	return value
}

// ReadVarUint64 reads canonical unsigned LEB128. Encodings longer than ten
// bytes, values exceeding uint64, and non-minimal encodings are rejected.
func (d *Decoder) ReadVarUint64() uint64 {
	p := d.pos
	if d.err != nil {
		return 0
	}
	var value uint64
	for index := 0; index < 10; index++ {
		if len(d.data)-p <= index {
			d.fail(ErrUnexpectedEOF, index+1)
			return 0
		}
		current := d.data[p+index]
		if index == 9 && current > 1 {
			d.fail(ErrOverflow, index+1)
			return 0
		}
		value |= uint64(current&0x7f) << (7 * index)
		if current < 0x80 {
			if index > 0 && current == 0 {
				d.fail(ErrNonCanonical, index+1)
				return 0
			}
			d.pos = p + index + 1
			return value
		}
	}
	d.fail(ErrOverflow, 10)
	return 0
}
