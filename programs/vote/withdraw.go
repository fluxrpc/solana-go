package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type Withdraw struct {
	Lamports uint64
	instruction
}

func NewWithdrawInstruction(lamports uint64, voteAccount, recipientAccount, withdrawAuthority solana.PublicKey) *Withdraw {
	return &Withdraw{Lamports: lamports, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(recipientAccount, true, false),
		solana.NewAccountMeta(withdrawAuthority, false, true),
	}}}
}
func (inst *Withdraw) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint32(uint32(WithdrawInstruction))
	enc.WriteUint64(inst.Lamports)
	return enc.Bytes(), enc.Err()
}
