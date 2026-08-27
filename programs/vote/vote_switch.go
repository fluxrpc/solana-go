package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type VoteSwitch struct {
	Slots     []uint64
	VoteHash  solana.Hash
	Timestamp *int64
	ProofHash solana.Hash
	instruction
}

func NewVoteSwitchInstruction(slots []uint64, voteHash solana.Hash, timestamp *int64, proofHash solana.Hash, voteAccount, voteAuthority solana.PublicKey) *VoteSwitch {
	return &VoteSwitch{Slots: slots, VoteHash: voteHash, Timestamp: timestamp, ProofHash: proofHash, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(solana.SysVarSlotHashesPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(voteAuthority, false, true),
	}}}
}
func (inst *VoteSwitch) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 77+len(inst.Slots)*8))
	enc.WriteUint32(uint32(VoteSwitchInstruction))
	if err := (voteData{Slots: inst.Slots, Hash: inst.VoteHash, Timestamp: inst.Timestamp}).write(enc); err != nil {
		return nil, err
	}
	enc.WriteHash(inst.ProofHash)
	return enc.Bytes(), enc.Err()
}
