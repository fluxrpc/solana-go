package loaderv2

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type instruction struct{ solana.AccountMetaSlice }

func (*instruction) ProgramID() solana.PublicKey          { return ProgramID }
func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

type InstructionType uint32

func (typ InstructionType) String() string {
	switch typ {
	case WriteInstruction:
		return "Write"
	case FinalizeInstruction:
		return "Finalize"
	default:
		return fmt.Sprintf("InstructionType(%d)", uint32(typ))
	}
}

type DecodedInstruction struct {
	Type     InstructionType
	Write    *Write
	Finalize *Finalize
}
