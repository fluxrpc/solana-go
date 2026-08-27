package loaderv3

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
	case InitializeBufferInstruction:
		return "InitializeBuffer"
	case WriteInstruction:
		return "Write"
	case DeployWithMaxDataLenInstruction:
		return "DeployWithMaxDataLen"
	case UpgradeInstruction:
		return "Upgrade"
	case SetAuthorityInstruction:
		return "SetAuthority"
	case CloseInstruction:
		return "Close"
	case ExtendProgramInstruction:
		return "ExtendProgram"
	case SetAuthorityCheckedInstruction:
		return "SetAuthorityChecked"
	default:
		return fmt.Sprintf("InstructionType(%d)", uint32(typ))
	}
}

type DecodedInstruction struct {
	Type                 InstructionType
	InitializeBuffer     *InitializeBuffer
	Write                *Write
	DeployWithMaxDataLen *DeployWithMaxDataLen
	Upgrade              *Upgrade
	SetAuthority         *SetAuthority
	Close                *Close
	ExtendProgram        *ExtendProgram
	SetAuthorityChecked  *SetAuthorityChecked
}
