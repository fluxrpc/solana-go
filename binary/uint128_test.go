package binary

import (
	"bytes"
	"math"
	"math/big"
	"testing"
)

func TestUint128WireFormat(t *testing.T) {
	v := Uint128{Lo: 0x1122334455667788, Hi: 0x99aabbccddeeff00}
	e := NewEncoder(nil)
	if err := v.MarshalWithEncoder(e); err != nil {
		t.Fatal(err)
	}

	var want []byte
	want = append(want, le64(v.Lo)...)
	want = append(want, le64(v.Hi)...)
	if !bytes.Equal(e.Bytes(), want) {
		t.Fatalf("Bytes = % x, want % x", e.Bytes(), want)
	}

	var got Uint128
	if err := got.UnmarshalWithDecoder(NewDecoder(e.Bytes())); err != nil {
		t.Fatal(err)
	}
	if got != v {
		t.Fatalf("round trip = %+v, want %+v", got, v)
	}
}

func TestInt128WireFormat(t *testing.T) {
	// Int128 must round-trip identically to Uint128 for the same bit pattern (two's complement
	// is already correct as raw little-endian bytes).
	v := Int128{Lo: 0x1122334455667788, Hi: 0x99aabbccddeeff00}
	e := NewEncoder(nil)
	if err := v.MarshalWithEncoder(e); err != nil {
		t.Fatal(err)
	}

	var want []byte
	want = append(want, le64(v.Lo)...)
	want = append(want, le64(v.Hi)...)
	if !bytes.Equal(e.Bytes(), want) {
		t.Fatalf("Bytes = % x, want % x", e.Bytes(), want)
	}

	var got Int128
	if err := got.UnmarshalWithDecoder(NewDecoder(e.Bytes())); err != nil {
		t.Fatal(err)
	}
	if got != v {
		t.Fatalf("round trip = %+v, want %+v", got, v)
	}
}

func TestUint128BigIntRoundTrip(t *testing.T) {
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))
	for _, want := range []*big.Int{big.NewInt(0), big.NewInt(1), max} {
		got := Uint128FromBigInt(want).BigInt()
		if got.Cmp(want) != 0 {
			t.Errorf("Uint128FromBigInt(%s).BigInt() = %s, want %s", want, got, want)
		}
		if s := Uint128FromBigInt(want).String(); s != want.String() {
			t.Errorf("Uint128FromBigInt(%s).String() = %s, want %s", want, s, want)
		}
	}
}

func TestInt128BigIntRoundTrip(t *testing.T) {
	min := new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), 127))
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 127), big.NewInt(1))
	for _, want := range []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(-1), min, max} {
		got := Int128FromBigInt(want).BigInt()
		if got.Cmp(want) != 0 {
			t.Errorf("Int128FromBigInt(%s).BigInt() = %s, want %s", want, got, want)
		}
		if s := Int128FromBigInt(want).String(); s != want.String() {
			t.Errorf("Int128FromBigInt(%s).String() = %s, want %s", want, s, want)
		}
	}
}

func TestUint128FromUint64(t *testing.T) {
	for _, v := range []uint64{0, 1, math.MaxUint64} {
		want := new(big.Int).SetUint64(v)
		got := Uint128FromUint64(v).BigInt()
		if got.Cmp(want) != 0 {
			t.Errorf("Uint128FromUint64(%d).BigInt() = %s, want %s", v, got, want)
		}
	}
}

func TestInt128FromInt64(t *testing.T) {
	for _, v := range []int64{0, 1, -1, math.MinInt64, math.MaxInt64} {
		want := big.NewInt(v)
		got := Int128FromInt64(v).BigInt()
		if got.Cmp(want) != 0 {
			t.Errorf("Int128FromInt64(%d).BigInt() = %s, want %s", v, got, want)
		}
	}
}
