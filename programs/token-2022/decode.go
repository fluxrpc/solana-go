package token2022

import (
	"bytes"
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
	token "github.com/fluxrpc/solana-go/programs/token"
)

func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	if len(data) == 0 {
		return DecodedInstruction{}, fmt.Errorf("token-2022 instruction type: %w", bin.ErrUnexpectedEOF)
	}
	if data[0] <= uint8(InitializeMint2Instruction) {
		base, err := token.DecodeInstructionWithProgramID(accounts, data, ProgramID)
		if err != nil {
			return DecodedInstruction{}, err
		}
		return DecodedInstruction{Base: &base}, nil
	}
	if bytes.HasPrefix(data, []byte{210, 225, 30, 162, 88, 184, 77, 141}) {
		dec := bin.NewDecoder(data[8:])
		inst := &InitializeMetadata{extensionInstruction: extensionInstruction{accounts}, Name: dec.ReadBorshString(), Symbol: dec.ReadBorshString(), URI: dec.ReadBorshString()}
		if err := dec.Err(); err != nil {
			return DecodedInstruction{}, fmt.Errorf("decode token-2022 initialize metadata: %w", err)
		}
		return DecodedInstruction{InitializeMetadata: inst}, nil
	}
	dec := bin.NewDecoder(data[1:])
	out := DecodedInstruction{}
	switch InstructionType(data[0]) {
	case GetAccountDataSizeInstruction:
		inst := &GetAccountDataSize{extensionInstruction: extensionInstruction{accounts}}
		for dec.Remaining() >= 2 {
			inst.ExtensionTypes = append(inst.ExtensionTypes, ExtensionType(dec.ReadUint16()))
		}
		out.GetAccountDataSize = inst
	case InitializeImmutableOwnerInstruction:
		out.InitializeImmutableOwner = &InitializeImmutableOwner{extensionInstruction{accounts}}
	case AmountToUIAmountInstruction:
		out.AmountToUIAmount = &AmountToUIAmount{extensionInstruction: extensionInstruction{accounts}, Amount: dec.ReadUint64()}
	case UIAmountToAmountInstruction:
		out.UIAmountToAmount = &UIAmountToAmount{extensionInstruction: extensionInstruction{accounts}, UIAmount: string(dec.ReadBytes(dec.Remaining()))}
	case InitializeMintCloseAuthorityInstruction:
		inst := &InitializeMintCloseAuthority{extensionInstruction: extensionInstruction{accounts}}
		if dec.ReadOption() {
			authority := dec.ReadPublicKey()
			inst.CloseAuthority = &authority
		}
		out.InitializeMintCloseAuthority = inst
	case TransferFeeExtensionInstruction:
		switch sub := dec.ReadUint8(); sub {
		case 0:
			inst := &InitializeTransferFeeConfig{extensionInstruction: extensionInstruction{accounts}}
			hasConfigAuthority := dec.ReadOption()
			authority := dec.ReadPublicKey()
			if hasConfigAuthority {
				inst.TransferFeeConfigAuthority = &authority
			}
			hasWithdrawAuthority := dec.ReadOption()
			withdrawAuthority := dec.ReadPublicKey()
			if hasWithdrawAuthority {
				inst.WithdrawWithheldAuthority = &withdrawAuthority
			}
			inst.TransferFeeBasisPoints = dec.ReadUint16()
			inst.MaximumFee = dec.ReadUint64()
			out.InitializeTransferFeeConfig = inst
		case 1:
			out.TransferCheckedWithFee = &TransferCheckedWithFee{extensionInstruction: extensionInstruction{accounts}, Amount: dec.ReadUint64(), Decimals: dec.ReadUint8(), Fee: dec.ReadUint64()}
		case 2:
			out.WithdrawWithheldTokensFromMint = &WithdrawWithheldTokensFromMint{extensionInstruction{accounts}}
		case 3:
			out.WithdrawWithheldTokensFromAccounts = &WithdrawWithheldTokensFromAccounts{extensionInstruction: extensionInstruction{accounts}, NumTokenAccounts: dec.ReadUint8()}
		case 4:
			out.HarvestWithheldTokensToMint = &HarvestWithheldTokensToMint{extensionInstruction{accounts}}
		case 5:
			out.SetTransferFee = &SetTransferFee{extensionInstruction: extensionInstruction{accounts}, TransferFeeBasisPoints: dec.ReadUint16(), MaximumFee: dec.ReadUint64()}
		default:
			return DecodedInstruction{}, fmt.Errorf("token-2022: unknown transfer-fee instruction: %d", sub)
		}
	case MetadataPointerExtensionInstruction:
		sub := dec.ReadUint8()
		if sub > 1 {
			return DecodedInstruction{}, fmt.Errorf("token-2022: unknown metadata-pointer instruction: %d", sub)
		}
		inst := &InitializeMetadataPointer{SubInstruction: sub, extensionInstruction: extensionInstruction{accounts}}
		if sub == 0 {
			authority := dec.ReadPublicKey()
			if !authority.IsZero() {
				inst.Authority = &authority
			}
		}
		inst.MetadataAddress = dec.ReadPublicKey()
		out.InitializeMetadataPointer = inst
	default:
		return DecodedInstruction{}, fmt.Errorf("token-2022: unknown instruction: %d", data[0])
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode token-2022: %w", err)
	}
	return out, nil
}
