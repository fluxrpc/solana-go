package loaderv2

import solana "github.com/fluxrpc/solana-go"

type Finalize struct{ instruction }

func NewFinalizeInstruction(program solana.PublicKey) *Finalize {
	return &Finalize{instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(program, true, true),
		solana.NewAccountMeta(solana.SysVarRentPubkey, false, false),
	}}}
}

func (*Finalize) Data() ([]byte, error) {
	return []byte{byte(FinalizeInstruction), 0, 0, 0}, nil
}
