package rpc

import (
	"context"

	solana "github.com/fluxrpc/solana-go"
)

// commitmentParam returns the {"commitment": ...} param object, or nothing
// when the commitment is empty (node default).
func commitmentParam(commitment CommitmentType) []any {
	if commitment == "" {
		return nil
	}
	return []any{M{"commitment": commitment}}
}

// withCommitment appends the optional commitment object to params.
func withCommitment(params []any, commitment CommitmentType) []any {
	return append(params, commitmentParam(commitment)...)
}

// --- Accounts ---

// GetAccountInfo returns all information associated with the account,
// requesting base64 encoding. Returns ErrNotFound if the account does not
// exist.
func (c *Client) GetAccountInfo(ctx context.Context, account solana.PublicKey) (*GetAccountInfoResult, error) {
	return c.GetAccountInfoWithOpts(ctx, account, &GetAccountInfoOpts{Encoding: solana.EncodingBase64})
}

func (c *Client) GetAccountInfoWithOpts(ctx context.Context, account solana.PublicKey, opts *GetAccountInfoOpts) (*GetAccountInfoResult, error) {
	params := []any{account}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetAccountInfoResult](ctx, c, "getAccountInfo", params...)
	if err != nil {
		return nil, err
	}
	if result.Value == nil {
		return nil, ErrNotFound
	}
	return &result, nil
}

func (c *Client) GetBalance(ctx context.Context, account solana.PublicKey, commitment CommitmentType) (*GetBalanceResult, error) {
	result, err := call[GetBalanceResult](ctx, c, "getBalance", withCommitment([]any{account}, commitment)...)
	return &result, err
}

func (c *Client) GetMultipleAccounts(ctx context.Context, accounts ...solana.PublicKey) (*GetMultipleAccountsResult, error) {
	return c.GetMultipleAccountsWithOpts(ctx, accounts, &GetMultipleAccountsOpts{Encoding: solana.EncodingBase64})
}

func (c *Client) GetMultipleAccountsWithOpts(ctx context.Context, accounts []solana.PublicKey, opts *GetMultipleAccountsOpts) (*GetMultipleAccountsResult, error) {
	params := []any{accounts}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetMultipleAccountsResult](ctx, c, "getMultipleAccounts", params...)
	return &result, err
}

func (c *Client) GetProgramAccounts(ctx context.Context, program solana.PublicKey) (GetProgramAccountsResult, error) {
	return c.GetProgramAccountsWithOpts(ctx, program, nil)
}

func (c *Client) GetProgramAccountsWithOpts(ctx context.Context, program solana.PublicKey, opts *GetProgramAccountsOpts) (GetProgramAccountsResult, error) {
	params := []any{program}
	if opts != nil {
		params = append(params, opts)
	}
	return call[GetProgramAccountsResult](ctx, c, "getProgramAccounts", params...)
}

func (c *Client) GetLargestAccounts(ctx context.Context, commitment CommitmentType, filter LargestAccountsFilterType) (*GetLargestAccountsResult, error) {
	opts := M{}
	if commitment != "" {
		opts["commitment"] = commitment
	}
	if filter != "" {
		opts["filter"] = filter
	}
	params := []any{}
	if len(opts) > 0 {
		params = append(params, opts)
	}
	result, err := call[GetLargestAccountsResult](ctx, c, "getLargestAccounts", params...)
	return &result, err
}

func (c *Client) GetMinimumBalanceForRentExemption(ctx context.Context, dataSize uint64, commitment CommitmentType) (uint64, error) {
	return call[uint64](ctx, c, "getMinimumBalanceForRentExemption", withCommitment([]any{dataSize}, commitment)...)
}

// --- Blocks ---

// GetBlock returns identity and transaction information about a confirmed
// block. Unlike upstream, maxSupportedTransactionVersion defaults to 0 so
// blocks containing versioned transactions do not error.
func (c *Client) GetBlock(ctx context.Context, slot uint64) (*GetBlockResult, error) {
	version := uint64(0)
	return c.GetBlockWithOpts(ctx, slot, &GetBlockOpts{MaxSupportedTransactionVersion: &version})
}

func (c *Client) GetBlockWithOpts(ctx context.Context, slot uint64, opts *GetBlockOpts) (*GetBlockResult, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	params := []any{slot}
	if opts != nil {
		params = append(params, opts)
	}
	return callNullable[GetBlockResult](ctx, c, "getBlock", params...)
}

// GetParsedBlock fetches the block with jsonParsed encoding.
func (c *Client) GetParsedBlock(ctx context.Context, slot uint64, opts *GetBlockOpts) (*GetParsedBlockResult, error) {
	if opts == nil {
		version := uint64(0)
		opts = &GetBlockOpts{MaxSupportedTransactionVersion: &version}
	}
	opts.Encoding = solana.EncodingJSONParsed
	params := []any{slot, opts}
	return callNullable[GetParsedBlockResult](ctx, c, "getBlock", params...)
}

func (c *Client) GetBlockHeight(ctx context.Context, commitment CommitmentType) (uint64, error) {
	return call[uint64](ctx, c, "getBlockHeight", commitmentParam(commitment)...)
}

func (c *Client) GetBlockCommitment(ctx context.Context, slot uint64) (*GetBlockCommitmentResult, error) {
	result, err := call[GetBlockCommitmentResult](ctx, c, "getBlockCommitment", slot)
	return &result, err
}

func (c *Client) GetBlockProduction(ctx context.Context) (*GetBlockProductionResult, error) {
	return c.GetBlockProductionWithOpts(ctx, nil)
}

func (c *Client) GetBlockProductionWithOpts(ctx context.Context, opts *GetBlockProductionOpts) (*GetBlockProductionResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetBlockProductionResult](ctx, c, "getBlockProduction", params...)
	return &result, err
}

func (c *Client) GetBlockTime(ctx context.Context, slot uint64) (solana.UnixTimeSeconds, error) {
	return call[solana.UnixTimeSeconds](ctx, c, "getBlockTime", slot)
}

func (c *Client) GetBlocks(ctx context.Context, startSlot uint64, endSlot *uint64, commitment CommitmentType) (BlocksResult, error) {
	params := []any{startSlot}
	if endSlot != nil {
		params = append(params, *endSlot)
	}
	return call[BlocksResult](ctx, c, "getBlocks", withCommitment(params, commitment)...)
}

func (c *Client) GetBlocksWithLimit(ctx context.Context, startSlot uint64, limit uint64, commitment CommitmentType) (BlocksResult, error) {
	return call[BlocksResult](ctx, c, "getBlocksWithLimit", withCommitment([]any{startSlot, limit}, commitment)...)
}

func (c *Client) GetFirstAvailableBlock(ctx context.Context) (uint64, error) {
	return call[uint64](ctx, c, "getFirstAvailableBlock")
}

// --- Transactions ---

// GetTransaction returns details for a confirmed transaction, requesting
// base64 encoding with maxSupportedTransactionVersion 0. Returns
// ErrNotFound for unknown transactions.
func (c *Client) GetTransaction(ctx context.Context, signature solana.Signature) (*GetTransactionResult, error) {
	version := uint64(0)
	return c.GetTransactionWithOpts(ctx, signature, &GetTransactionOpts{
		Encoding:                       solana.EncodingBase64,
		MaxSupportedTransactionVersion: &version,
	})
}

func (c *Client) GetTransactionWithOpts(ctx context.Context, signature solana.Signature, opts *GetTransactionOpts) (*GetTransactionResult, error) {
	params := []any{signature}
	if opts != nil {
		params = append(params, opts)
	}
	return callNullable[GetTransactionResult](ctx, c, "getTransaction", params...)
}

// GetParsedTransaction fetches the transaction with jsonParsed encoding.
func (c *Client) GetParsedTransaction(ctx context.Context, signature solana.Signature, opts *GetParsedTransactionOpts) (*GetParsedTransactionResult, error) {
	version := uint64(0)
	obj := M{"encoding": solana.EncodingJSONParsed, "maxSupportedTransactionVersion": &version}
	if opts != nil {
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.MaxSupportedTransactionVersion != nil {
			obj["maxSupportedTransactionVersion"] = opts.MaxSupportedTransactionVersion
		}
	}
	return callNullable[GetParsedTransactionResult](ctx, c, "getTransaction", signature, obj)
}

func (c *Client) GetTransactionCount(ctx context.Context, commitment CommitmentType) (uint64, error) {
	return call[uint64](ctx, c, "getTransactionCount", commitmentParam(commitment)...)
}

func (c *Client) GetSignaturesForAddress(ctx context.Context, account solana.PublicKey) ([]*TransactionSignature, error) {
	return c.GetSignaturesForAddressWithOpts(ctx, account, nil)
}

func (c *Client) GetSignaturesForAddressWithOpts(ctx context.Context, account solana.PublicKey, opts *GetSignaturesForAddressOpts) ([]*TransactionSignature, error) {
	params := []any{account}
	if opts != nil {
		params = append(params, opts)
	}
	return call[[]*TransactionSignature](ctx, c, "getSignaturesForAddress", params...)
}

func (c *Client) GetSignatureStatuses(ctx context.Context, searchTransactionHistory bool, signatures ...solana.Signature) (*GetSignatureStatusesResult, error) {
	params := []any{signatures}
	if searchTransactionHistory {
		params = append(params, M{"searchTransactionHistory": true})
	}
	result, err := call[GetSignatureStatusesResult](ctx, c, "getSignatureStatuses", params...)
	return &result, err
}

// --- Sending ---

// SendTransaction submits a signed transaction (base64-encoded wire form)
// and returns its signature. Preflight checks are performed.
func (c *Client) SendTransaction(ctx context.Context, tx *solana.Transaction) (solana.Signature, error) {
	return c.SendTransactionWithOpts(ctx, tx, TransactionOpts{})
}

func (c *Client) SendTransactionWithOpts(ctx context.Context, tx *solana.Transaction, opts TransactionOpts) (solana.Signature, error) {
	raw, err := tx.MarshalBinary()
	if err != nil {
		return solana.Signature{}, err
	}
	return c.SendRawTransactionWithOpts(ctx, raw, opts)
}

func (c *Client) SendRawTransactionWithOpts(ctx context.Context, raw []byte, opts TransactionOpts) (solana.Signature, error) {
	return call[solana.Signature](ctx, c, "sendTransaction", solana.Base64(raw).String(), opts.ToMap())
}

// SendEncodedTransaction submits an already base64-encoded transaction.
func (c *Client) SendEncodedTransaction(ctx context.Context, encoded string) (solana.Signature, error) {
	var opts TransactionOpts
	return call[solana.Signature](ctx, c, "sendTransaction", encoded, opts.ToMap())
}

func (c *Client) SimulateTransaction(ctx context.Context, tx *solana.Transaction) (*SimulateTransactionResponse, error) {
	return c.SimulateTransactionWithOpts(ctx, tx, nil)
}

func (c *Client) SimulateTransactionWithOpts(ctx context.Context, tx *solana.Transaction, opts *SimulateTransactionOpts) (*SimulateTransactionResponse, error) {
	raw, err := tx.MarshalBinary()
	if err != nil {
		return nil, err
	}
	obj := M{"encoding": solana.EncodingBase64}
	if opts != nil {
		if opts.SigVerify {
			obj["sigVerify"] = true
		}
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if opts.ReplaceRecentBlockhash {
			obj["replaceRecentBlockhash"] = true
		}
		if opts.Accounts != nil {
			obj["accounts"] = opts.Accounts
		}
		if opts.MinContextSlot != nil {
			obj["minContextSlot"] = opts.MinContextSlot
		}
	}
	result, err := call[SimulateTransactionResponse](ctx, c, "simulateTransaction", solana.Base64(raw).String(), obj)
	return &result, err
}

func (c *Client) RequestAirdrop(ctx context.Context, account solana.PublicKey, lamports uint64, commitment CommitmentType) (solana.Signature, error) {
	return call[solana.Signature](ctx, c, "requestAirdrop", withCommitment([]any{account, lamports}, commitment)...)
}

// --- Blockhashes & fees ---

func (c *Client) GetLatestBlockhash(ctx context.Context, commitment CommitmentType) (*GetLatestBlockhashResult, error) {
	result, err := call[GetLatestBlockhashResult](ctx, c, "getLatestBlockhash", commitmentParam(commitment)...)
	return &result, err
}

func (c *Client) IsBlockhashValid(ctx context.Context, blockhash solana.Hash, commitment CommitmentType) (*IsValidBlockhashResult, error) {
	result, err := call[IsValidBlockhashResult](ctx, c, "isBlockhashValid", withCommitment([]any{blockhash}, commitment)...)
	return &result, err
}

func (c *Client) GetFeeForMessage(ctx context.Context, messageBase64 string, commitment CommitmentType) (*GetFeeForMessageResult, error) {
	result, err := call[GetFeeForMessageResult](ctx, c, "getFeeForMessage", withCommitment([]any{messageBase64}, commitment)...)
	return &result, err
}

func (c *Client) GetRecentPrioritizationFees(ctx context.Context, accounts []solana.PublicKey) ([]PriorizationFeeResult, error) {
	params := []any{}
	if len(accounts) > 0 {
		params = append(params, accounts)
	}
	return call[[]PriorizationFeeResult](ctx, c, "getRecentPrioritizationFees", params...)
}

// --- Cluster & network ---

func (c *Client) GetSlot(ctx context.Context, commitment CommitmentType) (uint64, error) {
	return call[uint64](ctx, c, "getSlot", commitmentParam(commitment)...)
}

func (c *Client) GetSlotLeader(ctx context.Context, commitment CommitmentType) (solana.PublicKey, error) {
	return call[solana.PublicKey](ctx, c, "getSlotLeader", commitmentParam(commitment)...)
}

func (c *Client) GetSlotLeaders(ctx context.Context, start uint64, limit uint64) ([]solana.PublicKey, error) {
	return call[[]solana.PublicKey](ctx, c, "getSlotLeaders", start, limit)
}

func (c *Client) GetClusterNodes(ctx context.Context) ([]GetClusterNodesResult, error) {
	return call[[]GetClusterNodesResult](ctx, c, "getClusterNodes")
}

func (c *Client) GetVersion(ctx context.Context) (*GetVersionResult, error) {
	result, err := call[GetVersionResult](ctx, c, "getVersion")
	return &result, err
}

func (c *Client) GetHealth(ctx context.Context) (string, error) {
	return call[string](ctx, c, "getHealth")
}

func (c *Client) GetIdentity(ctx context.Context) (*GetIdentityResult, error) {
	result, err := call[GetIdentityResult](ctx, c, "getIdentity")
	return &result, err
}

func (c *Client) GetGenesisHash(ctx context.Context) (solana.Hash, error) {
	return call[solana.Hash](ctx, c, "getGenesisHash")
}

func (c *Client) GetHighestSnapshotSlot(ctx context.Context) (*GetHighestSnapshotSlotResult, error) {
	result, err := call[GetHighestSnapshotSlotResult](ctx, c, "getHighestSnapshotSlot")
	return &result, err
}

func (c *Client) GetMaxRetransmitSlot(ctx context.Context) (uint64, error) {
	return call[uint64](ctx, c, "getMaxRetransmitSlot")
}

func (c *Client) GetMaxShredInsertSlot(ctx context.Context) (uint64, error) {
	return call[uint64](ctx, c, "getMaxShredInsertSlot")
}

func (c *Client) MinimumLedgerSlot(ctx context.Context) (uint64, error) {
	return call[uint64](ctx, c, "minimumLedgerSlot")
}

func (c *Client) GetRecentPerformanceSamples(ctx context.Context, limit *uint) ([]GetRecentPerformanceSamplesResult, error) {
	params := []any{}
	if limit != nil {
		params = append(params, *limit)
	}
	return call[[]GetRecentPerformanceSamplesResult](ctx, c, "getRecentPerformanceSamples", params...)
}

// --- Epoch, inflation & supply ---

func (c *Client) GetEpochInfo(ctx context.Context, commitment CommitmentType) (*GetEpochInfoResult, error) {
	result, err := call[GetEpochInfoResult](ctx, c, "getEpochInfo", commitmentParam(commitment)...)
	return &result, err
}

func (c *Client) GetEpochSchedule(ctx context.Context) (*GetEpochScheduleResult, error) {
	result, err := call[GetEpochScheduleResult](ctx, c, "getEpochSchedule")
	return &result, err
}

func (c *Client) GetInflationGovernor(ctx context.Context, commitment CommitmentType) (*GetInflationGovernorResult, error) {
	result, err := call[GetInflationGovernorResult](ctx, c, "getInflationGovernor", commitmentParam(commitment)...)
	return &result, err
}

func (c *Client) GetInflationRate(ctx context.Context) (*GetInflationRateResult, error) {
	result, err := call[GetInflationRateResult](ctx, c, "getInflationRate")
	return &result, err
}

func (c *Client) GetInflationReward(ctx context.Context, addresses []solana.PublicKey, opts *GetInflationRewardOpts) ([]*GetInflationRewardResult, error) {
	params := []any{addresses}
	if opts != nil {
		params = append(params, opts)
	}
	return call[[]*GetInflationRewardResult](ctx, c, "getInflationReward", params...)
}

func (c *Client) GetSupply(ctx context.Context, commitment CommitmentType) (*GetSupplyResult, error) {
	return c.GetSupplyWithOpts(ctx, &GetSupplyOpts{Commitment: commitment})
}

func (c *Client) GetSupplyWithOpts(ctx context.Context, opts *GetSupplyOpts) (*GetSupplyResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetSupplyResult](ctx, c, "getSupply", params...)
	return &result, err
}

func (c *Client) GetLeaderSchedule(ctx context.Context) (GetLeaderScheduleResult, error) {
	return c.GetLeaderScheduleWithOpts(ctx, nil)
}

func (c *Client) GetLeaderScheduleWithOpts(ctx context.Context, opts *GetLeaderScheduleOpts) (GetLeaderScheduleResult, error) {
	params := []any{nil}
	if opts != nil {
		if opts.Epoch != nil {
			params[0] = opts.Epoch
		}
		obj := M{}
		if opts.Identity != nil {
			obj["identity"] = opts.Identity
		}
		if opts.Commitment != "" {
			obj["commitment"] = opts.Commitment
		}
		if len(obj) > 0 {
			params = append(params, obj)
		}
	}
	return call[GetLeaderScheduleResult](ctx, c, "getLeaderSchedule", params...)
}

func (c *Client) GetVoteAccounts(ctx context.Context, opts *GetVoteAccountsOpts) (*GetVoteAccountsResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetVoteAccountsResult](ctx, c, "getVoteAccounts", params...)
	return &result, err
}

func (c *Client) GetStakeMinimumDelegation(ctx context.Context, commitment CommitmentType) (*GetStakeMinimumDelegationResult, error) {
	result, err := call[GetStakeMinimumDelegationResult](ctx, c, "getStakeMinimumDelegation", commitmentParam(commitment)...)
	return &result, err
}

// --- Tokens ---

func (c *Client) GetTokenAccountBalance(ctx context.Context, account solana.PublicKey, commitment CommitmentType) (*GetTokenAccountBalanceResult, error) {
	result, err := call[GetTokenAccountBalanceResult](ctx, c, "getTokenAccountBalance", withCommitment([]any{account}, commitment)...)
	return &result, err
}

func (c *Client) GetTokenAccountsByOwner(ctx context.Context, owner solana.PublicKey, config *GetTokenAccountsConfig, opts *GetTokenAccountsOpts) (*GetTokenAccountsResult, error) {
	return c.getTokenAccountsBy(ctx, "getTokenAccountsByOwner", owner, config, opts)
}

func (c *Client) GetTokenAccountsByDelegate(ctx context.Context, delegate solana.PublicKey, config *GetTokenAccountsConfig, opts *GetTokenAccountsOpts) (*GetTokenAccountsResult, error) {
	return c.getTokenAccountsBy(ctx, "getTokenAccountsByDelegate", delegate, config, opts)
}

func (c *Client) getTokenAccountsBy(ctx context.Context, method string, key solana.PublicKey, config *GetTokenAccountsConfig, opts *GetTokenAccountsOpts) (*GetTokenAccountsResult, error) {
	params := []any{key}
	if config != nil {
		params = append(params, config)
	}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetTokenAccountsResult](ctx, c, method, params...)
	return &result, err
}

func (c *Client) GetTokenLargestAccounts(ctx context.Context, mint solana.PublicKey, commitment CommitmentType) (*GetTokenLargestAccountsResult, error) {
	result, err := call[GetTokenLargestAccountsResult](ctx, c, "getTokenLargestAccounts", withCommitment([]any{mint}, commitment)...)
	return &result, err
}

func (c *Client) GetTokenSupply(ctx context.Context, mint solana.PublicKey, commitment CommitmentType) (*GetTokenSupplyResult, error) {
	result, err := call[GetTokenSupplyResult](ctx, c, "getTokenSupply", withCommitment([]any{mint}, commitment)...)
	return &result, err
}
