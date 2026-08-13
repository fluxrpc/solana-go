package rpc

// GetAccountInfoResult is the response of the getAccountInfo RPC method.
type GetAccountInfoResult struct {
	RPCContext
	Value *Account `json:"value"`
}

// GetBinary returns the binary representation of the account data.
func (a *GetAccountInfoResult) GetBinary() []byte {
	if a == nil {
		return nil
	}
	if a.Value == nil {
		return nil
	}
	if a.Value.Data == nil {
		return nil
	}
	return a.Value.Data.GetBinary()
}

// Bytes returns the binary representation of the account data.
func (a *GetAccountInfoResult) Bytes() []byte {
	return a.GetBinary()
}

// GetBalanceResult is the response of the getBalance RPC method.
type GetBalanceResult struct {
	RPCContext
	Value uint64 `json:"value"`
}

// GetStakeMinimumDelegationResult is the response of the
// getStakeMinimumDelegation RPC method.
type GetStakeMinimumDelegationResult struct {
	RPCContext
	Value uint64 `json:"value"`
}

// IsValidBlockhashResult is the response of the isBlockhashValid RPC method.
type IsValidBlockhashResult struct {
	RPCContext
	Value bool `json:"value"` // True if the blockhash is still valid.
}
