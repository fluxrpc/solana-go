package token2022

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type GetAccountDataSize struct {
	extensionInstruction
	ExtensionTypes []ExtensionType
}

func NewGetAccountDataSizeInstruction(extensionTypes []ExtensionType, mint solana.PublicKey) *GetAccountDataSize {
	return &GetAccountDataSize{ExtensionTypes: extensionTypes, extensionInstruction: extensionInstruction{solana.AccountMetaSlice{solana.NewAccountMeta(mint, false, false)}}}
}

func (inst *GetAccountDataSize) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 1, 1+2*len(inst.ExtensionTypes)))
	enc.Bytes()[0] = byte(GetAccountDataSizeInstruction)
	for _, typ := range inst.ExtensionTypes {
		enc.WriteUint16(uint16(typ))
	}
	return enc.Bytes(), enc.Err()
}

type InitializeImmutableOwner struct{ extensionInstruction }

func NewInitializeImmutableOwnerInstruction(account solana.PublicKey) *InitializeImmutableOwner {
	return &InitializeImmutableOwner{extensionInstruction{solana.AccountMetaSlice{solana.NewAccountMeta(account, true, false)}}}
}

func (*InitializeImmutableOwner) Data() ([]byte, error) {
	return []byte{byte(InitializeImmutableOwnerInstruction)}, nil
}

type AmountToUIAmount struct {
	extensionInstruction
	Amount uint64
}

func NewAmountToUiAmountInstruction(amount uint64, mint solana.PublicKey) *AmountToUIAmount {
	return &AmountToUIAmount{Amount: amount, extensionInstruction: extensionInstruction{solana.AccountMetaSlice{solana.NewAccountMeta(mint, false, false)}}}
}

func (inst *AmountToUIAmount) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 9))
	enc.WriteUint8(uint8(AmountToUIAmountInstruction))
	enc.WriteUint64(inst.Amount)
	return enc.Bytes(), enc.Err()
}

type UIAmountToAmount struct {
	extensionInstruction
	UIAmount string
}

func NewUiAmountToAmountInstruction(uiAmount string, mint solana.PublicKey) *UIAmountToAmount {
	return &UIAmountToAmount{UIAmount: uiAmount, extensionInstruction: extensionInstruction{solana.AccountMetaSlice{solana.NewAccountMeta(mint, false, false)}}}
}

func (inst *UIAmountToAmount) Data() ([]byte, error) {
	data := make([]byte, 1, 1+len(inst.UIAmount))
	data[0] = byte(UIAmountToAmountInstruction)
	data = append(data, inst.UIAmount...)
	return data, nil
}

type InitializeMintCloseAuthority struct {
	extensionInstruction
	CloseAuthority *solana.PublicKey
}

func NewInitializeMintCloseAuthorityInstruction(closeAuthority, mint solana.PublicKey) *InitializeMintCloseAuthority {
	inst := &InitializeMintCloseAuthority{extensionInstruction: extensionInstruction{solana.AccountMetaSlice{solana.NewAccountMeta(mint, true, false)}}}
	if !closeAuthority.IsZero() {
		inst.CloseAuthority = &closeAuthority
	}
	return inst
}

func (inst *InitializeMintCloseAuthority) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 34))
	enc.WriteUint8(uint8(InitializeMintCloseAuthorityInstruction))
	enc.WriteOption(inst.CloseAuthority != nil)
	if inst.CloseAuthority != nil {
		enc.WritePublicKey(*inst.CloseAuthority)
	}
	return enc.Bytes(), enc.Err()
}
