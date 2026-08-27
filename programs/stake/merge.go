package stake

import solana "github.com/fluxrpc/solana-go"

type Merge struct{ instruction }

func NewMergeInstruction(destinationStake, sourceStake, stakeAuthority solana.PublicKey) *Merge {
	return &Merge{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(destinationStake, true, false),
		solana.NewAccountMeta(sourceStake, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarStakeHistoryPubkey, false, false),
		solana.NewAccountMeta(stakeAuthority, false, true),
	}}}
}

func (*Merge) Data() ([]byte, error) {
	return []byte{byte(MergeInstruction), 0, 0, 0}, nil
}
