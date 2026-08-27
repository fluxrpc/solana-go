package token

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type amountInstruction struct {
	instruction
	Amount uint64
}

func (inst *amountInstruction) data(typ InstructionType) ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 9))
	enc.WriteUint8(uint8(typ))
	enc.WriteUint64(inst.Amount)
	return enc.Bytes(), enc.Err()
}

type Transfer struct{ amountInstruction }

func NewTransferInstruction(amount uint64, source, destination, owner solana.PublicKey, signers []solana.PublicKey) *Transfer {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(destination, true, false),
	}
	return &Transfer{amountInstruction{Amount: amount, instruction: instruction{AccountMetaSlice: appendAuthority(accounts, owner, signers)}}}
}
func (inst *Transfer) tokenInstruction() *instruction { return &inst.instruction }
func (inst *Transfer) Data() ([]byte, error)          { return inst.data(TransferInstruction) }

type Approve struct{ amountInstruction }

func NewApproveInstruction(amount uint64, source, delegate, owner solana.PublicKey, signers []solana.PublicKey) *Approve {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(delegate, false, false),
	}
	return &Approve{amountInstruction{Amount: amount, instruction: instruction{AccountMetaSlice: appendAuthority(accounts, owner, signers)}}}
}
func (inst *Approve) tokenInstruction() *instruction { return &inst.instruction }
func (inst *Approve) Data() ([]byte, error)          { return inst.data(ApproveInstruction) }

type MintTo struct{ amountInstruction }

func NewMintToInstruction(amount uint64, mint, destination, authority solana.PublicKey, signers []solana.PublicKey) *MintTo {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(mint, true, false),
		solana.NewAccountMeta(destination, true, false),
	}
	return &MintTo{amountInstruction{Amount: amount, instruction: instruction{AccountMetaSlice: appendAuthority(accounts, authority, signers)}}}
}
func (inst *MintTo) tokenInstruction() *instruction { return &inst.instruction }
func (inst *MintTo) Data() ([]byte, error)          { return inst.data(MintToInstruction) }

type Burn struct{ amountInstruction }

func NewBurnInstruction(amount uint64, source, mint, owner solana.PublicKey, signers []solana.PublicKey) *Burn {
	accounts := solana.AccountMetaSlice{
		solana.NewAccountMeta(source, true, false),
		solana.NewAccountMeta(mint, true, false),
	}
	return &Burn{amountInstruction{Amount: amount, instruction: instruction{AccountMetaSlice: appendAuthority(accounts, owner, signers)}}}
}
func (inst *Burn) tokenInstruction() *instruction { return &inst.instruction }
func (inst *Burn) Data() ([]byte, error)          { return inst.data(BurnInstruction) }
