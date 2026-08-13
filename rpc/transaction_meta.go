package rpc

import ()

// TransactionMeta is the transaction status metadata object returned with
// binary/base64/base58/json transaction encodings.
type TransactionMeta struct {
	// Error if transaction failed, null if transaction succeeded.
	// https://github.com/solana-labs/solana/blob/master/sdk/src/transaction.rs#L24
	Err any `json:"err"`

	// Fee this transaction was charged
	Fee uint64 `json:"fee"`

	// Array of u64 account balances from before the transaction was processed
	PreBalances []uint64 `json:"preBalances"`

	// Array of u64 account balances after the transaction was processed
	PostBalances []uint64 `json:"postBalances"`

	// List of inner instructions or omitted if inner instruction recording
	// was not yet enabled during this transaction
	InnerInstructions []InnerInstruction `json:"innerInstructions"`

	// List of token balances from before the transaction was processed
	// or omitted if token balance recording was not yet enabled during this transaction
	PreTokenBalances []TokenBalance `json:"preTokenBalances"`

	// List of token balances from after the transaction was processed
	// or omitted if token balance recording was not yet enabled during this transaction
	PostTokenBalances []TokenBalance `json:"postTokenBalances"`

	// Array of string log messages or omitted if log message
	// recording was not yet enabled during this transaction
	LogMessages []string `json:"logMessages"`

	// DEPRECATED: Transaction status.
	Status DeprecatedTransactionMetaStatus `json:"status"`

	Rewards []BlockReward `json:"rewards"`

	LoadedAddresses LoadedAddresses `json:"loadedAddresses"`

	ReturnData ReturnData `json:"returnData"`

	ComputeUnitsConsumed *uint64 `json:"computeUnitsConsumed"`

	// Transaction cost units, as reported by newer validators.
	CostUnits *uint64 `json:"costUnits,omitempty"`
}

// DeprecatedTransactionMetaStatus is the deprecated transaction status:
//
//	Ok  interface{} `json:"Ok"`  // <null> Transaction was successful
//	Err interface{} `json:"Err"` // Transaction failed with TransactionError
type DeprecatedTransactionMetaStatus M
