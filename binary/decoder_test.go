package binary

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func le16(v uint16) []byte { b := make([]byte, 2); binary.LittleEndian.PutUint16(b, v); return b }
func le32(v uint32) []byte { b := make([]byte, 4); binary.LittleEndian.PutUint32(b, v); return b }
func le64(v uint64) []byte { b := make([]byte, 8); binary.LittleEndian.PutUint64(b, v); return b }

func seq(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i + 1)
	}
	return b
}

func TestReadFixedWidth(t *testing.T) {
	var buf []byte
	buf = append(buf, 0xAB)
	buf = append(buf, le16(0xBEEF)...)
	buf = append(buf, le32(0xDEADBEEF)...)
	buf = append(buf, le64(0x1122334455667788)...)
	buf = append(buf, le64(uint64(0xFFFFFFFFFFFFFFFF))...) // int64 -1

	d := NewDecoder(buf)
	if got := d.ReadUint8(); got != 0xAB {
		t.Errorf("ReadUint8 = %#x", got)
	}
	if got := d.ReadUint16(); got != 0xBEEF {
		t.Errorf("ReadUint16 = %#x", got)
	}
	if got := d.ReadUint32(); got != 0xDEADBEEF {
		t.Errorf("ReadUint32 = %#x", got)
	}
	if got := d.ReadUint64(); got != 0x1122334455667788 {
		t.Errorf("ReadUint64 = %#x", got)
	}
	if got := d.ReadInt64(); got != -1 {
		t.Errorf("ReadInt64 = %d", got)
	}
	if err := d.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}
	if d.Remaining() != 0 || d.Pos() != len(buf) || d.Len() != len(buf) {
		t.Errorf("Remaining/Pos/Len = %d/%d/%d", d.Remaining(), d.Pos(), d.Len())
	}
}

func TestTruncation(t *testing.T) {
	// Each case: minimum buffer length the read needs, and the read itself.
	cases := []struct {
		name string
		need int
		read func(*Decoder)
	}{
		{"Uint8", 1, func(d *Decoder) { d.ReadUint8() }},
		{"Uint16", 2, func(d *Decoder) { d.ReadUint16() }},
		{"Uint32", 4, func(d *Decoder) { d.ReadUint32() }},
		{"Uint64", 8, func(d *Decoder) { d.ReadUint64() }},
		{"Int64", 8, func(d *Decoder) { d.ReadInt64() }},
		{"Bool", 1, func(d *Decoder) { d.ReadBool() }},
		{"Bytes", 5, func(d *Decoder) { d.ReadBytes(5) }},
		{"BytesCopy", 5, func(d *Decoder) { d.ReadBytesCopy(5) }},
		{"PublicKey", 32, func(d *Decoder) { d.ReadPublicKey() }},
		{"Signature", 64, func(d *Decoder) { d.ReadSignature() }},
		{"Hash", 32, func(d *Decoder) { d.ReadHash() }},
		{"COption", 4, func(d *Decoder) { d.ReadCOption() }},
		{"Skip", 3, func(d *Decoder) { d.Skip(3) }},
		{"BorshString", 5, func(d *Decoder) { d.ReadBorshString() }}, // u32 len 1 + 1 byte
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// A full-size zeroed buffer succeeds.
			full := make([]byte, tc.need)
			if tc.name == "BorshString" {
				copy(full, le32(1)) // length 1, then 1 zero byte
			}
			d := NewDecoder(full)
			tc.read(d)
			if err := d.Err(); err != nil {
				t.Fatalf("full buffer: %v", err)
			}
			// Every shorter buffer fails with ErrUnexpectedEOF.
			for n := 0; n < tc.need; n++ {
				d := NewDecoder(full[:n])
				tc.read(d)
				if !errors.Is(d.Err(), ErrUnexpectedEOF) {
					t.Fatalf("len %d: Err = %v, want ErrUnexpectedEOF", n, d.Err())
				}
			}
		})
	}
}

func TestStickyError(t *testing.T) {
	d := NewDecoder([]byte{1, 2})
	d.ReadUint64() // fails
	first := d.Err()
	if !errors.Is(first, ErrUnexpectedEOF) {
		t.Fatalf("Err = %v", first)
	}
	if d.Pos() != 0 {
		t.Errorf("failed read advanced pos to %d", d.Pos())
	}
	// Subsequent reads return zero values and preserve the first error.
	if got := d.ReadUint8(); got != 0 {
		t.Errorf("ReadUint8 after error = %d", got)
	}
	if got := d.ReadBytes(1); got != nil {
		t.Errorf("ReadBytes after error = %v", got)
	}
	if got := d.ReadPublicKey(); !got.IsZero() {
		t.Errorf("ReadPublicKey after error = %v", got)
	}
	if d.Err() != first {
		t.Errorf("first error not preserved: %v", d.Err())
	}
}

func TestReadBytesAliasing(t *testing.T) {
	buf := seq(8)
	d := NewDecoder(buf)

	view := d.ReadBytes(4)
	if !bytes.Equal(view, []byte{1, 2, 3, 4}) {
		t.Fatalf("view = %v", view)
	}
	buf[0] = 0xFF
	if view[0] != 0xFF {
		t.Error("ReadBytes did not alias the input buffer")
	}
	// The view's capacity is clipped: appending must not clobber the buffer.
	_ = append(view, 0xEE)
	if buf[4] != 5 {
		t.Error("append through view clobbered the buffer")
	}

	cp := d.ReadBytesCopy(4)
	if !bytes.Equal(cp, []byte{5, 6, 7, 8}) {
		t.Fatalf("copy = %v", cp)
	}
	buf[4] = 0xFF
	if cp[0] != 5 {
		t.Error("ReadBytesCopy aliases the input buffer")
	}
}

func TestReadBytesNegative(t *testing.T) {
	d := NewDecoder(seq(4))
	if got := d.ReadBytes(-1); got != nil {
		t.Errorf("ReadBytes(-1) = %v", got)
	}
	if !errors.Is(d.Err(), ErrOverflow) {
		t.Errorf("Err = %v, want ErrOverflow", d.Err())
	}

	d = NewDecoder(seq(4))
	d.Skip(-1)
	if !errors.Is(d.Err(), ErrOverflow) {
		t.Errorf("Skip(-1) Err = %v, want ErrOverflow", d.Err())
	}
}

func TestReadTypedKeys(t *testing.T) {
	pk := seq(32)
	sig := seq(64)
	d := NewDecoder(append(append(append([]byte{}, pk...), sig...), pk...))

	if got := d.ReadPublicKey(); !bytes.Equal(got.Bytes(), pk) {
		t.Errorf("ReadPublicKey = %v", got)
	}
	if got := d.ReadSignature(); !bytes.Equal(got[:], sig) {
		t.Errorf("ReadSignature = %v", got)
	}
	if got := d.ReadHash(); !bytes.Equal(got[:], pk) {
		t.Errorf("ReadHash = %v", got)
	}
	if err := d.Err(); err != nil {
		t.Fatal(err)
	}

	// The returned key is a copy, not a view.
	buf := seq(32)
	d = NewDecoder(buf)
	got := d.ReadPublicKey()
	buf[0] = 0xFF
	if got[0] != 1 {
		t.Error("ReadPublicKey aliases the input buffer")
	}
}

func TestBorshString(t *testing.T) {
	var buf []byte
	buf = append(buf, le32(5)...)
	buf = append(buf, "hello"...)
	buf = append(buf, le32(0)...)

	d := NewDecoder(buf)
	if got := d.ReadBorshString(); got != "hello" {
		t.Errorf("ReadBorshString = %q", got)
	}
	if got := d.ReadBorshString(); got != "" {
		t.Errorf("empty ReadBorshString = %q", got)
	}
	if err := d.Err(); err != nil {
		t.Fatal(err)
	}

	d = NewDecoder(buf)
	if got := d.ReadBorshStringCopy(); got != "hello" {
		t.Errorf("ReadBorshStringCopy = %q", got)
	}

	// Length prefix larger than the remaining bytes fails.
	d = NewDecoder(le32(100))
	_ = d.ReadBorshString()
	if !errors.Is(d.Err(), ErrUnexpectedEOF) {
		t.Errorf("oversized length: Err = %v", d.Err())
	}
}

func TestTags(t *testing.T) {
	d := NewDecoder([]byte{0, 1, 2})
	if d.ReadBool() != false || d.ReadBool() != true {
		t.Error("ReadBool values wrong")
	}
	pos := d.Pos()
	d.ReadBool()
	if !errors.Is(d.Err(), ErrInvalidTag) {
		t.Errorf("tag 2: Err = %v", d.Err())
	}
	if d.Pos() != pos {
		t.Errorf("invalid tag advanced pos")
	}

	d = NewDecoder(append(append(le32(0), le32(1)...), le32(7)...))
	if d.ReadCOption() != false || d.ReadCOption() != true {
		t.Error("ReadCOption values wrong")
	}
	pos = d.Pos()
	d.ReadCOption()
	if !errors.Is(d.Err(), ErrInvalidTag) {
		t.Errorf("COption tag 7: Err = %v", d.Err())
	}
	if d.Pos() != pos {
		t.Errorf("invalid COption tag advanced pos")
	}

	// ReadOption is the Borsh single-byte tag.
	d = NewDecoder([]byte{1})
	if d.ReadOption() != true || d.Err() != nil {
		t.Error("ReadOption(1) wrong")
	}
}

func TestCompactU16(t *testing.T) {
	valid := []struct {
		in   []byte
		want int
	}{
		{[]byte{0x00}, 0},
		{[]byte{0x7F}, 0x7F},
		{[]byte{0x80, 0x01}, 0x80},
		{[]byte{0xFF, 0x7F}, 0x3FFF},
		{[]byte{0x80, 0x80, 0x01}, 0x4000},
		{[]byte{0xFF, 0xFF, 0x03}, 0xFFFF},
	}
	for _, tc := range valid {
		d := NewDecoder(tc.in)
		if got := d.ReadCompactU16(); got != tc.want || d.Err() != nil {
			t.Errorf("ReadCompactU16(%v) = %d, %v; want %d", tc.in, got, d.Err(), tc.want)
		}
		if d.Pos() != len(tc.in) {
			t.Errorf("ReadCompactU16(%v) consumed %d bytes", tc.in, d.Pos())
		}
	}

	invalid := []struct {
		in  []byte
		err error
	}{
		{[]byte{}, ErrUnexpectedEOF},
		{[]byte{0x80}, ErrUnexpectedEOF},
		{[]byte{0x80, 0x80}, ErrUnexpectedEOF},
		{[]byte{0x80, 0x00}, ErrNonCanonical},     // zero continuation
		{[]byte{0x80, 0x80, 0x00}, ErrNonCanonical},
		{[]byte{0xFF, 0xFF, 0x04}, ErrOverflow},   // 0x10000
		{[]byte{0x80, 0x80, 0x80}, ErrOverflow},   // third continuation bit
	}
	for _, tc := range invalid {
		d := NewDecoder(tc.in)
		d.ReadCompactU16()
		if !errors.Is(d.Err(), tc.err) {
			t.Errorf("ReadCompactU16(%v) Err = %v, want %v", tc.in, d.Err(), tc.err)
		}
	}
}

// TestDecodeSPLTokenAccount decodes a hand-built 165-byte SPL token account,
// the layout dto/json_parsed and token_2022_go read on the hot path.
func TestDecodeSPLTokenAccount(t *testing.T) {
	mint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	owner := solana.MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
	delegate := solana.MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")

	var buf []byte
	buf = append(buf, mint[:]...)
	buf = append(buf, owner[:]...)
	buf = append(buf, le64(123_456_789)...) // amount
	buf = append(buf, le32(1)...)           // delegate COption: Some
	buf = append(buf, delegate[:]...)
	buf = append(buf, 1)                   // state: Initialized
	buf = append(buf, le32(0)...)          // isNative COption: None
	buf = append(buf, le64(0)...)          // isNative value (still present)
	buf = append(buf, le64(42)...)         // delegated amount
	buf = append(buf, le32(0)...)          // closeAuthority COption: None
	buf = append(buf, make([]byte, 32)...) // closeAuthority value
	if len(buf) != 165 {
		t.Fatalf("fixture length = %d, want 165", len(buf))
	}

	d := NewDecoder(buf)
	gotMint := d.ReadPublicKey()
	gotOwner := d.ReadPublicKey()
	amount := d.ReadUint64()
	hasDelegate := d.ReadCOption()
	gotDelegate := d.ReadPublicKey()
	state := d.ReadUint8()
	isNative := d.ReadCOption()
	d.Skip(8)
	delegated := d.ReadUint64()
	hasClose := d.ReadCOption()
	d.Skip(32)
	if err := d.Err(); err != nil {
		t.Fatal(err)
	}
	if d.Remaining() != 0 {
		t.Errorf("Remaining = %d", d.Remaining())
	}
	if gotMint != mint || gotOwner != owner || amount != 123_456_789 ||
		!hasDelegate || gotDelegate != delegate || state != 1 ||
		isNative || delegated != 42 || hasClose {
		t.Error("decoded fields do not match fixture")
	}
}

// FuzzDecoder drives every read method over arbitrary input and asserts the
// decoder never panics, never reads past the buffer, and only moves forward.
func FuzzDecoder(f *testing.F) {
	f.Add([]byte{}, []byte{0})
	f.Add(seq(200), []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13})
	f.Add(le32(5), []byte{9, 9, 9})
	f.Fuzz(func(t *testing.T, data []byte, script []byte) {
		d := NewDecoder(data)
		for _, op := range script {
			prev := d.Pos()
			errBefore := d.Err()
			switch op % 14 {
			case 0:
				d.ReadUint8()
			case 1:
				d.ReadUint16()
			case 2:
				d.ReadUint32()
			case 3:
				d.ReadUint64()
			case 4:
				d.ReadInt64()
			case 5:
				d.ReadBool()
			case 6:
				d.ReadBytes(int(op))
			case 7:
				d.ReadBytesCopy(int(op) % 7)
			case 8:
				d.ReadPublicKey()
			case 9:
				d.ReadSignature()
			case 10:
				d.ReadBorshString()
			case 11:
				d.ReadCOption()
			case 12:
				d.ReadCompactU16()
			case 13:
				d.Skip(int(op) % 5)
			}
			if d.Pos() < prev || d.Pos() > len(data) {
				t.Fatalf("pos moved from %d to %d (len %d)", prev, d.Pos(), len(data))
			}
			if errBefore != nil && d.Pos() != prev {
				t.Fatalf("read advanced pos despite prior error %v", errBefore)
			}
		}
	})
}
