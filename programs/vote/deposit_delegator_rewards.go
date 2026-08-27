package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type DepositDelegatorRewards struct {
	Deposit uint64
	instruction
}

func NewDepositDelegatorRewardsInstruction(deposit uint64, voteAccount, depositor solana.PublicKey) *DepositDelegatorRewards {
	return &DepositDelegatorRewards{Deposit: deposit, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(depositor, true, true),
	}}}
}
func (inst *DepositDelegatorRewards) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint32(uint32(DepositDelegatorRewardsInstruction))
	enc.WriteUint64(inst.Deposit)
	return enc.Bytes(), enc.Err()
}
