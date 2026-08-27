package loaderv3

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Close struct {
	Tombstone bool
	instruction
}

func NewCloseInstruction(tombstone bool, closeAccount, recipient, authority solana.PublicKey) *Close {
	return &Close{Tombstone: tombstone, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(closeAccount, true, false),
		solana.NewAccountMeta(recipient, true, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func NewCloseProgramInstruction(
	tombstone bool,
	programData, recipient, authority, program solana.PublicKey,
) *Close {
	return &Close{Tombstone: tombstone, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(programData, true, false),
		solana.NewAccountMeta(recipient, true, false),
		solana.NewAccountMeta(authority, false, true),
		solana.NewAccountMeta(program, true, false),
	}}}
}

func NewCloseUninitializedInstruction(tombstone bool, closeAccount, recipient solana.PublicKey) *Close {
	return &Close{Tombstone: tombstone, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(closeAccount, true, false),
		solana.NewAccountMeta(recipient, true, false),
	}}}
}

func (inst *Close) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 5))
	enc.WriteUint32(uint32(CloseInstruction))
	enc.WriteBool(inst.Tombstone)
	return enc.Bytes(), enc.Err()
}
