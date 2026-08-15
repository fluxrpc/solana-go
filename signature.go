package solana_go

import (
	"fmt"

	"github.com/fluxrpc/base58"
	voied25519 "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

const (
	// SignatureLength is the number of bytes in a signature.
	SignatureLength = 64
)

// Signature is a 64-byte ed25519 transaction signature.
type Signature [SignatureLength]byte

// SignatureFromBase58 decodes a base58 string into a Signature.
func SignatureFromBase58(in string) (out Signature, err error) {
	if err = base58.Decode64(in, (*[64]byte)(&out)); err != nil {
		return Signature{}, fmt.Errorf("decode: %w", err)
	}
	return
}

// MustSignatureFromBase58 decodes a base58 string into a Signature.
// Panics on error.
func MustSignatureFromBase58(in string) Signature {
	out, err := SignatureFromBase58(in)
	if err != nil {
		panic(err)
	}
	return out
}

// SignatureFromBytes copies up to SignatureLength bytes from in into a
// Signature; shorter input leaves the remaining bytes zero.
func SignatureFromBytes(in []byte) (out Signature) {
	copy(out[:], in)
	return
}

// IsZero reports whether the signature is all zeros (a placeholder).
func (s Signature) IsZero() bool {
	return s == Signature{}
}

// Equals reports whether two signatures are the same.
func (s Signature) Equals(other Signature) bool {
	return s == other
}

// Bytes returns the signature as a byte slice backed by a copy.
func (s Signature) Bytes() []byte {
	return s[:]
}

// verifyOptsStdLib makes voi match crypto/ed25519 verification semantics.
var verifyOptsStdLib = &voied25519.Options{
	Verify: voied25519.VerifyOptionsStdLib,
}

// Verify reports whether the signature is valid for the given public key and
// message.
func (s Signature) Verify(pubkey PublicKey, msg []byte) bool {
	return voied25519.VerifyWithOptions(pubkey[:], msg, s[:], verifyOptsStdLib)
}

// MarshalJSON implements json.Marshaler, encoding the signature as a
// base58 JSON string.
func (s Signature) MarshalJSON() ([]byte, error) {
	// Base58 characters never need JSON escaping, so write the quoted string
	// directly instead of going through json.Marshal.
	buf := make([]byte, 0, base58.EncodedMaxLen64+2)
	buf = append(buf, '"')
	buf = base58.AppendEncode64(buf, (*[64]byte)(&s))
	buf = append(buf, '"')
	return buf, nil
}

// UnmarshalJSON implements json.Unmarshaler, decoding a base58 JSON
// string.
func (s *Signature) UnmarshalJSON(data []byte) error {
	str, err := jsonUnquote(data)
	if err != nil {
		return err
	}

	decoded, err := SignatureFromBase58(str)
	if err != nil {
		return fmt.Errorf("invalid signature %q: %w", str, err)
	}
	*s = decoded
	return nil
}

// String returns the base58 representation of the signature.
func (s Signature) String() string {
	return base58.Encode64((*[64]byte)(&s))
}
