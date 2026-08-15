package rpc

import (
	"fmt"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
)

// TransactionWithMeta is a confirmed transaction plus its status metadata,
// as returned by getTransaction and inside getBlock responses.
type TransactionWithMeta struct {
	// The slot this transaction was processed in.
	Slot uint64 `json:"slot"`

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch)
	// of when the transaction was processed.
	// Nil if not available.
	BlockTime *solana.UnixTimeSeconds `json:"blockTime"`

	Transaction *DataBytesOrJSON `json:"transaction"`

	// Transaction status metadata object
	Meta    *TransactionMeta   `json:"meta,omitempty"`
	Version TransactionVersion `json:"version"`
}

// GetParsedTransaction decodes the transaction from its jsonParsed form.
func (twm TransactionWithMeta) GetParsedTransaction() (*solana.Transaction, error) {
	if twm.Transaction == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	raw := twm.Transaction.GetRawJSON()
	if raw == nil {
		return nil, fmt.Errorf("data is not in JSONParsed encoding")
	}
	var parsedTransaction solana.Transaction
	if err := sonic.Unmarshal(raw, &parsedTransaction); err != nil {
		return nil, err
	}
	return &parsedTransaction, nil
}

// MustGetTransaction decodes the transaction, panicking on error.
func (twm TransactionWithMeta) MustGetTransaction() *solana.Transaction {
	tx, err := twm.GetTransaction()
	if err != nil {
		panic(err)
	}
	return tx
}

// GetTransaction decodes the transaction from either its JSON or its binary
// (base64/base58) representation, whichever the RPC returned.
func (twm TransactionWithMeta) GetTransaction() (*solana.Transaction, error) {
	if twm.Transaction == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	// EncodingJSON: the RPC returned a JSON object — unmarshal directly.
	if raw := twm.Transaction.GetRawJSON(); raw != nil {
		var tx solana.Transaction
		if err := sonic.Unmarshal(raw, &tx); err != nil {
			return nil, err
		}
		if tx.Message.AccountKeys == nil {
			return nil, fmt.Errorf("transaction has no message: block may have been fetched with transactionDetails=accounts; use GetAccountKeys instead")
		}
		return &tx, nil
	}
	return solana.TransactionFromBytes(twm.Transaction.GetBinary())
}

// GetAccountKeys returns the account keys when the block was fetched with
// transactionDetails=accounts. In this mode the transaction field contains
// {"signatures": [...], "accountKeys": [...]} instead of the full encoded transaction.
func (twm TransactionWithMeta) GetAccountKeys() (*TransactionAccountKeys, error) {
	if twm.Transaction == nil {
		return nil, fmt.Errorf("transaction is nil")
	}
	raw := twm.Transaction.GetRawJSON()
	if raw == nil {
		return nil, fmt.Errorf("transaction is not JSON (accounts mode requires transactionDetails=accounts)")
	}
	var out TransactionAccountKeys
	if err := sonic.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("failed to unmarshal account keys: %w", err)
	}
	return &out, nil
}

// TransactionAccountKeys is the transaction representation returned when
// transactionDetails is "accounts". Instead of the full message, it contains
// only the signatures and the list of account keys with their roles.
type TransactionAccountKeys struct {
	Signatures  []solana.Signature `json:"signatures"`
	AccountKeys []AccountKey       `json:"accountKeys"`
}

// AccountKey represents a single account involved in a transaction,
// as returned in the "accounts" transaction detail mode.
type AccountKey struct {
	// The account's public key.
	Pubkey solana.PublicKey `json:"pubkey"`

	// Whether this account signed the transaction.
	Signer bool `json:"signer"`

	// Whether the transaction marks this account as writable.
	Writable bool `json:"writable"`

	// The source of the account key: "transaction" for keys from the
	// message itself, "lookupTable" for keys resolved from address
	// lookup tables. Nil for legacy transactions.
	Source *AccountKeySource `json:"source,omitempty"`
}

// AccountKeySource says where an account key was loaded from.
type AccountKeySource string

const (
	// AccountKeySourceTransaction marks a key listed in the transaction
	// message itself.
	AccountKeySourceTransaction AccountKeySource = "transaction"
	// AccountKeySourceLookupTable marks a key resolved from an address
	// lookup table.
	AccountKeySourceLookupTable AccountKeySource = "lookupTable"
)

// TransactionParsed is a decoded transaction plus its status metadata.
type TransactionParsed struct {
	Meta        *TransactionMeta    `json:"meta,omitempty"`
	Transaction *solana.Transaction `json:"transaction"`
}
