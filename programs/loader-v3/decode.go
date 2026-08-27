package loaderv3

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := binary.NewDecoder(data)
	typ := InstructionType(dec.ReadUint32())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("loader-v3 instruction type: %w", err)
	}
	out := DecodedInstruction{Type: typ}
	switch typ {
	case InitializeBufferInstruction:
		out.InitializeBuffer = &InitializeBuffer{instruction{accounts}}
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
	case DeployWithMaxDataLenInstruction:
		maxDataLen := dec.ReadUint64()
		closeBuffer := true
		if dec.Remaining() > 0 {
			closeBuffer = dec.ReadBool()
		}
		out.DeployWithMaxDataLen = &DeployWithMaxDataLen{
			MaxDataLen: maxDataLen, CloseBuffer: closeBuffer, instruction: instruction{AccountMetaSlice: accounts},
		}
	case UpgradeInstruction:
		closeBuffer := true
		if dec.Remaining() > 0 {
			closeBuffer = dec.ReadBool()
		}
		out.Upgrade = &Upgrade{CloseBuffer: closeBuffer, instruction: instruction{AccountMetaSlice: accounts}}
	case SetAuthorityInstruction:
		out.SetAuthority = &SetAuthority{instruction{accounts}}
	case CloseInstruction:
		tombstone := false
		if dec.Remaining() > 0 {
			tombstone = dec.ReadBool()
		}
		out.Close = &Close{Tombstone: tombstone, instruction: instruction{AccountMetaSlice: accounts}}
	case ExtendProgramInstruction:
		out.ExtendProgram = &ExtendProgram{AdditionalBytes: dec.ReadUint32(), instruction: instruction{AccountMetaSlice: accounts}}
	case SetAuthorityCheckedInstruction:
		out.SetAuthorityChecked = &SetAuthorityChecked{instruction{accounts}}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, typ)
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode loader-v3 %s: %w", typ, err)
	}
	return out, nil
}
