package loaderv2

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := binary.NewDecoder(data)
	typ := InstructionType(dec.ReadUint32())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("loader-v2 instruction type: %w", err)
	}
	out := DecodedInstruction{Type: typ}
	switch typ {
	case WriteInstruction:
		offset := dec.ReadUint32()
		length := dec.ReadUint64()
		if length > uint64(dec.Remaining()) {
			dec.ReadBytes(dec.Remaining() + 1)
		} else {
			out.Write = &Write{
				Offset:      offset,
				Bytes:       dec.ReadBytes(int(length)),
				instruction: instruction{AccountMetaSlice: accounts},
			}
		}
	case FinalizeInstruction:
		out.Finalize = &Finalize{instruction: instruction{AccountMetaSlice: accounts}}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, typ)
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode loader-v2 %s: %w", typ, err)
	}
	return out, nil
}
