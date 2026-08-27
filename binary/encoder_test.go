package binary

import (
	"bytes"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestEncoderFixedWidth(t *testing.T) {
	e := NewEncoder(make([]byte, 0, 24))
	e.WriteUint8(0xAB)
	e.WriteUint16(0xBEEF)
	e.WriteUint32(0xDEADBEEF)
	e.WriteUint64(0x1122334455667788)
	e.WriteInt64(-1)

	var want []byte
	want = append(want, 0xAB)
	want = append(want, le16(0xBEEF)...)
	want = append(want, le32(0xDEADBEEF)...)
	want = append(want, le64(0x1122334455667788)...)
	want = append(want, le64(^uint64(0))...)
	if err := e.Err(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(e.Bytes(), want) {
		t.Fatalf("Bytes = %x, want %x", e.Bytes(), want)
	}
	if e.Len() != len(want) || e.Pos() != len(want) {
		t.Fatalf("Len/Pos = %d/%d, want %d", e.Len(), e.Pos(), len(want))
	}
}

func TestEncoderTypedValuesAndTags(t *testing.T) {
	var key solana.PublicKey
	copy(key[:], seq(solana.PublicKeyLength))
	var signature solana.Signature
	copy(signature[:], seq(solana.SignatureLength))
	hash := solana.Hash(key)

	e := NewEncoder(make([]byte, 0, 32+64+32+12))
	e.WritePublicKey(key)
	e.WriteSignature(signature)
	e.WriteHash(hash)
	e.WriteBool(false)
	e.WriteBool(true)
	e.WriteOption(false)
	e.WriteOption(true)
	e.WriteCOption(false)
	e.WriteCOption(true)

	var want []byte
	want = append(want, key[:]...)
	want = append(want, signature[:]...)
	want = append(want, hash[:]...)
	want = append(want, 0, 1, 0, 1)
	want = append(want, 0, 0, 0, 0, 1, 0, 0, 0)
	if err := e.Err(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(e.Bytes(), want) {
		t.Fatalf("Bytes = %x, want %x", e.Bytes(), want)
	}
}

func TestEncoderBytesStringAndReset(t *testing.T) {
	prefix := []byte{0xAA, 0xBB}
	e := NewEncoder(prefix[:1])
	e.WriteBytes([]byte{1, 2, 3})
	e.WriteBorshString("hello")
	e.WriteBorshString("")
	e.WriteBincodeString("rust")

	want := []byte{0xAA, 1, 2, 3}
	want = append(want, le32(5)...)
	want = append(want, "hello"...)
	want = append(want, le32(0)...)
	want = append(want, le64(4)...)
	want = append(want, "rust"...)
	if !bytes.Equal(e.Bytes(), want) {
		t.Fatalf("Bytes = %x, want %x", e.Bytes(), want)
	}

	buf := e.Bytes()
	e.Reset(buf[:0])
	e.WriteUint8(9)
	if e.Err() != nil || !bytes.Equal(e.Bytes(), []byte{9}) {
		t.Fatalf("after Reset: Bytes = %x, Err = %v", e.Bytes(), e.Err())
	}
}

func TestEncoderRejectsInvalidBincodeString(t *testing.T) {
	e := NewEncoder(nil)
	e.WriteBincodeString(string([]byte{0xff}))
	if !errors.Is(e.Err(), ErrInvalidUTF8) {
		t.Fatalf("Err = %v", e.Err())
	}
	if len(e.Bytes()) != 0 {
		t.Fatalf("invalid string wrote %x", e.Bytes())
	}
}

func TestEncoderCompactU16(t *testing.T) {
	cases := []struct {
		value int
		want  []byte
	}{
		{0, []byte{0x00}},
		{0x7F, []byte{0x7F}},
		{0x80, []byte{0x80, 0x01}},
		{0x3FFF, []byte{0xFF, 0x7F}},
		{0x4000, []byte{0x80, 0x80, 0x01}},
		{0xFFFF, []byte{0xFF, 0xFF, 0x03}},
	}
	for _, tc := range cases {
		e := NewEncoder(nil)
		e.WriteCompactU16(tc.value)
		if e.Err() != nil || !bytes.Equal(e.Bytes(), tc.want) {
			t.Errorf("WriteCompactU16(%d) = %v, %v; want %v", tc.value, e.Bytes(), e.Err(), tc.want)
		}
		d := NewDecoder(e.Bytes())
		if got := d.ReadCompactU16(); got != tc.value || d.Err() != nil {
			t.Errorf("round trip %d = %d, %v", tc.value, got, d.Err())
		}
	}

	// Exercise every value, including both encoding boundaries.
	e := NewEncoder(make([]byte, 0, 3))
	for value := 0; value <= 0xFFFF; value++ {
		e.Reset(e.Bytes()[:0])
		e.WriteCompactU16(value)
		d := NewDecoder(e.Bytes())
		if got := d.ReadCompactU16(); got != value || d.Err() != nil || d.Remaining() != 0 {
			t.Fatalf("value %d: got %d, decode Err %v, remaining %d", value, got, d.Err(), d.Remaining())
		}
	}
}

func TestEncoderVarUint64(t *testing.T) {
	cases := []struct {
		value uint64
		want  []byte
	}{
		{0, []byte{0x00}},
		{0x7f, []byte{0x7f}},
		{0x80, []byte{0x80, 0x01}},
		{0x4000, []byte{0x80, 0x80, 0x01}},
		{^uint64(0), []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}},
	}
	for _, test := range cases {
		e := NewEncoder(nil)
		e.WriteVarUint64(test.value)
		if err := e.Err(); err != nil {
			t.Fatalf("WriteVarUint64(%d): %v", test.value, err)
		}
		if !bytes.Equal(e.Bytes(), test.want) {
			t.Errorf("WriteVarUint64(%d) = %x, want %x", test.value, e.Bytes(), test.want)
		}
	}
}

func TestEncoderStickyOverflow(t *testing.T) {
	for _, value := range []int{-1, 1 << 16} {
		e := NewEncoder([]byte{0xAA})
		e.WriteCompactU16(value)
		first := e.Err()
		if !errors.Is(first, ErrOverflow) {
			t.Fatalf("WriteCompactU16(%d) Err = %v", value, first)
		}
		if e.Pos() != 1 {
			t.Errorf("failed write changed position to %d", e.Pos())
		}
		e.WriteUint64(1)
		e.WriteBytes([]byte{2, 3})
		if e.Err() != first {
			t.Errorf("first error not preserved: %v", e.Err())
		}
		if !bytes.Equal(e.Bytes(), []byte{0xAA}) {
			t.Errorf("write after error changed buffer: %x", e.Bytes())
		}

		e.Reset(e.Bytes()[:0])
		e.WriteCompactU16(7)
		if e.Err() != nil || !bytes.Equal(e.Bytes(), []byte{7}) {
			t.Errorf("Reset did not clear error: %x, %v", e.Bytes(), e.Err())
		}
	}
}

func TestEncoderDecoderRoundTrip(t *testing.T) {
	var key solana.PublicKey
	copy(key[:], seq(solana.PublicKeyLength))
	var signature solana.Signature
	copy(signature[:], seq(solana.SignatureLength))

	e := NewEncoder(make([]byte, 0, 160))
	e.WriteUint8(7)
	e.WriteUint16(0x1234)
	e.WriteUint32(0x12345678)
	e.WriteUint64(0x123456789ABCDEF0)
	e.WriteInt64(-42)
	e.WriteBool(true)
	e.WriteBytes([]byte{8, 9, 10})
	e.WritePublicKey(key)
	e.WriteSignature(signature)
	e.WriteHash(solana.Hash(key))
	e.WriteBorshString("solana")
	e.WriteOption(false)
	e.WriteCOption(true)
	e.WriteCompactU16(0xFFFF)
	if err := e.Err(); err != nil {
		t.Fatal(err)
	}

	d := NewDecoder(e.Bytes())
	if d.ReadUint8() != 7 || d.ReadUint16() != 0x1234 ||
		d.ReadUint32() != 0x12345678 || d.ReadUint64() != 0x123456789ABCDEF0 ||
		d.ReadInt64() != -42 || !d.ReadBool() ||
		!bytes.Equal(d.ReadBytes(3), []byte{8, 9, 10}) ||
		d.ReadPublicKey() != key || d.ReadSignature() != signature ||
		d.ReadHash() != solana.Hash(key) || d.ReadBorshString() != "solana" ||
		d.ReadOption() || !d.ReadCOption() || d.ReadCompactU16() != 0xFFFF {
		t.Fatal("round-trip value mismatch")
	}
	if err := d.Err(); err != nil {
		t.Fatal(err)
	}
	if d.Remaining() != 0 {
		t.Fatalf("Remaining = %d", d.Remaining())
	}
}

func TestEncoderPreallocatedIsAllocationFree(t *testing.T) {
	var key solana.PublicKey
	buf := make([]byte, 0, 165)
	e := NewEncoder(buf)
	allocs := testing.AllocsPerRun(1000, func() {
		e.Reset(buf[:0])
		e.WritePublicKey(key)
		e.WritePublicKey(key)
		e.WriteUint64(123)
		e.WriteCOption(true)
		e.WritePublicKey(key)
		e.WriteUint8(1)
		e.WriteCOption(false)
		e.WriteUint64(0)
		e.WriteUint64(42)
		e.WriteCOption(false)
		e.WritePublicKey(key)
		buf = e.Bytes()
	})
	if allocs != 0 {
		t.Fatalf("allocations = %v, want 0", allocs)
	}
	if e.Len() != 165 {
		t.Fatalf("encoded length = %d, want 165", e.Len())
	}
}
