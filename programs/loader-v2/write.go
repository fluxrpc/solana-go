package loaderv2

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

type Write struct {
	Offset uint32
	Bytes  []byte
	instruction
}

func NewWriteInstruction(offset uint32, data []byte, program solana.PublicKey) *Write {
	return &Write{
		Offset: offset,
		Bytes:  data,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(program, true, true),
		}},
	}
}

func (inst *Write) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 4+4+8+len(inst.Bytes)))
	enc.WriteUint32(uint32(WriteInstruction))
	enc.WriteUint32(inst.Offset)
	enc.WriteUint64(uint64(len(inst.Bytes)))
	enc.WriteBytes(inst.Bytes)
	return enc.Bytes(), enc.Err()
}
