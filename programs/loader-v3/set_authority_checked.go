package loaderv3

import solana "github.com/fluxrpc/solana-go"

type SetAuthorityChecked struct{ instruction }

func NewSetAuthorityCheckedInstruction(
	target, currentAuthority, newAuthority solana.PublicKey,
) *SetAuthorityChecked {
	return &SetAuthorityChecked{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(target, true, false),
		solana.NewAccountMeta(currentAuthority, false, true),
		solana.NewAccountMeta(newAuthority, false, true),
	}}}
}

func (*SetAuthorityChecked) Data() ([]byte, error) {
	return []byte{byte(SetAuthorityCheckedInstruction), 0, 0, 0}, nil
}
