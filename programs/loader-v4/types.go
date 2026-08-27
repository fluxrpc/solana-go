package loaderv4

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type instruction struct{ solana.AccountMetaSlice }

func (*instruction) ProgramID() solana.PublicKey          { return ProgramID }
func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

// InstructionType is encoded by bincode as a little-endian uint32, despite
// the Rust enum's repr(u8) annotation.
type InstructionType uint32

func (typ InstructionType) String() string {
	switch typ {
	case WriteInstruction:
		return "Write"
	case CopyInstruction:
		return "Copy"
	case SetProgramLengthInstruction:
		return "SetProgramLength"
	case DeployInstruction:
		return "Deploy"
	case RetractInstruction:
		return "Retract"
	case TransferAuthorityInstruction:
		return "TransferAuthority"
	case FinalizeInstruction:
		return "Finalize"
	default:
		return fmt.Sprintf("InstructionType(%d)", uint32(typ))
	}
}

type DecodedInstruction struct {
	Type              InstructionType
	Write             *Write
	Copy              *Copy
	SetProgramLength  *SetProgramLength
	Deploy            *Deploy
	Retract           *Retract
	TransferAuthority *TransferAuthority
	Finalize          *Finalize
}
