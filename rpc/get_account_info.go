package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetAccountInfoOpts is the optional configuration object for the
// getAccountInfo RPC method.
type GetAccountInfoOpts struct {
	// Encoding for Account data.
	// Either "base58" (slow), "base64", "base64+zstd", or "jsonParsed".
	// - "base58" is limited to Account data of less than 129 bytes.
	// - "base64" will return base64 encoded data for Account data of any size.
	// - "base64+zstd" compresses the Account data using Zstandard and base64-encodes the result.
	// - "jsonParsed" encoding attempts to use program-specific state parsers to return more
	// 	 human-readable and explicit account state data. If "jsonParsed" is requested but a parser
	//   cannot be found, the field falls back to "base64" encoding,
	//   detectable when the data field is type <string>.
	//
	// This parameter is optional.
	Encoding solana.EncodingType

	// Commitment requirement.
	//
	// This parameter is optional.
	Commitment CommitmentType

	// dataSlice parameters for limiting returned account data:
	// Limits the returned account data using the provided offset and length fields;
	// only available for "base58", "base64" or "base64+zstd" encodings.
	//
	// This parameter is optional.
	DataSlice *DataSlice

	// The minimum slot that the request can be evaluated at.
	// This parameter is optional.
	MinContextSlot *uint64
}
