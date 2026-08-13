package rpc

// GetTokenSupplyResult is the result of the getTokenSupply method.
type GetTokenSupplyResult struct {
	RPCContext
	Value *UiTokenAmount `json:"value"`
}
