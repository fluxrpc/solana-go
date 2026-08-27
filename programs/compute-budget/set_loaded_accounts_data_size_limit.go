package computebudget

import (
	"github.com/fluxrpc/solana-go/binary"
)

// SetLoadedAccountsDataSizeLimit sets the transaction-wide account-data load
// limit in bytes.
type SetLoadedAccountsDataSizeLimit struct {
	instruction
	Bytes uint32
}

func NewSetLoadedAccountsDataSizeLimitInstruction(bytes uint32) *SetLoadedAccountsDataSizeLimit {
	return &SetLoadedAccountsDataSizeLimit{Bytes: bytes}
}

func (inst *SetLoadedAccountsDataSizeLimit) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 5))
	enc.WriteUint8(uint8(SetLoadedAccountsDataSizeLimitInstruction))
	enc.WriteUint32(inst.Bytes)
	return enc.Bytes(), enc.Err()
}
