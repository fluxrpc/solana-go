package stake

import solana "github.com/fluxrpc/solana-go"

type InitializeChecked struct{ instruction }

func NewInitializeCheckedInstruction(authorized Authorized, stakeAccount solana.PublicKey) *InitializeChecked {
	return &InitializeChecked{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(stakeAccount, true, false),
		solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
		solana.NewAccountMeta(authorized.Staker, false, false),
		solana.NewAccountMeta(authorized.Withdrawer, false, true),
	}}}
}

func (*InitializeChecked) Data() ([]byte, error) {
	return []byte{byte(InitializeCheckedInstruction), 0, 0, 0}, nil
}
