package addresslookuptable

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type instruction struct{ solana.AccountMetaSlice }

func (*instruction) ProgramID() solana.PublicKey          { return ProgramID }
func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

// InstructionType is the little-endian uint32 bincode enum tag at the start
// of every Address Lookup Table instruction.
type InstructionType uint32

func (typ InstructionType) String() string {
	switch typ {
	case CreateLookupTableInstruction:
		return "CreateLookupTable"
	case FreezeLookupTableInstruction:
		return "FreezeLookupTable"
	case ExtendLookupTableInstruction:
		return "ExtendLookupTable"
	case DeactivateLookupTableInstruction:
		return "DeactivateLookupTable"
	case CloseLookupTableInstruction:
		return "CloseLookupTable"
	default:
		return fmt.Sprintf("InstructionType(%d)", uint32(typ))
	}
}

// DecodedInstruction is the fully typed result of DecodeInstruction. Type
// identifies the one non-nil instruction field.
type DecodedInstruction struct {
	Type InstructionType

	CreateLookupTable     *CreateLookupTable
	FreezeLookupTable     *FreezeLookupTable
	ExtendLookupTable     *ExtendLookupTable
	DeactivateLookupTable *DeactivateLookupTable
	CloseLookupTable      *CloseLookupTable
}
