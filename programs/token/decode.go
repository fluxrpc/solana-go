package token

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	return DecodeInstructionWithProgramID(accounts, data, ProgramID)
}

// DecodeInstructionWithProgramID decodes the common Token instruction set for
// the original program, Token-2022, or a compatible deployment.
func DecodeInstructionWithProgramID(accounts solana.AccountMetaSlice, data []byte, programID solana.PublicKey) (DecodedInstruction, error) {
	dec := bin.NewDecoder(data)
	typ := InstructionType(dec.ReadUint8())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("token instruction type: %w", err)
	}
	base := instruction{AccountMetaSlice: accounts, programID: programID}
	out := DecodedInstruction{Type: typ}
	switch typ {
	case InitializeMintInstruction, InitializeMint2Instruction:
		value := InitializeMint{instruction: base, Decimals: dec.ReadUint8(), MintAuthority: dec.ReadPublicKey()}
		if dec.ReadOption() {
			authority := dec.ReadPublicKey()
			value.FreezeAuthority = &authority
		}
		if typ == InitializeMintInstruction {
			out.InitializeMint = &value
		} else {
			out.InitializeMint2 = &InitializeMint2{InitializeMint: value}
		}
	case InitializeAccountInstruction:
		out.InitializeAccount = &InitializeAccount{instruction: base}
	case InitializeMultisigInstruction:
		out.InitializeMultisig = &InitializeMultisig{instruction: base, M: dec.ReadUint8()}
	case TransferInstruction:
		out.Transfer = &Transfer{amountInstruction{instruction: base, Amount: dec.ReadUint64()}}
	case ApproveInstruction:
		out.Approve = &Approve{amountInstruction{instruction: base, Amount: dec.ReadUint64()}}
	case RevokeInstruction:
		out.Revoke = &Revoke{instruction: base}
	case SetAuthorityInstruction:
		value := SetAuthority{instruction: base, AuthorityType: AuthorityType(dec.ReadUint8())}
		if dec.ReadOption() {
			authority := dec.ReadPublicKey()
			value.NewAuthority = &authority
		}
		out.SetAuthority = &value
	case MintToInstruction:
		out.MintTo = &MintTo{amountInstruction{instruction: base, Amount: dec.ReadUint64()}}
	case BurnInstruction:
		out.Burn = &Burn{amountInstruction{instruction: base, Amount: dec.ReadUint64()}}
	case CloseAccountInstruction:
		out.CloseAccount = &CloseAccount{instruction: base}
	case FreezeAccountInstruction:
		out.FreezeAccount = &FreezeAccount{instruction: base}
	case ThawAccountInstruction:
		out.ThawAccount = &ThawAccount{instruction: base}
	case TransferCheckedInstruction:
		out.TransferChecked = &TransferChecked{checkedInstruction{instruction: base, Amount: dec.ReadUint64(), Decimals: dec.ReadUint8()}}
	case ApproveCheckedInstruction:
		out.ApproveChecked = &ApproveChecked{checkedInstruction{instruction: base, Amount: dec.ReadUint64(), Decimals: dec.ReadUint8()}}
	case MintToCheckedInstruction:
		out.MintToChecked = &MintToChecked{checkedInstruction{instruction: base, Amount: dec.ReadUint64(), Decimals: dec.ReadUint8()}}
	case BurnCheckedInstruction:
		out.BurnChecked = &BurnChecked{checkedInstruction{instruction: base, Amount: dec.ReadUint64(), Decimals: dec.ReadUint8()}}
	case InitializeAccount2Instruction:
		out.InitializeAccount2 = &InitializeAccount2{instruction: base, Owner: dec.ReadPublicKey()}
	case SyncNativeInstruction:
		out.SyncNative = &SyncNative{instruction: base}
	case InitializeAccount3Instruction:
		out.InitializeAccount3 = &InitializeAccount3{instruction: base, Owner: dec.ReadPublicKey()}
	case InitializeMultisig2Instruction:
		out.InitializeMultisig2 = &InitializeMultisig2{InitializeMultisig: InitializeMultisig{instruction: base, M: dec.ReadUint8()}}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, uint8(typ))
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode token %s: %w", typ, err)
	}
	return out, nil
}
