package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type Vote struct {
	Slots     []uint64
	Hash      solana.Hash
	Timestamp *int64
	instruction
}

func NewVoteInstruction(slots []uint64, hash solana.Hash, timestamp *int64, voteAccount, voteAuthority solana.PublicKey) *Vote {
	return &Vote{Slots: slots, Hash: hash, Timestamp: timestamp, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(solana.SysVarSlotHashesPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(voteAuthority, false, true),
	}}}
}
func (inst *Vote) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 45+len(inst.Slots)*8))
	enc.WriteUint32(uint32(VoteInstruction))
	if err := (voteData{Slots: inst.Slots, Hash: inst.Hash, Timestamp: inst.Timestamp}).write(enc); err != nil {
		return nil, err
	}
	return enc.Bytes(), enc.Err()
}
