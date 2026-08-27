package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// Allocate allocates data space in an account without funding it.
type Allocate struct {
	Space uint64
	instruction
}

// NewAllocateInstruction creates an Allocate instruction.
func NewAllocateInstruction(space uint64, newAccount solana.PublicKey) *Allocate {
	return &Allocate{
		Space: space,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(newAccount, true, true),
		}},
	}
}

// Data encodes the instruction data.
func (inst *Allocate) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint32(uint32(AllocateInstruction))
	enc.WriteUint64(inst.Space)
	return enc.Bytes(), enc.Err()
}
