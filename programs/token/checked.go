package token

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type checkedInstruction struct {
	instruction
	Amount   uint64
	Decimals uint8
}

func (inst *checkedInstruction) data(typ InstructionType) ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 10))
	enc.WriteUint8(uint8(typ))
	enc.WriteUint64(inst.Amount)
	enc.WriteUint8(inst.Decimals)
	return enc.Bytes(), enc.Err()
}

type TransferChecked struct{ checkedInstruction }

func NewTransferCheckedInstruction(amount uint64, decimals uint8, source, mint, destination, owner solana.PublicKey, signers []solana.PublicKey) *TransferChecked {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(destination, true, false),
	}
	return &TransferChecked{checkedInstruction{Amount: amount, Decimals: decimals, instruction: instruction{AccountMetaSlice: appendAuthority(accounts, owner, signers)}}}
}
func (inst *TransferChecked) tokenInstruction() *instruction { return &inst.instruction }
func (inst *TransferChecked) Data() ([]byte, error)          { return inst.data(TransferCheckedInstruction) }

type ApproveChecked struct{ checkedInstruction }

func NewApproveCheckedInstruction(amount uint64, decimals uint8, source, mint, delegate, owner solana.PublicKey, signers []solana.PublicKey) *ApproveChecked {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(mint, false, false),
		solana.NewAccountMeta(delegate, false, false),
	}
	return &ApproveChecked{checkedInstruction{Amount: amount, Decimals: decimals, instruction: instruction{AccountMetaSlice: appendAuthority(accounts, owner, signers)}}}
}
func (inst *ApproveChecked) tokenInstruction() *instruction { return &inst.instruction }
func (inst *ApproveChecked) Data() ([]byte, error)          { return inst.data(ApproveCheckedInstruction) }

type MintToChecked struct{ checkedInstruction }

func NewMintToCheckedInstruction(amount uint64, decimals uint8, mint, destination, authority solana.PublicKey, signers []solana.PublicKey) *MintToChecked {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(mint, true, false),
		solana.NewAccountMeta(destination, true, false),
	}
	return &MintToChecked{checkedInstruction{Amount: amount, Decimals: decimals, instruction: instruction{AccountMetaSlice: appendAuthority(accounts, authority, signers)}}}
}
func (inst *MintToChecked) tokenInstruction() *instruction { return &inst.instruction }
func (inst *MintToChecked) Data() ([]byte, error)          { return inst.data(MintToCheckedInstruction) }

type BurnChecked struct{ checkedInstruction }

func NewBurnCheckedInstruction(amount uint64, decimals uint8, source, mint, owner solana.PublicKey, signers []solana.PublicKey) *BurnChecked {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(mint, true, false),
	}
	return &BurnChecked{checkedInstruction{Amount: amount, Decimals: decimals, instruction: instruction{AccountMetaSlice: appendAuthority(accounts, owner, signers)}}}
}
func (inst *BurnChecked) tokenInstruction() *instruction { return &inst.instruction }
func (inst *BurnChecked) Data() ([]byte, error)          { return inst.data(BurnCheckedInstruction) }
