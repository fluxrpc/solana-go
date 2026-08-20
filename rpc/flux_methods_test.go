package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestClientFluxMethodsWireFormat(t *testing.T) {
	ctx := context.Background()
	key := solana.PublicKey{1}

	t.Run("getPriorityFeeEstimate", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"priorityFeeEstimate":42,"estimatedPriorityLamports":7,"priorityFeeLevels":{"min":1,"low":2,"medium":3,"high":4,"veryHigh":5,"unsafeMax":6},"jitoTipEstimate":{"low":8,"medium":9,"high":10}}`)
		result, err := client.GetPriorityFeeEstimate(ctx, GetPriorityFeeEstimateRequest{
			AccountKeys: []solana.PublicKey{key},
			Options: &GetPriorityFeeEstimateOpts{
				PriorityLevel:               PriorityLevelHigh,
				IncludeAllPriorityFeeLevels: true,
				IncludeJito:                 true,
				LookbackSlots:               100,
				IncludeVote:                 true,
				Recommended:                 true,
				EvaluateEmptySlotAsZero:     true,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.PriorityFeeEstimate != 42 || result.EstimatedPriorityLamports == nil || *result.EstimatedPriorityLamports != 7 || result.PriorityFeeLevels == nil || result.JitoTipEstimate == nil {
			t.Fatalf("result = %+v", result)
		}
		requireLastRequest(t, ts, "getPriorityFeeEstimate", `[{"accountKeys":["4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM"],"options":{"priorityLevel":"High","includeAllPriorityFeeLevels":true,"includeJito":true,"lookbackSlots":100,"includeVote":true,"recommended":true,"evaluateEmptySlotAsZero":true}}]`)
	})

	t.Run("getTransactionsForAddress", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"data":[],"paginationToken":"12:3"}`)
		limit := 25
		before := solana.Signature{2}
		result, err := client.GetTransactionsForAddress(ctx, key, &GetTransactionsForAddressOpts{
			TransactionDetails: TransactionDetailsFull,
			SortOrder:          TransactionsForAddressSortAsc,
			Limit:              &limit,
			Before:             &before,
			Encoding:           solana.EncodingBase64,
		})
		if err != nil {
			t.Fatal(err)
		}
		if result.PaginationToken == nil || *result.PaginationToken != "12:3" {
			t.Fatalf("result = %+v", result)
		}
		requireLastRequest(t, ts, "getTransactionsForAddress", `[
			"4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM",
			{"transactionDetails":"full","sortOrder":"asc","limit":25,"before":"3KWq19hjnoKF7rXwBCvs4oiyYSeDzukeyLLLcADXzrTWpH5PNYKB56KL2pRmJEsfHP6y5PcRYMqsWTcLHUDKBp3","encoding":"base64"}
		]`)
	})

	t.Run("getTokenAccounts", func(t *testing.T) {
		client, ts := newTestClient(t)
		account := solana.PublicKey{2}
		owner := solana.PublicKey{3}
		ts.serveRaw(t, `{"context":{"slot":99},"value":{"accounts":[{"pubKey":"`+account.String()+`","owner":"`+owner.String()+`","amount":123}],"cursor":"next"}}`)
		limit := 10
		result, err := client.GetTokenAccounts(ctx, key, &GetTokenAccountsIndexOpts{Limit: &limit, Cursor: "previous"})
		if err != nil {
			t.Fatal(err)
		}
		if result.Context.Slot != 99 || len(result.Value.Accounts) != 1 || result.Value.Accounts[0].Amount != 123 || result.Value.Cursor != "next" {
			t.Fatalf("result = %+v", result)
		}
		requireLastRequest(t, ts, "getTokenAccounts", `["4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM",{"limit":10,"cursor":"previous"}]`)
	})

	t.Run("getTokenAccountsCount", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.result = uint64(1234)
		result, err := client.GetTokenAccountsCount(ctx, key, &GetTokenAccountsCountOpts{ExcludeZero: true})
		if err != nil || result != 1234 {
			t.Fatalf("result = %d, err = %v", result, err)
		}
		requireLastRequest(t, ts, "getTokenAccountsCount", `["4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM",{"excludeZero":true}]`)
	})

	t.Run("getUpcomingLeaders", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"slot":100,"leaders":[{"slots":[101,102],"clusterInfo":{"pubkey":"4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM"}}]}`)
		result, err := client.GetUpcomingLeaders(ctx, 2)
		if err != nil {
			t.Fatal(err)
		}
		if result.Slot != 100 || len(result.Leaders) != 1 || len(result.Leaders[0].Slots) != 2 || result.Leaders[0].ClusterInfo == nil {
			t.Fatalf("result = %+v", result)
		}
		requireLastRequest(t, ts, "getUpcomingLeaders", `[2]`)
	})
}

func TestGetTransactionsForAddressLegacyArray(t *testing.T) {
	var result GetTransactionsForAddressResult
	if err := json.Unmarshal([]byte(`[{"slot":10,"blockTime":null,"transaction":null,"meta":null}]`), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Data) != 1 || result.Data[0].Slot != 10 || result.PaginationToken != nil {
		t.Fatalf("result = %+v", result)
	}
}

func TestClientProgramAccountFluxOptions(t *testing.T) {
	client, ts := newTestClient(t)
	ts.result = []any{}
	changedSince := uint64(100)
	changedUntil := uint64(200)
	cursor := solana.PublicKey{2}
	sorted := true
	opts := &GetProgramAccountsOpts{
		CursorFilter:     &ProgramAccountsCursorFilter{Cursor: &cursor, Limit: 50},
		ChangedSinceSlot: &changedSince,
		ChangedUntilSlot: &changedUntil,
		Async:            true,
		SortedResults:    &sorted,
	}
	if _, err := client.GetProgramAccountsWithOpts(context.Background(), solana.PublicKey{1}, opts); err != nil {
		t.Fatal(err)
	}
	requireLastRequest(t, ts, "getProgramAccounts", `[
		"4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM",
		{"cursorFilter":{"cursor":"8opHzTAnfzRpPEx21XtnrVTX28YQuCpAjcn1PczScKh","limit":50},"changedSinceSlot":100,"changedUntilSlot":200,"async":true,"sortedResults":true}
	]`)
}

func TestClientMultipleAccountsAsyncOption(t *testing.T) {
	client, ts := newTestClient(t)
	ts.serveRaw(t, `{"context":{"slot":1},"value":[]}`)
	if _, err := client.GetMultipleAccountsWithOpts(context.Background(), []solana.PublicKey{{1}}, &GetMultipleAccountsOpts{Async: true}); err != nil {
		t.Fatal(err)
	}
	requireLastRequest(t, ts, "getMultipleAccounts", `[["4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM"],{"async":true}]`)
}

func TestClientProgramAccountsContextDoesNotMutateOpts(t *testing.T) {
	client, ts := newTestClient(t)
	ts.serveRaw(t, `{"context":{"slot":5},"value":[]}`)
	withContext := false
	opts := &GetProgramAccountsOpts{WithContext: &withContext}
	result, err := client.GetProgramAccountsWithContext(context.Background(), solana.PublicKey{1}, opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Context.Slot != 5 || *opts.WithContext {
		t.Fatalf("result = %+v, opts = %+v", result, opts)
	}
	requireLastRequest(t, ts, "getProgramAccounts", `["4uQeVj5tqViQh7yWWGStvkEG1Zmhx6uasJtWCJziofM",{"withContext":true}]`)
}

func TestTokenAccountConfigValidation(t *testing.T) {
	client, _ := newTestClient(t)
	ctx := context.Background()
	if _, err := client.GetTokenAccountsByOwner(ctx, solana.PublicKey{1}, nil, nil); err == nil {
		t.Fatal("expected nil config error")
	}
	mint, program := solana.PublicKey{2}, solana.PublicKey{3}
	if _, err := client.GetTokenAccountsByOwner(ctx, solana.PublicKey{1}, &GetTokenAccountsConfig{Mint: &mint, ProgramId: &program}, nil); err == nil {
		t.Fatal("expected ambiguous config error")
	}
}

func TestNullableScalarAndMapResults(t *testing.T) {
	client, ts := newTestClient(t)
	ts.result = nil
	if _, err := client.GetBlockTime(context.Background(), 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetBlockTime err = %v", err)
	}
	if _, err := client.GetLeaderSchedule(context.Background()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetLeaderSchedule err = %v", err)
	}
}

func TestClientOptionHelpersDoNotMutateCallers(t *testing.T) {
	t.Run("getParsedBlock", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.result = nil
		opts := &GetBlockOpts{Encoding: solana.EncodingBase64}
		if _, err := client.GetParsedBlock(context.Background(), 1, opts); !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v", err)
		}
		if opts.Encoding != solana.EncodingBase64 {
			t.Fatalf("caller encoding mutated to %q", opts.Encoding)
		}
		requireLastRequest(t, ts, "getBlock", `[1,{"encoding":"jsonParsed"}]`)
	})

	t.Run("getProgramAccountsStream", func(t *testing.T) {
		client, ts := newTestClient(t)
		ts.serveRaw(t, `{"context":{"slot":1},"value":[]}`)
		withContext := false
		opts := &GetProgramAccountsOpts{WithContext: &withContext}
		if _, err := client.GetProgramAccountsStream(context.Background(), solana.PublicKey{1}, opts, func(*KeyedAccount) error { return nil }); err != nil {
			t.Fatal(err)
		}
		if *opts.WithContext {
			t.Fatal("caller withContext was mutated")
		}
	})
}

func TestSimulateTransactionIncludesInnerInstructionsOption(t *testing.T) {
	client, ts := newTestClient(t)
	ts.serveRaw(t, `{"context":{"slot":1},"value":{"accounts":null,"innerInstructions":[]}}`)
	tx := testUnsignedTransactionRPC(testPrivateKey(t))
	if _, err := client.SimulateTransactionWithOpts(context.Background(), tx, &SimulateTransactionOpts{InnerInstructions: true}); err != nil {
		t.Fatal(err)
	}
	params := ts.params(t)
	var opts map[string]any
	if err := json.Unmarshal(params[1], &opts); err != nil {
		t.Fatal(err)
	}
	if opts["innerInstructions"] != true {
		t.Fatalf("opts = %v", opts)
	}
}
