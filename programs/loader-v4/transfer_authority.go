package loaderv4

import solana "github.com/fluxrpc/solana-go"

type TransferAuthority struct{ instruction }

func NewTransferAuthorityInstruction(
	program, currentAuthority, newAuthority solana.PublicKey,
) *TransferAuthority {
	return &TransferAuthority{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(currentAuthority, false, true),
		solana.NewAccountMeta(newAuthority, false, true),
	}}}
}

func (*TransferAuthority) Data() ([]byte, error) {
	return []byte{byte(TransferAuthorityInstruction), 0, 0, 0}, nil
}
