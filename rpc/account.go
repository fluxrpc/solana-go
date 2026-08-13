package rpc

import (
	"math/big"

	solana "github.com/fluxrpc/solana-go"
)

// Account describes an on-chain account as returned by the RPC.
type Account struct {
	// Number of lamports assigned to this account.
	Lamports uint64 `json:"lamports"`

	// Pubkey of the program this account has been assigned to.
	Owner solana.PublicKey `json:"owner"`

	// Data associated with the account, either as encoded binary data or in
	// JSON format {<program>: <state>}, depending on the encoding parameter.
	Data *DataBytesOrJSON `json:"data"`

	// Boolean indicating if the account contains a program (and is strictly read-only).
	Executable bool `json:"executable"`

	// The epoch at which this account will next owe rent.
	RentEpoch *big.Int `json:"rentEpoch"`

	// The amount of storage space required to store the token account.
	Space uint64 `json:"space"`
}

// KeyedAccount pairs an account with its public key.
type KeyedAccount struct {
	Pubkey  solana.PublicKey `json:"pubkey"`
	Account *Account         `json:"account"`
}
