package token2022

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type extensionInstruction struct{ solana.AccountMetaSlice }

func (*extensionInstruction) ProgramID() solana.PublicKey          { return ProgramID }
func (inst *extensionInstruction) Accounts() []*solana.AccountMeta { return inst.AccountMetaSlice }

type InitializeTransferFeeConfig struct {
	extensionInstruction
	TransferFeeConfigAuthority *solana.PublicKey
	WithdrawWithheldAuthority  *solana.PublicKey
	TransferFeeBasisPoints     uint16
	MaximumFee                 uint64
}

func NewInitializeTransferFeeConfig(mint solana.PublicKey, configAuthority, withdrawAuthority *solana.PublicKey, basisPoints uint16, maximumFee uint64) *InitializeTransferFeeConfig {
	return &InitializeTransferFeeConfig{
		extensionInstruction:       extensionInstruction{solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}},
		TransferFeeConfigAuthority: configAuthority,
		WithdrawWithheldAuthority:  withdrawAuthority,
		TransferFeeBasisPoints:     basisPoints,
		MaximumFee:                 maximumFee,
	}
}

func NewInitializeTransferFeeConfigInstruction(configAuthority, withdrawAuthority *solana.PublicKey, basisPoints uint16, maximumFee uint64, mint solana.PublicKey) *InitializeTransferFeeConfig {
	return NewInitializeTransferFeeConfig(mint, configAuthority, withdrawAuthority, basisPoints, maximumFee)
}

func (inst *InitializeTransferFeeConfig) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 80))
	enc.WriteUint8(uint8(TransferFeeExtensionInstruction))
	enc.WriteUint8(0)
	enc.WriteOption(inst.TransferFeeConfigAuthority != nil)
	if inst.TransferFeeConfigAuthority == nil {
		enc.WritePublicKey(solana.PublicKey{})
	} else {
		enc.WritePublicKey(*inst.TransferFeeConfigAuthority)
	}
	enc.WriteOption(inst.WithdrawWithheldAuthority != nil)
	if inst.WithdrawWithheldAuthority == nil {
		enc.WritePublicKey(solana.PublicKey{})
	} else {
		enc.WritePublicKey(*inst.WithdrawWithheldAuthority)
	}
	enc.WriteUint16(inst.TransferFeeBasisPoints)
	enc.WriteUint64(inst.MaximumFee)
	return enc.Bytes(), enc.Err()
}

type TransferCheckedWithFee struct {
	extensionInstruction
	Amount   uint64
	Decimals uint8
	Fee      uint64
}

func NewTransferCheckedWithFeeInstruction(amount uint64, decimals uint8, fee uint64, source, mint, destination, authority solana.PublicKey, signers []solana.PublicKey) *TransferCheckedWithFee {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(destination, true, false),
	}
	accounts = append(accounts, tokenAuthorityAccounts(authority, signers)...)
	return &TransferCheckedWithFee{extensionInstruction: extensionInstruction{accounts}, Amount: amount, Decimals: decimals, Fee: fee}
}

func (inst *TransferCheckedWithFee) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 19))
	enc.WriteUint8(uint8(TransferFeeExtensionInstruction))
	enc.WriteUint8(1)
	enc.WriteUint64(inst.Amount)
	enc.WriteUint8(inst.Decimals)
	enc.WriteUint64(inst.Fee)
	return enc.Bytes(), enc.Err()
}

type WithdrawWithheldTokensFromMint struct{ extensionInstruction }

func NewWithdrawWithheldTokensFromMint(mint, authority, destination solana.PublicKey) *WithdrawWithheldTokensFromMint {
	return &WithdrawWithheldTokensFromMint{extensionInstruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(mint, true, false),
		solana.NewAccountMeta(destination, true, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}
func (inst *WithdrawWithheldTokensFromMint) Data() ([]byte, error) {
	return []byte{uint8(TransferFeeExtensionInstruction), 2}, nil
}

func NewWithdrawWithheldTokensFromMintInstruction(mint, destination, authority solana.PublicKey, signers []solana.PublicKey) *WithdrawWithheldTokensFromMint {
	inst := NewWithdrawWithheldTokensFromMint(mint, authority, destination)
	inst.AccountMetaSlice = append(inst.AccountMetaSlice[:2], tokenAuthorityAccounts(authority, signers)...)
	return inst
}

type WithdrawWithheldTokensFromAccounts struct {
	extensionInstruction
	NumTokenAccounts uint8
}

func NewWithdrawWithheldTokensFromAccountsInstruction(mint, destination, authority solana.PublicKey, signers []solana.PublicKey, sourceAccounts ...solana.PublicKey) *WithdrawWithheldTokensFromAccounts {
	metas := solana.AccountMetaSlice{
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(destination, true, false),
	}
	metas = append(metas, tokenAuthorityAccounts(authority, signers)...)
	for _, account := range sourceAccounts {
		metas = append(metas, solana.NewAccountMeta(account, true, false))
	}
	return &WithdrawWithheldTokensFromAccounts{extensionInstruction: extensionInstruction{metas}, NumTokenAccounts: uint8(len(sourceAccounts))}
}

func NewWithdrawWithheldTokensFromAccounts(mint, authority, destination solana.PublicKey, accounts []solana.PublicKey) *WithdrawWithheldTokensFromAccounts {
	metas := solana.AccountMetaSlice{
		solana.NewAccountMeta(mint, true, false),
		solana.NewAccountMeta(destination, true, false),
		solana.NewAccountMeta(authority, false, true),
	}
	for _, account := range accounts {
		metas = append(metas, solana.NewAccountMeta(account, true, false))
	}
	return &WithdrawWithheldTokensFromAccounts{extensionInstruction: extensionInstruction{metas}, NumTokenAccounts: uint8(len(accounts))}
}
func (inst *WithdrawWithheldTokensFromAccounts) Data() ([]byte, error) {
	return []byte{uint8(TransferFeeExtensionInstruction), 3, inst.NumTokenAccounts}, nil
}

type HarvestWithheldTokensToMint struct{ extensionInstruction }

func NewHarvestWithheldTokensToMint(mint solana.PublicKey, accounts []solana.PublicKey) *HarvestWithheldTokensToMint {
	metas := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	for _, account := range accounts {
		metas = append(metas, solana.NewAccountMeta(account, true, false))
	}
	return &HarvestWithheldTokensToMint{extensionInstruction{metas}}
}
func (inst *HarvestWithheldTokensToMint) Data() ([]byte, error) {
	return []byte{uint8(TransferFeeExtensionInstruction), 4}, nil
}

func NewHarvestWithheldTokensToMintInstruction(mint solana.PublicKey, sourceAccounts ...solana.PublicKey) *HarvestWithheldTokensToMint {
	return NewHarvestWithheldTokensToMint(mint, sourceAccounts)
}

type SetTransferFee struct {
	extensionInstruction
	TransferFeeBasisPoints uint16
	MaximumFee             uint64
}

func NewSetTransferFeeInstruction(basisPoints uint16, maximumFee uint64, mint, authority solana.PublicKey, signers []solana.PublicKey) *SetTransferFee {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	accounts = append(accounts, tokenAuthorityAccounts(authority, signers)...)
	return &SetTransferFee{extensionInstruction: extensionInstruction{accounts}, TransferFeeBasisPoints: basisPoints, MaximumFee: maximumFee}
}

func (inst *SetTransferFee) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint8(uint8(TransferFeeExtensionInstruction))
	enc.WriteUint8(5)
	enc.WriteUint16(inst.TransferFeeBasisPoints)
	enc.WriteUint64(inst.MaximumFee)
	return enc.Bytes(), enc.Err()
}

type InitializeMetadataPointer struct {
	extensionInstruction
	SubInstruction  uint8
	Authority       *solana.PublicKey
	MetadataAddress solana.PublicKey
}

func NewInitializeMetadataPointer(mint, metadataAddress solana.PublicKey, authority *solana.PublicKey) *InitializeMetadataPointer {
	return &InitializeMetadataPointer{
		extensionInstruction: extensionInstruction{solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}},
		Authority:            authority,
		MetadataAddress:      metadataAddress,
	}
}

func NewInitializeMetadataPointerInstruction(authority, metadataAddress *solana.PublicKey, mint solana.PublicKey) *InitializeMetadataPointer {
	address := solana.PublicKey{}
	if metadataAddress != nil {
		address = *metadataAddress
	}
	return NewInitializeMetadataPointer(mint, address, authority)
}

func NewUpdateMetadataPointerInstruction(metadataAddress *solana.PublicKey, mint, authority solana.PublicKey, signers []solana.PublicKey) *InitializeMetadataPointer {
	address := solana.PublicKey{}
	if metadataAddress != nil {
		address = *metadataAddress
	}
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}
	accounts = append(accounts, tokenAuthorityAccounts(authority, signers)...)
	return &InitializeMetadataPointer{SubInstruction: 1, MetadataAddress: address, extensionInstruction: extensionInstruction{accounts}}
}
func (inst *InitializeMetadataPointer) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 66))
	enc.WriteUint8(uint8(MetadataPointerExtensionInstruction))
	enc.WriteUint8(inst.SubInstruction)
	if inst.SubInstruction == 0 {
		if inst.Authority == nil {
			enc.WritePublicKey(solana.PublicKey{})
		} else {
			enc.WritePublicKey(*inst.Authority)
		}
	}
	enc.WritePublicKey(inst.MetadataAddress)
	return enc.Bytes(), enc.Err()
}

func tokenAuthorityAccounts(authority solana.PublicKey, signers []solana.PublicKey) solana.AccountMetaSlice {
	accounts := solana.AccountMetaSlice{solana.NewAccountMeta(authority, false, len(signers) == 0)}
	for _, signer := range signers {
		accounts = append(accounts, solana.NewAccountMeta(signer, false, true))
	}
	return accounts
}

type InitializeMetadata struct {
	extensionInstruction
	Name   string
	Symbol string
	URI    string
}

func NewInitializeMetadata(mint, metadata, mintAuthority solana.PublicKey, updateAuthority *solana.PublicKey, name, symbol, uri string) *InitializeMetadata {
	update := solana.PublicKey{}
	if updateAuthority != nil {
		update = *updateAuthority
	}
	return &InitializeMetadata{
		extensionInstruction: extensionInstruction{solana.AccountMetaSlice{
			solana.NewAccountMeta(metadata, true, false),
			solana.NewAccountMeta(update, false, false),
			solana.NewAccountMeta(mint, false, false),
			solana.NewAccountMeta(mintAuthority, false, true),
		}},
		Name: name, Symbol: symbol, URI: uri,
	}
}
func (inst *InitializeMetadata) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 20+len(inst.Name)+len(inst.Symbol)+len(inst.URI)))
	enc.WriteBytes([]byte{210, 225, 30, 162, 88, 184, 77, 141})
	enc.WriteBorshString(inst.Name)
	enc.WriteBorshString(inst.Symbol)
	enc.WriteBorshString(inst.URI)
	return enc.Bytes(), enc.Err()
}
