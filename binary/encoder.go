package binary

import (
	"encoding/binary"
	"fmt"
	"math"

	solana "github.com/fluxrpc/solana-go"
)

// Encoder appends little-endian and Borsh values to a byte slice.
//
// The zero value is ready to use. NewEncoder is useful when appending to an
// existing buffer or a preallocated destination. Writes are allocation-free
// while that destination has enough capacity.
//
// Errors are sticky, like Decoder errors: the first failed write records the
// error and subsequent writes do nothing. Check Err once after writing all
// fields. An Encoder must not be used from multiple goroutines concurrently.
type Encoder struct {
	data []byte

	err     error
	failPos int
	failArg int
}

// NewEncoder returns an Encoder that appends to dst. Existing bytes in dst are
// retained. Pass make([]byte, 0, size) to encode into a preallocated buffer.
func NewEncoder(dst []byte) *Encoder {
	return &Encoder{data: dst}
}

// Reset clears the Encoder's error and makes it append to dst. Existing bytes
// in dst are retained, making Reset(buf[:0]) convenient for buffer reuse.
func (e *Encoder) Reset(dst []byte) {
	e.data = dst
	e.err = nil
}

// Bytes returns the encoded buffer. The returned slice aliases the Encoder's
// storage and remains valid until the next write reallocates it.
func (e *Encoder) Bytes() []byte { return e.data }

// Len returns the current encoded buffer length.
func (e *Encoder) Len() int { return len(e.data) }

// Pos returns the offset at which the next byte will be written.
func (e *Encoder) Pos() int { return len(e.data) }

// Err returns the first error encountered by any write, or nil. The returned
// error wraps a package sentinel and adds offset context.
func (e *Encoder) Err() error {
	if e.err == nil {
		return nil
	}
	if e.err == ErrOverflow {
		e.err = fmt.Errorf("length %d at offset %d: %w", e.failArg, e.failPos, ErrOverflow)
	}
	return e.err
}

// fail records the first failure without changing the output buffer.
func (e *Encoder) fail(sentinel error, arg int) {
	if e.err == nil {
		e.err = sentinel
		e.failPos = len(e.data)
		e.failArg = arg
	}
}

// WriteUint8 writes one byte.
func (e *Encoder) WriteUint8(v uint8) {
	if e.err != nil {
		return
	}
	e.data = append(e.data, v)
}

// WriteUint16 writes a little-endian uint16.
func (e *Encoder) WriteUint16(v uint16) {
	if e.err != nil {
		return
	}
	e.data = binary.LittleEndian.AppendUint16(e.data, v)
}

// WriteUint32 writes a little-endian uint32.
func (e *Encoder) WriteUint32(v uint32) {
	if e.err != nil {
		return
	}
	e.data = binary.LittleEndian.AppendUint32(e.data, v)
}

// WriteUint64 writes a little-endian uint64.
func (e *Encoder) WriteUint64(v uint64) {
	if e.err != nil {
		return
	}
	e.data = binary.LittleEndian.AppendUint64(e.data, v)
}

// WriteInt64 writes a little-endian int64 (Borsh i64, e.g. Unix timestamps).
func (e *Encoder) WriteInt64(v int64) {
	e.WriteUint64(uint64(v))
}

// WriteBool writes a Borsh bool as 0 or 1.
func (e *Encoder) WriteBool(v bool) {
	if e.err != nil {
		return
	}
	if v {
		e.data = append(e.data, 1)
	} else {
		e.data = append(e.data, 0)
	}
}

// WriteBytes appends raw bytes without a length prefix.
func (e *Encoder) WriteBytes(v []byte) {
	if e.err != nil {
		return
	}
	e.data = append(e.data, v...)
}

// WritePublicKey writes a 32-byte public key.
func (e *Encoder) WritePublicKey(v solana.PublicKey) {
	if e.err != nil {
		return
	}
	e.data = append(e.data, v[:]...)
}

// WriteSignature writes a 64-byte signature.
func (e *Encoder) WriteSignature(v solana.Signature) {
	if e.err != nil {
		return
	}
	e.data = append(e.data, v[:]...)
}

// WriteHash writes a 32-byte hash.
func (e *Encoder) WriteHash(v solana.Hash) {
	if e.err != nil {
		return
	}
	e.data = append(e.data, v[:]...)
}

// WriteBorshString writes a uint32 byte length followed by the string bytes.
func (e *Encoder) WriteBorshString(v string) {
	if e.err != nil {
		return
	}
	if uint64(len(v)) > math.MaxUint32 {
		e.fail(ErrOverflow, len(v))
		return
	}
	e.data = binary.LittleEndian.AppendUint32(e.data, uint32(len(v)))
	e.data = append(e.data, v...)
}

// WriteOption writes a Borsh Option tag: one byte containing 0 (None) or 1
// (Some).
func (e *Encoder) WriteOption(v bool) {
	e.WriteBool(v)
}

// WriteCOption writes an SPL Token COption tag: a little-endian uint32
// containing 0 (None) or 1 (Some).
func (e *Encoder) WriteCOption(v bool) {
	if e.err != nil {
		return
	}
	if v {
		e.data = append(e.data, 1, 0, 0, 0)
	} else {
		e.data = append(e.data, 0, 0, 0, 0)
	}
}

// WriteCompactU16 writes a canonical Solana compact-u16 ("shortvec") value.
// Values outside the uint16 range record ErrOverflow and leave the buffer
// unchanged.
func (e *Encoder) WriteCompactU16(v int) {
	if e.err != nil {
		return
	}
	if v < 0 || v > math.MaxUint16 {
		e.fail(ErrOverflow, v)
		return
	}
	if v < 1<<7 {
		e.data = append(e.data, byte(v))
		return
	}
	if v < 1<<14 {
		e.data = append(e.data, byte(v)|0x80, byte(v>>7))
		return
	}
	e.data = append(e.data, byte(v)|0x80, byte(v>>7)|0x80, byte(v>>14))
}
