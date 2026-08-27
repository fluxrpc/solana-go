package computebudget

import (
	"github.com/fluxrpc/solana-go/binary"
)

// SetComputeUnitPrice sets the transaction compute-unit price in
// micro-lamports.
type SetComputeUnitPrice struct {
	instruction
	MicroLamports uint64
}

func NewSetComputeUnitPriceInstruction(microLamports uint64) *SetComputeUnitPrice {
	return &SetComputeUnitPrice{MicroLamports: microLamports}
}

func (inst *SetComputeUnitPrice) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 9))
	enc.WriteUint8(uint8(SetComputeUnitPriceInstruction))
	enc.WriteUint64(inst.MicroLamports)
	return enc.Bytes(), enc.Err()
}
