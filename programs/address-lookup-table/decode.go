package addresslookuptable

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

// DecodeInstruction decodes one Address Lookup Table instruction. The result
// retains the supplied account slice. Trailing bytes are accepted to match the
// native program's limited bincode deserialization.
func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	dec := bin.NewDecoder(data)
	typ := InstructionType(dec.ReadUint32())
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("address lookup instruction type: %w", err)
	}

	out := DecodedInstruction{Type: typ}
	switch typ {
	case CreateLookupTableInstruction:
		out.CreateLookupTable = &CreateLookupTable{
			RecentSlot:  dec.ReadUint64(),
			BumpSeed:    dec.ReadUint8(),
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case FreezeLookupTableInstruction:
		out.FreezeLookupTable = &FreezeLookupTable{instruction: instruction{AccountMetaSlice: accounts}}
	case ExtendLookupTableInstruction:
		count := dec.ReadUint64()
		if err := dec.Err(); err != nil {
			return DecodedInstruction{}, fmt.Errorf("decode address count: %w", err)
		}
		if count > LookupTableMaxAddresses {
			return DecodedInstruction{}, fmt.Errorf("%w: %d > %d", ErrTooManyAddresses, count, LookupTableMaxAddresses)
		}
		needed := int(count) * solana.PublicKeyLength
		if dec.Remaining() < needed {
			return DecodedInstruction{}, fmt.Errorf("decode %d lookup addresses: need %d bytes, have %d: %w", count, needed, dec.Remaining(), bin.ErrUnexpectedEOF)
		}
		addresses := make([]solana.PublicKey, int(count))
		for index := range addresses {
			addresses[index] = dec.ReadPublicKey()
		}
		out.ExtendLookupTable = &ExtendLookupTable{
			Addresses:   addresses,
			instruction: instruction{AccountMetaSlice: accounts},
		}
	case DeactivateLookupTableInstruction:
		out.DeactivateLookupTable = &DeactivateLookupTable{instruction: instruction{AccountMetaSlice: accounts}}
	case CloseLookupTableInstruction:
		out.CloseLookupTable = &CloseLookupTable{instruction: instruction{AccountMetaSlice: accounts}}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, uint32(typ))
	}
	if err := dec.Err(); err != nil {
		return DecodedInstruction{}, fmt.Errorf("decode address lookup %s: %w", typ, err)
	}
	return out, nil
}
