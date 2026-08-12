package solana_go

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/fluxrpc/base58"
)

// PrivateKey is a 64-byte ed25519 private key (seed followed by public key),
// the format used by Solana keypairs.
type PrivateKey []byte

// NewRandomPrivateKey generates a new random private key.
func NewRandomPrivateKey() (PrivateKey, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return PrivateKey(priv), nil
}

// PrivateKeyFromBase58 decodes a base58 string into a PrivateKey.
func PrivateKeyFromBase58(in string) (PrivateKey, error) {
	out, err := base58.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if len(out) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size, expected %d, got %d", ed25519.PrivateKeySize, len(out))
	}
	return out, nil
}

// MustPrivateKeyFromBase58 decodes a base58 string into a PrivateKey.
// Panics on error.
func MustPrivateKeyFromBase58(in string) PrivateKey {
	out, err := PrivateKeyFromBase58(in)
	if err != nil {
		panic(err)
	}
	return out
}

// IsValid reports whether the private key has the expected size.
func (k PrivateKey) IsValid() bool {
	return len(k) == ed25519.PrivateKeySize
}

// PublicKey returns the public key half of the keypair.
// Panics if the private key is not valid.
func (k PrivateKey) PublicKey() (out PublicKey) {
	if !k.IsValid() {
		panic(fmt.Errorf("invalid private key size, expected %d, got %d", ed25519.PrivateKeySize, len(k)))
	}

	copy(out[:], k[ed25519.SeedSize:])
	return
}

// Sign signs the given message with the private key.
func (k PrivateKey) Sign(msg []byte) (Signature, error) {
	if !k.IsValid() {
		return Signature{}, fmt.Errorf("invalid private key size, expected %d, got %d", ed25519.PrivateKeySize, len(k))
	}
	return SignatureFromBytes(ed25519.Sign(ed25519.PrivateKey(k), msg)), nil
}

func (k PrivateKey) Bytes() []byte {
	return k
}

func (k PrivateKey) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, len(k)*2+2)
	buf = append(buf, '"')
	buf = base58.AppendEncode(buf, k)
	buf = append(buf, '"')
	return buf, nil
}

func (k *PrivateKey) UnmarshalJSON(data []byte) error {
	s, err := jsonUnquote(data)
	if err != nil {
		return err
	}

	decoded, err := PrivateKeyFromBase58(s)
	if err != nil {
		return fmt.Errorf("invalid private key %q: %w", s, err)
	}
	*k = decoded
	return nil
}

func (k PrivateKey) String() string {
	return base58.Encode(k)
}
