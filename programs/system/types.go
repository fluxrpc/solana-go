package system

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type instruction struct{ solana.AccountMetaSlice }

func (*instruction) ProgramID() solana.PublicKey          { return ProgramID }
func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

// InstructionType is the little-endian uint32 variant encoded at the start of
// every System Program instruction.
type InstructionType uint32

func (typ InstructionType) String() string {
	switch typ {
	case CreateAccountInstruction:
		return "CreateAccount"
	case AssignInstruction:
		return "Assign"
	case TransferInstruction:
		return "Transfer"
	case CreateAccountWithSeedInstruction:
		return "CreateAccountWithSeed"
	case AdvanceNonceAccountInstruction:
		return "AdvanceNonceAccount"
	case WithdrawNonceAccountInstruction:
		return "WithdrawNonceAccount"
	case InitializeNonceAccountInstruction:
		return "InitializeNonceAccount"
	case AuthorizeNonceAccountInstruction:
		return "AuthorizeNonceAccount"
	case AllocateInstruction:
		return "Allocate"
	case AllocateWithSeedInstruction:
		return "AllocateWithSeed"
	case AssignWithSeedInstruction:
		return "AssignWithSeed"
	case TransferWithSeedInstruction:
		return "TransferWithSeed"
	case UpgradeNonceAccountInstruction:
		return "UpgradeNonceAccount"
	case CreateAccountAllowPrefundInstruction:
		return "CreateAccountAllowPrefund"
	default:
		return fmt.Sprintf("InstructionType(%d)", uint32(typ))
	}
}

// DecodedInstruction is the fully typed result of DecodeInstruction. Type
// identifies the one non-nil instruction field.
type DecodedInstruction struct {
	Type InstructionType

	CreateAccount             *CreateAccount
	Assign                    *Assign
	Transfer                  *Transfer
	CreateAccountWithSeed     *CreateAccountWithSeed
	AdvanceNonceAccount       *AdvanceNonceAccount
	WithdrawNonceAccount      *WithdrawNonceAccount
	InitializeNonceAccount    *InitializeNonceAccount
	AuthorizeNonceAccount     *AuthorizeNonceAccount
	Allocate                  *Allocate
	AllocateWithSeed          *AllocateWithSeed
	AssignWithSeed            *AssignWithSeed
	TransferWithSeed          *TransferWithSeed
	UpgradeNonceAccount       *UpgradeNonceAccount
	CreateAccountAllowPrefund *CreateAccountAllowPrefund
}
