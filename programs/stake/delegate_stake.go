package stake

import solana "github.com/fluxrpc/solana-go"

type DelegateStake struct{ instruction }

func NewDelegateStakeInstruction(voteAccount, stakeAuthority, stakeAccount solana.PublicKey) *DelegateStake {
	return &DelegateStake{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(voteAccount, false, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarStakeHistoryPubkey, false, false),
		solana.NewAccountMeta(solana.SysVarStakeConfigPubkey, false, false),
		solana.NewAccountMeta(stakeAuthority, false, true),
	}}}
}

func (*DelegateStake) Data() ([]byte, error) {
	return []byte{byte(DelegateStakeInstruction), 0, 0, 0}, nil
}
