package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// AllocateWithSeed allocates and assigns an account derived from a base and
// seed.
type AllocateWithSeed struct {
	Base  solana.PublicKey
	Seed  string
	Space uint64
	Owner solana.PublicKey
	instruction
}

// NewAllocateWithSeedInstruction creates an AllocateWithSeed instruction.
func NewAllocateWithSeedInstruction(
	base solana.PublicKey,
	seed string,
	space uint64,
	owner solana.PublicKey,
	allocatedAccount solana.PublicKey,
	baseAccount solana.PublicKey,
) (*AllocateWithSeed, error) {
	if len(seed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	return &AllocateWithSeed{
		Base:  base,
		Seed:  seed,
		Space: space,
		Owner: owner,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(allocatedAccount, true, false),
			solana.NewAccountMeta(baseAccount, false, true),
		}},
	}, nil
}

// Data encodes the instruction data.
func (inst *AllocateWithSeed) Data() ([]byte, error) {
	if len(inst.Seed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	enc := binary.NewEncoder(make([]byte, 0, 84+len(inst.Seed)))
	enc.WriteUint32(uint32(AllocateWithSeedInstruction))
	enc.WritePublicKey(inst.Base)
	enc.WriteBincodeString(inst.Seed)
	enc.WriteUint64(inst.Space)
	enc.WritePublicKey(inst.Owner)
	return enc.Bytes(), enc.Err()
}
