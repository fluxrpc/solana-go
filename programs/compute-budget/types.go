package computebudget

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type instruction struct{}

func (*instruction) ProgramID() solana.PublicKey     { return ProgramID }
func (*instruction) Accounts() []*solana.AccountMeta { return nil }

// InstructionType is the one-byte Borsh enum tag at the start of every
// Compute Budget instruction.
type InstructionType uint8

func (typ InstructionType) String() string {
	switch typ {
	case UnusedInstruction:
		return "Unused"
	case RequestHeapFrameInstruction:
		return "RequestHeapFrame"
	case SetComputeUnitLimitInstruction:
		return "SetComputeUnitLimit"
	case SetComputeUnitPriceInstruction:
		return "SetComputeUnitPrice"
	case SetLoadedAccountsDataSizeLimitInstruction:
		return "SetLoadedAccountsDataSizeLimit"
	default:
		return fmt.Sprintf("InstructionType(%d)", uint8(typ))
	}
}

// DecodedInstruction is the fully typed result of DecodeInstruction. Type
// identifies the one non-nil instruction field.
type DecodedInstruction struct {
	Type InstructionType

	Unused                         *Unused
	RequestUnitsDeprecated         *RequestUnitsDeprecated
	RequestHeapFrame               *RequestHeapFrame
	SetComputeUnitLimit            *SetComputeUnitLimit
	SetComputeUnitPrice            *SetComputeUnitPrice
	SetLoadedAccountsDataSizeLimit *SetLoadedAccountsDataSizeLimit
}
