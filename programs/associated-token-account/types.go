package associatedtokenaccount

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type instruction struct {
	solana.AccountMetaSlice
}

func (*instruction) ProgramID() solana.PublicKey { return ProgramID }

func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

// InstructionType is the one-byte tag at the start of every Associated Token
// Account instruction.
type InstructionType uint8

func (typ InstructionType) String() string {
	switch typ {
	case CreateInstruction:
		return "Create"
	case CreateIdempotentInstruction:
		return "CreateIdempotent"
	case RecoverNestedInstruction:
		return "RecoverNested"
	default:
		return fmt.Sprintf("InstructionType(%d)", uint8(typ))
	}
}

// DecodedInstruction is the fully typed result of DecodeInstruction. Type
// identifies the one non-nil instruction field.
type DecodedInstruction struct {
	Type InstructionType

	Create           *Create
	CreateIdempotent *CreateIdempotent
	RecoverNested    *RecoverNested
}
