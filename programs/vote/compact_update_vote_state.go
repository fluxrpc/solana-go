package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type CompactUpdateVoteState struct {
	VoteStateUpdate
	instruction
}

func NewCompactUpdateVoteStateInstruction(update VoteStateUpdate, voteAccount, voteAuthority solana.PublicKey) *CompactUpdateVoteState {
	return &CompactUpdateVoteState{VoteStateUpdate: update, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(voteAuthority, false, true),
	}}}
}
func (inst *CompactUpdateVoteState) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 55+len(inst.Lockouts)*3))
	enc.WriteUint32(uint32(CompactUpdateVoteStateInstruction))
	if err := inst.VoteStateUpdate.writeCompact(enc); err != nil {
		return nil, err
	}
	return enc.Bytes(), enc.Err()
}
