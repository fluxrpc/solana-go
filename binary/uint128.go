package binary

import (
	"math"
	"math/big"
)

// Uint128 is a 128-bit unsigned integer stored as two 64-bit halves. Lo holds the low 64 bits,
// Hi the high 64 bits — MarshalWithEncoder/UnmarshalWithDecoder write/read them as two
// consecutive little-endian uint64s, which is byte-identical to Borsh's 16-byte little-endian
// u128 wire format.
type Uint128 struct {
	Lo uint64
	Hi uint64
}

// Uint128FromBigInt builds a Uint128 from an arbitrary-precision integer, truncating to the low
// 128 bits if i is out of range.
func Uint128FromBigInt(i *big.Int) Uint128 {
	lo, hi := bigIntToLoHi(i)
	return Uint128{Lo: lo, Hi: hi}
}

// Uint128FromUint64 widens a uint64 into a Uint128 — the common case, with no big.Int involved.
func Uint128FromUint64(v uint64) Uint128 {
	return Uint128{Lo: v}
}

// BigInt returns v as an arbitrary-precision integer, for arithmetic that doesn't fit in 64 bits.
func (v Uint128) BigInt() *big.Int {
	i := new(big.Int).SetUint64(v.Hi)
	i.Lsh(i, 64)
	return i.Or(i, new(big.Int).SetUint64(v.Lo))
}

func (v Uint128) String() string {
	return v.BigInt().String()
}

func (v Uint128) MarshalWithEncoder(e *Encoder) error {
	e.WriteUint64(v.Lo)
	e.WriteUint64(v.Hi)
	return e.Err()
}

func (v *Uint128) UnmarshalWithDecoder(d *Decoder) error {
	v.Lo = d.ReadUint64()
	v.Hi = d.ReadUint64()
	return d.Err()
}

// Int128 shares Uint128's layout — two's complement means MarshalWithEncoder/UnmarshalWithDecoder
// are identical; only BigInt/String need to interpret the sign bit (Hi's top bit).
type Int128 Uint128

// Int128FromBigInt builds an Int128 from an arbitrary-precision integer, truncating to the low
// 128 bits (as two's complement) if i is out of range.
func Int128FromBigInt(i *big.Int) Int128 {
	lo, hi := bigIntToLoHi(i)
	return Int128{Lo: lo, Hi: hi}
}

// Int128FromInt64 widens an int64 into an Int128, sign-extending into the upper 64 bits.
func Int128FromInt64(v int64) Int128 {
	var hi uint64
	if v < 0 {
		hi = math.MaxUint64
	}
	return Int128{Lo: uint64(v), Hi: hi}
}

func (v Int128) BigInt() *big.Int {
	i := Uint128(v).BigInt()
	if v.Hi&(1<<63) != 0 {
		i.Sub(i, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	return i
}

func (v Int128) String() string {
	return v.BigInt().String()
}

func (v Int128) MarshalWithEncoder(e *Encoder) error {
	e.WriteUint64(v.Lo)
	e.WriteUint64(v.Hi)
	return e.Err()
}

func (v *Int128) UnmarshalWithDecoder(d *Decoder) error {
	v.Lo = d.ReadUint64()
	v.Hi = d.ReadUint64()
	return d.Err()
}

// bigIntToLoHi extracts the low 128 bits of i's two's-complement representation. math/big's
// bitwise ops (And, Rsh) operate on negative numbers as if infinitely sign-extended in two's
// complement, so this is correct for both signed and unsigned callers.
func bigIntToLoHi(i *big.Int) (lo, hi uint64) {
	mask := new(big.Int).SetUint64(math.MaxUint64)
	lo = new(big.Int).And(i, mask).Uint64()
	hi = new(big.Int).And(new(big.Int).Rsh(i, 64), mask).Uint64()
	return
}
