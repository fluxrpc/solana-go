package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// WithdrawNonceAccount withdraws lamports from a durable nonce account.
type WithdrawNonceAccount struct {
	Lamports uint64
	instruction
}

// NewWithdrawNonceAccountInstruction creates an instruction that withdraws
// lamports from a durable nonce account.
func NewWithdrawNonceAccountInstruction(
	lamports uint64,
	nonceAccount solana.PublicKey,
	recipientAccount solana.PublicKey,
	recentBlockhashesSysvar solana.PublicKey,
	rentSysvar solana.PublicKey,
	nonceAuthority solana.PublicKey,
) *WithdrawNonceAccount {
	return &WithdrawNonceAccount{
		Lamports: lamports,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(nonceAccount, true, false),
			solana.NewAccountMeta(recipientAccount, true, false),
			solana.NewAccountMeta(recentBlockhashesSysvar, false, false),
			solana.NewAccountMeta(rentSysvar, false, false),
			solana.NewAccountMeta(nonceAuthority, false, true),
		}},
	}
}

// Data returns the encoded System Program instruction data.
func (inst *WithdrawNonceAccount) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 12))
	enc.WriteUint32(uint32(WithdrawNonceAccountInstruction))
	enc.WriteUint64(inst.Lamports)
	return enc.Bytes(), enc.Err()
}
