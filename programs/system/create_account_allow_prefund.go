package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// CreateAccountAllowPrefund creates an account without requiring its existing
// lamport balance to be zero.
type CreateAccountAllowPrefund struct {
	// Lamports is the number of additional lamports transferred to the account.
	Lamports uint64
	// Space is the number of bytes allocated to the account.
	Space uint64
	// Owner is the program that will own the account.
	Owner solana.PublicKey
	// [0] New account: writable, signer.
	// [1] Optional funding account: writable, signer.
	instruction
}

// NewCreateAccountAllowPrefundInstruction creates a System Program
// CreateAccountAllowPrefund instruction which transfers lamports from the
// funding account.
func NewCreateAccountAllowPrefundInstruction(
	lamports uint64,
	space uint64,
	owner solana.PublicKey,
	newAccount solana.PublicKey,
	fundingAccount solana.PublicKey,
) *CreateAccountAllowPrefund {
	return &CreateAccountAllowPrefund{
		Lamports: lamports,
		Space:    space,
		Owner:    owner,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(newAccount, true, true),
			solana.NewAccountMeta(fundingAccount, true, true),
		}},
	}
}

// NewCreateAccountAllowPrefundWithoutFundingInstruction creates a System
// Program CreateAccountAllowPrefund instruction which does not transfer
// additional lamports and therefore omits the funding account.
func NewCreateAccountAllowPrefundWithoutFundingInstruction(
	space uint64,
	owner solana.PublicKey,
	newAccount solana.PublicKey,
) *CreateAccountAllowPrefund {
	return &CreateAccountAllowPrefund{
		Space: space,
		Owner: owner,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(newAccount, true, true),
		}},
	}
}

// Data returns the instruction's binary-encoded data.
func (inst *CreateAccountAllowPrefund) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 4+8+8+solana.PublicKeyLength))
	enc.WriteUint32(uint32(CreateAccountAllowPrefundInstruction))
	enc.WriteUint64(inst.Lamports)
	enc.WriteUint64(inst.Space)
	enc.WritePublicKey(inst.Owner)
	return enc.Bytes(), enc.Err()
}
