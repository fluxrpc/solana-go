package benchcmp

import (
	"bytes"
	"encoding/binary"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	fluxbin "github.com/fluxrpc/solana-go/binary"
	gaglbin "github.com/gagliardetto/binary"
	gagl "github.com/gagliardetto/solana-go"
)

// ---- Fixtures ----------------------------------------------------------

func binLE32(v uint32) []byte { b := make([]byte, 4); binary.LittleEndian.PutUint32(b, v); return b }
func binLE64(v uint64) []byte { b := make([]byte, 8); binary.LittleEndian.PutUint64(b, v); return b }

// splTokenAccountBytes is a 165-byte SPL token account: the layout decoded
// on the jsonParsed hot path for every token account response.
var splTokenAccountBytes = func() []byte {
	mint := gagl.MPK("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	owner := gagl.MPK("So11111111111111111111111111111111111111112")
	delegate := gagl.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")

	var buf []byte
	buf = append(buf, mint[:]...)
	buf = append(buf, owner[:]...)
	buf = append(buf, binLE64(123_456_789)...) // amount
	buf = append(buf, binLE32(1)...)           // delegate COption: Some
	buf = append(buf, delegate[:]...)
	buf = append(buf, 1)                   // state: Initialized
	buf = append(buf, binLE32(0)...)       // isNative COption: None
	buf = append(buf, binLE64(0)...)       // isNative value
	buf = append(buf, binLE64(42)...)      // delegated amount
	buf = append(buf, binLE32(0)...)       // closeAuthority COption: None
	buf = append(buf, make([]byte, 32)...) // closeAuthority value
	if len(buf) != 165 {
		panic("bad token account fixture")
	}
	return buf
}()

// splMintBytes is an 82-byte SPL mint.
var splMintBytes = func() []byte {
	authority := gagl.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")

	var buf []byte
	buf = append(buf, binLE32(1)...) // mintAuthority COption: Some
	buf = append(buf, authority[:]...)
	buf = append(buf, binLE64(1_000_000_000_000)...) // supply
	buf = append(buf, 6)                             // decimals
	buf = append(buf, 1)                             // isInitialized
	buf = append(buf, binLE32(0)...)                 // freezeAuthority COption: None
	buf = append(buf, make([]byte, 32)...)
	if len(buf) != 82 {
		panic("bad mint fixture")
	}
	return buf
}()

// metadataBytes is a Metaplex token-metadata prefix in Borsh: key,
// update authority, mint, name/symbol/uri strings, seller fee, creators.
var metadataBytes = func() []byte {
	updateAuth := gagl.MPK("So11111111111111111111111111111111111111112")
	mint := gagl.MPK("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
	creator := gagl.MPK("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")

	appendString := func(buf []byte, s string) []byte {
		buf = append(buf, binLE32(uint32(len(s)))...)
		return append(buf, s...)
	}

	var buf []byte
	buf = append(buf, 4) // key: MetadataV1
	buf = append(buf, updateAuth[:]...)
	buf = append(buf, mint[:]...)
	buf = appendString(buf, "Wrapped Solana Token Example")
	buf = appendString(buf, "WSOL")
	buf = appendString(buf, "https://example.com/metadata/wsol.json")
	buf = append(buf, 0xF4, 0x01) // sellerFeeBasisPoints: 500
	buf = append(buf, 1)          // creators Option: Some
	buf = append(buf, binLE32(2)...)
	for i := 0; i < 2; i++ {
		buf = append(buf, creator[:]...)
		buf = append(buf, byte(i)) // verified
		buf = append(buf, 50)      // share
	}
	return buf
}()

// ---- Decoded shapes ----------------------------------------------------

type tokenAccount struct {
	Mint            [32]byte
	Owner           [32]byte
	Amount          uint64
	HasDelegate     bool
	Delegate        [32]byte
	State           uint8
	IsNative        bool
	NativeValue     uint64
	DelegatedAmount uint64
	HasClose        bool
	CloseAuthority  [32]byte
}

type metadataCreator struct {
	Address  [32]byte
	Verified bool
	Share    uint8
}

type metadata struct {
	Key             uint8
	UpdateAuthority [32]byte
	Mint            [32]byte
	Name            string
	Symbol          string
	URI             string
	SellerFee       uint16
	Creators        []metadataCreator
}

// ---- flux decode paths -------------------------------------------------

func fluxDecodeTokenAccount(data []byte) (tokenAccount, error) {
	var out tokenAccount
	d := fluxbin.NewDecoder(data)
	out.Mint = d.ReadPublicKey()
	out.Owner = d.ReadPublicKey()
	out.Amount = d.ReadUint64()
	out.HasDelegate = d.ReadCOption()
	out.Delegate = d.ReadPublicKey()
	out.State = d.ReadUint8()
	out.IsNative = d.ReadCOption()
	out.NativeValue = d.ReadUint64()
	out.DelegatedAmount = d.ReadUint64()
	out.HasClose = d.ReadCOption()
	out.CloseAuthority = d.ReadPublicKey()
	return out, d.Err()
}

func fluxDecodeMetadata(data []byte, copyStrings bool) (metadata, error) {
	var out metadata
	d := fluxbin.NewDecoder(data)
	out.Key = d.ReadUint8()
	out.UpdateAuthority = d.ReadPublicKey()
	out.Mint = d.ReadPublicKey()
	if copyStrings {
		out.Name = d.ReadBorshStringCopy()
		out.Symbol = d.ReadBorshStringCopy()
		out.URI = d.ReadBorshStringCopy()
	} else {
		out.Name = d.ReadBorshString()
		out.Symbol = d.ReadBorshString()
		out.URI = d.ReadBorshString()
	}
	out.SellerFee = d.ReadUint16()
	if d.ReadOption() {
		n := int(d.ReadUint32())
		if d.Err() != nil {
			return out, d.Err()
		}
		out.Creators = make([]metadataCreator, n)
		for i := range out.Creators {
			out.Creators[i].Address = d.ReadPublicKey()
			out.Creators[i].Verified = d.ReadBool()
			out.Creators[i].Share = d.ReadUint8()
		}
	}
	return out, d.Err()
}

// ---- gagl decode paths -------------------------------------------------

// gaglDecodeTokenAccount mirrors token_2022_go's UnmarshalWithDecoder style:
// explicit reads on a bin.Decoder, no reflection.
func gaglDecodeTokenAccount(data []byte) (out tokenAccount, err error) {
	d := gaglbin.NewBinDecoder(data)
	read32 := func(dst *[32]byte) {
		if err != nil {
			return
		}
		var b []byte
		if b, err = d.ReadNBytes(32); err == nil {
			copy(dst[:], b)
		}
	}
	read32(&out.Mint)
	read32(&out.Owner)
	if err == nil {
		out.Amount, err = d.ReadUint64(binary.LittleEndian)
	}
	if err == nil {
		out.HasDelegate, err = d.ReadCOption()
	}
	read32(&out.Delegate)
	if err == nil {
		out.State, err = d.ReadUint8()
	}
	if err == nil {
		out.IsNative, err = d.ReadCOption()
	}
	if err == nil {
		out.NativeValue, err = d.ReadUint64(binary.LittleEndian)
	}
	if err == nil {
		out.DelegatedAmount, err = d.ReadUint64(binary.LittleEndian)
	}
	if err == nil {
		out.HasClose, err = d.ReadCOption()
	}
	read32(&out.CloseAuthority)
	return out, err
}

// gaglDecodeMetadata mirrors dto/json_parsed/spl_metadata.go: explicit reads
// on a Borsh decoder with ReadString.
func gaglDecodeMetadata(data []byte) (out metadata, err error) {
	d := gaglbin.NewBorshDecoder(data)
	if out.Key, err = d.ReadUint8(); err != nil {
		return
	}
	var b []byte
	if b, err = d.ReadNBytes(32); err != nil {
		return
	}
	copy(out.UpdateAuthority[:], b)
	if b, err = d.ReadNBytes(32); err != nil {
		return
	}
	copy(out.Mint[:], b)
	if out.Name, err = d.ReadString(); err != nil {
		return
	}
	if out.Symbol, err = d.ReadString(); err != nil {
		return
	}
	if out.URI, err = d.ReadString(); err != nil {
		return
	}
	if out.SellerFee, err = d.ReadUint16(binary.LittleEndian); err != nil {
		return
	}
	var has bool
	if has, err = d.ReadOption(); err != nil || !has {
		return
	}
	var n uint32
	if n, err = d.ReadUint32(binary.LittleEndian); err != nil {
		return
	}
	out.Creators = make([]metadataCreator, n)
	for i := range out.Creators {
		if b, err = d.ReadNBytes(32); err != nil {
			return
		}
		copy(out.Creators[i].Address[:], b)
		if out.Creators[i].Verified, err = d.ReadBool(); err != nil {
			return
		}
		if out.Creators[i].Share, err = d.ReadUint8(); err != nil {
			return
		}
	}
	return out, nil
}

// ---- Parity ------------------------------------------------------------

// TestBinaryParity decodes each fixture with both libraries and requires
// identical results, so the benchmarks compare equivalent work.
func TestBinaryParity(t *testing.T) {
	fa, err := fluxDecodeTokenAccount(splTokenAccountBytes)
	if err != nil {
		t.Fatal(err)
	}
	ga, err := gaglDecodeTokenAccount(splTokenAccountBytes)
	if err != nil {
		t.Fatal(err)
	}
	if fa != ga {
		t.Errorf("token account mismatch:\nflux %+v\ngagl %+v", fa, ga)
	}

	for _, copyStrings := range []bool{false, true} {
		fm, err := fluxDecodeMetadata(metadataBytes, copyStrings)
		if err != nil {
			t.Fatal(err)
		}
		gm, err := gaglDecodeMetadata(metadataBytes)
		if err != nil {
			t.Fatal(err)
		}
		if fm.Key != gm.Key || fm.UpdateAuthority != gm.UpdateAuthority ||
			fm.Mint != gm.Mint || fm.Name != gm.Name || fm.Symbol != gm.Symbol ||
			fm.URI != gm.URI || fm.SellerFee != gm.SellerFee ||
			len(fm.Creators) != len(gm.Creators) {
			t.Errorf("metadata mismatch (copy=%v):\nflux %+v\ngagl %+v", copyStrings, fm, gm)
		}
		for i := range fm.Creators {
			if fm.Creators[i] != gm.Creators[i] {
				t.Errorf("creator %d mismatch", i)
			}
		}
	}

	// Compact-u16 parity across the full value range.
	for v := 0; v <= 0xFFFF; v++ {
		var enc []byte
		if err := gaglbin.EncodeCompactU16Length(&enc, v); err != nil {
			t.Fatal(err)
		}
		gv, gn, err := gaglbin.DecodeCompactU16(enc)
		if err != nil {
			t.Fatal(err)
		}
		d := fluxbin.NewDecoder(enc)
		fv := d.ReadCompactU16()
		if d.Err() != nil {
			t.Fatalf("value %d: %v", v, d.Err())
		}
		if fv != gv || d.Pos() != gn {
			t.Fatalf("value %d: flux (%d, %d) vs gagl (%d, %d)", v, fv, d.Pos(), gv, gn)
		}
	}

	// Truncated and malformed inputs must fail in both.
	for _, in := range [][]byte{{}, {0x80}, {0x80, 0x80}, {0xFF, 0xFF, 0x04}} {
		if _, _, err := gaglbin.DecodeCompactU16(in); err == nil {
			t.Errorf("gagl accepted %v", in)
		}
		d := fluxbin.NewDecoder(in)
		d.ReadCompactU16()
		if d.Err() == nil {
			t.Errorf("flux accepted %v", in)
		}
	}

	// Zero-copy reads must alias the source buffer.
	src := bytes.Clone(metadataBytes)
	d := fluxbin.NewDecoder(src)
	d.Skip(1 + 32 + 32)
	name := d.ReadBorshString()
	src[1+32+32+4] = 'X'
	if name[0] != 'X' {
		t.Error("ReadBorshString did not alias the buffer")
	}
}

// ---- Benchmarks --------------------------------------------------------

var (
	sinkTokenAccount tokenAccount
	sinkMetadata     metadata
	sinkInt          int
)

func BenchmarkBinary_TokenAccount(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkTokenAccount, sinkErr = fluxDecodeTokenAccount(splTokenAccountBytes)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkTokenAccount, sinkErr = gaglDecodeTokenAccount(splTokenAccountBytes)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkBinary_Metadata(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkMetadata, sinkErr = fluxDecodeMetadata(metadataBytes, false)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkMetadata, sinkErr = gaglDecodeMetadata(metadataBytes)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkBinary_MetadataCopy(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkMetadata, sinkErr = fluxDecodeMetadata(metadataBytes, true)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkMetadata, sinkErr = gaglDecodeMetadata(metadataBytes)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

var (
	compactU16Fixture      = []byte{0xFF, 0xFF, 0x03} // worst case: 3 bytes
	compactU16ShortFixture = []byte{0x05}             // common case: 1 byte
)

func BenchmarkBinary_CompactU16Short(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			d := fluxbin.NewDecoder(compactU16ShortFixture)
			sinkInt = d.ReadCompactU16()
			sinkErr = d.Err()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkInt, _, sinkErr = gaglbin.DecodeCompactU16(compactU16ShortFixture)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkBinary_CompactU16(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			d := fluxbin.NewDecoder(compactU16Fixture)
			sinkInt = d.ReadCompactU16()
			sinkErr = d.Err()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkInt, _, sinkErr = gaglbin.DecodeCompactU16(compactU16Fixture)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

// Ensure the flux sink types stay referenced even if benchmarks change.
var _ = flux.PublicKey{}
