package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// TransactionOpts are the options for sending a transaction.
type TransactionOpts struct {
	Encoding            solana.EncodingType `json:"encoding,omitempty"`
	SkipPreflight       bool                `json:"skipPreflight,omitempty"`
	PreflightCommitment CommitmentType      `json:"preflightCommitment,omitempty"`
	MaxRetries          *uint               `json:"maxRetries"`
	MinContextSlot      *uint64             `json:"minContextSlot"`
}

// ToMap converts the options into the RPC parameter object.
func (opts *TransactionOpts) ToMap() M {
	obj := M{}

	if opts.Encoding == "" {
		// default to base64 encoding
		obj["encoding"] = "base64"
	} else {
		obj["encoding"] = opts.Encoding
	}

	obj["skipPreflight"] = opts.SkipPreflight

	if opts.PreflightCommitment != "" {
		obj["preflightCommitment"] = opts.PreflightCommitment
	}

	if opts.MaxRetries != nil {
		obj["maxRetries"] = *opts.MaxRetries
	}

	if opts.MinContextSlot != nil {
		obj["minContextSlot"] = *opts.MinContextSlot
	}

	return obj
}
