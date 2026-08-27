package computebudget

import (
	"github.com/fluxrpc/solana-go/binary"
)

// RequestUnitsDeprecated is the retired request-units instruction retained
// for decoding and compatibility with historical transactions.
type RequestUnitsDeprecated struct {
	instruction
	Units         uint32
	AdditionalFee uint32
}

func NewRequestUnitsDeprecatedInstruction(units, additionalFee uint32) *RequestUnitsDeprecated {
	return &RequestUnitsDeprecated{Units: units, AdditionalFee: additionalFee}
}

func (inst *RequestUnitsDeprecated) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 9))
	enc.WriteUint8(uint8(RequestUnitsDeprecatedInstruction))
	enc.WriteUint32(inst.Units)
	enc.WriteUint32(inst.AdditionalFee)
	return enc.Bytes(), enc.Err()
}
