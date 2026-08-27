package solana_go

import (
	"fmt"

	"github.com/fluxrpc/base58"
)

// Hash is a 32-byte value, such as a block hash.
type Hash PublicKey

// HashFromBase58 decodes a base58 string into a Hash.
func HashFromBase58(in string) (Hash, error) {
	out, err := PublicKeyFromBase58(in)
	if err != nil {
		return Hash{}, err
	}
	return Hash(out), nil
}

// MustHashFromBase58 decodes a base58 string into a Hash.
// Panics on error.
func MustHashFromBase58(in string) Hash {
	out, err := HashFromBase58(in)
	if err != nil {
		panic(err)
	}
	return out
}

// HashFromBytes creates a Hash from a byte slice that must be 32 bytes long.
func HashFromBytes(in []byte) Hash {
	return Hash(PublicKeyFromBytes(in))
}

// IsZero reports whether the hash is all zeros.
func (h Hash) IsZero() bool {
	return h == Hash{}
}

// Equals reports whether two hashes are the same.
func (h Hash) Equals(other Hash) bool {
	return h == other
}

// Bytes returns the hash as a byte slice backed by a copy.
func (h Hash) Bytes() []byte {
	return h[:]
}

// MarshalJSON implements json.Marshaler, encoding the hash as a base58
// JSON string.
func (h Hash) MarshalJSON() ([]byte, error) {
	return PublicKey(h).MarshalJSON()
}

// MarshalText implements encoding.TextMarshaler, encoding the hash as base58.
// This also allows Hash to be used as a JSON object key.
func (h Hash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

// UnmarshalJSON implements json.Unmarshaler, decoding a base58 JSON
// string.
func (h *Hash) UnmarshalJSON(data []byte) error {
	s, err := jsonUnquote(data)
	if err != nil {
		return err
	}

	decoded, err := HashFromBase58(s)
	if err != nil {
		return fmt.Errorf("invalid hash %q: %w", s, err)
	}
	*h = decoded
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler, decoding a base58 hash.
// This also allows Hash to be used as a JSON object key.
func (h *Hash) UnmarshalText(text []byte) error {
	decoded, err := HashFromBase58(unsafeString(text))
	if err != nil {
		return fmt.Errorf("invalid hash %q: %w", text, err)
	}
	*h = decoded
	return nil
}

// String returns the base58 representation of the hash.
func (h Hash) String() string {
	return base58.EncodeCached32((*[32]byte)(&h))
}
