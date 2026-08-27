package loaderv3

import solana "github.com/fluxrpc/solana-go"

type SetAuthority struct{ instruction }

func NewSetAuthorityInstruction(
	target, currentAuthority, newAuthority solana.PublicKey,
) *SetAuthority {
	return &SetAuthority{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(target, true, false),
		solana.NewAccountMeta(currentAuthority, false, true),
		solana.NewAccountMeta(newAuthority, false, false),
	}}}
}

func NewRemoveAuthorityInstruction(target, currentAuthority solana.PublicKey) *SetAuthority {
	return &SetAuthority{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(target, true, false),
		solana.NewAccountMeta(currentAuthority, false, true),
	}}}
}

func (*SetAuthority) Data() ([]byte, error) {
	return []byte{byte(SetAuthorityInstruction), 0, 0, 0}, nil
}
