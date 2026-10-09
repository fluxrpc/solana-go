package associatedtokenaccount

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

// DecodeInstruction decodes one Associated Token Account instruction. Empty
// data is the legacy encoding of Create.
func DecodeInstruction(accounts solana.AccountMetaSlice, data []byte) (DecodedInstruction, error) {
	typ := CreateInstruction
	if len(data) > 0 {
		typ = InstructionType(data[0])
	}

	base := instruction{AccountMetaSlice: accounts}
	out := DecodedInstruction{Type: typ}
	switch typ {
	case CreateInstruction:
		out.Create = &Create{base}
	case CreateIdempotentInstruction:
		out.CreateIdempotent = &CreateIdempotent{base}
	case RecoverNestedInstruction:
		out.RecoverNested = &RecoverNested{base}
	default:
		return DecodedInstruction{}, fmt.Errorf("%w: %d", ErrUnknownInstruction, uint8(typ))
	}
	return out, nil
}
