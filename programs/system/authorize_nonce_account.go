package system

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/binary"
)

// AuthorizeNonceAccount changes the authority of a durable nonce account.
type AuthorizeNonceAccount struct {
	Authorized solana.PublicKey
	instruction
}

// NewAuthorizeNonceAccountInstruction creates an instruction that changes a
// durable nonce account's authority.
func NewAuthorizeNonceAccountInstruction(
	authorized solana.PublicKey,
	nonceAccount solana.PublicKey,
	nonceAuthority solana.PublicKey,
) *AuthorizeNonceAccount {
	return &AuthorizeNonceAccount{
		Authorized: authorized,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(nonceAccount, true, false),
			solana.NewAccountMeta(nonceAuthority, false, true),
		}},
	}
}

// Data returns the encoded System Program instruction data.
func (inst *AuthorizeNonceAccount) Data() ([]byte, error) {
	enc := binary.NewEncoder(make([]byte, 0, 36))
	enc.WriteUint32(uint32(AuthorizeNonceAccountInstruction))
	enc.WritePublicKey(inst.Authorized)
	return enc.Bytes(), enc.Err()
}
