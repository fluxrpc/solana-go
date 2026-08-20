package rpc

import (
	"bytes"
	"encoding/json"

	solana "github.com/fluxrpc/solana-go"
)

// TransactionsForAddressSortOrder controls the order in which
// getTransactionsForAddress returns results.
type TransactionsForAddressSortOrder string

const (
	// TransactionsForAddressSortDesc returns the newest transactions first
	// (the default when no sort order is provided).
	TransactionsForAddressSortDesc TransactionsForAddressSortOrder = "desc"
	// TransactionsForAddressSortAsc returns the oldest transactions first.
	TransactionsForAddressSortAsc TransactionsForAddressSortOrder = "asc"
)

// GetTransactionsForAddressOpts is the optional configuration object for the
// getTransactionsForAddress method.
type GetTransactionsForAddressOpts struct {
	// Level of transaction detail to return: "signatures" (default) or "full".
	TransactionDetails TransactionDetailsType `json:"transactionDetails,omitempty"`

	// Order of the returned transactions: "desc" (newest first, the default)
	// or "asc" (oldest first).
	SortOrder TransactionsForAddressSortOrder `json:"sortOrder,omitempty"`

	// (optional) Maximum number of transactions to return (between 1 and 1,000,
	// default: 1,000).
	Limit *int `json:"limit,omitempty"`

	// (optional) Cursor returned as PaginationToken by a previous response,
	// in the "slot:position" format, used to fetch the next page.
	PaginationToken string `json:"paginationToken,omitempty"`

	// Before requests transactions older than this signature. This is
	// supported by FluxRPC's rpc_shard compatibility endpoint.
	Before *solana.Signature `json:"before,omitempty"`

	// Until requests transactions newer than this signature. This is
	// supported by FluxRPC's rpc_shard compatibility endpoint.
	Until *solana.Signature `json:"until,omitempty"`

	// (optional) Commitment; "processed" is not supported.
	// If not provided, the default is "finalized".
	Commitment CommitmentType `json:"commitment,omitempty"`

	// (optional) Encoding for the returned transactions when TransactionDetails
	// is "full": "json", "jsonParsed", "base58" (slow), or "base64".
	Encoding solana.EncodingType `json:"encoding,omitempty"`

	// (optional) The max transaction version to return in responses.
	// Set to 0 to return all (legacy and versioned) transactions; if a returned
	// block contains a transaction with a higher version, an error is returned.
	MaxSupportedTransactionVersion *uint64 `json:"maxSupportedTransactionVersion,omitempty"`

	// (optional) The minimum slot that the request can be evaluated at.
	MinContextSlot *uint64 `json:"minContextSlot,omitempty"`

	// (optional) Server-side filters narrowing the returned transactions.
	Filters *TransactionsForAddressFilters `json:"filters,omitempty"`
}

// TransactionStatus filters returned transactions by execution status.
type TransactionStatus string

const (
	// TransactionStatusSucceeded matches transactions that executed successfully.
	TransactionStatusSucceeded TransactionStatus = "succeeded"
	// TransactionStatusFailed matches transactions that failed on chain.
	TransactionStatusFailed TransactionStatus = "failed"
	// TransactionStatusAny matches transactions regardless of status (the default).
	TransactionStatusAny TransactionStatus = "any"
)

// TokenAccountsFilter narrows results by the queried address's token-account
// activity within each transaction.
type TokenAccountsFilter string

const (
	// TokenAccountsNone matches transactions with no token-account activity.
	TokenAccountsNone TokenAccountsFilter = "none"
	// TokenAccountsBalanceChanged matches transactions where a token balance changed.
	TokenAccountsBalanceChanged TokenAccountsFilter = "balanceChanged"
	// TokenAccountsAll matches transactions with any token-account activity.
	TokenAccountsAll TokenAccountsFilter = "all"
)

// TokenTransferDirection constrains a token-transfer filter to a direction
// relative to the queried address.
type TokenTransferDirection string

const (
	// TokenTransferIn matches transfers into the queried address.
	TokenTransferIn TokenTransferDirection = "in"
	// TokenTransferOut matches transfers out of the queried address.
	TokenTransferOut TokenTransferDirection = "out"
	// TokenTransferAny matches transfers in either direction (the default).
	TokenTransferAny TokenTransferDirection = "any"
)

// TransactionsForAddressFilters narrows the result set of
// getTransactionsForAddress on the server side.
type TransactionsForAddressFilters struct {
	// Slot-number range comparisons.
	Slot *RangeFilterUint64 `json:"slot,omitempty"`

	// Block-time (Unix timestamp) range comparisons.
	BlockTime *RangeFilterInt64 `json:"blockTime,omitempty"`

	// Signature-string range comparisons.
	Signature *RangeFilterString `json:"signature,omitempty"`

	// Execution status: "succeeded", "failed", or "any".
	Status TransactionStatus `json:"status,omitempty"`

	// Token-account activity: "none", "balanceChanged", or "all".
	TokenAccounts TokenAccountsFilter `json:"tokenAccounts,omitempty"`

	// Token-transfer constraints.
	TokenTransfer *TokenTransferFilter `json:"tokenTransfer,omitempty"`
}

// RangeFilterUint64 expresses greater/less-than(-or-equal) bounds on a uint64.
type RangeFilterUint64 struct {
	Gte *uint64 `json:"gte,omitempty"`
	Gt  *uint64 `json:"gt,omitempty"`
	Lte *uint64 `json:"lte,omitempty"`
	Lt  *uint64 `json:"lt,omitempty"`
}

// RangeFilterInt64 expresses greater/less-than(-or-equal) and equality bounds
// on an int64 (used for Unix timestamps).
type RangeFilterInt64 struct {
	Gte *int64 `json:"gte,omitempty"`
	Gt  *int64 `json:"gt,omitempty"`
	Lte *int64 `json:"lte,omitempty"`
	Lt  *int64 `json:"lt,omitempty"`
	Eq  *int64 `json:"eq,omitempty"`
}

// RangeFilterString expresses greater/less-than(-or-equal) bounds on a string.
type RangeFilterString struct {
	Gte string `json:"gte,omitempty"`
	Gt  string `json:"gt,omitempty"`
	Lte string `json:"lte,omitempty"`
	Lt  string `json:"lt,omitempty"`
}

// TokenTransferFilter constrains results to transactions matching a token
// transfer against a counterparty, mint, direction and/or amount.
type TokenTransferFilter struct {
	// Counterparty address.
	With string `json:"with,omitempty"`
	// Transfer direction relative to the queried address: "in", "out", or "any".
	Direction TokenTransferDirection `json:"direction,omitempty"`
	// Token mint address.
	Mint string `json:"mint,omitempty"`
	// Amount range comparisons.
	Amount *RangeFilterUint64 `json:"amount,omitempty"`
}

// GetTransactionsForAddressResult is the response of getTransactionsForAddress.
type GetTransactionsForAddressResult struct {
	// The transactions matching the query, in the requested sort order.
	Data []*TransactionForAddress `json:"data"`

	// Opaque cursor ("slot:position") for the next page, or nil when there are
	// no more results. Pass it back via GetTransactionsForAddressOpts.PaginationToken.
	PaginationToken *string `json:"paginationToken"`
}

// UnmarshalJSON accepts both FluxRPC response generations: the documented
// {data,paginationToken} object and rpc_shard's legacy top-level array.
func (r *GetTransactionsForAddressResult) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '[' {
		var transactions []*TransactionForAddress
		if err := json.Unmarshal(data, &transactions); err != nil {
			return err
		}
		r.Data = transactions
		r.PaginationToken = nil
		return nil
	}
	type plain GetTransactionsForAddressResult
	return json.Unmarshal(data, (*plain)(r))
}

// TransactionForAddress is a single entry of a getTransactionsForAddress response.
// Which fields are populated depends on the requested TransactionDetails level.
type TransactionForAddress struct {
	// The slot that contains the block with the transaction.
	Slot uint64 `json:"slot"`

	// The transaction's index within the block. This field is specific to
	// getTransactionsForAddress.
	TransactionIndex uint64 `json:"transactionIndex"`

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch)
	// of when the transaction was processed. Nil if not available.
	BlockTime *solana.UnixTimeSeconds `json:"blockTime"`

	// Fields below are present when TransactionDetails is "signatures" (default):

	// Transaction signature.
	Signature solana.Signature `json:"signature,omitempty"`

	// Error if the transaction failed, nil if it succeeded.
	Err any `json:"err,omitempty"`

	// Memo associated with the transaction, nil if none.
	Memo *string `json:"memo,omitempty"`

	// The transaction's cluster confirmation status.
	ConfirmationStatus ConfirmationStatusType `json:"confirmationStatus,omitempty"`

	// Fields below are present when TransactionDetails is "full":

	// The decoded transaction, honoring the requested Encoding.
	Transaction *DataBytesOrJSON `json:"transaction,omitempty"`

	// Transaction status metadata object.
	Meta *TransactionMeta `json:"meta,omitempty"`
}

// GetParsedTransactionsForAddressResult is the jsonParsed form of
// GetTransactionsForAddressResult.
type GetParsedTransactionsForAddressResult struct {
	Data            []*ParsedTransactionForAddress `json:"data"`
	PaginationToken *string                        `json:"paginationToken"`
}

// UnmarshalJSON accepts both the documented object response and rpc_shard's
// legacy top-level array.
func (r *GetParsedTransactionsForAddressResult) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) > 0 && data[0] == '[' {
		var transactions []*ParsedTransactionForAddress
		if err := json.Unmarshal(data, &transactions); err != nil {
			return err
		}
		r.Data = transactions
		r.PaginationToken = nil
		return nil
	}
	type plain GetParsedTransactionsForAddressResult
	return json.Unmarshal(data, (*plain)(r))
}

// ParsedTransactionForAddress is one jsonParsed transaction-history entry.
type ParsedTransactionForAddress struct {
	Slot             uint64                  `json:"slot"`
	TransactionIndex uint64                  `json:"transactionIndex"`
	BlockTime        *solana.UnixTimeSeconds `json:"blockTime"`

	Signature          solana.Signature       `json:"signature,omitempty"`
	Err                any                    `json:"err,omitempty"`
	Memo               *string                `json:"memo,omitempty"`
	ConfirmationStatus ConfirmationStatusType `json:"confirmationStatus,omitempty"`

	Transaction *ParsedTransaction     `json:"transaction,omitempty"`
	Meta        *ParsedTransactionMeta `json:"meta,omitempty"`
}
