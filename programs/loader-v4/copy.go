package loaderv4

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Copy struct {
	DestinationOffset uint32
	SourceOffset      uint32
	Length            uint32
	instruction
}

func NewCopyInstruction(
	destinationOffset, sourceOffset, length uint32,
	program, authority, source solana.PublicKey,
) *Copy {
	return &Copy{
		DestinationOffset: destinationOffset,
		SourceOffset:      sourceOffset,
		Length:            length,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(program, true, false),
			solana.NewAccountMeta(authority, false, true),
			solana.NewAccountMeta(source, false, false),
		}},
	}
}

func (inst *Copy) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 16))
	enc.WriteUint32(uint32(CopyInstruction))
	enc.WriteUint32(inst.DestinationOffset)
	enc.WriteUint32(inst.SourceOffset)
	enc.WriteUint32(inst.Length)
	return enc.Bytes(), enc.Err()
}
