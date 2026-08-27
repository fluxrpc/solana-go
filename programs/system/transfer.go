package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// Transfer moves lamports between accounts.
type Transfer struct {
	// Lamports is the number of lamports transferred to the recipient.
	Lamports uint64
	// [0] Funding account: writable, signer.
	// [1] Recipient account: writable.
	instruction
}

// NewTransferInstruction creates a System Program Transfer instruction.
func NewTransferInstruction(
	lamports uint64,
	fundingAccount solana.PublicKey,
	recipientAccount solana.PublicKey,
) *Transfer {
	return &Transfer{
		Lamports: lamports,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(fundingAccount, true, true),
			solana.NewAccountMeta(recipientAccount, true, false),
		}},
	}
}

// Data returns the instruction's binary-encoded data.
func (inst *Transfer) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 4+8))
	enc.WriteUint32(uint32(TransferInstruction))
	enc.WriteUint64(inst.Lamports)
	return enc.Bytes(), enc.Err()
}
