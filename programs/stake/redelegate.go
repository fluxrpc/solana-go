package stake

import solana "github.com/fluxrpc/solana-go"

// Redelegate is retained for wire compatibility. The native variant is not
// enabled by the current Stake Program.
type Redelegate struct{ instruction }

func NewRedelegateInstruction(stakeAccount, uninitializedStakeAccount, voteAccount, stakeAuthority solana.PublicKey) *Redelegate {
	return &Redelegate{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(uninitializedStakeAccount, true, false),
		solana.NewAccountMeta(voteAccount, false, false),
		solana.NewAccountMeta(solana.SysVarStakeConfigPubkey, false, false),
		solana.NewAccountMeta(stakeAuthority, false, true),
	}}}
}

func (*Redelegate) Data() ([]byte, error) {
	return []byte{byte(RedelegateInstruction), 0, 0, 0}, nil
}
