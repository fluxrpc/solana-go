package loaderv4

import solana "github.com/fluxrpc/solana-go"

type Retract struct{ instruction }

func NewRetractInstruction(program, authority solana.PublicKey) *Retract {
	return &Retract{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func (*Retract) Data() ([]byte, error) {
	return []byte{byte(RetractInstruction), 0, 0, 0}, nil
}
