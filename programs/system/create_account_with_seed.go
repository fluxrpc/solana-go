package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// CreateAccountWithSeed creates and funds an account derived from a base and seed.
type CreateAccountWithSeed struct {
	// Base is the public key used to derive the created account.
	Base solana.PublicKey
	// Seed is the derivation seed. Its byte length cannot exceed solana.MaxSeedLength.
	Seed string
	// Lamports is the number of lamports transferred to the created account.
	Lamports uint64
	// Space is the number of bytes allocated to the created account.
	Space uint64
	// Owner is the program that will own the created account.
	Owner solana.PublicKey
	// [0] Funding account: writable, signer.
	// [1] Created account: writable.
	// [2] Base account: signer.
	instruction
}

// NewCreateAccountWithSeedInstruction creates a System Program
// CreateAccountWithSeed instruction.
func NewCreateAccountWithSeedInstruction(
	base solana.PublicKey,
	seed string,
	lamports uint64,
	space uint64,
	owner solana.PublicKey,
	fundingAccount solana.PublicKey,
	createdAccount solana.PublicKey,
	baseAccount solana.PublicKey,
) (*CreateAccountWithSeed, error) {
	if len(seed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}

	return &CreateAccountWithSeed{
		Base:     base,
		Seed:     seed,
		Lamports: lamports,
		Space:    space,
		Owner:    owner,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(fundingAccount, true, true),
			solana.NewAccountMeta(createdAccount, true, false),
			solana.NewAccountMeta(baseAccount, false, true),
		}},
	}, nil
}

// Data returns the instruction's binary-encoded data.
func (inst *CreateAccountWithSeed) Data() ([]byte, error) {
	if len(inst.Seed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}

	enc := binary.NewEncoder(make([]byte, 0, 4+solana.PublicKeyLength+8+len(inst.Seed)+8+8+solana.PublicKeyLength))
	enc.WriteUint32(uint32(CreateAccountWithSeedInstruction))
	enc.WritePublicKey(inst.Base)
	enc.WriteBincodeString(inst.Seed)
	enc.WriteUint64(inst.Lamports)
	enc.WriteUint64(inst.Space)
	enc.WritePublicKey(inst.Owner)
	return enc.Bytes(), enc.Err()
}
