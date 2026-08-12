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

func (h Hash) IsZero() bool {
	return h == Hash{}
}

func (h Hash) Equals(other Hash) bool {
	return h == other
}

func (h Hash) Bytes() []byte {
	return h[:]
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return PublicKey(h).MarshalJSON()
}

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

func (h Hash) String() string {
	return base58.Encode32((*[32]byte)(&h))
}
