package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Withdraw struct {
	Lamports uint64
	instruction
}

func NewWithdrawInstruction(lamports uint64, stakeAccount, recipient, withdrawAuthority solana.PublicKey, custodian *solana.PublicKey) *Withdraw {
	accounts := make(solana.AccountMetaSlice, 0, 6)
	accounts = append(accounts,
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(recipient, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarStakeHistoryPubkey, false, false),
		solana.NewAccountMeta(withdrawAuthority, false, true),
	)
	if custodian != nil {
		accounts = append(
			accounts,
			solana.NewAccountMeta(*custodian, false, true),
		)
	}
	return &Withdraw{Lamports: lamports, instruction: instruction{AccountMetaSlice: accounts}}
}

func (inst *Withdraw) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint32(uint32(WithdrawInstruction))
	enc.WriteUint64(inst.Lamports)
	return enc.Bytes(), enc.Err()
}
