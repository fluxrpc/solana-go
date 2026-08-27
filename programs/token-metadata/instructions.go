package tokenmetadata

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type instruction struct {
	solana.AccountMetaSlice
	typ InstructionType
}

func (*instruction) ProgramID() solana.PublicKey          { return ProgramID }
func (inst *instruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }
func (inst *instruction) Data() ([]byte, error)           { return []byte{byte(inst.typ)}, nil }

type CreateMetadataAccount struct {
	instruction
	Args CreateMetadataAccountArgs
}

func NewCreateMetadataAccountInstruction(args CreateMetadataAccountArgs, metadata, mint, mintAuthority, payer, updateAuthority, system, rent solana.PublicKey) *CreateMetadataAccount {
	return &CreateMetadataAccount{Args: args, instruction: instruction{typ: CreateMetadataAccountInstruction, AccountMetaSlice: metadataAccounts(metadata, mint, mintAuthority, payer, updateAuthority, system, rent)}}
}
func (inst *CreateMetadataAccount) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 128))
	enc.WriteUint8(uint8(CreateMetadataAccountInstruction))
	inst.Args.Data.encode(enc)
	enc.WriteBool(inst.Args.IsMutable)
	return enc.Bytes(), enc.Err()
}

type CreateMetadataAccountV2 struct {
	instruction
	Args CreateMetadataAccountArgsV2
}

func NewCreateMetadataAccountV2Instruction(args CreateMetadataAccountArgsV2, metadata, mint, mintAuthority, payer, updateAuthority, system, rent solana.PublicKey) *CreateMetadataAccountV2 {
	return &CreateMetadataAccountV2{Args: args, instruction: instruction{typ: CreateMetadataAccountV2Instruction, AccountMetaSlice: metadataAccounts(metadata, mint, mintAuthority, payer, updateAuthority, system, rent)}}
}
func (inst *CreateMetadataAccountV2) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 160))
	enc.WriteUint8(uint8(CreateMetadataAccountV2Instruction))
	inst.Args.Data.encode(enc)
	enc.WriteBool(inst.Args.IsMutable)
	return enc.Bytes(), enc.Err()
}

func metadataAccounts(metadata, mint, mintAuthority, payer, updateAuthority, system, rent solana.PublicKey) solana.AccountMetaSlice {
	return solana.AccountMetaSlice{
		solana.NewAccountMeta(metadata, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(mintAuthority, false, true),
		solana.NewAccountMeta(payer, true, true),
		solana.NewAccountMeta(updateAuthority, false, false),
		solana.NewAccountMeta(system, false, false),
		solana.NewAccountMeta(rent, false, false),
	}
}

type UpdateMetadataAccount struct {
	instruction
	Args UpdateMetadataAccountArgs
}

func NewUpdateMetadataAccountInstruction(args UpdateMetadataAccountArgs, metadata, updateAuthority solana.PublicKey) *UpdateMetadataAccount {
	return &UpdateMetadataAccount{Args: args, instruction: instruction{typ: UpdateMetadataAccountInstruction, AccountMetaSlice: updateAccounts(metadata, updateAuthority)}}
}
func (inst *UpdateMetadataAccount) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 128))
	enc.WriteUint8(uint8(UpdateMetadataAccountInstruction))
	enc.WriteOption(inst.Args.Data != nil)
	if inst.Args.Data != nil {
		inst.Args.Data.encode(enc)
	}
	writeOptionalPublicKey(enc, inst.Args.UpdateAuthority)
	writeOptionalBool(enc, inst.Args.PrimarySaleHappened)
	return enc.Bytes(), enc.Err()
}

type UpdateMetadataAccountV2 struct {
	instruction
	Args UpdateMetadataAccountArgsV2
}

func NewUpdateMetadataAccountV2Instruction(args UpdateMetadataAccountArgsV2, metadata, updateAuthority solana.PublicKey) *UpdateMetadataAccountV2 {
	return &UpdateMetadataAccountV2{Args: args, instruction: instruction{typ: UpdateMetadataAccountV2Instruction, AccountMetaSlice: updateAccounts(metadata, updateAuthority)}}
}
func (inst *UpdateMetadataAccountV2) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 160))
	enc.WriteUint8(uint8(UpdateMetadataAccountV2Instruction))
	enc.WriteOption(inst.Args.Data != nil)
	if inst.Args.Data != nil {
		inst.Args.Data.encode(enc)
	}
	writeOptionalPublicKey(enc, inst.Args.UpdateAuthority)
	writeOptionalBool(enc, inst.Args.PrimarySaleHappened)
	writeOptionalBool(enc, inst.Args.IsMutable)
	return enc.Bytes(), enc.Err()
}

func updateAccounts(metadata, updateAuthority solana.PublicKey) solana.AccountMetaSlice {
	return solana.AccountMetaSlice{solana.NewAccountMeta(metadata, true, false), solana.NewAccountMeta(updateAuthority, false, true)}
}

type CreateMasterEdition struct {
	instruction
	Args CreateMasterEditionArgs
}

func NewCreateMasterEditionInstruction(args CreateMasterEditionArgs, edition, mint, updateAuthority, mintAuthority, payer, metadata, tokenProgram, system, rent solana.PublicKey) *CreateMasterEdition {
	return &CreateMasterEdition{Args: args, instruction: instruction{typ: CreateMasterEditionInstruction, AccountMetaSlice: masterEditionAccounts(edition, mint, updateAuthority, mintAuthority, payer, metadata, tokenProgram, system, rent)}}
}
func (inst *CreateMasterEdition) Data() ([]byte, error) {
	return masterEditionData(CreateMasterEditionInstruction, inst.Args.MaxSupply)
}

type CreateMasterEditionV3 struct {
	instruction
	Args CreateMasterEditionArgs
}

func NewCreateMasterEditionV3Instruction(args CreateMasterEditionArgs, edition, mint, updateAuthority, mintAuthority, payer, metadata, tokenProgram, system, rent solana.PublicKey) *CreateMasterEditionV3 {
	return &CreateMasterEditionV3{Args: args, instruction: instruction{typ: CreateMasterEditionV3Instruction, AccountMetaSlice: masterEditionAccounts(edition, mint, updateAuthority, mintAuthority, payer, metadata, tokenProgram, system, rent)}}
}
func (inst *CreateMasterEditionV3) Data() ([]byte, error) {
	return masterEditionData(CreateMasterEditionV3Instruction, inst.Args.MaxSupply)
}

func masterEditionData(typ InstructionType, maxSupply *uint64) ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 10))
	enc.WriteUint8(uint8(typ))
	enc.WriteOption(maxSupply != nil)
	if maxSupply != nil {
		enc.WriteUint64(*maxSupply)
	}
	return enc.Bytes(), enc.Err()
}
func masterEditionAccounts(edition, mint, updateAuthority, mintAuthority, payer, metadata, tokenProgram, system, rent solana.PublicKey) solana.AccountMetaSlice {
	return solana.AccountMetaSlice{
		solana.NewAccountMeta(edition, true, false), solana.NewAccountMeta(mint, true, false),
		solana.NewAccountMeta(updateAuthority, false, true), solana.NewAccountMeta(mintAuthority, false, true),
		solana.NewAccountMeta(payer, true, true), solana.NewAccountMeta(metadata, false, false),
		solana.NewAccountMeta(tokenProgram, false, false), solana.NewAccountMeta(system, false, false),
		solana.NewAccountMeta(rent, false, false),
	}
}

type SignMetadata struct{ instruction }

func NewSignMetadataInstruction(metadata, creator solana.PublicKey) *SignMetadata {
	return &SignMetadata{instruction{typ: SignMetadataInstruction, AccountMetaSlice: solana.AccountMetaSlice{solana.NewAccountMeta(metadata, true, false), solana.NewAccountMeta(creator, false, true)}}}
}

type PuffMetadata struct{ instruction }

func NewPuffMetadataInstruction(metadata solana.PublicKey) *PuffMetadata {
	return &PuffMetadata{instruction{typ: PuffMetadataInstruction, AccountMetaSlice: solana.AccountMetaSlice{solana.NewAccountMeta(metadata, true, false)}}}
}

type UpdatePrimarySaleHappenedViaToken struct{ instruction }

func NewUpdatePrimarySaleHappenedViaTokenInstruction(metadata, owner, tokenAccount solana.PublicKey) *UpdatePrimarySaleHappenedViaToken {
	return &UpdatePrimarySaleHappenedViaToken{instruction{typ: UpdatePrimarySaleHappenedViaTokenInstruction, AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(metadata, true, false), solana.NewAccountMeta(owner, false, true), solana.NewAccountMeta(tokenAccount, false, false),
	}}}
}

type VerifyCollection struct{ instruction }

func NewVerifyCollectionInstruction(metadata, collectionAuthority, payer, collectionMint, collectionMetadata, masterEdition solana.PublicKey) *VerifyCollection {
	return &VerifyCollection{instruction{typ: VerifyCollectionInstruction, AccountMetaSlice: collectionAccounts(metadata, collectionAuthority, payer, collectionMint, collectionMetadata, masterEdition)}}
}

type UnverifyCollection struct{ instruction }

func NewUnverifyCollectionInstruction(metadata, collectionAuthority, payer, collectionMint, collectionMetadata, masterEdition solana.PublicKey) *UnverifyCollection {
	return &UnverifyCollection{instruction{typ: UnverifyCollectionInstruction, AccountMetaSlice: collectionAccounts(metadata, collectionAuthority, payer, collectionMint, collectionMetadata, masterEdition)}}
}
func collectionAccounts(metadata, authority, payer, mint, metadataCollection, edition solana.PublicKey) solana.AccountMetaSlice {
	return solana.AccountMetaSlice{
		solana.NewAccountMeta(metadata, true, false), solana.NewAccountMeta(authority, false, true),
		solana.NewAccountMeta(payer, true, true), solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(metadataCollection, false, false), solana.NewAccountMeta(edition, false, false),
	}
}
