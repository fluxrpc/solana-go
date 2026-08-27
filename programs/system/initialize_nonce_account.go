package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// InitializeNonceAccount initializes a durable nonce account and records its
// authority.
type InitializeNonceAccount struct {
	Authorized solana.PublicKey
	instruction
}

// NewInitializeNonceAccountInstruction creates an instruction that initializes
// a durable nonce account.
func NewInitializeNonceAccountInstruction(
	authorized solana.PublicKey,
	nonceAccount solana.PublicKey,
	recentBlockhashesSysvar solana.PublicKey,
	rentSysvar solana.PublicKey,
) *InitializeNonceAccount {
	return &InitializeNonceAccount{
		Authorized: authorized,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(nonceAccount, true, false),
			solana.NewAccountMeta(recentBlockhashesSysvar, false, false),
			solana.NewAccountMeta(rentSysvar, false, false),
		}},
	}
}

// Data returns the encoded System Program instruction data.
func (inst *InitializeNonceAccount) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 36))
	enc.WriteUint32(uint32(InitializeNonceAccountInstruction))
	enc.WritePublicKey(inst.Authorized)
	return enc.Bytes(), enc.Err()
}
