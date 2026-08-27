package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type CompactUpdateVoteStateSwitch struct {
	VoteStateUpdate
	ProofHash solana.Hash
	instruction
}

func NewCompactUpdateVoteStateSwitchInstruction(update VoteStateUpdate, proofHash solana.Hash, voteAccount, voteAuthority solana.PublicKey) *CompactUpdateVoteStateSwitch {
	return &CompactUpdateVoteStateSwitch{VoteStateUpdate: update, ProofHash: proofHash, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(voteAuthority, false, true),
	}}}
}
func (inst *CompactUpdateVoteStateSwitch) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 87+len(inst.Lockouts)*3))
	enc.WriteUint32(uint32(CompactUpdateVoteStateSwitchInstruction))
	if err := inst.VoteStateUpdate.writeCompact(enc); err != nil {
		return nil, err
	}
	enc.WriteHash(inst.ProofHash)
	return enc.Bytes(), enc.Err()
}
