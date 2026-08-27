package loaderv4

import solana "github.com/fluxrpc/solana-go"

type Finalize struct{ instruction }

func NewFinalizeInstruction(
	program, authority, nextVersionProgram solana.PublicKey,
) *Finalize {
	return &Finalize{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(authority, false, true),
		solana.NewAccountMeta(nextVersionProgram, false, false),
	}}}
}

func (*Finalize) Data() ([]byte, error) {
	return []byte{byte(FinalizeInstruction), 0, 0, 0}, nil
}
