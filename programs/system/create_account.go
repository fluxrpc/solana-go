package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// CreateAccount creates and funds an account owned by a program.
type CreateAccount struct {
	// Lamports is the number of lamports transferred to the new account.
	Lamports uint64
	// Space is the number of bytes allocated to the new account.
	Space uint64
	// Owner is the program that will own the new account.
	Owner solana.PublicKey
	// [0] Funding account: writable, signer.
	// [1] New account: writable, signer.
	instruction
}

// NewCreateAccountInstruction creates a System Program CreateAccount instruction.
func NewCreateAccountInstruction(
	lamports uint64,
	space uint64,
	owner solana.PublicKey,
	fundingAccount solana.PublicKey,
	newAccount solana.PublicKey,
) *CreateAccount {
	return &CreateAccount{
		Lamports: lamports,
		Space:    space,
		Owner:    owner,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(fundingAccount, true, true),
			solana.NewAccountMeta(newAccount, true, true),
		}},
	}
}

// Data returns the instruction's binary-encoded data.
func (inst *CreateAccount) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 4+8+8+solana.PublicKeyLength))
	enc.WriteUint32(uint32(CreateAccountInstruction))
	enc.WriteUint64(inst.Lamports)
	enc.WriteUint64(inst.Space)
	enc.WritePublicKey(inst.Owner)
	return enc.Bytes(), enc.Err()
}
