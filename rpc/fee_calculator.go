package rpc

// FeeCalculator holds the network fee rate.
type FeeCalculator struct {
	LamportsPerSignature uint64 `json:"lamportsPerSignature"`
}
