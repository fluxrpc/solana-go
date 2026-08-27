package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// AssignWithSeed assigns an account derived from a base and seed to a program.
type AssignWithSeed struct {
	Base  solana.PublicKey
	Seed  string
	Owner solana.PublicKey
	instruction
}

// NewAssignWithSeedInstruction creates an AssignWithSeed instruction.
func NewAssignWithSeedInstruction(
	base solana.PublicKey,
	seed string,
	owner solana.PublicKey,
	assignedAccount solana.PublicKey,
	baseAccount solana.PublicKey,
) (*AssignWithSeed, error) {
	if len(seed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	return &AssignWithSeed{
		Base:  base,
		Seed:  seed,
		Owner: owner,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(assignedAccount, true, false),
			solana.NewAccountMeta(baseAccount, false, true),
		}},
	}, nil
}

// Data encodes the instruction data.
func (inst *AssignWithSeed) Data() ([]byte, error) {
	if len(inst.Seed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	enc := binary.NewEncoder(make([]byte, 0, 76+len(inst.Seed)))
	enc.WriteUint32(uint32(AssignWithSeedInstruction))
	enc.WritePublicKey(inst.Base)
	enc.WriteBincodeString(inst.Seed)
	enc.WritePublicKey(inst.Owner)
	return enc.Bytes(), enc.Err()
}
