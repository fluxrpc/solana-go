package tokenmetadata

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type RawInstruction struct {
	instruction
	DataBytes []byte
}

func (inst *RawInstruction) Data() ([]byte, error) { return inst.DataBytes, nil }

type DecodedInstruction struct {
	Type                              InstructionType
	CreateMetadataAccount             *CreateMetadataAccount
	UpdateMetadataAccount             *UpdateMetadataAccount
	UpdatePrimarySaleHappenedViaToken *UpdatePrimarySaleHappenedViaToken
	SignMetadata                      *SignMetadata
	CreateMasterEdition               *CreateMasterEdition
	PuffMetadata                      *PuffMetadata
	UpdateMetadataAccountV2           *UpdateMetadataAccountV2
	CreateMetadataAccountV2           *CreateMetadataAccountV2
	CreateMasterEditionV3             *CreateMasterEditionV3
	VerifyCollection                  *VerifyCollection
	UnverifyCollection                *UnverifyCollection
	Raw                               *RawInstruction
}

func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := bin.NewDecoder(data)
	typ := InstructionType(dec.ReadUint8())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("token metadata instruction type: %w", err)
	}
	if typ > RevokeCollectionAuthorityInstruction {
		return DecodedInstruction{}, fmt.Errorf("token metadata: unknown instruction: %d", typ)
	}
	base := instruction{typ: typ, AccountMetaSlice: accounts}
	out := DecodedInstruction{Type: typ}
	switch typ {
	case CreateMetadataAccountInstruction:
		inst := &CreateMetadataAccount{instruction: base}
		inst.Args.Data.decode(dec)
		inst.Args.IsMutable = dec.ReadBool()
		out.CreateMetadataAccount = inst
	case UpdateMetadataAccountInstruction:
		inst := &UpdateMetadataAccount{instruction: base}
		decodeUpdateArgs(dec, &inst.Args)
		out.UpdateMetadataAccount = inst
	case UpdatePrimarySaleHappenedViaTokenInstruction:
		out.UpdatePrimarySaleHappenedViaToken = &UpdatePrimarySaleHappenedViaToken{instruction: base}
	case SignMetadataInstruction:
		out.SignMetadata = &SignMetadata{instruction: base}
	case CreateMasterEditionInstruction:
		inst := &CreateMasterEdition{instruction: base}
		inst.Args.MaxSupply = readOptionalUint64(dec)
		out.CreateMasterEdition = inst
	case PuffMetadataInstruction:
		out.PuffMetadata = &PuffMetadata{instruction: base}
	case UpdateMetadataAccountV2Instruction:
		inst := &UpdateMetadataAccountV2{instruction: base}
		decodeUpdateArgsV2(dec, &inst.Args)
		out.UpdateMetadataAccountV2 = inst
	case CreateMetadataAccountV2Instruction:
		inst := &CreateMetadataAccountV2{instruction: base}
		inst.Args.Data.decode(dec)
		inst.Args.IsMutable = dec.ReadBool()
		out.CreateMetadataAccountV2 = inst
	case CreateMasterEditionV3Instruction:
		inst := &CreateMasterEditionV3{instruction: base}
		inst.Args.MaxSupply = readOptionalUint64(dec)
		out.CreateMasterEditionV3 = inst
	case VerifyCollectionInstruction:
		out.VerifyCollection = &VerifyCollection{instruction: base}
	case UnverifyCollectionInstruction:
		out.UnverifyCollection = &UnverifyCollection{instruction: base}
	default:
		out.Raw = &RawInstruction{instruction: base, DataBytes: data}
		return out, nil
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode token metadata %s: %w", typ, err)
	}
	return out, nil
}

func decodeUpdateArgs(dec *bin.Decoder, args *UpdateMetadataAccountArgs) {
	if dec.ReadOption() {
		value := new(Data)
		value.decode(dec)
		args.Data = value
	}
	if dec.ReadOption() {
		value := dec.ReadPublicKey()
		args.UpdateAuthority = &value
	}
	if dec.ReadOption() {
		value := dec.ReadBool()
		args.PrimarySaleHappened = &value
	}
}

func decodeUpdateArgsV2(dec *bin.Decoder, args *UpdateMetadataAccountArgsV2) {
	if dec.ReadOption() {
		value := new(DataV2)
		value.decode(dec)
		args.Data = value
	}
	if dec.ReadOption() {
		value := dec.ReadPublicKey()
		args.UpdateAuthority = &value
	}
	if dec.ReadOption() {
		value := dec.ReadBool()
		args.PrimarySaleHappened = &value
	}
	if dec.ReadOption() {
		value := dec.ReadBool()
		args.IsMutable = &value
	}
}

func readOptionalUint64(dec *bin.Decoder) *uint64 {
	if !dec.ReadOption() {
		return nil
	}
	value := dec.ReadUint64()
	return &value
}
