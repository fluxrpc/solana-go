package loaderv4

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type SetProgramLength struct {
	NewSize uint32
	instruction
}

func NewSetProgramLengthInstruction(
	newSize uint32,
	program, authority, recipient solana.PublicKey,
) *SetProgramLength {
	return &SetProgramLength{NewSize: newSize, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(authority, false, true),
		solana.NewAccountMeta(recipient, true, false),
	}}}
}

func NewSetProgramLengthWithoutRecipientInstruction(
	newSize uint32,
	program, authority solana.PublicKey,
) *SetProgramLength {
	return &SetProgramLength{NewSize: newSize, instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(program, true, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func (inst *SetProgramLength) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 8))
	enc.WriteUint32(uint32(SetProgramLengthInstruction))
	enc.WriteUint32(inst.NewSize)
	return enc.Bytes(), enc.Err()
}
