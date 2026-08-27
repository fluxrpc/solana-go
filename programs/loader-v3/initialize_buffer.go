package loaderv3

import solana "github.com/fluxrpc/solana-go"

type InitializeBuffer struct{ instruction }

func NewInitializeBufferInstruction(buffer, authority solana.PublicKey) *InitializeBuffer {
	return &InitializeBuffer{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(buffer, true, false),
		solana.NewAccountMeta(authority, false, false),
	}}}
}

func NewInitializeImmutableBufferInstruction(buffer solana.PublicKey) *InitializeBuffer {
	return &InitializeBuffer{instruction{solana.AccountMetaSlice{
		solana.NewAccountMeta(buffer, true, false),
	}}}
}

func (*InitializeBuffer) Data() ([]byte, error) {
	return []byte{byte(InitializeBufferInstruction), 0, 0, 0}, nil
}
