package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// TransferWithSeed transfers lamports from an address derived from a base and
// seed.
type TransferWithSeed struct {
	Lamports  uint64
	FromSeed  string
	FromOwner solana.PublicKey
	instruction
}

// NewTransferWithSeedInstruction creates a TransferWithSeed instruction.
func NewTransferWithSeedInstruction(
	lamports uint64,
	fromSeed string,
	fromOwner solana.PublicKey,
	fundingAccount solana.PublicKey,
	baseForFundingAccount solana.PublicKey,
	recipientAccount solana.PublicKey,
) (*TransferWithSeed, error) {
	if len(fromSeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	return &TransferWithSeed{
		Lamports:  lamports,
		FromSeed:  fromSeed,
		FromOwner: fromOwner,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(fundingAccount, true, false),
			solana.NewAccountMeta(baseForFundingAccount, false, true),
			solana.NewAccountMeta(recipientAccount, true, false),
		}},
	}, nil
}

// Data encodes the instruction data.
func (inst *TransferWithSeed) Data() ([]byte, error) {
	if len(inst.FromSeed) > solana.MaxSeedLength {
		return nil, solana.ErrMaxSeedLengthExceeded
	}
	enc := binary.NewEncoder(make([]byte, 0, 52+len(inst.FromSeed)))
	enc.WriteUint32(uint32(TransferWithSeedInstruction))
	enc.WriteUint64(inst.Lamports)
	enc.WriteBincodeString(inst.FromSeed)
	enc.WritePublicKey(inst.FromOwner)
	return enc.Bytes(), enc.Err()
}
