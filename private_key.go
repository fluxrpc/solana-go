package solana_go

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"os"

	"github.com/fluxrpc/base58"
	voied25519 "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
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

// PrivateKeyFromSeed derives the keypair deterministically from a 32-byte
// ed25519 seed.
func PrivateKeyFromSeed(seed []byte) (PrivateKey, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid seed size, expected %d, got %d", ed25519.SeedSize, len(seed))
	}
	return PrivateKey(voied25519.NewKeyFromSeed(seed)), nil
}

// PrivateKeyFromSolanaKeygenFile loads a keypair from a file in the JSON
// byte-array format written by `solana-keygen new` (e.g. `[12,34,...]`,
// 64 values).
func PrivateKeyFromSolanaKeygenFile(file string) (PrivateKey, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read keygen file: %w", err)
	}
	key, err := PrivateKeyFromSolanaKeygenFileBytes(content)
	if err != nil {
		return nil, fmt.Errorf("keygen file %s: %w", file, err)
	}
	return key, nil
}

// PrivateKeyFromSolanaKeygenFileBytes parses the content of a solana-keygen
// keypair file: a JSON array of 64 byte values. The key is validated with
// Validate before being returned.
func PrivateKeyFromSolanaKeygenFileBytes(content []byte) (PrivateKey, error) {
	pos := 0
	skipSpace := func() {
		for pos < len(content) {
			switch content[pos] {
			case ' ', '\t', '\n', '\r':
				pos++
			default:
				return
			}
		}
	}

	skipSpace()
	if pos == len(content) || content[pos] != '[' {
		return nil, errors.New("expected a JSON array of byte values")
	}
	pos++

	key := make(PrivateKey, 0, ed25519.PrivateKeySize)
	skipSpace()
	if pos < len(content) && content[pos] == ']' {
		pos++
	} else {
		for {
			skipSpace()
			start := pos
			value := 0
			for pos < len(content) && content[pos] >= '0' && content[pos] <= '9' {
				value = value*10 + int(content[pos]-'0')
				if value > 255 {
					return nil, fmt.Errorf("byte value out of range at offset %d", start)
				}
				pos++
			}
			if pos == start {
				return nil, fmt.Errorf("expected a byte value at offset %d", pos)
			}
			key = append(key, byte(value))
			skipSpace()
			if pos == len(content) {
				return nil, errors.New("unexpected end of input")
			}
			if content[pos] == ',' {
				pos++
				continue
			}
			if content[pos] == ']' {
				pos++
				break
			}
			return nil, fmt.Errorf("unexpected character %q at offset %d", content[pos], pos)
		}
	}
	skipSpace()
	if pos != len(content) {
		return nil, fmt.Errorf("trailing data at offset %d", pos)
	}

	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length %d, expected %d", len(key), ed25519.PrivateKeySize)
	}
	if err := key.Validate(); err != nil {
		return nil, err
	}
	return key, nil
}

// IsValid reports whether the private key has the expected size.
func (k PrivateKey) IsValid() bool {
	return len(k) == ed25519.PrivateKeySize
}

// Validate checks that the key is 64 bytes and that its public-key half
// matches the key derived from its seed half.
func (k PrivateKey) Validate() error {
	if len(k) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid private key size, expected %d, got %d", ed25519.PrivateKeySize, len(k))
	}
	derived := voied25519.NewKeyFromSeed(k[:ed25519.SeedSize])
	if !bytes.Equal(derived, []byte(k)) {
		if !IsOnCurve(k[ed25519.SeedSize:]) {
			return errors.New("invalid private key: seed/public key mismatch (public key half is not on the ed25519 curve)")
		}
		return errors.New("invalid private key: seed/public key mismatch")
	}
	return nil
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
	return SignatureFromBytes(voied25519.Sign(voied25519.PrivateKey(k), msg)), nil
}

// Bytes returns the raw 64-byte keypair. The returned slice shares the
// key's backing storage.
func (k PrivateKey) Bytes() []byte {
	return k
}

// MarshalJSON implements json.Marshaler, encoding the key as a base58
// JSON string.
func (k PrivateKey) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, len(k)*2+2)
	buf = append(buf, '"')
	buf = base58.AppendEncode(buf, k)
	buf = append(buf, '"')
	return buf, nil
}

// UnmarshalJSON implements json.Unmarshaler, decoding a base58 JSON
// string and validating the key size.
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

// String returns the base58 representation of the keypair. Handle with
// care: this is secret material.
func (k PrivateKey) String() string {
	return base58.Encode(k)
}
