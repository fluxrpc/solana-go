package benchcmp

import (
	"bytes"
	"math/big"
	"testing"

	fluxbin "github.com/fluxrpc/solana-go/binary"
	gaglbin "github.com/gagliardetto/binary"
)

// ---- Fixtures ----------------------------------------------------------

var (
	uint128Fixture      = fluxbin.Uint128{Lo: 0x1122334455667788, Hi: 0x99aabbccddeeff00}
	uint128SmallFixture = fluxbin.Uint128FromUint64(123456789)
	int128Fixture       = fluxbin.Int128{Lo: 0x1122334455667788, Hi: 0x99aabbccddeeff00}

	gaglUint128Fixture      = gaglbin.Uint128{Lo: uint128Fixture.Lo, Hi: uint128Fixture.Hi}
	gaglUint128SmallFixture = gaglbin.Uint128{Lo: uint128SmallFixture.Lo, Hi: uint128SmallFixture.Hi}
	gaglInt128Fixture       = gaglbin.Int128{Lo: int128Fixture.Lo, Hi: int128Fixture.Hi}

	uint128Bytes = func() []byte {
		e := fluxbin.NewEncoder(nil)
		if err := uint128Fixture.MarshalWithEncoder(e); err != nil {
			panic(err)
		}
		return e.Bytes()
	}()
	int128Bytes = func() []byte {
		e := fluxbin.NewEncoder(nil)
		if err := int128Fixture.MarshalWithEncoder(e); err != nil {
			panic(err)
		}
		return e.Bytes()
	}()
)

// ---- Parity --------------------------------------------------------------

// TestUint128Parity confirms flux's and gagl's u128 types agree on wire
// format and BigInt interpretation, so the benchmarks below compare
// equivalent work.
func TestUint128Parity(t *testing.T) {
	fe := fluxbin.NewEncoder(nil)
	if err := uint128Fixture.MarshalWithEncoder(fe); err != nil {
		t.Fatal(err)
	}
	var gbuf bytes.Buffer
	ge := gaglbin.NewBorshEncoder(&gbuf)
	if err := gaglUint128Fixture.MarshalWithEncoder(ge); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fe.Bytes(), gbuf.Bytes()) {
		t.Fatalf("Uint128 wire mismatch: flux % x, gagl % x", fe.Bytes(), gbuf.Bytes())
	}

	fe.Reset(nil)
	if err := int128Fixture.MarshalWithEncoder(fe); err != nil {
		t.Fatal(err)
	}
	gbuf.Reset()
	ge = gaglbin.NewBorshEncoder(&gbuf)
	if err := gaglInt128Fixture.MarshalWithEncoder(ge); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fe.Bytes(), gbuf.Bytes()) {
		t.Fatalf("Int128 wire mismatch: flux % x, gagl % x", fe.Bytes(), gbuf.Bytes())
	}

	if uint128Fixture.BigInt().Cmp(gaglUint128Fixture.BigInt()) != 0 {
		t.Fatalf("Uint128 BigInt mismatch: flux %s, gagl %s", uint128Fixture.BigInt(), gaglUint128Fixture.BigInt())
	}
	if uint128SmallFixture.BigInt().Cmp(gaglUint128SmallFixture.BigInt()) != 0 {
		t.Fatalf("Uint128 (small) BigInt mismatch: flux %s, gagl %s", uint128SmallFixture.BigInt(), gaglUint128SmallFixture.BigInt())
	}
	if int128Fixture.BigInt().Cmp(gaglInt128Fixture.BigInt()) != 0 {
		t.Fatalf("Int128 BigInt mismatch: flux %s, gagl %s", int128Fixture.BigInt(), gaglInt128Fixture.BigInt())
	}
	if int128Fixture.BigInt().Sign() >= 0 {
		t.Fatal("int128Fixture must be negative to exercise two's-complement handling")
	}

	var gotGagl gaglbin.Uint128
	if err := gotGagl.UnmarshalWithDecoder(gaglbin.NewBorshDecoder(uint128Bytes)); err != nil {
		t.Fatal(err)
	}
	if gotGagl != gaglUint128Fixture {
		t.Fatalf("gagl Uint128 decode = %+v, want %+v", gotGagl, gaglUint128Fixture)
	}
}

// ---- Benchmarks ------------------------------------------------------------

var (
	sinkUint128     fluxbin.Uint128
	sinkGaglUint128 gaglbin.Uint128
	sinkInt128      fluxbin.Int128
	sinkGaglInt128  gaglbin.Int128
	sinkBigInt      *big.Int
)

func BenchmarkUint128_Encode(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		dst := make([]byte, 0, 16)
		e := fluxbin.NewEncoder(dst)
		b.ReportAllocs()
		for b.Loop() {
			e.Reset(dst[:0])
			sinkErr = uint128Fixture.MarshalWithEncoder(e)
			dst = e.Bytes()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkBinaryBytes = dst
	})
	b.Run("gagl", func(b *testing.B) {
		var dst bytes.Buffer
		dst.Grow(16)
		e := gaglbin.NewBorshEncoder(&dst)
		b.ReportAllocs()
		for b.Loop() {
			dst.Reset()
			sinkErr = gaglUint128Fixture.MarshalWithEncoder(e)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkBinaryBytes = dst.Bytes()
	})
}

func BenchmarkUint128_Decode(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkErr = sinkUint128.UnmarshalWithDecoder(fluxbin.NewDecoder(uint128Bytes))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkErr = sinkGaglUint128.UnmarshalWithDecoder(gaglbin.NewBorshDecoder(uint128Bytes))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkInt128_Encode(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		dst := make([]byte, 0, 16)
		e := fluxbin.NewEncoder(dst)
		b.ReportAllocs()
		for b.Loop() {
			e.Reset(dst[:0])
			sinkErr = int128Fixture.MarshalWithEncoder(e)
			dst = e.Bytes()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkBinaryBytes = dst
	})
	b.Run("gagl", func(b *testing.B) {
		var dst bytes.Buffer
		dst.Grow(16)
		e := gaglbin.NewBorshEncoder(&dst)
		b.ReportAllocs()
		for b.Loop() {
			dst.Reset()
			sinkErr = gaglInt128Fixture.MarshalWithEncoder(e)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkBinaryBytes = dst.Bytes()
	})
}

func BenchmarkInt128_Decode(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkErr = sinkInt128.UnmarshalWithDecoder(fluxbin.NewDecoder(int128Bytes))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkErr = sinkGaglInt128.UnmarshalWithDecoder(gaglbin.NewBorshDecoder(int128Bytes))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkUint128_BigInt(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBigInt = uint128Fixture.BigInt()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBigInt = gaglUint128Fixture.BigInt()
		}
	})
}

func BenchmarkUint128_BigIntSmall(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBigInt = uint128SmallFixture.BigInt()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBigInt = gaglUint128SmallFixture.BigInt()
		}
	})
}

func BenchmarkInt128_BigInt(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBigInt = int128Fixture.BigInt()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBigInt = gaglInt128Fixture.BigInt()
		}
	})
}

func BenchmarkUint128_String(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = uint128Fixture.String()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = gaglUint128Fixture.String()
		}
	})
}

func BenchmarkUint128_StringSmall(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = uint128SmallFixture.String()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = gaglUint128SmallFixture.String()
		}
	})
}

func BenchmarkInt128_String(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = int128Fixture.String()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkString = gaglInt128Fixture.String()
		}
	})
}
