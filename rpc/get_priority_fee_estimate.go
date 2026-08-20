package rpc

import solana "github.com/fluxrpc/solana-go"

// PriorityLevel selects the percentile used by FluxRPC's priority-fee
// estimator.
type PriorityLevel string

const (
	PriorityLevelMin       PriorityLevel = "Min"
	PriorityLevelLow       PriorityLevel = "Low"
	PriorityLevelMedium    PriorityLevel = "Medium"
	PriorityLevelHigh      PriorityLevel = "High"
	PriorityLevelVeryHigh  PriorityLevel = "VeryHigh"
	PriorityLevelUnsafeMax PriorityLevel = "UnsafeMax"
)

// GetPriorityFeeEstimateRequest is the request object accepted by FluxRPC's
// getPriorityFeeEstimate method. At least one of Transaction or AccountKeys
// must be supplied; FluxRPC can use both when both are present.
type GetPriorityFeeEstimateRequest struct {
	// Transaction is a base58- or base64-encoded wire transaction.
	Transaction string `json:"transaction,omitempty"`

	// AccountKeys are accounts whose local fee markets should be considered.
	AccountKeys []solana.PublicKey `json:"accountKeys,omitempty"`

	Options *GetPriorityFeeEstimateOpts `json:"options,omitempty"`
}

// GetPriorityFeeEstimateOpts controls FluxRPC's priority-fee estimator.
type GetPriorityFeeEstimateOpts struct {
	TransactionEncoding solana.EncodingType `json:"transactionEncoding,omitempty"`
	PriorityLevel       PriorityLevel       `json:"priorityLevel,omitempty"`

	IncludeAllPriorityFeeLevels bool `json:"includeAllPriorityFeeLevels,omitempty"`
	IncludeJito                 bool `json:"includeJito,omitempty"`
	LookbackSlots               int  `json:"lookbackSlots,omitempty"`
	IncludeVote                 bool `json:"includeVote,omitempty"`
	Recommended                 bool `json:"recommended,omitempty"`
	EvaluateEmptySlotAsZero     bool `json:"evaluateEmptySlotAsZero,omitempty"`
}

// GetPriorityFeeEstimateResult is FluxRPC's priority-fee estimate in
// micro-lamports per compute unit, with optional level and Jito estimates.
type GetPriorityFeeEstimateResult struct {
	PriorityFeeEstimate uint64 `json:"priorityFeeEstimate"`

	// EstimatedPriorityLamports is included when a transaction supplies a
	// compute-unit limit and the total estimate can be calculated.
	EstimatedPriorityLamports *uint64 `json:"estimatedPriorityLamports,omitempty"`

	PriorityFeeLevels *PriorityFeeLevels `json:"priorityFeeLevels,omitempty"`
	JitoTipEstimate   *JitoTipEstimate   `json:"jitoTipEstimate,omitempty"`
}

// PriorityFeeLevels contains FluxRPC's fee percentiles in micro-lamports per
// compute unit.
type PriorityFeeLevels struct {
	Min       uint64 `json:"min"`
	Low       uint64 `json:"low"`
	Medium    uint64 `json:"medium"`
	High      uint64 `json:"high"`
	VeryHigh  uint64 `json:"veryHigh"`
	UnsafeMax uint64 `json:"unsafeMax"`
}

// JitoTipEstimate contains estimated Jito tips in lamports.
type JitoTipEstimate struct {
	Low    uint64 `json:"low"`
	Medium uint64 `json:"medium"`
	High   uint64 `json:"high"`
}
