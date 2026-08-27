package loaderv4

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := binary.NewDecoder(data)
	typ := InstructionType(dec.ReadUint32())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("loader-v4 instruction type: %w", err)
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
	case CopyInstruction:
		out.Copy = &Copy{
			DestinationOffset: dec.ReadUint32(),
			SourceOffset:      dec.ReadUint32(),
			Length:            dec.ReadUint32(),
			instruction:       instruction{AccountMetaSlice: accounts},
		}
	case SetProgramLengthInstruction:
		out.SetProgramLength = &SetProgramLength{NewSize: dec.ReadUint32(), instruction: instruction{AccountMetaSlice: accounts}}
	case DeployInstruction:
		out.Deploy = &Deploy{instruction{accounts}}
	case RetractInstruction:
		out.Retract = &Retract{instruction{accounts}}
	case TransferAuthorityInstruction:
		out.TransferAuthority = &TransferAuthority{instruction{accounts}}
	case FinalizeInstruction:
		out.Finalize = &Finalize{instruction{accounts}}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, typ)
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode loader-v4 %s: %w", typ, err)
	}
	return out, nil
}
