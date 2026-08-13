package rpc

// GetMultipleAccountsResult is the result of the getMultipleAccounts method.
type GetMultipleAccountsResult struct {
	RPCContext
	Value []*Account `json:"value"`
}

// GetMultipleAccountsOpts is the optional configuration object for
// getMultipleAccounts; it accepts the same fields as getAccountInfo.
type GetMultipleAccountsOpts GetAccountInfoOpts
