package stake

import solana "github.com/fluxrpc/solana-go"

type Deactivate struct{ instruction }

func NewDeactivateInstruction(stakeAccount, stakeAuthority solana.PublicKey) *Deactivate {
	return &Deactivate{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(solana.SysVarClockPubkey, false, false),
		solana.NewAccountMeta(stakeAuthority, false, true),
	}}}
}

func (*Deactivate) Data() ([]byte, error) {
	return []byte{byte(DeactivateInstruction), 0, 0, 0}, nil
}
