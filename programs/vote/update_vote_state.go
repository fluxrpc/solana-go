package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type UpdateVoteState struct {
	VoteStateUpdate
	instruction
}

func NewUpdateVoteStateInstruction(update VoteStateUpdate, voteAccount, voteAuthority solana.PublicKey) *UpdateVoteState {
	return &UpdateVoteState{VoteStateUpdate: update, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(voteAuthority, false, true),
	}}}
}
func (inst *UpdateVoteState) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 47+len(inst.Lockouts)*12))
	enc.WriteUint32(uint32(UpdateVoteStateInstruction))
	if err := inst.VoteStateUpdate.write(enc); err != nil {
		return nil, err
	}
	return enc.Bytes(), enc.Err()
}
