package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// Assign changes the program that owns an account.
type Assign struct {
	// Owner is the program that will own the assigned account.
	Owner solana.PublicKey
	// [0] Assigned account: writable, signer.
	instruction
}

// NewAssignInstruction creates a System Program Assign instruction.
func NewAssignInstruction(owner solana.PublicKey, assignedAccount solana.PublicKey) *Assign {
	return &Assign{
		Owner: owner,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(assignedAccount, true, true),
		}},
	}
}

// Data returns the instruction's binary-encoded data.
func (inst *Assign) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 4+solana.PublicKeyLength))
	enc.WriteUint32(uint32(AssignInstruction))
	enc.WritePublicKey(inst.Owner)
	return enc.Bytes(), enc.Err()
}
