package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type MoveStake struct {
	Lamports uint64
	instruction
}

func NewMoveStakeInstruction(lamports uint64, sourceStakeAccount, destinationStakeAccount, stakeAuthority solana.PublicKey) *MoveStake {
	return &MoveStake{Lamports: lamports, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(sourceStakeAccount, true, false),
		solana.NewAccountMeta(destinationStakeAccount, true, false),
		solana.NewAccountMeta(stakeAuthority, false, true),
	}}}
}

func (inst *MoveStake) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint32(uint32(MoveStakeInstruction))
	enc.WriteUint64(inst.Lamports)
	return enc.Bytes(), enc.Err()
}
