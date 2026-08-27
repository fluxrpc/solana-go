package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Split struct {
	Lamports uint64
	instruction
}

func NewSplitInstruction(lamports uint64, stakeAccount, splitStakeAccount, stakeAuthority solana.PublicKey) *Split {
	return &Split{Lamports: lamports, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(splitStakeAccount, true, false),
		solana.NewAccountMeta(stakeAuthority, false, true),
	}}}
}

func (inst *Split) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint32(uint32(SplitInstruction))
	enc.WriteUint64(inst.Lamports)
	return enc.Bytes(), enc.Err()
}
