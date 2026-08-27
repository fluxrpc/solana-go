package loaderv4

import solana "github.com/fluxrpc/solana-go"

type Deploy struct{ instruction }

func NewDeployInstruction(program, authority solana.PublicKey) *Deploy {
	return &Deploy{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func NewDeployFromSourceInstruction(program, authority, source solana.PublicKey) *Deploy {
	return &Deploy{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(authority, false, true),
		solana.NewAccountMeta(source, true, false),
	}}}
}

func (*Deploy) Data() ([]byte, error) {
	return []byte{byte(DeployInstruction), 0, 0, 0}, nil
}
