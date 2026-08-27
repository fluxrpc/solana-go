package vote

import (
	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type InitializeAccountV2 struct {
	VoteInitV2
	instruction
}

func NewInitializeAccountV2Instruction(voteInit VoteInitV2, voteAccount, inflationRewardsCollector, blockRevenueCollector solana.PublicKey) *InitializeAccountV2 {
	return &InitializeAccountV2{VoteInitV2: voteInit, instruction: instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(voteAccount, true, false),
		solana.NewAccountMeta(voteInit.NodePubkey, false, true),
		solana.NewAccountMeta(inflationRewardsCollector, true, false),
		solana.NewAccountMeta(blockRevenueCollector, true, false),
	}}}
}
func (inst *InitializeAccountV2) Data() ([]byte, error) {
	enc := bin.NewEncoder(make([]byte, 0, 248))
	enc.WriteUint32(uint32(InitializeAccountV2Instruction))
	inst.VoteInitV2.write(enc)
	return enc.Bytes(), enc.Err()
}
