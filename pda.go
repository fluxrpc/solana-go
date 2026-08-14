package solana_go

import (
	"crypto/sha256"
	"errors"

	"github.com/oasisprotocol/curve25519-voi/curve"
)

const (
	// MaxSeedLength is the maximum length of a derivation seed.
	MaxSeedLength = 32
	// MaxSeeds is the maximum number of seeds (including the bump seed).
	MaxSeeds = 16

	// PDA_MARKER is appended to the derivation preimage so program-derived
	// addresses can never collide with curve points derived elsewhere.
	PDA_MARKER = "ProgramDerivedAddress"
)

var (
	ErrMaxSeedLengthExceeded = errors.New("max seed length exceeded")
	ErrNoValidProgramAddress = errors.New("unable to find a valid program address")

	errAddressOnCurve = errors.New("invalid seeds; address must fall off the curve")
)

// pdaPreimageSize is the largest possible derivation preimage: all seeds at
// their maximum, the program ID and the marker. Small enough to live on the
// stack, so derivation does not allocate.
const pdaPreimageSize = MaxSeeds*MaxSeedLength + PublicKeyLength + len(PDA_MARKER)

// IsOnCurve reports whether b is a valid compressed ed25519 curve point.
func IsOnCurve(b []byte) bool {
	if len(b) != PublicKeyLength {
		return false
	}
	var compressed curve.CompressedEdwardsY
	if _, err := compressed.SetBytes(b); err != nil {
		return false
	}
	var point curve.EdwardsPoint
	_, err := point.SetCompressedY(&compressed)
	return err == nil
}

// IsOnCurve reports whether the public key is a valid ed25519 curve point.
// Program-derived addresses are never on the curve.
func (p PublicKey) IsOnCurve() bool {
	return IsOnCurve(p[:])
}

// CreateWithSeed derives the address base+seed owned by owner, per
// Pubkey::create_with_seed.
func CreateWithSeed(base PublicKey, seed string, owner PublicKey) (PublicKey, error) {
	if len(seed) > MaxSeedLength {
		return PublicKey{}, ErrMaxSeedLengthExceeded
	}
	var arr [2*PublicKeyLength + MaxSeedLength]byte
	buf := append(arr[:0], base[:]...)
	buf = append(buf, seed...)
	buf = append(buf, owner[:]...)
	return PublicKey(sha256.Sum256(buf)), nil
}

// appendPDAPreimage validates the seeds and assembles the shared prefix of
// the derivation preimage into buf.
func appendPDAPreimage(buf []byte, seeds [][]byte) ([]byte, error) {
	for _, seed := range seeds {
		if len(seed) > MaxSeedLength {
			return nil, ErrMaxSeedLengthExceeded
		}
		buf = append(buf, seed...)
	}
	return buf, nil
}

// CreateProgramAddress derives the program address for the given seeds, per
// Pubkey::create_program_address. It errors if the derived address lands on
// the ed25519 curve.
func CreateProgramAddress(seeds [][]byte, programID PublicKey) (PublicKey, error) {
	if len(seeds) > MaxSeeds {
		return PublicKey{}, ErrMaxSeedLengthExceeded
	}

	var arr [pdaPreimageSize]byte
	buf, err := appendPDAPreimage(arr[:0], seeds)
	if err != nil {
		return PublicKey{}, err
	}
	buf = append(buf, programID[:]...)
	buf = append(buf, PDA_MARKER...)

	hash := sha256.Sum256(buf)
	if IsOnCurve(hash[:]) {
		return PublicKey{}, errAddressOnCurve
	}
	return PublicKey(hash), nil
}

// FindProgramAddress finds the highest bump seed (255 down to 1) producing a
// valid (off-curve) program address for the given seeds, returning the
// address and the bump. The preimage is assembled once and only the bump
// byte changes between attempts, so the search does not allocate.
func FindProgramAddress(seeds [][]byte, programID PublicKey) (PublicKey, uint8, error) {
	if len(seeds) >= MaxSeeds { // the bump seed must also fit
		return PublicKey{}, 0, ErrMaxSeedLengthExceeded
	}

	var arr [pdaPreimageSize]byte
	buf, err := appendPDAPreimage(arr[:0], seeds)
	if err != nil {
		return PublicKey{}, 0, err
	}
	bumpIdx := len(buf)
	buf = append(buf, 0)
	buf = append(buf, programID[:]...)
	buf = append(buf, PDA_MARKER...)

	for bump := 255; bump > 0; bump-- {
		buf[bumpIdx] = byte(bump)
		hash := sha256.Sum256(buf)
		if !IsOnCurve(hash[:]) {
			return PublicKey(hash), uint8(bump), nil
		}
	}
	return PublicKey{}, 0, ErrNoValidProgramAddress
}

// FindAssociatedTokenAddress derives the associated token account of wallet
// for mint under the SPL Token program.
func FindAssociatedTokenAddress(wallet PublicKey, mint PublicKey) (PublicKey, uint8, error) {
	return FindAssociatedTokenAddressWithProgram(wallet, mint, TokenProgramID)
}

// FindAssociatedTokenAddressWithProgram derives the associated token account
// of wallet for mint under the given token program (e.g. Token2022ProgramID).
func FindAssociatedTokenAddressWithProgram(wallet, mint, tokenProgram PublicKey) (PublicKey, uint8, error) {
	return FindProgramAddress(
		[][]byte{wallet[:], tokenProgram[:], mint[:]},
		SPLAssociatedTokenAccountProgramID,
	)
}

// FindTokenMetadataAddress derives the Metaplex token metadata address for
// the given mint.
func FindTokenMetadataAddress(mint PublicKey) (PublicKey, uint8, error) {
	return FindProgramAddress(
		[][]byte{[]byte("metadata"), TokenMetadataProgramID[:], mint[:]},
		TokenMetadataProgramID,
	)
}
