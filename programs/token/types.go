package token

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

type instruction struct {
	solana.AccountMetaSlice
	programID solana.PublicKey
}

func (inst *instruction) ProgramID() solana.PublicKey {
	if inst.programID.IsZero() {
		return ProgramID
	}
	return inst.programID
}

func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

// SetProgramID changes this instruction's program without mutating package state.
func (inst *instruction) SetProgramID(programID solana.PublicKey) { inst.programID = programID }

type InstructionType uint8

func (typ InstructionType) String() string {
	names := [...]string{
		"InitializeMint", "InitializeAccount", "InitializeMultisig", "Transfer",
		"Approve", "Revoke", "SetAuthority", "MintTo", "Burn", "CloseAccount",
		"FreezeAccount", "ThawAccount", "TransferChecked", "ApproveChecked",
		"MintToChecked", "BurnChecked", "InitializeAccount2", "SyncNative",
		"InitializeAccount3", "InitializeMultisig2", "InitializeMint2",
	}
	if int(typ) < len(names) {
		return names[typ]
	}
	return fmt.Sprintf("InstructionType(%d)", uint8(typ))
}

type AuthorityType uint8

const (
	AuthorityMintTokens AuthorityType = iota
	AuthorityFreezeAccount
	AuthorityAccountOwner
	AuthorityCloseAccount
)

type AccountState uint8

const (
	Uninitialized AccountState = iota
	Initialized
	Frozen
)

func authorityAccounts(authority solana.PublicKey, signers []solana.PublicKey) solana.AccountMetaSlice {
	accounts := make(solana.AccountMetaSlice, 1, 1+len(signers))
	accounts[0] = solana.NewAccountMeta(authority, false, len(signers) == 0)
	for _, signer := range signers {
		accounts = append(accounts, solana.NewAccountMeta(signer, false, true))
	}
	return accounts
}

func appendAuthority(accounts solana.AccountMetaSlice, authority solana.PublicKey, signers []solana.PublicKey) solana.AccountMetaSlice {
	return append(accounts, authorityAccounts(authority, signers)...)
}

// DecodedInstruction is the typed result of decoding one Token instruction.
type DecodedInstruction struct {
	Type                InstructionType
	InitializeMint      *InitializeMint
	InitializeAccount   *InitializeAccount
	InitializeMultisig  *InitializeMultisig
	Transfer            *Transfer
	Approve             *Approve
	Revoke              *Revoke
	SetAuthority        *SetAuthority
	MintTo              *MintTo
	Burn                *Burn
	CloseAccount        *CloseAccount
	FreezeAccount       *FreezeAccount
	ThawAccount         *ThawAccount
	TransferChecked     *TransferChecked
	ApproveChecked      *ApproveChecked
	MintToChecked       *MintToChecked
	BurnChecked         *BurnChecked
	InitializeAccount2  *InitializeAccount2
	SyncNative          *SyncNative
	InitializeAccount3  *InitializeAccount3
	InitializeMultisig2 *InitializeMultisig2
	InitializeMint2     *InitializeMint2
}

func (out *DecodedInstruction) SetProgramID(programID solana.PublicKey) {
	switch out.Type {
	case InitializeMintInstruction:
		out.InitializeMint.SetProgramID(programID)
	case InitializeAccountInstruction:
		out.InitializeAccount.SetProgramID(programID)
	case InitializeMultisigInstruction:
		out.InitializeMultisig.SetProgramID(programID)
	case TransferInstruction:
		out.Transfer.SetProgramID(programID)
	case ApproveInstruction:
		out.Approve.SetProgramID(programID)
	case RevokeInstruction:
		out.Revoke.SetProgramID(programID)
	case SetAuthorityInstruction:
		out.SetAuthority.SetProgramID(programID)
	case MintToInstruction:
		out.MintTo.SetProgramID(programID)
	case BurnInstruction:
		out.Burn.SetProgramID(programID)
	case CloseAccountInstruction:
		out.CloseAccount.SetProgramID(programID)
	case FreezeAccountInstruction:
		out.FreezeAccount.SetProgramID(programID)
	case ThawAccountInstruction:
		out.ThawAccount.SetProgramID(programID)
	case TransferCheckedInstruction:
		out.TransferChecked.SetProgramID(programID)
	case ApproveCheckedInstruction:
		out.ApproveChecked.SetProgramID(programID)
	case MintToCheckedInstruction:
		out.MintToChecked.SetProgramID(programID)
	case BurnCheckedInstruction:
		out.BurnChecked.SetProgramID(programID)
	case InitializeAccount2Instruction:
		out.InitializeAccount2.SetProgramID(programID)
	case SyncNativeInstruction:
		out.SyncNative.SetProgramID(programID)
	case InitializeAccount3Instruction:
		out.InitializeAccount3.SetProgramID(programID)
	case InitializeMultisig2Instruction:
		out.InitializeMultisig2.SetProgramID(programID)
	case InitializeMint2Instruction:
		out.InitializeMint2.SetProgramID(programID)
	}
}
