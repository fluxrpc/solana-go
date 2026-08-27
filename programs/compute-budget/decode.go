package computebudget

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

// DecodeInstruction decodes one Compute Budget instruction. Compute Budget
// instructions use no accounts. Trailing bytes are accepted to match the
// runtime's unchecked Borsh deserialization.
func DecodeInstruction(_ solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := bin.NewDecoder(data)
	typ := InstructionType(dec.ReadUint8())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("compute budget instruction type: %w", err)
	}

	out := DecodedInstruction{Type: typ}
	switch typ {
	case UnusedInstruction:
		if dec.Remaining() >= 8 {
			out.RequestUnitsDeprecated = &RequestUnitsDeprecated{
				Units:         dec.ReadUint32(),
				AdditionalFee: dec.ReadUint32(),
			}
		} else {
			out.Unused = &Unused{}
		}
	case RequestHeapFrameInstruction:
		out.RequestHeapFrame = &RequestHeapFrame{HeapSize: dec.ReadUint32()}
	case SetComputeUnitLimitInstruction:
		out.SetComputeUnitLimit = &SetComputeUnitLimit{Units: dec.ReadUint32()}
	case SetComputeUnitPriceInstruction:
		out.SetComputeUnitPrice = &SetComputeUnitPrice{MicroLamports: dec.ReadUint64()}
	case SetLoadedAccountsDataSizeLimitInstruction:
		out.SetLoadedAccountsDataSizeLimit = &SetLoadedAccountsDataSizeLimit{Bytes: dec.ReadUint32()}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, uint8(typ))
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode compute budget %s: %w", typ, err)
	}
	return out, nil
}
