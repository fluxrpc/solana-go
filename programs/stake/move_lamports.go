package stake

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type MoveLamports struct {
	Lamports uint64
	instruction
}

func NewMoveLamportsInstruction(lamports uint64, sourceStakeAccount, destinationStakeAccount, stakeAuthority solana.PublicKey) *MoveLamports {
	return &MoveLamports{Lamports: lamports, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(sourceStakeAccount, true, false),
		solana.NewAccountMeta(destinationStakeAccount, true, false),
		solana.NewAccountMeta(stakeAuthority, false, true),
	}}}
}

func (inst *MoveLamports) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint32(uint32(MoveLamportsInstruction))
	enc.WriteUint64(inst.Lamports)
	return enc.Bytes(), enc.Err()
}
