package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type UpdateVoteStateSwitch struct {
	VoteStateUpdate
	ProofHash solana.Hash
	instruction
}

func NewUpdateVoteStateSwitchInstruction(update VoteStateUpdate, proofHash solana.Hash, voteAccount, voteAuthority solana.PublicKey) *UpdateVoteStateSwitch {
	return &UpdateVoteStateSwitch{VoteStateUpdate: update, ProofHash: proofHash, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(voteAuthority, false, true),
	}}}
}
func (inst *UpdateVoteStateSwitch) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 79+len(inst.Lockouts)*12))
	enc.WriteUint32(uint32(UpdateVoteStateSwitchInstruction))
	if err := inst.VoteStateUpdate.write(enc); err != nil {
		return nil, err
	}
	enc.WriteHash(inst.ProofHash)
	return enc.Bytes(), enc.Err()
}
