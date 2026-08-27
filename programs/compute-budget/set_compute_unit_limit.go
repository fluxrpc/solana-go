package computebudget

import (
	"github.com/fluxrpc/solana-go/binary"
)

// SetComputeUnitLimit sets the transaction compute-unit limit.
type SetComputeUnitLimit struct {
	instruction
	Units uint32
}

func NewSetComputeUnitLimitInstruction(units uint32) *SetComputeUnitLimit {
	return &SetComputeUnitLimit{Units: units}
}

func (inst *SetComputeUnitLimit) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 5))
	enc.WriteUint8(uint8(SetComputeUnitLimitInstruction))
	enc.WriteUint32(inst.Units)
	return enc.Bytes(), enc.Err()
}
