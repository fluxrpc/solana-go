package solana_go

import (
	"fmt"

	"github.com/fluxrpc/base58"
)

const (
	// PublicKeyLength is the number of bytes in a public key.
	PublicKeyLength = 32

	// A 32-byte value needs at least 32 Base58 characters: every leading zero
	// byte is represented by a leading '1'.
	publicKeyMinEncodedLength = PublicKeyLength
)

type PublicKey [PublicKeyLength]byte

// PublicKeyFromBytes creates a PublicKey from a byte slice that must be 32 bytes long.
// NOTE: it will accept on- and off-curve pubkeys.
func PublicKeyFromBytes(in []byte) (out PublicKey) {
	if len(in) != PublicKeyLength {
		panic(fmt.Errorf("invalid public key size, expected %d, got %d", PublicKeyLength, len(in)))
	}

	copy(out[:], in)
	return
}

// MustPublicKeyFromBase58 parses a Base58-encoded public key and panics if the
// input is invalid.
func MustPublicKeyFromBase58(in string) PublicKey {
	out, err := PublicKeyFromBase58(in)
	if err != nil {
		panic(err)
	}
	return out
}

// PublicKeyFromBase58 creates a PublicKey from a base58 encoded string.
// NOTE: it will accept on- and off-curve pubkeys.
func PublicKeyFromBase58(in string) (out PublicKey, err error) {
	// Reject impossible lengths before invoking either decoder. Besides being
	// cheaper, this bounds the work performed for attacker-controlled input.
	if len(in) < publicKeyMinEncodedLength || len(in) > base58.EncodedMaxLen32 {
		return out, fmt.Errorf(
			"invalid encoded length, expected %d to %d, got %d",
			publicKeyMinEncodedLength,
			base58.EncodedMaxLen32,
			len(in),
		)
	}

	// Decode into a temporary so callers always receive the zero PublicKey on
	// failure, even if the decoder wrote output before detecting an error.
	var decoded PublicKey
	if err = base58.Decode32(in, (*[32]byte)(&decoded)); err != nil {
		// Fall back to the variable-length decoder to produce a more
		// informative error when the input decodes to a wrong-length value.
		val, decErr := base58.Decode(in)
		if decErr != nil {
			return out, fmt.Errorf("decode: %w", decErr)
		}
		if len(val) != PublicKeyLength {
			return out, fmt.Errorf("invalid length, expected %v, got %d", PublicKeyLength, len(val))
		}
		return out, fmt.Errorf("decode: %w", err)
	}
	return decoded, nil
}

func (p PublicKey) MarshalJSON() ([]byte, error) {
	// Write directly into a JSON-quoted buffer. Base58 characters are all ASCII
	// and never contain JSON-escape characters, so we can skip json.Marshal.
	buf := make([]byte, 0, base58.EncodedMaxLen32+2)
	buf = append(buf, '"')
	buf = base58.AppendEncode32(buf, (*[32]byte)(&p))
	buf = append(buf, '"')
	return buf, nil
}

func (p *PublicKey) UnmarshalJSON(data []byte) (err error) {
	s, err := jsonUnquote(data)
	if err != nil {
		return err
	}

	decoded, err := PublicKeyFromBase58(s)
	if err != nil {
		return fmt.Errorf("invalid public key %q: %w", s, err)
	}
	*p = decoded
	return nil
}

func (p PublicKey) Equals(pb PublicKey) bool {
	return p == pb
}

func (p PublicKey) Bytes() []byte {
	return p[:]
}

// IsZero returns whether the public key is zero.
// NOTE: the System Program public key is also zero.
func (p PublicKey) IsZero() bool {
	return p == (PublicKey{})
}

func (p PublicKey) String() string {
	return base58.Encode32((*[32]byte)(&p))
}
