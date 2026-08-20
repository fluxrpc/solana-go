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
// exist. With the cache enabled (see EnableCache) the account is served
// locally whenever it is streamed, immutable or fresh.
func (c *Client) GetAccountInfo(ctx context.Context, account solana.PublicKey) (*GetAccountInfoResult, error) {
	if cache := c.cache.Load(); cache != nil {
		return c.cachedGetAccountInfo(ctx, account, cache)
	}
	return c.GetAccountInfoWithOpts(ctx, account, &GetAccountInfoOpts{Encoding: solana.EncodingBase64})
}

// GetAccountInfoWithOpts is GetAccountInfo with explicit options. Returns
// ErrNotFound if the account does not exist.
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

// GetBalance returns the lamport balance of the account via getBalance.
// Commitment "" uses the node default.
func (c *Client) GetBalance(ctx context.Context, account solana.PublicKey, commitment CommitmentType) (*GetBalanceResult, error) {
	var opts *GetBalanceOpts
	if commitment != "" {
		opts = &GetBalanceOpts{Commitment: commitment}
	}
	return c.GetBalanceWithOpts(ctx, account, opts)
}

// GetBalanceWithOpts is GetBalance with explicit options.
func (c *Client) GetBalanceWithOpts(ctx context.Context, account solana.PublicKey, opts *GetBalanceOpts) (*GetBalanceResult, error) {
	params := []any{account}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetBalanceResult](ctx, c, "getBalance", params...)
	return &result, err
}

// GetMultipleAccounts returns the information of multiple accounts via
// getMultipleAccounts, requesting base64 encoding. Accounts that do not
// exist come back as nil entries in the result. With the cache enabled
// (see EnableCache) as many accounts as possible are served locally and
// only the deduplicated misses are fetched, in a single call.
func (c *Client) GetMultipleAccounts(ctx context.Context, accounts ...solana.PublicKey) (*GetMultipleAccountsResult, error) {
	if cache := c.cache.Load(); cache != nil {
		return c.cachedGetMultipleAccounts(ctx, accounts, cache)
	}
	return c.GetMultipleAccountsWithOpts(ctx, accounts, &GetMultipleAccountsOpts{Encoding: solana.EncodingBase64})
}

// GetMultipleAccountsWithOpts is GetMultipleAccounts with explicit options.
func (c *Client) GetMultipleAccountsWithOpts(ctx context.Context, accounts []solana.PublicKey, opts *GetMultipleAccountsOpts) (*GetMultipleAccountsResult, error) {
	params := []any{accounts}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetMultipleAccountsResult](ctx, c, "getMultipleAccounts", params...)
	return &result, err
}

// GetProgramAccounts returns all accounts owned by the program via
// getProgramAccounts, with the node's default options.
func (c *Client) GetProgramAccounts(ctx context.Context, program solana.PublicKey) (GetProgramAccountsResult, error) {
	return c.GetProgramAccountsWithOpts(ctx, program, nil)
}

// GetProgramAccountsWithOpts is GetProgramAccounts with explicit options
// (filters, encoding, data slice, commitment). For very large result sets
// consider GetProgramAccountsStream instead.
func (c *Client) GetProgramAccountsWithOpts(ctx context.Context, program solana.PublicKey, opts *GetProgramAccountsOpts) (GetProgramAccountsResult, error) {
	if opts != nil && opts.WithContext != nil && *opts.WithContext {
		result, err := c.GetProgramAccountsWithContext(ctx, program, opts)
		if err != nil {
			return nil, err
		}
		return result.Value, nil
	}
	params := []any{program}
	if opts != nil {
		params = append(params, opts)
	}
	return call[GetProgramAccountsResult](ctx, c, "getProgramAccounts", params...)
}

// GetProgramAccountsWithContext returns the context-wrapped form of
// getProgramAccounts. The caller's options are copied before withContext is
// enabled.
func (c *Client) GetProgramAccountsWithContext(ctx context.Context, program solana.PublicKey, opts *GetProgramAccountsOpts) (*GetProgramAccountsWithContextResult, error) {
	withContext := true
	requestOpts := GetProgramAccountsOpts{WithContext: &withContext}
	if opts != nil {
		requestOpts = *opts
		requestOpts.WithContext = &withContext
	}
	result, err := call[GetProgramAccountsWithContextResult](ctx, c, "getProgramAccounts", program, &requestOpts)
	return &result, err
}

// GetLargestAccounts returns the 20 largest accounts by lamport balance via
// getLargestAccounts. Empty commitment and filter are omitted from the
// request so the node defaults apply.
func (c *Client) GetLargestAccounts(ctx context.Context, commitment CommitmentType, filter LargestAccountsFilterType) (*GetLargestAccountsResult, error) {
	var opts *GetLargestAccountsOpts
	if commitment != "" || filter != "" {
		opts = &GetLargestAccountsOpts{Commitment: commitment, Filter: filter}
	}
	return c.GetLargestAccountsWithOpts(ctx, opts)
}

// GetLargestAccountsWithOpts is GetLargestAccounts with explicit options.
func (c *Client) GetLargestAccountsWithOpts(ctx context.Context, opts *GetLargestAccountsOpts) (*GetLargestAccountsResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetLargestAccountsResult](ctx, c, "getLargestAccounts", params...)
	return &result, err
}

// GetMinimumBalanceForRentExemption returns the minimum lamport balance that
// makes an account of the given data size rent exempt. Commitment "" uses
// the node default.
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

// GetBlockWithOpts is GetBlock with explicit options; the configured
// encoding is validated first. Returns ErrNotFound when there is no
// confirmed block at the slot.
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
	requestOpts := GetBlockOpts{}
	if opts == nil {
		version := uint64(0)
		requestOpts.MaxSupportedTransactionVersion = &version
	} else {
		requestOpts = *opts
	}
	requestOpts.Encoding = solana.EncodingJSONParsed
	params := []any{slot, &requestOpts}
	return callNullable[GetParsedBlockResult](ctx, c, "getBlock", params...)
}

// GetBlockHeight returns the current block height via getBlockHeight.
// Commitment "" uses the node default (finalized). With the cache enabled
// and fed by streamed block metadata (see the yellowstone package's
// Pipe), the height is served locally while fresh.
func (c *Client) GetBlockHeight(ctx context.Context, commitment CommitmentType) (uint64, error) {
	cache := c.cache.Load()
	if cache != nil {
		if height, ok := cache.lookupBlockHeight(commitment); ok {
			return height, nil
		}
	}
	var opts *GetBlockHeightOpts
	if commitment != "" {
		opts = &GetBlockHeightOpts{Commitment: commitment}
	}
	height, err := c.GetBlockHeightWithOpts(ctx, opts)
	if err == nil && cache != nil {
		cache.storeBlockHeight(commitment, height)
	}
	return height, err
}

// GetBlockHeightWithOpts is GetBlockHeight with explicit options. It always
// queries the RPC endpoint, bypassing the chain-head cache.
func (c *Client) GetBlockHeightWithOpts(ctx context.Context, opts *GetBlockHeightOpts) (uint64, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	return call[uint64](ctx, c, "getBlockHeight", params...)
}

// GetBlockCommitment returns the amount of cluster stake that voted on the
// block at the given slot, via getBlockCommitment.
func (c *Client) GetBlockCommitment(ctx context.Context, slot uint64) (*GetBlockCommitmentResult, error) {
	result, err := call[GetBlockCommitmentResult](ctx, c, "getBlockCommitment", slot)
	return &result, err
}

// GetBlockProduction returns recent block production information for the
// current epoch via getBlockProduction.
func (c *Client) GetBlockProduction(ctx context.Context) (*GetBlockProductionResult, error) {
	return c.GetBlockProductionWithOpts(ctx, nil)
}

// GetBlockProductionWithOpts is GetBlockProduction with explicit options
// (identity, slot range, commitment).
func (c *Client) GetBlockProductionWithOpts(ctx context.Context, opts *GetBlockProductionOpts) (*GetBlockProductionResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetBlockProductionResult](ctx, c, "getBlockProduction", params...)
	return &result, err
}

// GetBlockTime returns the estimated production time of the block at the
// given slot via getBlockTime.
func (c *Client) GetBlockTime(ctx context.Context, slot uint64) (solana.UnixTimeSeconds, error) {
	result, err := callNullable[solana.UnixTimeSeconds](ctx, c, "getBlockTime", slot)
	if err != nil {
		return 0, err
	}
	return *result, nil
}

// GetBlocks returns the confirmed blocks between startSlot and endSlot
// (inclusive) via getBlocks. A nil endSlot leaves the upper bound to the
// node; commitment "" uses the node default.
func (c *Client) GetBlocks(ctx context.Context, startSlot uint64, endSlot *uint64, commitment CommitmentType) (BlocksResult, error) {
	var opts *GetBlocksOpts
	if commitment != "" {
		opts = &GetBlocksOpts{Commitment: commitment}
	}
	return c.GetBlocksWithOpts(ctx, startSlot, endSlot, opts)
}

// GetBlocksWithOpts is GetBlocks with explicit options.
func (c *Client) GetBlocksWithOpts(ctx context.Context, startSlot uint64, endSlot *uint64, opts *GetBlocksOpts) (BlocksResult, error) {
	params := []any{startSlot}
	if endSlot != nil {
		params = append(params, *endSlot)
	}
	if opts != nil {
		params = append(params, opts)
	}
	return call[BlocksResult](ctx, c, "getBlocks", params...)
}

// GetBlocksWithLimit returns up to limit confirmed blocks starting at
// startSlot, via getBlocksWithLimit.
func (c *Client) GetBlocksWithLimit(ctx context.Context, startSlot uint64, limit uint64, commitment CommitmentType) (BlocksResult, error) {
	var opts *GetBlocksOpts
	if commitment != "" {
		opts = &GetBlocksOpts{Commitment: commitment}
	}
	return c.GetBlocksWithLimitWithOpts(ctx, startSlot, limit, opts)
}

// GetBlocksWithLimitWithOpts is GetBlocksWithLimit with explicit options.
func (c *Client) GetBlocksWithLimitWithOpts(ctx context.Context, startSlot uint64, limit uint64, opts *GetBlocksOpts) (BlocksResult, error) {
	params := []any{startSlot, limit}
	if opts != nil {
		params = append(params, opts)
	}
	return call[BlocksResult](ctx, c, "getBlocksWithLimit", params...)
}

// GetFirstAvailableBlock returns the slot of the lowest confirmed block
// still present in the node's ledger, via getFirstAvailableBlock.
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

// GetTransactionWithOpts is GetTransaction with explicit options. Returns
// ErrNotFound for unknown transactions.
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

// GetTransactionCount returns the current transaction count from the ledger
// via getTransactionCount. Commitment "" uses the node default.
func (c *Client) GetTransactionCount(ctx context.Context, commitment CommitmentType) (uint64, error) {
	var opts *GetTransactionCountOpts
	if commitment != "" {
		opts = &GetTransactionCountOpts{Commitment: commitment}
	}
	return c.GetTransactionCountWithOpts(ctx, opts)
}

// GetTransactionCountWithOpts is GetTransactionCount with explicit options.
func (c *Client) GetTransactionCountWithOpts(ctx context.Context, opts *GetTransactionCountOpts) (uint64, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	return call[uint64](ctx, c, "getTransactionCount", params...)
}

// GetSignaturesForAddress returns signatures of confirmed transactions that
// involve the account, newest first, via getSignaturesForAddress with the
// node's default options.
func (c *Client) GetSignaturesForAddress(ctx context.Context, account solana.PublicKey) ([]*TransactionSignature, error) {
	return c.GetSignaturesForAddressWithOpts(ctx, account, nil)
}

// GetSignaturesForAddressWithOpts is GetSignaturesForAddress with explicit
// options (limit, before/until cursors, commitment).
func (c *Client) GetSignaturesForAddressWithOpts(ctx context.Context, account solana.PublicKey, opts *GetSignaturesForAddressOpts) ([]*TransactionSignature, error) {
	params := []any{account}
	if opts != nil {
		params = append(params, opts)
	}
	return call[[]*TransactionSignature](ctx, c, "getSignaturesForAddress", params...)
}

// GetSignatureStatuses returns the processing statuses of the given
// signatures via getSignatureStatuses. Unless searchTransactionHistory is
// set, the node only consults its recent status cache; unknown signatures
// come back as nil entries.
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

// SendTransactionWithOpts serializes the transaction and submits it via
// sendTransaction with the given options; it is SendTransaction with
// explicit options.
func (c *Client) SendTransactionWithOpts(ctx context.Context, tx *solana.Transaction, opts TransactionOpts) (solana.Signature, error) {
	raw, err := tx.MarshalBinary()
	if err != nil {
		return solana.Signature{}, err
	}
	return c.SendRawTransactionWithOpts(ctx, raw, opts)
}

// SendRawTransactionWithOpts submits an already-serialized transaction,
// base64-encoding it for the sendTransaction call.
func (c *Client) SendRawTransactionWithOpts(ctx context.Context, raw []byte, opts TransactionOpts) (solana.Signature, error) {
	return call[solana.Signature](ctx, c, "sendTransaction", solana.Base64(raw).String(), opts.ToMap())
}

// SendEncodedTransaction submits an already base64-encoded transaction.
func (c *Client) SendEncodedTransaction(ctx context.Context, encoded string) (solana.Signature, error) {
	var opts TransactionOpts
	return call[solana.Signature](ctx, c, "sendTransaction", encoded, opts.ToMap())
}

// SimulateTransaction simulates sending the transaction via
// simulateTransaction, with the node's default options.
func (c *Client) SimulateTransaction(ctx context.Context, tx *solana.Transaction) (*SimulateTransactionResponse, error) {
	return c.SimulateTransactionWithOpts(ctx, tx, nil)
}

// SimulateTransactionWithOpts is SimulateTransaction with explicit options.
// The transaction is sent base64-encoded; unset options are omitted from the
// request so the node defaults apply.
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
		if opts.InnerInstructions {
			obj["innerInstructions"] = true
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

// RequestAirdrop requests an airdrop of lamports to the account via
// requestAirdrop and returns the signature of the airdrop transaction.
func (c *Client) RequestAirdrop(ctx context.Context, account solana.PublicKey, lamports uint64, commitment CommitmentType) (solana.Signature, error) {
	var opts *RequestAirdropOpts
	if commitment != "" {
		opts = &RequestAirdropOpts{Commitment: commitment}
	}
	return c.RequestAirdropWithOpts(ctx, account, lamports, opts)
}

// RequestAirdropWithOpts is RequestAirdrop with explicit options.
func (c *Client) RequestAirdropWithOpts(ctx context.Context, account solana.PublicKey, lamports uint64, opts *RequestAirdropOpts) (solana.Signature, error) {
	params := []any{account, lamports}
	if opts != nil {
		params = append(params, opts)
	}
	return call[solana.Signature](ctx, c, "requestAirdrop", params...)
}

// --- Blockhashes & fees ---

// GetLatestBlockhash returns the latest blockhash and the last block height
// at which it will be valid, via getLatestBlockhash. Commitment "" uses the
// node default (finalized). With the cache enabled and fed by streamed
// block metadata (see the yellowstone package's Pipe), the blockhash is
// served locally while fresh.
func (c *Client) GetLatestBlockhash(ctx context.Context, commitment CommitmentType) (*GetLatestBlockhashResult, error) {
	cache := c.cache.Load()
	if cache != nil {
		if cached, ok := cache.lookupBlockhash(commitment); ok {
			return cached, nil
		}
	}
	var opts *GetLatestBlockhashOpts
	if commitment != "" {
		opts = &GetLatestBlockhashOpts{Commitment: commitment}
	}
	result, err := c.GetLatestBlockhashWithOpts(ctx, opts)
	if err == nil && cache != nil && result.Value != nil {
		cache.storeBlockhash(commitment, result.Value.Blockhash, result.Value.LastValidBlockHeight, result.Context.Slot)
	}
	return result, err
}

// GetLatestBlockhashWithOpts is GetLatestBlockhash with explicit options. It
// always queries the RPC endpoint, bypassing the chain-head cache.
func (c *Client) GetLatestBlockhashWithOpts(ctx context.Context, opts *GetLatestBlockhashOpts) (*GetLatestBlockhashResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetLatestBlockhashResult](ctx, c, "getLatestBlockhash", params...)
	return &result, err
}

// IsBlockhashValid reports whether the blockhash is still valid for
// submitting transactions, via isBlockhashValid. With the cache enabled, a
// hash matching the fresh streamed latest blockhash is confirmed valid
// locally; anything else asks the network (a cache can prove validity,
// never invalidity).
func (c *Client) IsBlockhashValid(ctx context.Context, blockhash solana.Hash, commitment CommitmentType) (*IsValidBlockhashResult, error) {
	if cache := c.cache.Load(); cache != nil {
		if cached, ok := cache.lookupBlockhash(commitment); ok && cached.Value.Blockhash == blockhash {
			result := &IsValidBlockhashResult{Value: true}
			result.Context.Slot = cached.Context.Slot
			return result, nil
		}
	}
	var opts *IsBlockhashValidOpts
	if commitment != "" {
		opts = &IsBlockhashValidOpts{Commitment: commitment}
	}
	return c.IsBlockhashValidWithOpts(ctx, blockhash, opts)
}

// IsBlockhashValidWithOpts is IsBlockhashValid with explicit options. It
// always queries the RPC endpoint, bypassing the chain-head cache.
func (c *Client) IsBlockhashValidWithOpts(ctx context.Context, blockhash solana.Hash, opts *IsBlockhashValidOpts) (*IsValidBlockhashResult, error) {
	params := []any{blockhash}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[IsValidBlockhashResult](ctx, c, "isBlockhashValid", params...)
	return &result, err
}

// GetFeeForMessage returns the fee the network will charge for the given
// base64-encoded message, via getFeeForMessage.
func (c *Client) GetFeeForMessage(ctx context.Context, messageBase64 string, commitment CommitmentType) (*GetFeeForMessageResult, error) {
	var opts *GetFeeForMessageOpts
	if commitment != "" {
		opts = &GetFeeForMessageOpts{Commitment: commitment}
	}
	return c.GetFeeForMessageWithOpts(ctx, messageBase64, opts)
}

// GetFeeForMessageWithOpts is GetFeeForMessage with explicit options.
func (c *Client) GetFeeForMessageWithOpts(ctx context.Context, messageBase64 string, opts *GetFeeForMessageOpts) (*GetFeeForMessageResult, error) {
	params := []any{messageBase64}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetFeeForMessageResult](ctx, c, "getFeeForMessage", params...)
	return &result, err
}

// GetRecentPrioritizationFees returns per-slot prioritization fees from
// recent blocks via getRecentPrioritizationFees. When accounts are given,
// fees reflect transactions that lock all of them.
func (c *Client) GetRecentPrioritizationFees(ctx context.Context, accounts []solana.PublicKey) ([]PriorizationFeeResult, error) {
	params := []any{}
	if len(accounts) > 0 {
		params = append(params, accounts)
	}
	return call[[]PriorizationFeeResult](ctx, c, "getRecentPrioritizationFees", params...)
}

// --- Cluster & network ---

// GetSlot returns the slot the node has reached for the given commitment,
// via getSlot. Commitment "" uses the node default (finalized). With the
// cache enabled and fed by streamed slot updates (see the yellowstone
// package's Pipe), the slot is served locally while fresh.
func (c *Client) GetSlot(ctx context.Context, commitment CommitmentType) (uint64, error) {
	cache := c.cache.Load()
	if cache != nil {
		if slot, ok := cache.lookupSlot(commitment); ok {
			return slot, nil
		}
	}
	var opts *GetSlotOpts
	if commitment != "" {
		opts = &GetSlotOpts{Commitment: commitment}
	}
	slot, err := c.GetSlotWithOpts(ctx, opts)
	if err == nil && cache != nil {
		cache.storeSlot(commitment, slot)
	}
	return slot, err
}

// GetSlotWithOpts is GetSlot with explicit options. It always queries the RPC
// endpoint, bypassing the chain-head cache.
func (c *Client) GetSlotWithOpts(ctx context.Context, opts *GetSlotOpts) (uint64, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	return call[uint64](ctx, c, "getSlot", params...)
}

// GetSlotLeader returns the identity of the current slot leader via
// getSlotLeader.
func (c *Client) GetSlotLeader(ctx context.Context, commitment CommitmentType) (solana.PublicKey, error) {
	var opts *GetSlotLeaderOpts
	if commitment != "" {
		opts = &GetSlotLeaderOpts{Commitment: commitment}
	}
	return c.GetSlotLeaderWithOpts(ctx, opts)
}

// GetSlotLeaderWithOpts is GetSlotLeader with explicit options.
func (c *Client) GetSlotLeaderWithOpts(ctx context.Context, opts *GetSlotLeaderOpts) (solana.PublicKey, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	return call[solana.PublicKey](ctx, c, "getSlotLeader", params...)
}

// GetSlotLeaders returns the slot leaders for limit slots starting at slot
// start, via getSlotLeaders.
func (c *Client) GetSlotLeaders(ctx context.Context, start uint64, limit uint64) ([]solana.PublicKey, error) {
	return call[[]solana.PublicKey](ctx, c, "getSlotLeaders", start, limit)
}

// GetClusterNodes returns information about all the nodes participating in
// the cluster, via getClusterNodes.
func (c *Client) GetClusterNodes(ctx context.Context) ([]GetClusterNodesResult, error) {
	return call[[]GetClusterNodesResult](ctx, c, "getClusterNodes")
}

// GetVersion returns the solana-core software version running on the node,
// via getVersion.
func (c *Client) GetVersion(ctx context.Context) (*GetVersionResult, error) {
	result, err := call[GetVersionResult](ctx, c, "getVersion")
	return &result, err
}

// GetHealth returns the node's health via getHealth: "ok" when healthy, an
// RPC error otherwise.
func (c *Client) GetHealth(ctx context.Context) (string, error) {
	return call[string](ctx, c, "getHealth")
}

// GetIdentity returns the identity public key of the node via getIdentity.
func (c *Client) GetIdentity(ctx context.Context) (*GetIdentityResult, error) {
	result, err := call[GetIdentityResult](ctx, c, "getIdentity")
	return &result, err
}

// GetGenesisHash returns the cluster's genesis hash via getGenesisHash.
func (c *Client) GetGenesisHash(ctx context.Context) (solana.Hash, error) {
	return call[solana.Hash](ctx, c, "getGenesisHash")
}

// GetHighestSnapshotSlot returns the highest slots the node has full and
// incremental snapshots for, via getHighestSnapshotSlot.
func (c *Client) GetHighestSnapshotSlot(ctx context.Context) (*GetHighestSnapshotSlotResult, error) {
	result, err := call[GetHighestSnapshotSlotResult](ctx, c, "getHighestSnapshotSlot")
	return &result, err
}

// GetMaxRetransmitSlot returns the max slot seen from the retransmit stage,
// via getMaxRetransmitSlot.
func (c *Client) GetMaxRetransmitSlot(ctx context.Context) (uint64, error) {
	return call[uint64](ctx, c, "getMaxRetransmitSlot")
}

// GetMaxShredInsertSlot returns the max slot seen after shred insert, via
// getMaxShredInsertSlot.
func (c *Client) GetMaxShredInsertSlot(ctx context.Context) (uint64, error) {
	return call[uint64](ctx, c, "getMaxShredInsertSlot")
}

// MinimumLedgerSlot returns the lowest slot the node has information about
// in its ledger, via minimumLedgerSlot.
func (c *Client) MinimumLedgerSlot(ctx context.Context) (uint64, error) {
	return call[uint64](ctx, c, "minimumLedgerSlot")
}

// GetRecentPerformanceSamples returns recent per-slot performance samples
// via getRecentPerformanceSamples. A nil limit leaves the sample count to
// the node.
func (c *Client) GetRecentPerformanceSamples(ctx context.Context, limit *uint) ([]GetRecentPerformanceSamplesResult, error) {
	params := []any{}
	if limit != nil {
		params = append(params, *limit)
	}
	return call[[]GetRecentPerformanceSamplesResult](ctx, c, "getRecentPerformanceSamples", params...)
}

// --- Epoch, inflation & supply ---

// GetEpochInfo returns information about the current epoch via getEpochInfo.
// Commitment "" uses the node default.
func (c *Client) GetEpochInfo(ctx context.Context, commitment CommitmentType) (*GetEpochInfoResult, error) {
	var opts *GetEpochInfoOpts
	if commitment != "" {
		opts = &GetEpochInfoOpts{Commitment: commitment}
	}
	return c.GetEpochInfoWithOpts(ctx, opts)
}

// GetEpochInfoWithOpts is GetEpochInfo with explicit options.
func (c *Client) GetEpochInfoWithOpts(ctx context.Context, opts *GetEpochInfoOpts) (*GetEpochInfoResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetEpochInfoResult](ctx, c, "getEpochInfo", params...)
	return &result, err
}

// GetEpochSchedule returns the cluster's epoch schedule configuration via
// getEpochSchedule.
func (c *Client) GetEpochSchedule(ctx context.Context) (*GetEpochScheduleResult, error) {
	result, err := call[GetEpochScheduleResult](ctx, c, "getEpochSchedule")
	return &result, err
}

// GetInflationGovernor returns the current inflation governor parameters via
// getInflationGovernor.
func (c *Client) GetInflationGovernor(ctx context.Context, commitment CommitmentType) (*GetInflationGovernorResult, error) {
	result, err := call[GetInflationGovernorResult](ctx, c, "getInflationGovernor", commitmentParam(commitment)...)
	return &result, err
}

// GetInflationRate returns the inflation values for the current epoch via
// getInflationRate.
func (c *Client) GetInflationRate(ctx context.Context) (*GetInflationRateResult, error) {
	result, err := call[GetInflationRateResult](ctx, c, "getInflationRate")
	return &result, err
}

// GetInflationReward returns the inflation (staking) reward for the given
// addresses for an epoch, via getInflationReward. Addresses without a reward
// come back as nil entries.
func (c *Client) GetInflationReward(ctx context.Context, addresses []solana.PublicKey, opts *GetInflationRewardOpts) ([]*GetInflationRewardResult, error) {
	params := []any{addresses}
	if opts != nil {
		params = append(params, opts)
	}
	return call[[]*GetInflationRewardResult](ctx, c, "getInflationReward", params...)
}

// GetSupply returns information about the current lamport supply via
// getSupply. Commitment "" uses the node default.
func (c *Client) GetSupply(ctx context.Context, commitment CommitmentType) (*GetSupplyResult, error) {
	return c.GetSupplyWithOpts(ctx, &GetSupplyOpts{Commitment: commitment})
}

// GetSupplyWithOpts is GetSupply with explicit options (commitment,
// excluding the non-circulating accounts list).
func (c *Client) GetSupplyWithOpts(ctx context.Context, opts *GetSupplyOpts) (*GetSupplyResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetSupplyResult](ctx, c, "getSupply", params...)
	return &result, err
}

// GetLeaderSchedule returns the leader schedule for the current epoch via
// getLeaderSchedule.
func (c *Client) GetLeaderSchedule(ctx context.Context) (GetLeaderScheduleResult, error) {
	return c.GetLeaderScheduleWithOpts(ctx, nil)
}

// GetLeaderScheduleWithOpts is GetLeaderSchedule with explicit options: an
// epoch (nil means the current one) and an optional identity filter and
// commitment.
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
	result, err := callNullable[GetLeaderScheduleResult](ctx, c, "getLeaderSchedule", params...)
	if err != nil {
		return nil, err
	}
	return *result, nil
}

// GetVoteAccounts returns the current and delinquent vote accounts via
// getVoteAccounts. A nil opts uses the node defaults.
func (c *Client) GetVoteAccounts(ctx context.Context, opts *GetVoteAccountsOpts) (*GetVoteAccountsResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetVoteAccountsResult](ctx, c, "getVoteAccounts", params...)
	return &result, err
}

// GetStakeMinimumDelegation returns the stake minimum delegation in
// lamports, via getStakeMinimumDelegation.
func (c *Client) GetStakeMinimumDelegation(ctx context.Context, commitment CommitmentType) (*GetStakeMinimumDelegationResult, error) {
	var opts *GetStakeMinimumDelegationOpts
	if commitment != "" {
		opts = &GetStakeMinimumDelegationOpts{Commitment: commitment}
	}
	return c.GetStakeMinimumDelegationWithOpts(ctx, opts)
}

// GetStakeMinimumDelegationWithOpts is GetStakeMinimumDelegation with
// explicit options.
func (c *Client) GetStakeMinimumDelegationWithOpts(ctx context.Context, opts *GetStakeMinimumDelegationOpts) (*GetStakeMinimumDelegationResult, error) {
	params := []any{}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetStakeMinimumDelegationResult](ctx, c, "getStakeMinimumDelegation", params...)
	return &result, err
}

// --- Tokens ---

// GetTokenAccountBalance returns the token balance of an SPL token account
// via getTokenAccountBalance. Commitment "" uses the node default.
func (c *Client) GetTokenAccountBalance(ctx context.Context, account solana.PublicKey, commitment CommitmentType) (*GetTokenAccountBalanceResult, error) {
	var opts *GetTokenAccountBalanceOpts
	if commitment != "" {
		opts = &GetTokenAccountBalanceOpts{Commitment: commitment}
	}
	return c.GetTokenAccountBalanceWithOpts(ctx, account, opts)
}

// GetTokenAccountBalanceWithOpts is GetTokenAccountBalance with explicit
// options.
func (c *Client) GetTokenAccountBalanceWithOpts(ctx context.Context, account solana.PublicKey, opts *GetTokenAccountBalanceOpts) (*GetTokenAccountBalanceResult, error) {
	params := []any{account}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetTokenAccountBalanceResult](ctx, c, "getTokenAccountBalance", params...)
	return &result, err
}

// GetTokenAccountsByOwner returns all SPL token accounts of the given owner
// via getTokenAccountsByOwner, filtered by the mint or token program in
// config.
func (c *Client) GetTokenAccountsByOwner(ctx context.Context, owner solana.PublicKey, config *GetTokenAccountsConfig, opts *GetTokenAccountsOpts) (*GetTokenAccountsResult, error) {
	return c.getTokenAccountsBy(ctx, "getTokenAccountsByOwner", owner, config, opts)
}

// GetTokenAccountsByDelegate returns all SPL token accounts with the given
// approved delegate via getTokenAccountsByDelegate, filtered by the mint or
// token program in config.
func (c *Client) GetTokenAccountsByDelegate(ctx context.Context, delegate solana.PublicKey, config *GetTokenAccountsConfig, opts *GetTokenAccountsOpts) (*GetTokenAccountsResult, error) {
	return c.getTokenAccountsBy(ctx, "getTokenAccountsByDelegate", delegate, config, opts)
}

func (c *Client) getTokenAccountsBy(ctx context.Context, method string, key solana.PublicKey, config *GetTokenAccountsConfig, opts *GetTokenAccountsOpts) (*GetTokenAccountsResult, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	params := []any{key, config}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetTokenAccountsResult](ctx, c, method, params...)
	return &result, err
}

// GetTokenLargestAccounts returns the 20 largest accounts of an SPL token
// mint via getTokenLargestAccounts. Commitment "" uses the node default.
func (c *Client) GetTokenLargestAccounts(ctx context.Context, mint solana.PublicKey, commitment CommitmentType) (*GetTokenLargestAccountsResult, error) {
	result, err := call[GetTokenLargestAccountsResult](ctx, c, "getTokenLargestAccounts", withCommitment([]any{mint}, commitment)...)
	return &result, err
}

// GetTokenSupply returns the total supply of an SPL token mint via
// getTokenSupply. Commitment "" uses the node default.
func (c *Client) GetTokenSupply(ctx context.Context, mint solana.PublicKey, commitment CommitmentType) (*GetTokenSupplyResult, error) {
	result, err := call[GetTokenSupplyResult](ctx, c, "getTokenSupply", withCommitment([]any{mint}, commitment)...)
	return &result, err
}

// --- FluxRPC extensions ---

// GetTransactionsForAddress returns transactions involving address using
// FluxRPC's indexed transaction-history endpoint.
func (c *Client) GetTransactionsForAddress(ctx context.Context, address solana.PublicKey, opts *GetTransactionsForAddressOpts) (*GetTransactionsForAddressResult, error) {
	params := []any{address}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetTransactionsForAddressResult](ctx, c, "getTransactionsForAddress", params...)
	return &result, err
}

// GetParsedTransactionsForAddress returns full transaction history with
// jsonParsed transactions and metadata. The caller's options are copied.
func (c *Client) GetParsedTransactionsForAddress(ctx context.Context, address solana.PublicKey, opts *GetTransactionsForAddressOpts) (*GetParsedTransactionsForAddressResult, error) {
	requestOpts := GetTransactionsForAddressOpts{}
	if opts != nil {
		requestOpts = *opts
	}
	requestOpts.TransactionDetails = TransactionDetailsFull
	requestOpts.Encoding = solana.EncodingJSONParsed
	result, err := call[GetParsedTransactionsForAddressResult](ctx, c, "getTransactionsForAddress", address, &requestOpts)
	return &result, err
}

// GetPriorityFeeEstimate estimates a compute-unit price using FluxRPC's local
// fee-market history.
func (c *Client) GetPriorityFeeEstimate(ctx context.Context, request GetPriorityFeeEstimateRequest) (*GetPriorityFeeEstimateResult, error) {
	result, err := call[GetPriorityFeeEstimateResult](ctx, c, "getPriorityFeeEstimate", request)
	return &result, err
}

// GetTokenAccounts returns FluxRPC's holder-index entries for a token mint.
// A nil opts performs a full scan; a non-nil Limit requests one page.
func (c *Client) GetTokenAccounts(ctx context.Context, mint solana.PublicKey, opts *GetTokenAccountsIndexOpts) (*GetTokenAccountsIndexResult, error) {
	params := []any{mint}
	if opts != nil {
		params = append(params, opts)
	}
	result, err := call[GetTokenAccountsIndexResult](ctx, c, "getTokenAccounts", params...)
	return &result, err
}

// GetTokenAccountsCount returns FluxRPC's indexed token-account count for a
// mint.
func (c *Client) GetTokenAccountsCount(ctx context.Context, mint solana.PublicKey, opts *GetTokenAccountsCountOpts) (uint64, error) {
	params := []any{mint}
	if opts != nil {
		params = append(params, opts)
	}
	return call[uint64](ctx, c, "getTokenAccountsCount", params...)
}

// GetUpcomingLeaders returns the next amount leader groups from FluxRPC.
func (c *Client) GetUpcomingLeaders(ctx context.Context, amount uint64) (*GetUpcomingLeadersResult, error) {
	result, err := call[GetUpcomingLeadersResult](ctx, c, "getUpcomingLeaders", amount)
	return &result, err
}
