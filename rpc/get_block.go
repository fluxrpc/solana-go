package rpc

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
)

// TransactionDetailsType is the level of transaction detail to return with
// a block.
type TransactionDetailsType string

const (
	TransactionDetailsFull       TransactionDetailsType = "full"
	TransactionDetailsSignatures TransactionDetailsType = "signatures"
	TransactionDetailsNone       TransactionDetailsType = "none"
	TransactionDetailsAccounts   TransactionDetailsType = "accounts"
)

// GetBlockOpts is the optional configuration object for the getBlock RPC
// method.
type GetBlockOpts struct {
	// Encoding for each returned Transaction, either "json", "jsonParsed", "base58" (slow), "base64".
	// If parameter not provided, the default encoding is "json".
	// - "jsonParsed" encoding attempts to use program-specific instruction parsers to return
	//   more human-readable and explicit data in the transaction.message.instructions list.
	// - If "jsonParsed" is requested but a parser cannot be found, the instruction falls back
	//   to regular JSON encoding (accounts, data, and programIdIndex fields).
	//
	// This parameter is optional.
	Encoding solana.EncodingType

	// Level of transaction detail to return.
	// If parameter not provided, the default detail level is "full".
	//
	// This parameter is optional.
	TransactionDetails TransactionDetailsType

	// Whether to populate the rewards array.
	// If parameter not provided, the default includes rewards.
	//
	// This parameter is optional.
	Rewards *bool

	// "processed" is not supported.
	// If parameter not provided, the default is "finalized".
	//
	// This parameter is optional.
	Commitment CommitmentType

	// Max transaction version to return in responses.
	// If the requested block contains a transaction with a higher version, an error will be returned.
	MaxSupportedTransactionVersion *uint64
}

var (
	MaxSupportedTransactionVersion0 uint64 = 0
	MaxSupportedTransactionVersion1 uint64 = 1
)

// Validate checks that the configured encoding is one the getBlock RPC
// method supports.
func (opts *GetBlockOpts) Validate() error {
	if opts == nil || opts.Encoding == "" {
		return nil
	}
	if !solana.IsAnyOfEncodingType(
		opts.Encoding,
		// Valid encodings:
		solana.EncodingJSON,
		solana.EncodingJSONParsed,
		solana.EncodingBase58,
		solana.EncodingBase64,
		solana.EncodingBase64Zstd,
	) {
		return fmt.Errorf("provided encoding is not supported: %s", opts.Encoding)
	}
	return nil
}

// GetBlockResult is the response of the getBlock RPC method.
type GetBlockResult struct {
	// The blockhash of this block.
	Blockhash solana.Hash `json:"blockhash"`

	// The blockhash of this block's parent;
	// if the parent block is not available due to ledger cleanup,
	// this field will return "11111111111111111111111111111111".
	PreviousBlockhash solana.Hash `json:"previousBlockhash"`

	// The slot index of this block's parent.
	ParentSlot uint64 `json:"parentSlot"`

	// Present if "full" transaction details are requested.
	Transactions []TransactionWithMeta `json:"transactions"`

	// Present if "signatures" are requested for transaction details;
	// an array of signatures, corresponding to the transaction order in the block.
	Signatures []solana.Signature `json:"signatures"`

	// Present if rewards are requested.
	Rewards []BlockReward `json:"rewards"`

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch).
	// Nil if not available.
	BlockTime *solana.UnixTimeSeconds `json:"blockTime"`

	// The number of blocks beneath this block.
	BlockHeight *uint64 `json:"blockHeight"`

	// The number of reward partitions.
	// Present for the first block in the epoch otherwise Nil.
	NumRewardPartitions *uint64 `json:"numRewardPartitions"`
}

// GetParsedBlockResult is the response of the getBlock RPC method when the
// jsonParsed encoding is requested.
type GetParsedBlockResult struct {
	// The blockhash of this block.
	Blockhash solana.Hash `json:"blockhash"`

	// The blockhash of this block's parent;
	// if the parent block is not available due to ledger cleanup,
	// this field will return "11111111111111111111111111111111".
	PreviousBlockhash solana.Hash `json:"previousBlockhash"`

	// The slot index of this block's parent.
	ParentSlot uint64 `json:"parentSlot"`

	// Present if "full" transaction details are requested.
	Transactions []ParsedTransactionWithMeta `json:"transactions"`

	// Present if "signatures" are requested for transaction details;
	// an array of signatures, corresponding to the transaction order in the block.
	Signatures []solana.Signature `json:"signatures"`

	// Present if rewards are requested.
	Rewards []BlockReward `json:"rewards"`

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch).
	// Nil if not available.
	BlockTime *solana.UnixTimeSeconds `json:"blockTime"`

	// The number of blocks beneath this block.
	BlockHeight *uint64 `json:"blockHeight"`

	// The number of reward partitions.
	// Present for the first block in the epoch otherwise Nil.
	NumRewardPartitions *uint64 `json:"numRewardPartitions"`
}

// ParsedTransactionWithMeta is a jsonParsed-encoded transaction plus its
// status metadata.
type ParsedTransactionWithMeta struct {
	Slot        uint64
	BlockTime   *solana.UnixTimeSeconds
	Transaction *ParsedTransaction
	Meta        *ParsedTransactionMeta
}
