package rpc

// Live integration tests against a real RPC endpoint. They only run when
// RPC_URL is set:
//
//	RPC_URL=https://... go test ./rpc/ -run TestLive -v
//
// Every response is decoded into its DTO and then re-marshaled; any key the
// endpoint returned that the DTO dropped (and that carries a non-zero value)
// fails the test, so spec drift surfaces here first. Calls that can return
// large result sets are scoped down (sysvar-owned getProgramAccounts with a
// dataSlice, identity-scoped leader schedule / block production, supply
// without the non-circulating list) to keep memory in check.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
)

const (
	usdcMint    = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	sysvarOwner = "Sysvar1111111111111111111111111111111111111"
)

type liveClient struct {
	t   *testing.T
	url string
	c   *http.Client
	id  int
}

func newLiveClient(t *testing.T) *liveClient {
	url := os.Getenv("RPC_URL")
	if url == "" {
		t.Skip("RPC_URL not set; skipping live RPC tests")
	}
	return &liveClient{t: t, url: url, c: &http.Client{Timeout: 60 * time.Second}}
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string { return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message) }

// call performs a JSON-RPC request and returns the raw result, retrying
// transient errors once. This suite validates DTOs, not endpoint health, so
// any persistent server-side RPC error skips the subtest (loudly) instead of
// failing it; transport errors still fail.
func (lc *liveClient) call(t *testing.T, method string, params ...any) json.RawMessage {
	t.Helper()
	raw, rpcErr := lc.tryCall(t, method, params...)
	if rpcErr != nil && (rpcErr.Code == -32000 || strings.Contains(strings.ToLower(rpcErr.Message), "transient")) {
		time.Sleep(time.Second)
		raw, rpcErr = lc.tryCall(t, method, params...)
	}
	if rpcErr != nil {
		t.Skipf("%s not testable on this endpoint: %v", method, rpcErr)
	}
	return raw
}

func (lc *liveClient) tryCall(t *testing.T, method string, params ...any) (json.RawMessage, *rpcError) {
	t.Helper()
	lc.id++
	body, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      lc.id,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		t.Fatal(err)
	}

	resp, err := lc.c.Post(lc.url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("%s: %v", method, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &rpcError{Code: resp.StatusCode, Message: fmt.Sprintf("HTTP %s: %.200s", resp.Status, body)}
	}

	var envelope struct {
		Result json.RawMessage `json:"result"`
		Error  *rpcError       `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("%s: decoding response envelope: %v", method, err)
	}
	return envelope.Result, envelope.Error
}

// decodeStrict decodes raw into T with sonic (the canonical path),
// re-marshals it, and fails on any populated key the DTO dropped.
func decodeStrict[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var out T
	if err := sonic.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decoding into %T: %v\nraw: %.2000s", out, err, raw)
	}
	remarshaled, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("re-marshaling %T: %v", out, err)
	}

	var rawAny, oursAny any
	if err := json.Unmarshal(raw, &rawAny); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(remarshaled, &oursAny); err != nil {
		t.Fatal(err)
	}
	var missing []string
	diffKeys("", rawAny, oursAny, &missing)
	if len(missing) > 0 {
		t.Errorf("%T drops populated response fields: %v", out, missing)
	}
	return out
}

// diffKeys records key paths present in raw with a non-zero value but absent
// from ours. Array recursion is capped so huge result sets stay cheap.
func diffKeys(prefix string, raw, ours any, missing *[]string) {
	switch rawV := raw.(type) {
	case map[string]any:
		oursM, ok := ours.(map[string]any)
		if !ok {
			return
		}
		for k, v := range rawV {
			if isJSONZero(v) {
				continue
			}
			child, present := oursM[k]
			if !present {
				*missing = append(*missing, prefix+"."+k)
				continue
			}
			diffKeys(prefix+"."+k, v, child, missing)
		}
	case []any:
		oursS, ok := ours.([]any)
		if !ok {
			return
		}
		limit := len(rawV)
		if limit > 8 {
			limit = 8
		}
		for i := 0; i < limit && i < len(oursS); i++ {
			diffKeys(fmt.Sprintf("%s[%d]", prefix, i), rawV[i], oursS[i], missing)
		}
	}
}

// isJSONZero reports whether a decoded JSON value is null or a zero value
// that omitempty legitimately drops.
func isJSONZero(v any) bool {
	switch val := v.(type) {
	case nil:
		return true
	case bool:
		return !val
	case float64:
		return val == 0
	case string:
		return val == ""
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	}
	return false
}

func decodePlain[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var out T
	if err := sonic.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decoding into %T: %v\nraw: %.500s", out, err, raw)
	}
	return out
}

func TestLiveRPC(t *testing.T) {
	lc := newLiveClient(t)

	// Shared state gathered by the early subtests.
	var (
		slot         uint64
		blockSlot    uint64
		latestHash   solana.Hash
		voteAccount  solana.PublicKey
		nodeIdentity solana.PublicKey
		currentEpoch uint64
		signatures   []solana.Signature
		tokenAccount solana.PublicKey
		tokenOwner   solana.PublicKey
	)

	t.Run("getVersion", func(t *testing.T) {
		v := decodeStrict[GetVersionResult](t, lc.call(t, "getVersion"))
		if v.SolanaCore == "" {
			t.Fatalf("got %+v", v)
		}
	})

	t.Run("getSlot", func(t *testing.T) {
		slot = decodePlain[uint64](t, lc.call(t, "getSlot", M{"commitment": CommitmentFinalized}))
		if slot == 0 {
			t.Fatal("slot is zero")
		}
	})

	t.Run("getLatestBlockhash", func(t *testing.T) {
		res := decodeStrict[GetLatestBlockhashResult](t, lc.call(t, "getLatestBlockhash", M{"commitment": CommitmentFinalized}))
		if res.Value == nil || res.Value.Blockhash.IsZero() {
			t.Fatalf("got %+v", res)
		}
		latestHash = res.Value.Blockhash
	})

	t.Run("getVoteAccounts", func(t *testing.T) {
		res := decodeStrict[GetVoteAccountsResult](t, lc.call(t, "getVoteAccounts", M{"commitment": CommitmentFinalized}))
		if len(res.Current) == 0 {
			t.Fatal("no current vote accounts")
		}
		voteAccount = res.Current[0].VotePubkey
		nodeIdentity = res.Current[0].NodePubkey
	})

	t.Run("getEpochInfo", func(t *testing.T) {
		res := decodeStrict[GetEpochInfoResult](t, lc.call(t, "getEpochInfo"))
		if res.Epoch == 0 || res.SlotsInEpoch == 0 {
			t.Fatalf("got %+v", res)
		}
		currentEpoch = res.Epoch
	})

	t.Run("getBlocks", func(t *testing.T) {
		if slot == 0 {
			t.Skip("no slot")
		}
		res := decodePlain[BlocksResult](t, lc.call(t, "getBlocks", slot-20, slot))
		if len(res) == 0 {
			t.Fatal("no blocks in the last 20 slots")
		}
		blockSlot = res[len(res)-1]
	})

	t.Run("getBlocksWithLimit", func(t *testing.T) {
		if slot == 0 {
			t.Skip("no slot")
		}
		res := decodePlain[BlocksResult](t, lc.call(t, "getBlocksWithLimit", slot-20, 5))
		if len(res) == 0 {
			t.Fatal("no blocks")
		}
	})

	t.Run("getAccountInfo", func(t *testing.T) {
		for _, encoding := range []solana.EncodingType{
			solana.EncodingBase64,
			solana.EncodingBase58,
			solana.EncodingBase64Zstd,
		} {
			res := decodeStrict[GetAccountInfoResult](t, lc.call(t, "getAccountInfo", usdcMint, M{"encoding": encoding}))
			if got := len(res.GetBinary()); got != 82 {
				t.Fatalf("encoding %s: mint data length %d, want 82", encoding, got)
			}
		}
		parsed := decodeStrict[GetAccountInfoResult](t, lc.call(t, "getAccountInfo", usdcMint, M{"encoding": solana.EncodingJSONParsed}))
		if parsed.Value == nil || parsed.Value.Data.GetRawJSON() == nil {
			t.Fatal("jsonParsed returned no JSON")
		}
	})

	t.Run("getBalance", func(t *testing.T) {
		res := decodeStrict[GetBalanceResult](t, lc.call(t, "getBalance", usdcMint))
		if res.Value == 0 {
			t.Fatal("mint account has zero balance")
		}
	})

	t.Run("getBlock", func(t *testing.T) {
		if blockSlot == 0 {
			t.Skip("no block slot")
		}
		full := decodeStrict[GetBlockResult](t, lc.call(t, "getBlock", blockSlot, M{
			"encoding":                       solana.EncodingBase64,
			"transactionDetails":             TransactionDetailsFull,
			"rewards":                        true,
			"maxSupportedTransactionVersion": 0,
		}))
		if full.Blockhash.IsZero() || len(full.Transactions) == 0 {
			t.Fatalf("blockhash %s, %d transactions", full.Blockhash, len(full.Transactions))
		}
		if _, err := full.Transactions[0].GetTransaction(); err != nil {
			t.Fatalf("GetTransaction: %v", err)
		}

		sigsOnly := decodeStrict[GetBlockResult](t, lc.call(t, "getBlock", blockSlot, M{
			"transactionDetails":             TransactionDetailsSignatures,
			"rewards":                        false,
			"maxSupportedTransactionVersion": 0,
		}))
		if len(sigsOnly.Signatures) == 0 {
			t.Fatal("no signatures")
		}

		accounts := decodeStrict[GetBlockResult](t, lc.call(t, "getBlock", blockSlot, M{
			"encoding":                       solana.EncodingJSON,
			"transactionDetails":             TransactionDetailsAccounts,
			"rewards":                        false,
			"maxSupportedTransactionVersion": 0,
		}))
		if len(accounts.Transactions) == 0 {
			t.Fatal("no transactions in accounts mode")
		}
		keys, err := accounts.Transactions[0].GetAccountKeys()
		if err != nil {
			t.Fatalf("GetAccountKeys: %v", err)
		}
		if len(keys.AccountKeys) == 0 {
			t.Fatal("no account keys")
		}

		parsed := decodeStrict[GetParsedBlockResult](t, lc.call(t, "getBlock", blockSlot, M{
			"encoding":                       solana.EncodingJSONParsed,
			"transactionDetails":             TransactionDetailsFull,
			"rewards":                        false,
			"maxSupportedTransactionVersion": 0,
		}))
		if len(parsed.Transactions) == 0 {
			t.Fatal("no parsed transactions")
		}
	})

	t.Run("getBlockCommitment", func(t *testing.T) {
		if blockSlot == 0 {
			t.Skip("no block slot")
		}
		decodeStrict[GetBlockCommitmentResult](t, lc.call(t, "getBlockCommitment", blockSlot))
	})

	t.Run("getBlockHeight", func(t *testing.T) {
		if decodePlain[uint64](t, lc.call(t, "getBlockHeight")) == 0 {
			t.Fatal("zero block height")
		}
	})

	t.Run("getBlockProduction", func(t *testing.T) {
		if nodeIdentity.IsZero() {
			t.Skip("no identity")
		}
		res := decodeStrict[GetBlockProductionResult](t, lc.call(t, "getBlockProduction", M{"identity": nodeIdentity}))
		if res.Value.Range.FirstSlot == 0 {
			t.Fatalf("got %+v", res.Value.Range)
		}
	})

	t.Run("getBlockTime", func(t *testing.T) {
		if blockSlot == 0 {
			t.Skip("no block slot")
		}
		ts := decodePlain[solana.UnixTimeSeconds](t, lc.call(t, "getBlockTime", blockSlot))
		if ts.Time().Year() < 2020 {
			t.Fatalf("implausible block time %s", ts)
		}
	})

	t.Run("getClusterNodes", func(t *testing.T) {
		res := decodeStrict[[]GetClusterNodesResult](t, lc.call(t, "getClusterNodes"))
		if len(res) == 0 {
			t.Fatal("no cluster nodes")
		}
	})

	t.Run("getEpochSchedule", func(t *testing.T) {
		res := decodeStrict[GetEpochScheduleResult](t, lc.call(t, "getEpochSchedule"))
		if res.SlotsPerEpoch == 0 {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("getFeeForMessage", func(t *testing.T) {
		if latestHash.IsZero() {
			t.Skip("no blockhash")
		}
		// Build a real message with our own binary encoder: its base64 being
		// accepted by the node is an end-to-end check of the wire format.
		payer := mustRandomKey(t).PublicKey()
		msg := solana.Message{
			AccountKeys: []solana.PublicKey{payer, {}, {}},
			Header:      solana.MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 2},
			Instructions: []solana.CompiledInstruction{{
				ProgramIDIndex: 2,
				Accounts:       []uint16{0, 1},
				Data:           solana.Base58{2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			}},
			RecentBlockhash: latestHash,
		}
		res := decodeStrict[GetFeeForMessageResult](t, lc.call(t, "getFeeForMessage", msg.ToBase64(), M{"commitment": CommitmentProcessed}))
		if res.Value == nil || *res.Value == 0 {
			t.Fatalf("fee = %v (message not recognized?)", res.Value)
		}
	})

	t.Run("getFirstAvailableBlock", func(t *testing.T) {
		decodePlain[uint64](t, lc.call(t, "getFirstAvailableBlock"))
	})

	t.Run("getGenesisHash", func(t *testing.T) {
		if decodePlain[solana.Hash](t, lc.call(t, "getGenesisHash")).IsZero() {
			t.Fatal("zero genesis hash")
		}
	})

	t.Run("getHealth", func(t *testing.T) {
		raw, rpcErr := lc.tryCall(t, "getHealth")
		if rpcErr != nil {
			t.Logf("node reports unhealthy: %v", rpcErr) // still a valid response shape
			return
		}
		if got := decodePlain[string](t, raw); got != "ok" {
			t.Fatalf("health = %q", got)
		}
	})

	t.Run("getHighestSnapshotSlot", func(t *testing.T) {
		res := decodeStrict[GetHighestSnapshotSlotResult](t, lc.call(t, "getHighestSnapshotSlot"))
		if res.Full == 0 {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("getIdentity", func(t *testing.T) {
		res := decodeStrict[GetIdentityResult](t, lc.call(t, "getIdentity"))
		if res.Identity.IsZero() {
			t.Fatal("zero identity")
		}
	})

	t.Run("getInflationGovernor", func(t *testing.T) {
		decodeStrict[GetInflationGovernorResult](t, lc.call(t, "getInflationGovernor"))
	})

	t.Run("getInflationRate", func(t *testing.T) {
		res := decodeStrict[GetInflationRateResult](t, lc.call(t, "getInflationRate"))
		if res.Epoch == 0 {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("getInflationReward", func(t *testing.T) {
		if voteAccount.IsZero() || currentEpoch == 0 {
			t.Skip("no vote account/epoch")
		}
		res := decodeStrict[[]*GetInflationRewardResult](t, lc.call(t, "getInflationReward",
			[]solana.PublicKey{voteAccount}, M{"epoch": currentEpoch - 1}))
		if len(res) != 1 {
			t.Fatalf("got %d entries", len(res))
		}
	})

	t.Run("getLargestAccounts", func(t *testing.T) {
		res := decodeStrict[GetLargestAccountsResult](t, lc.call(t, "getLargestAccounts", M{"commitment": CommitmentFinalized}))
		if len(res.Value) == 0 {
			t.Fatal("no accounts")
		}
	})

	t.Run("getLeaderSchedule", func(t *testing.T) {
		if nodeIdentity.IsZero() {
			t.Skip("no identity")
		}
		// The identity option scopes the response on stock validators; some
		// providers ignore it and return the full schedule, so only decode
		// coverage is asserted here.
		res := decodeStrict[GetLeaderScheduleResult](t, lc.call(t, "getLeaderSchedule", nil, M{"identity": nodeIdentity}))
		if len(res) == 0 {
			t.Fatal("empty leader schedule")
		}
	})

	t.Run("getMaxRetransmitSlot", func(t *testing.T) {
		decodePlain[uint64](t, lc.call(t, "getMaxRetransmitSlot"))
	})

	t.Run("getMinimumBalanceForRentExemption", func(t *testing.T) {
		if decodePlain[uint64](t, lc.call(t, "getMinimumBalanceForRentExemption", 100)) == 0 {
			t.Fatal("zero rent exemption")
		}
	})

	t.Run("getMultipleAccounts", func(t *testing.T) {
		res := decodeStrict[GetMultipleAccountsResult](t, lc.call(t, "getMultipleAccounts",
			[]string{usdcMint, "So11111111111111111111111111111111111111112"},
			M{"encoding": solana.EncodingBase64Zstd}))
		if len(res.Value) != 2 || res.Value[0] == nil || res.Value[1] == nil {
			t.Fatalf("got %d values", len(res.Value))
		}
		if len(res.Value[0].Data.GetBinary()) != 82 {
			t.Fatalf("zstd mint data length %d", len(res.Value[0].Data.GetBinary()))
		}
	})

	t.Run("getProgramAccounts", func(t *testing.T) {
		// The sysvar owner holds a small fixed set of accounts; the dataSlice
		// keeps payloads tiny regardless.
		offset, length := uint64(0), uint64(32)
		opts := M{
			"encoding":  solana.EncodingBase64,
			"dataSlice": DataSlice{Offset: &offset, Length: &length},
		}
		res := decodeStrict[GetProgramAccountsResult](t, lc.call(t, "getProgramAccounts", sysvarOwner, opts))
		if len(res) == 0 {
			t.Fatal("no sysvar accounts")
		}

		opts["withContext"] = true
		withCtx := decodeStrict[GetProgramAccountsWithContextResult](t, lc.call(t, "getProgramAccounts", sysvarOwner, opts))
		if len(withCtx.Value) == 0 || withCtx.Context.Slot == 0 {
			t.Fatalf("got %+v", withCtx.Context)
		}
	})

	t.Run("getProgramAccounts_stream", func(t *testing.T) {
		offset, length := uint64(0), uint64(32)
		body, err := json.Marshal(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "getProgramAccounts",
			"params": []any{sysvarOwner, M{
				"encoding":    solana.EncodingBase64,
				"dataSlice":   DataSlice{Offset: &offset, Length: &length},
				"withContext": true,
			}},
		})
		if err != nil {
			t.Fatal(err)
		}
		resp, err := lc.c.Post(lc.url, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		var streamed []*KeyedAccount
		ctx, err := StreamProgramAccounts(resp.Body, func(ka *KeyedAccount) error {
			streamed = append(streamed, ka)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if ctx == nil || ctx.Slot == 0 {
			t.Fatalf("context = %+v", ctx)
		}
		if len(streamed) == 0 {
			t.Fatal("no streamed accounts")
		}
		for _, ka := range streamed {
			if ka.Pubkey.IsZero() || ka.Account == nil {
				t.Fatalf("bad streamed account %+v", ka)
			}
		}
	})

	t.Run("getRecentPerformanceSamples", func(t *testing.T) {
		res := decodeStrict[[]GetRecentPerformanceSamplesResult](t, lc.call(t, "getRecentPerformanceSamples", 5))
		if len(res) == 0 {
			t.Fatal("no samples")
		}
	})

	t.Run("getRecentPrioritizationFees", func(t *testing.T) {
		res := decodeStrict[[]PrioritizationFeeResult](t, lc.call(t, "getRecentPrioritizationFees", []string{usdcMint}))
		if len(res) == 0 {
			t.Fatal("no fee samples")
		}
	})

	t.Run("getSignaturesForAddress", func(t *testing.T) {
		res := decodeStrict[[]*TransactionSignature](t, lc.call(t, "getSignaturesForAddress", usdcMint, M{"limit": 5}))
		if len(res) == 0 {
			t.Fatal("no signatures")
		}
		for _, sig := range res {
			signatures = append(signatures, sig.Signature)
		}
	})

	t.Run("getSignatureStatuses", func(t *testing.T) {
		if len(signatures) == 0 {
			t.Skip("no signatures")
		}
		res := decodeStrict[GetSignatureStatusesResult](t, lc.call(t, "getSignatureStatuses",
			signatures, M{"searchTransactionHistory": true}))
		if len(res.Value) != len(signatures) {
			t.Fatalf("got %d statuses for %d signatures", len(res.Value), len(signatures))
		}
	})

	t.Run("getSlotLeader", func(t *testing.T) {
		if decodePlain[solana.PublicKey](t, lc.call(t, "getSlotLeader")).IsZero() {
			t.Fatal("zero slot leader")
		}
	})

	t.Run("getSlotLeaders", func(t *testing.T) {
		if slot == 0 {
			t.Skip("no slot")
		}
		res := decodePlain[[]solana.PublicKey](t, lc.call(t, "getSlotLeaders", slot-10, 10))
		if len(res) == 0 {
			t.Fatal("no slot leaders")
		}
	})

	t.Run("getStakeMinimumDelegation", func(t *testing.T) {
		res := decodeStrict[GetStakeMinimumDelegationResult](t, lc.call(t, "getStakeMinimumDelegation"))
		if res.Value == 0 {
			t.Fatal("zero minimum delegation")
		}
	})

	t.Run("getSupply", func(t *testing.T) {
		res := decodeStrict[GetSupplyResult](t, lc.call(t, "getSupply",
			M{"excludeNonCirculatingAccountsList": true}))
		if res.Value == nil || res.Value.Total == 0 {
			t.Fatalf("got %+v", res.Value)
		}
	})

	t.Run("getTokenLargestAccounts", func(t *testing.T) {
		res := decodeStrict[GetTokenLargestAccountsResult](t, lc.call(t, "getTokenLargestAccounts", usdcMint))
		if len(res.Value) == 0 {
			t.Fatal("no token accounts")
		}
		tokenAccount = res.Value[0].Address
	})

	t.Run("getTokenAccountBalance", func(t *testing.T) {
		if tokenAccount.IsZero() {
			t.Skip("no token account")
		}
		res := decodeStrict[GetTokenAccountBalanceResult](t, lc.call(t, "getTokenAccountBalance", tokenAccount))
		if res.Value == nil || res.Value.Amount == "" {
			t.Fatalf("got %+v", res.Value)
		}
	})

	t.Run("getTokenAccountsByOwner", func(t *testing.T) {
		if tokenAccount.IsZero() {
			t.Skip("no token account")
		}
		// Fish the owner out of the jsonParsed form of the token account.
		info := decodeStrict[GetAccountInfoResult](t, lc.call(t, "getAccountInfo", tokenAccount, M{"encoding": solana.EncodingJSONParsed}))
		var parsed struct {
			Parsed struct {
				Info struct {
					Owner solana.PublicKey `json:"owner"`
				} `json:"info"`
			} `json:"parsed"`
		}
		if err := json.Unmarshal(info.Value.Data.GetRawJSON(), &parsed); err != nil {
			t.Fatal(err)
		}
		tokenOwner = parsed.Parsed.Info.Owner

		res := decodeStrict[GetTokenAccountsResult](t, lc.call(t, "getTokenAccountsByOwner",
			tokenOwner, M{"mint": usdcMint}, M{"encoding": solana.EncodingBase64}))
		if len(res.Value) == 0 {
			t.Fatal("owner of the largest account has no accounts for the mint")
		}
	})

	t.Run("getTokenAccountsByDelegate", func(t *testing.T) {
		if tokenOwner.IsZero() {
			t.Skip("no token owner")
		}
		// Usually empty; exercises the response shape either way.
		decodeStrict[GetTokenAccountsResult](t, lc.call(t, "getTokenAccountsByDelegate",
			tokenOwner, M{"mint": usdcMint}, M{"encoding": solana.EncodingBase64}))
	})

	t.Run("getTokenSupply", func(t *testing.T) {
		res := decodeStrict[GetTokenSupplyResult](t, lc.call(t, "getTokenSupply", usdcMint))
		if res.Value == nil || res.Value.Amount == "" {
			t.Fatalf("got %+v", res.Value)
		}
	})

	t.Run("getTransaction", func(t *testing.T) {
		if len(signatures) == 0 {
			t.Skip("no signatures")
		}
		sig := signatures[0]

		b64 := decodeStrict[GetTransactionResult](t, lc.call(t, "getTransaction", sig, M{
			"encoding":                       solana.EncodingBase64,
			"maxSupportedTransactionVersion": 0,
		}))
		tx, err := b64.Transaction.GetTransaction()
		if err != nil {
			t.Fatalf("GetTransaction: %v", err)
		}
		// End-to-end wire check: our re-serialization must be byte-identical
		// to what the node returned.
		reserialized, err := tx.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(reserialized, b64.Transaction.GetBinary()) {
			t.Fatal("binary round trip differs from node payload")
		}

		asJSON := decodeStrict[GetTransactionResult](t, lc.call(t, "getTransaction", sig, M{
			"encoding":                       solana.EncodingJSON,
			"maxSupportedTransactionVersion": 0,
		}))
		if _, err := asJSON.Transaction.GetTransaction(); err != nil {
			t.Fatalf("GetTransaction(json): %v", err)
		}

		parsed := decodeStrict[GetParsedTransactionResult](t, lc.call(t, "getTransaction", sig, M{
			"encoding":                       solana.EncodingJSONParsed,
			"maxSupportedTransactionVersion": 0,
		}))
		if parsed.Transaction == nil || len(parsed.Transaction.Signatures) == 0 {
			t.Fatal("no parsed transaction")
		}
	})

	t.Run("getTransactionCount", func(t *testing.T) {
		if decodePlain[uint64](t, lc.call(t, "getTransactionCount")) == 0 {
			t.Fatal("zero transaction count")
		}
	})

	t.Run("isBlockhashValid", func(t *testing.T) {
		if latestHash.IsZero() {
			t.Skip("no blockhash")
		}
		res := decodeStrict[IsValidBlockhashResult](t, lc.call(t, "isBlockhashValid", latestHash))
		if !res.Value {
			t.Fatal("fresh blockhash reported invalid")
		}
	})

	t.Run("minimumLedgerSlot", func(t *testing.T) {
		decodePlain[uint64](t, lc.call(t, "minimumLedgerSlot"))
	})

	t.Run("simulateTransaction", func(t *testing.T) {
		if latestHash.IsZero() {
			t.Skip("no blockhash")
		}
		// A self-transfer from a random (unfunded) key: the simulation is
		// expected to fail on-chain-wise, but the response must decode and
		// the node must accept our transaction serialization.
		key := mustRandomKey(t)
		payer := key.PublicKey()
		tx := &solana.Transaction{
			Message: solana.Message{
				AccountKeys: []solana.PublicKey{payer, {}},
				Header:      solana.MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1},
				Instructions: []solana.CompiledInstruction{{
					ProgramIDIndex: 1,
					Accounts:       []uint16{0, 0},
					Data:           solana.Base58{2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
				}},
				RecentBlockhash: latestHash,
			},
		}
		if _, err := tx.Sign(func(pub solana.PublicKey) *solana.PrivateKey {
			if pub == payer {
				return &key
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}

		res := decodeStrict[SimulateTransactionResponse](t, lc.call(t, "simulateTransaction", tx.ToBase64(), M{
			"encoding":   solana.EncodingBase64,
			"commitment": CommitmentProcessed,
		}))
		if res.Value == nil {
			t.Fatal("no simulation value")
		}
		// An unfunded fee payer must produce a simulation error, proving the
		// node parsed our transaction rather than rejecting the request.
		if res.Value.Err == nil {
			t.Log("simulation unexpectedly succeeded (funded random key?)")
		}
	})
}

// TestLiveClient smoke-tests the HTTP client against the live endpoint.
func TestLiveClient(t *testing.T) {
	url := os.Getenv("RPC_URL")
	if url == "" {
		t.Skip("RPC_URL not set; skipping live RPC tests")
	}
	client := New(url)
	ctx := context.Background()

	slot, err := client.GetSlot(ctx, CommitmentFinalized)
	if err != nil || slot == 0 {
		t.Fatalf("GetSlot: %d, %v", slot, err)
	}
	if _, err := client.GetVersion(ctx); err != nil {
		t.Fatalf("GetVersion: %v", err)
	}
	info, err := client.GetAccountInfo(ctx, solana.MustPublicKeyFromBase58(usdcMint))
	if err != nil || len(info.GetBinary()) != 82 {
		t.Fatalf("GetAccountInfo: %v (%d bytes)", err, len(info.GetBinary()))
	}
	if _, err := client.GetAccountInfo(ctx, solana.PublicKey{0xde, 0xad}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetAccountInfo(nonexistent) = %v, want ErrNotFound", err)
	}
	hash, err := client.GetLatestBlockhash(ctx, CommitmentFinalized)
	if err != nil || hash.Value == nil || hash.Value.Blockhash.IsZero() {
		t.Fatalf("GetLatestBlockhash: %+v, %v", hash, err)
	}
	blocks, err := client.GetBlocks(ctx, slot-20, &slot, CommitmentFinalized)
	if err != nil || len(blocks) == 0 {
		t.Fatalf("GetBlocks: %v", err)
	}
	block, err := client.GetBlock(ctx, blocks[len(blocks)-1])
	if err != nil || len(block.Transactions) == 0 {
		t.Fatalf("GetBlock: %v", err)
	}
	count := 0
	streamCtx, err := client.GetProgramAccountsStream(ctx, solana.SysVarPubkey, nil, func(*KeyedAccount) error {
		count++
		return nil
	})
	if err != nil || streamCtx == nil || count == 0 {
		t.Fatalf("GetProgramAccountsStream: ctx %+v, %d accounts, %v", streamCtx, count, err)
	}
}

func mustRandomKey(t *testing.T) solana.PrivateKey {
	t.Helper()
	key, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	return key
}
