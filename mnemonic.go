package solana_go

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/sha256"
	"crypto/sha512"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	voied25519 "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

// SolanaDerivationPath is the default BIP-44 derivation path used by Phantom
// and most Solana wallets when deriving a key from a BIP-39 mnemonic:
// m/44'/501'/0'/0'.
const SolanaDerivationPath = "m/44'/501'/0'/0'"

// bip39EnglishRaw is the official BIP-39 English wordlist (2048 words,
// newline separated), embedded so mnemonic validation needs no dependency.
//
//go:embed bip39_english.txt
var bip39EnglishRaw string

var (
	bip39Once  sync.Once
	bip39Index map[string]uint16
)

func bip39WordIndex() map[string]uint16 {
	bip39Once.Do(func() {
		bip39Index = make(map[string]uint16, 2048)
		i := uint16(0)
		for word := range strings.Lines(bip39EnglishRaw) {
			// TrimRight also drops a stray \r from autocrlf checkouts.
			bip39Index[strings.TrimRight(word, "\r\n")] = i
			i++
		}
	})
	return bip39Index
}

// IsMnemonicValid reports whether the mnemonic is a well-formed English
// BIP-39 phrase: 12/15/18/21/24 words from the wordlist with a valid
// checksum.
func IsMnemonicValid(mnemonic string) bool {
	words := strings.Fields(mnemonic)
	switch len(words) {
	case 12, 15, 18, 21, 24:
	default:
		return false
	}
	index := bip39WordIndex()

	// Repack the 11-bit word indices into a bitstring: the leading bits are
	// entropy, the trailing len/3 bits are the checksum.
	totalBits := len(words) * 11
	checksumBits := totalBits / 33
	entropyBits := totalBits - checksumBits
	buf := make([]byte, (totalBits+7)/8)
	for i, word := range words {
		idx, ok := index[word]
		if !ok {
			return false
		}
		for b := 0; b < 11; b++ {
			if idx&(1<<(10-b)) != 0 {
				pos := i*11 + b
				buf[pos/8] |= 1 << (7 - pos%8)
			}
		}
	}

	hash := sha256.Sum256(buf[:entropyBits/8])
	for b := 0; b < checksumBits; b++ {
		pos := entropyBits + b
		if buf[pos/8]>>(7-pos%8)&1 != hash[b/8]>>(7-b%8)&1 {
			return false
		}
	}
	return true
}

// MnemonicToSeed derives the 64-byte BIP-39 seed from a mnemonic and
// passphrase (PBKDF2-HMAC-SHA512, 2048 rounds). The mnemonic is not
// validated; use IsMnemonicValid or the PrivateKeyFromMnemonic helpers for
// that.
func MnemonicToSeed(mnemonic, passphrase string) []byte {
	seed, err := pbkdf2.Key(sha512.New, mnemonic, []byte("mnemonic"+passphrase), 2048, 64)
	if err != nil {
		// Unreachable: pbkdf2.Key only fails on invalid parameters.
		panic(err)
	}
	return seed
}

// PrivateKeyFromMnemonic derives a PrivateKey from an English BIP-39
// mnemonic using the default Solana derivation path (m/44'/501'/0'/0').
// The passphrase may be empty; when set, it must match the passphrase used
// when the mnemonic was generated.
func PrivateKeyFromMnemonic(mnemonic, passphrase string) (PrivateKey, error) {
	return PrivateKeyFromMnemonicAtPath(mnemonic, passphrase, SolanaDerivationPath)
}

// PrivateKeyFromMnemonicAtPath derives a PrivateKey from an English BIP-39
// mnemonic at the given SLIP-0010 derivation path. All path segments must
// be hardened (suffixed with ' or h); SLIP-0010 does not define
// non-hardened derivation for ed25519.
func PrivateKeyFromMnemonicAtPath(mnemonic, passphrase, path string) (PrivateKey, error) {
	if !IsMnemonicValid(mnemonic) {
		return nil, errors.New("invalid mnemonic")
	}
	return PrivateKeyFromSeedAtPath(MnemonicToSeed(mnemonic, passphrase), path)
}

// PrivateKeyFromSeedAtPath derives a PrivateKey from a 16..64 byte seed
// (typically a 64-byte BIP-39 seed) using the given SLIP-0010 derivation
// path. All path segments must be hardened.
func PrivateKeyFromSeedAtPath(seed []byte, path string) (PrivateKey, error) {
	indices, err := parseDerivationPath(path)
	if err != nil {
		return nil, err
	}
	key, chainCode, err := slip10MasterKey(seed)
	if err != nil {
		return nil, err
	}
	for _, index := range indices {
		key, chainCode = slip10ChildKey(key, chainCode, index)
	}
	return PrivateKey(voied25519.NewKeyFromSeed(key)), nil
}

const slip10HardenedOffset = uint32(0x80000000)

// parseDerivationPath parses a SLIP-0010 derivation path of the form
// m/idx'/idx'/... Each index must be hardened (suffixed with ' or h). An
// empty path or "m" returns no indices, meaning the master key is used.
func parseDerivationPath(path string) ([]uint32, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" || trimmed == "m" || trimmed == "/" {
		return nil, nil
	}
	trimmed = strings.TrimPrefix(trimmed, "m")
	trimmed = strings.TrimPrefix(trimmed, "/")
	segments := strings.Split(trimmed, "/")
	indices := make([]uint32, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			return nil, fmt.Errorf("invalid derivation path %q: empty segment", path)
		}
		hardened := false
		switch seg[len(seg)-1] {
		case '\'', 'h', 'H':
			hardened = true
			seg = seg[:len(seg)-1]
		}
		if !hardened {
			return nil, fmt.Errorf("invalid derivation path %q: SLIP-0010 ed25519 requires hardened segments (suffix ' or h)", path)
		}
		n, err := strconv.ParseUint(seg, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid derivation path %q: %w", path, err)
		}
		if n >= uint64(slip10HardenedOffset) {
			return nil, fmt.Errorf("invalid derivation path %q: index %d out of range", path, n)
		}
		indices = append(indices, uint32(n)+slip10HardenedOffset)
	}
	return indices, nil
}

// slip10MasterKey derives the SLIP-0010 master key for ed25519 from a seed,
// per https://github.com/satoshilabs/slips/blob/master/slip-0010.md.
func slip10MasterKey(seed []byte) (key, chainCode []byte, err error) {
	if l := len(seed); l < 16 || l > 64 {
		return nil, nil, fmt.Errorf("invalid seed length %d (want 16..64 bytes)", l)
	}
	mac := hmac.New(sha512.New, []byte("ed25519 seed"))
	mac.Write(seed)
	sum := mac.Sum(nil)
	return sum[:32], sum[32:], nil
}

// slip10ChildKey derives a hardened SLIP-0010 child key for ed25519. The
// caller must ensure index >= slip10HardenedOffset; non-hardened derivation
// is not defined for ed25519.
func slip10ChildKey(parentKey, parentChainCode []byte, index uint32) (key, chainCode []byte) {
	var idx [4]byte
	binary.BigEndian.PutUint32(idx[:], index)
	mac := hmac.New(sha512.New, parentChainCode)
	mac.Write([]byte{0x00})
	mac.Write(parentKey)
	mac.Write(idx[:])
	sum := mac.Sum(nil)
	return sum[:32], sum[32:]
}
