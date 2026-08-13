package rpc

import (
	"fmt"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
)

// GetTransactionOpts is the optional configuration object for the
// getTransaction method.
type GetTransactionOpts struct {
	Encoding solana.EncodingType `json:"encoding,omitempty"`

	// Desired commitment. "processed" is not supported. If parameter not provided, the default is "finalized".
	Commitment CommitmentType `json:"commitment,omitempty"`

	// Max transaction version to return in responses.
	// If the requested block contains a transaction with a higher version, an error will be returned.
	MaxSupportedTransactionVersion *uint64
}

// GetTransactionResult is the result of the getTransaction method.
type GetTransactionResult struct {
	// The slot this transaction was processed in.
	Slot uint64 `json:"slot"`

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch)
	// of when the transaction was processed.
	// Nil if not available.
	BlockTime *solana.UnixTimeSeconds `json:"blockTime"`

	Transaction *TransactionResultEnvelope `json:"transaction"`
	Meta        *TransactionMeta           `json:"meta,omitempty"`
	Version     TransactionVersion         `json:"version"`
}

// TransactionResultEnvelope will contain a *solana.Transaction if the requested encoding is `solana.EncodingJSON`
// (which is also the default when the encoding is not specified),
// or a `solana.Data` in case of EncodingBase58, EncodingBase64.
type TransactionResultEnvelope struct {
	asDecodedBinary     solana.Data
	asParsedTransaction *solana.Transaction
}

func (wrap TransactionResultEnvelope) MarshalJSON() ([]byte, error) {
	if wrap.asParsedTransaction != nil {
		return sonic.Marshal(wrap.asParsedTransaction)
	}
	return wrap.asDecodedBinary.MarshalJSON()
}

func (wrap *TransactionResultEnvelope) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || (len(data) == 4 && string(data) == "null") {
		return nil
	}

	switch data[0] {
	// A JSON array is the ["<content>","<encoding>"] binary tuple.
	case '[':
		return wrap.asDecodedBinary.UnmarshalJSON(data)
	// A JSON object is the transaction in its JSON encoding.
	case '{':
		return sonic.Unmarshal(data, &wrap.asParsedTransaction)
	default:
		return fmt.Errorf("unknown kind: %v", data)
	}
}

// GetBinary returns the decoded bytes if the encoding is
// "base58", "base64".
func (dt *TransactionResultEnvelope) GetBinary() []byte {
	return dt.asDecodedBinary.Content
}

// GetData returns the transaction data plus its encoding.
func (dt *TransactionResultEnvelope) GetData() solana.Data {
	return dt.asDecodedBinary
}

// GetTransaction returns the decoded *solana.Transaction, decoding the binary
// representation if that is what the RPC returned.
func (dt *TransactionResultEnvelope) GetTransaction() (*solana.Transaction, error) {
	if dt.asDecodedBinary.Content != nil {
		return solana.TransactionFromBytes(dt.asDecodedBinary.Content)
	}
	return dt.asParsedTransaction, nil
}
