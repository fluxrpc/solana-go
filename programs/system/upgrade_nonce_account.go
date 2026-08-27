package system

import solana "github.com/fluxrpc/solana-go"

// UpgradeNonceAccount performs the one-time, idempotent upgrade of a legacy
// durable nonce account.
type UpgradeNonceAccount struct {
	instruction
}

// NewUpgradeNonceAccountInstruction creates an instruction that upgrades a
// legacy durable nonce account.
func NewUpgradeNonceAccountInstruction(nonceAccount solana.PublicKey) *UpgradeNonceAccount {
	return &UpgradeNonceAccount{instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(nonceAccount, true, false),
	}}}
}

// Data returns the encoded System Program instruction data.
func (*UpgradeNonceAccount) Data() ([]byte, error) {
	return []byte{byte(UpgradeNonceAccountInstruction), 0, 0, 0}, nil
}
