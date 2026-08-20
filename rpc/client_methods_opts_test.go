package rpc

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func requireLastRequest(t *testing.T, ts *testServer, method, wantParams string) {
	t.Helper()
	if ts.lastReq.Method != method {
		t.Fatalf("method = %q, want %q", ts.lastReq.Method, method)
	}
	var got, want any
	if err := json.Unmarshal(ts.lastReq.Params, &got); err != nil {
		t.Fatalf("decode actual params: %v", err)
	}
	if err := json.Unmarshal([]byte(wantParams), &want); err != nil {
		t.Fatalf("decode expected params: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("params = %s, want %s", ts.lastReq.Params, wantParams)
	}
}

func TestClientContextOptionsRequests(t *testing.T) {
	ctx := context.Background()
	minSlot := uint64(123)

	t.Run("getBalance", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"context":{"slot":123},"value":42}`)
		_, err := client.GetBalanceWithOpts(ctx, solana.PublicKey{1}, &GetBalanceOpts{
			Commitment: CommitmentConfirmed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getBalance", `[
			"4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM",
			{"commitment":"confirmed","minContextSlot":123}
		]`)
	})

	t.Run("getBlockHeight", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.result = uint64(456)
		_, err := client.GetBlockHeightWithOpts(ctx, &GetBlockHeightOpts{
			Commitment: CommitmentFinalized, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getBlockHeight", `[{"commitment":"finalized","minContextSlot":123}]`)
	})

	t.Run("getBlocks", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.result = []uint64{}
		end := uint64(20)
		_, err := client.GetBlocksWithOpts(ctx, 10, &end, &GetBlocksOpts{
			Commitment: CommitmentFinalized, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getBlocks", `[10,20,{"commitment":"finalized","minContextSlot":123}]`)
	})

	t.Run("getBlocksWithoutEnd", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.result = []uint64{}
		_, err := client.GetBlocksWithOpts(ctx, 10, nil, &GetBlocksOpts{MinContextSlot: &minSlot})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getBlocks", `[10,{"minContextSlot":123}]`)
	})

	t.Run("getBlocksWithLimit", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.result = []uint64{}
		_, err := client.GetBlocksWithLimitWithOpts(ctx, 10, 5, &GetBlocksOpts{
			Commitment: CommitmentConfirmed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getBlocksWithLimit", `[10,5,{"commitment":"confirmed","minContextSlot":123}]`)
	})

	t.Run("getTransactionCount", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.result = uint64(99)
		_, err := client.GetTransactionCountWithOpts(ctx, &GetTransactionCountOpts{
			Commitment: CommitmentProcessed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getTransactionCount", `[{"commitment":"processed","minContextSlot":123}]`)
	})

	t.Run("getEpochInfo", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"absoluteSlot":1,"blockHeight":1,"epoch":1,"slotIndex":1,"slotsInEpoch":1}`)
		_, err := client.GetEpochInfoWithOpts(ctx, &GetEpochInfoOpts{
			Commitment: CommitmentConfirmed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getEpochInfo", `[{"commitment":"confirmed","minContextSlot":123}]`)
	})

	t.Run("getSlotLeader", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.result = solana.PublicKey{2}.String()
		_, err := client.GetSlotLeaderWithOpts(ctx, &GetSlotLeaderOpts{
			Commitment: CommitmentConfirmed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getSlotLeader", `[{"commitment":"confirmed","minContextSlot":123}]`)
	})
}

func TestClientExistingOptionsArePlumbed(t *testing.T) {
	ctx := context.Background()
	minSlot := uint64(456)

	t.Run("getLatestBlockhash", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, getLatestBlockhashFixture)
		_, err := client.GetLatestBlockhashWithOpts(ctx, &GetLatestBlockhashOpts{
			Commitment: CommitmentProcessed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getLatestBlockhash", `[{"commitment":"processed","minContextSlot":456}]`)
	})

	t.Run("isBlockhashValid", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"context":{"slot":1},"value":true}`)
		hash := solana.Hash{3}
		_, err := client.IsBlockhashValidWithOpts(ctx, hash, &IsBlockhashValidOpts{
			Commitment: CommitmentConfirmed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "isBlockhashValid", `[
			"CiDwVBFgWV9E5MvXWoLgnEgn2hK7rJikbvfWavzAQz3",
			{"commitment":"confirmed","minContextSlot":456}
		]`)
	})

	t.Run("getFeeForMessage", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, getFeeForMessageFixture)
		_, err := client.GetFeeForMessageWithOpts(ctx, "message", &GetFeeForMessageOpts{
			Commitment: CommitmentProcessed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getFeeForMessage", `["message",{"commitment":"processed","minContextSlot":456}]`)
	})

	t.Run("getSlotBypassesCache", func(t *testing.T) {
		client, ts := newTestClient(t)
		client.EnableCache()
		client.CacheStoreSlot(CommitmentProcessed, 100)
		ts.result = uint64(200)
		got, err := client.GetSlotWithOpts(ctx, &GetSlotOpts{
			Commitment: CommitmentProcessed, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		if got != 200 {
			t.Fatalf("slot = %d, want RPC result 200", got)
		}
		requireLastRequest(t, ts, "getSlot", `[{"commitment":"processed","minContextSlot":456}]`)
	})

	t.Run("getStakeMinimumDelegation", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"context":{"slot":1},"value":1000000000}`)
		_, err := client.GetStakeMinimumDelegationWithOpts(ctx, &GetStakeMinimumDelegationOpts{
			Commitment: CommitmentFinalized, MinContextSlot: &minSlot,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getStakeMinimumDelegation", `[{"commitment":"finalized","minContextSlot":456}]`)
	})

	t.Run("requestAirdrop", func(t *testing.T) {
		client, ts := newTestClient(t)
		sig := solana.Signature{4}
		ts.result = sig.String()
		blockhash := solana.Hash{5}
		_, err := client.RequestAirdropWithOpts(ctx, solana.PublicKey{6}, 10, &RequestAirdropOpts{
			Commitment: CommitmentConfirmed, RecentBlockhash: &blockhash,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "requestAirdrop", `[
			"QRSsyMWN1yHT9ir42bgNZUNZ4PdEhcSWCrL2AryKpy5",10,
			{"commitment":"confirmed","recentBlockhash":"LX3EUdRUBUa3TbsYXLEUdj9J3prXkWXvLYSWyYyc2Jj"}
		]`)
	})

	t.Run("getLargestAccounts", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"context":{"slot":1},"value":[]}`)
		sortResults := true
		_, err := client.GetLargestAccountsWithOpts(ctx, &GetLargestAccountsOpts{
			Commitment:  CommitmentFinalized,
			Filter:      LargestAccountsFilterCirculating,
			SortResults: &sortResults,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getLargestAccounts", `[{"commitment":"finalized","filter":"circulating","sortResults":true}]`)
	})

	t.Run("getTokenAccountBalance", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"context":{"slot":1},"value":{"amount":"1","decimals":0,"uiAmount":1,"uiAmountString":"1"}}`)
		_, err := client.GetTokenAccountBalanceWithOpts(ctx, solana.PublicKey{7}, &GetTokenAccountBalanceOpts{
			Commitment: CommitmentConfirmed,
		})
		if err != nil {
			t.Fatal(err)
		}
		requireLastRequest(t, ts, "getTokenAccountBalance", `[
			"UKrXU5bFrTzrqqpZXs8GVDbp4xPweiM65ADXNAy3ddR",
			{"commitment":"confirmed"}
		]`)
	})
}
