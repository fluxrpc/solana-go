# solana-go

An optimized Go SDK for Solana: core chain types, a JSON-RPC HTTP client covering the full RPC spec, WebSocket pubsub subscriptions, and a Yellowstone gRPC Geyser client — built for performance with a deliberately small dependency tree.

- Base58 encoding/decoding via [fluxrpc/base58](https://github.com/fluxrpc/base58) (SIMD-accelerated, fixed-size fast paths for 32/64-byte values).
- Zero-allocation-conscious JSON marshaling (direct quoted-buffer writes, no `json.Marshal` round trips for base58/base64 strings); decoding via [bytedance/sonic](https://github.com/bytedance/sonic).
- Single-allocation binary (wire format) encoding with exact size precomputation; zero-allocation PDA derivation.
- Only supported serializations: `String`, `Bytes` (binary wire format) and JSON. No BSON, no text marshalers, no kitchen sink.
- Five dependencies, each earning its keep: `fluxrpc/base58` (base58), `bytedance/sonic` (JSON decoding), `oasisprotocol/curve25519-voi` (ed25519 sign/verify), `klauspost/compress` (base64+zstd account data), `gobwas/ws` (raw WebSocket frames). The gRPC stack lives only in the nested `yellowstone` module.

Full Solana spec compliance is a hard requirement: every RPC method and response field, all nine pubsub subscriptions, legacy, v0 and v1 ([SIMD-0385](https://github.com/solana-foundation/solana-improvement-documents/blob/main/proposals/0385-transaction-v1.md)) transactions, base64+zstd account data.

## Install

```bash
go get github.com/fluxrpc/solana-go
go get github.com/fluxrpc/solana-go/yellowstone   # optional; nested module, brings in gRPC
```

The root package is named `solana_go`; import it under an alias:

```go
import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)
```

## Quickstart

### Keys

```go
// Generate, or load from the formats solana-keygen and wallets use.
key, _ := solana.NewRandomPrivateKey()
key, err := solana.PrivateKeyFromSolanaKeygenFile("id.json")
key, err = solana.PrivateKeyFromBase58("...")
key, err = solana.PrivateKeyFromSeed(seed) // 32-byte ed25519 seed

fmt.Println(key.PublicKey()) // reads the stored public half, ~3ns

// Zero-allocation PDA derivation.
pda, bump, err := solana.FindProgramAddress(seeds, programID)
ata, bump, err := solana.FindAssociatedTokenAddress(wallet, usdcMint)
```

### Build, sign and send a transaction

```go
client := rpc.New("https://api.mainnet-beta.solana.com")
recent, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)

// System-program transfer: u32 instruction index 2, u64 lamports, little-endian.
data := make([]byte, 12)
binary.LittleEndian.PutUint32(data, 2)
binary.LittleEndian.PutUint64(data[4:], 1_000_000)

ix := solana.NewInstruction(
	solana.SystemProgramID,
	solana.AccountMetaSlice{
		solana.Meta(payer.PublicKey()).SIGNER().WRITE(),
		solana.Meta(recipient).WRITE(),
	},
	data,
)

tx, err := solana.NewTransaction([]solana.Instruction{ix}, recent.Value.Blockhash,
	solana.TransactionPayer(payer.PublicKey()))

_, err = tx.Sign(func(pub solana.PublicKey) *solana.PrivateKey {
	if pub == payer.PublicKey() {
		return &payer
	}
	return nil
})

sig, err := client.SendTransaction(ctx, tx)
```

For the upcoming v1 transaction format ([SIMD-0385](https://github.com/solana-foundation/solana-improvement-documents/blob/main/proposals/0385-transaction-v1.md), shipping with Agave 4.2): 4096-byte transactions with compute-budget requests carried in the header instead of ComputeBudgetProgram instructions:

```go
cuLimit := uint32(200_000)
fee := uint64(5_000)
txV1, err := solana.NewTransactionV1([]solana.Instruction{ix}, recent.Value.Blockhash,
	solana.TransactionConfig{ComputeUnitLimit: &cuLimit, PriorityFeeLamports: &fee},
	solana.TransactionPayer(payer.PublicKey()))
_, err = txV1.Sign(signerFn)
raw, err := txV1.MarshalBinary() // up to 4096 bytes, fully sanitized
```

`NewTransaction` compiles instructions into a legacy message, or a v0 message when address lookup tables are supplied via `solana.TransactionAddressTables`. To send and wait for confirmation in one call:

```go
import "github.com/fluxrpc/solana-go/confirm"

// With a ws client the signature subscription is armed BEFORE the send,
// so the confirmation can never be missed; pass nil to poll over RPC.
sig, err := confirm.SendAndConfirm(ctx, client, wsClient, tx)

var execErr *confirm.ExecutionError
if errors.As(err, &execErr) {
	// Confirmed on chain, but an instruction failed; execErr.Err has the cause.
}
```

### Read accounts

```go
client := rpc.New("https://your-endpoint")
client.SetHeader("X-Api-Key", "...") // if the endpoint is authenticated

account, err := client.GetAccountInfo(ctx, ata) // rpc.ErrNotFound if it doesn't exist
fmt.Println(account.Value.Owner, len(account.GetBinary()))

accounts, err := client.GetMultipleAccounts(ctx, key1, key2, key3)
slot, err := client.GetSlot(ctx, rpc.CommitmentFinalized)
```

Every RPC method is available; single-item lookups return `rpc.ErrNotFound` on a null result, and each method has a `WithOpts` variant exposing the full option set (the plain variants default `maxSupportedTransactionVersion=0` so versioned transactions decode out of the box).

### Streaming getProgramAccounts

```go
// Accounts are decoded and delivered while the response body is still
// downloading; memory stays bounded by the largest single account.
_, err := client.GetProgramAccountsStream(ctx, program, nil, func(ka *rpc.KeyedAccount) error {
	process(ka)
	return nil
})
```

### WebSocket subscriptions

```go
client, err := ws.Connect(ctx, "wss://your-endpoint")
sub, err := client.AccountSubscribe(ctx, account, rpc.CommitmentConfirmed)
defer sub.Unsubscribe(context.Background())
for {
	update, err := sub.Recv(ctx)
	if err != nil {
		break // connection died or unsubscribed; client.Err() has the cause
	}
	process(update)
}
```

All nine pubsub subscriptions are covered: account, program, logs, signature, slot, slotsUpdates, root, vote, block — plus `ParsedBlockSubscribe`, the jsonParsed variant of block.

### Yellowstone (gRPC Geyser)

```go
client, err := yellowstone.Connect(ctx, "https://your-geyser:443",
	yellowstone.WithToken("..."))

req := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED)
yellowstone.AddAccounts(req, "usdc", yellowstone.AccountsByOwner(tokenProgram.String()))
stream, err := client.Subscribe(ctx, req)
for {
	update, err := stream.Recv()
	...
}
```

`ConvertTransaction` and `ConvertAccount` map geyser protobuf payloads into this SDK's types; converted transactions re-serialize byte-identical to the on-chain wire form.

### Account cache

The `rpc.Client` has a built-in account cache — `EnableCache()` and account reads are served from an in-memory sharded cache, with realtime Yellowstone updates piped straight into it so reads for locally-tracked accounts never leave the process:

```go
client := rpc.New(endpoint)
client.EnableCache() // defaults: 800ms freshness, processed commitment; EnableCacheWithOpts to tune

// Mirror accounts, slots and block metadata into the cache...
stream, _ := ys.Subscribe(ctx, req) // yellowstone subscription: account filters + Slots() + BlocksMeta()
go yellowstone.Pipe(stream, rpc.CommitmentConfirmed, client)

// ...and these are all served locally:
account, _ := client.GetAccountInfo(ctx, watchedKey)
slot, _ := client.GetSlot(ctx, rpc.CommitmentProcessed)
blockhash, _ := client.GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
```

Entries are served when streamed (the feed keeps them current), immutable (`GetAccountInfoImmutable`, for data that never changes), or fetched within the freshness window. Writes are slot-ordered — a late RPC response can never overwrite a newer streamed update — and `GetMultipleAccounts` fetches only its cache misses, deduplicated, in one call. Chain-head data (slot, block height, latest blockhash per commitment level) is cached too, serving `GetSlot`, `GetBlockHeight`, `GetLatestBlockhash` and `IsBlockhashValid` — with its own short freshness window (2s default) so a dead feed can never serve a stale slot or an expiring blockhash. The `WithOpts` variants always bypass the cache; `DisableCache` reverts to pure passthrough. A janitor evicts idle entries; `CacheStats()` reports hits/misses.

## Packages

| Package | Purpose |
|---|---|
| [`solana-go`](https://pkg.go.dev/github.com/fluxrpc/solana-go) | Core chain types, codecs, keys and signing, transaction building, PDA derivation |
| [`solana-go/rpc`](https://pkg.go.dev/github.com/fluxrpc/solana-go/rpc) | JSON-RPC HTTP client + every RPC request/response type, streaming gPA, account cache |
| [`solana-go/ws`](https://pkg.go.dev/github.com/fluxrpc/solana-go/ws) | WebSocket pubsub subscriptions |
| [`solana-go/confirm`](https://pkg.go.dev/github.com/fluxrpc/solana-go/confirm) | Send-and-confirm: race-free WebSocket confirmation, RPC polling fallback |
| [`solana-go/yellowstone`](https://pkg.go.dev/github.com/fluxrpc/solana-go/yellowstone) | gRPC Geyser client (separate nested module) |

## Feature overview

### Core types

| Type | File | Notes |
|---|---|---|
| `PublicKey` | `public_key.go` | 32-byte account key |
| `Signature` | `signature.go` | 64-byte ed25519 signature, `Verify` |
| `Hash` | `hash.go` | 32-byte blockhash |
| `Base58` / `Base64` | `data.go` | byte slices with base58/base64 JSON encoding |
| `AccountMeta` | `account_meta.go` | account role in an instruction |
| `Instruction` / `CompiledInstruction` | `instruction.go` | instruction interface + compiled form |
| `PrivateKey` | `private_key.go` | 64-byte ed25519 keypair, `Sign`, seed/`solana-keygen` file loading |
| `Message` | `message.go` | legacy + v0 messages, binary & JSON |
| `Transaction` | `transaction.go` | signatures + message, binary & JSON |
| `EncodingType` / `Data` | `encoding.go` / `data.go` | RPC data encodings and the `["<content>","<encoding>"]` tuple |
| `NewTransaction` / `TransactionBuilder` | `transaction_builder.go` | compiles instructions into legacy/v0 messages (fee payer, dedup, lookup tables) |
| `TransactionV1` / `NewTransactionV1` | `transaction_v1.go` | SIMD-0385 v1 transactions: 4096-byte limit, header config mask, full sanitization |
| `Wallet` | `wallet.go` | keypair wrapper: random, base58, keygen file, mnemonic |

### RPC

The [`rpc`](rpc/) package implements every method of the Solana JSON-RPC API — one file per method (`rpc/get_block.go`, `rpc/get_transaction.go`, …) plus the shared response types (`TransactionMeta`, `Account`, `DataBytesOrJSON`, the `Parsed*` family). The client's generic call path decodes each response envelope in a single sonic pass into pooled buffers over a transport tuned for high per-endpoint concurrency. The response types themselves are plain JSON-tagged structs with custom codecs only where the wire format demands it, so they also work standalone with any JSON library — decode with `sonic.Unmarshal` for the numbers in the benchmarks below.

For `getProgramAccounts`, `rpc.StreamProgramAccounts(body, fn)` decodes accounts incrementally off the response body as it downloads — memory stays bounded by the largest account instead of the whole response, and decoding overlaps the transfer (built for constant-stream delivery such as fluxrpc's). With a paced-delivery benchmark (2000 accounts arriving as 32KB chunks every 250µs, `BenchmarkStreamProgramAccountsNetwork` vs `BenchmarkBufferedProgramAccountsNetwork`):

| | streamed | buffered (read all, then decode) | |
|---|---:|---:|---|
| total wall clock | 4.0 ms | 6.2 ms | 1.5x faster |
| time to first account | 0.09 ms | 6.2 ms | ~68x faster |
| memory per response | 686 KB | 2.1 MB | 3.0x less |
| allocations | 4.3 k | 14.1 k | 3.3x fewer |

### WebSocket subscriptions

The [`ws`](ws/) package covers all nine pubsub subscriptions (plus the jsonParsed `ParsedBlockSubscribe` variant of blockSubscribe) over [gobwas/ws](https://github.com/gobwas/ws) raw frames. One read loop reuses a single message buffer and routes exact-size payload copies to buffered per-subscription channels; typed decoding happens on the consumer's goroutine via generic `Subscription[T]`, so a slow consumer drops (and counts) its own notifications instead of stalling the socket. Subscription channels are registered inside the read loop's ack handling, so notifications arriving immediately behind the subscribe ack are never lost.

Notification pipeline throughput (20k account notifications flooded over a local socket, `BenchmarkWs_AccountNotifications`):

| per notification | fluxrpc `ws` | gagliardetto `rpc/ws` (gorilla) | |
|---|---:|---:|---|
| latency | 4.97 µs | 7.23 µs | 1.45x faster |
| memory | 1.66 KB | 2.20 KB | 1.3x less |
| allocations | 15.2 | 30.0 | 2.0x fewer |

Parsed-block notifications (2k jsonParsed block payloads, 3 parsed transactions each, `BenchmarkWs_ParsedBlockNotifications`): 21.8 µs vs 79.4 µs per notification — 3.6x faster, 2.2x fewer allocations. Under the same flood the upstream client aborts once its fixed 200-slot channel fills ("reached channel max capacity"); this client's drop-and-count backpressure keeps the socket alive.

### V1 transactions (SIMD-0385)

`TransactionV1` implements the [v1 transaction format](https://github.com/solana-foundation/solana-improvement-documents/blob/main/proposals/0385-transaction-v1.md) shipping with Agave 4.2: version byte 129, a 4096-byte size limit (up from 1232), fee/resource requests (priority fee, compute-unit limit, loaded-data limit, heap size) carried in a header config mask instead of ComputeBudgetProgram instructions, fixed-width counts instead of shortvec, signatures trailing the signed payload, and no address lookup tables (the full address list fits directly at 4096 bytes). `NewTransactionV1` reuses the transaction builder's account compilation; encode/decode enforce every sanitization rule in the SIMD (limits, duplicate addresses, index bounds, config-mask validity, heap bounds, trailing bytes).

There is no upstream implementation to compare against yet; self-measured (i7-9700K):

| operation | result |
|---|---:|
| `MarshalBinary` (incl. full sanitization) | 322 ns, 1 alloc |
| `TransactionV1FromBytes` (incl. full sanitization) | 459 ns, 8 allocs |
| `NewTransactionV1` (compile from instructions) | 1.5 µs, 10 allocs |

### Send and confirm

The [`confirm`](confirm/) package sends a transaction and waits for it to reach a commitment. With a WebSocket client, the signature subscription is registered **before** the transaction is sent, so a fast confirmation can never race the subscription — the upstream flow subscribes after sending and, if it loses that race, waits for the full timeout. Signature subscriptions are single-shot, so after the terminal notification only local state is dropped (`ws.Subscription.Release`) instead of an unsubscribe round trip. Without a WebSocket client, confirmation is polled over `getSignatureStatuses` with commitment-aware status ranking.

On a loopback end-to-end benchmark (`BenchmarkConfirm_SendAndConfirm`: connect, subscribe, send, notify) the pipeline runs 664 µs vs upstream's 567 µs with 18% fewer allocations — the difference is exactly the subscribe-ack round trip that buys race-freedom, a cost that disappears against real confirmation latencies (hundreds of ms) while the race it removes costs upstream a full timeout when lost.

### Yellowstone (gRPC Geyser)

The [`yellowstone`](yellowstone/) package is a separate nested Go module (`go get github.com/fluxrpc/solana-go/yellowstone`), so its gRPC/protobuf dependency tree never touches the core SDK. It wraps the Geyser protocol: `Connect` with `x-token` auth, TLS auto-detection, tuned keepalive and a 1GB receive limit; `Subscribe` with live filter updates and thin filter builders; unary `Ping`/`GetVersion`/`GetSlot`/`GetLatestBlockhash`/`GetBlockHeight`/`IsBlockhashValid`; and allocation-light converters from geyser protobuf into this SDK's types — converted transactions re-serialize byte-identical and pass signature verification.

| benchmark (bufconn, i7-9700K) | result |
|---|---:|
| subscribe throughput (incl. account conversion) | ~600k updates/sec |
| `ConvertTransaction` | 375 ns/op, 7 allocs |
| `ConvertAccount` | 55 ns/op, 1 alloc |

### Account cache

| benchmark | result |
|---|---:|
| `GetAccountInfo` cache hit | 96 ns, 1 alloc (vs ~125 µs for a localhost RPC round trip) |
| `CacheStoreStreamed` ingest | 76 ns, 0 allocs (~13M updates/sec) |
| `GetMultipleAccounts`, 100 cached accounts | 3.6 µs, 2 allocs |

### Not included yet

Program instruction builders (System, Token, and friends) are intentionally not part of the SDK yet — they are planned around a registry-based design rather than one hand-written package per program, and contributions there are welcome. Until then, `solana.NewInstruction` with hand-encoded instruction data (as in the quickstart) works against any program.

The tables and graph below are generated by the [`benchcmp`](benchcmp/) submodule (its own Go module, so the comparison target's dependency tree never touches this one), comparing identical operations against the widely-used [gagliardetto/solana-go](https://github.com/gagliardetto/solana-go):

```bash
cd benchcmp
./run.sh '.*'      # run all comparisons, append to results.txt
./splice.sh '.*'   # regenerate the README section below
```

<!-- BENCHMARKS:BEGIN -->
## Benchmarks

Comparing `github.com/fluxrpc/solana-go` (with `github.com/fluxrpc/base58`) against upstream `github.com/gagliardetto/solana-go` (with its bundled base58) for identical operations.

Machine: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz. Go: go1.26.4 (linux/amd64).

### Base58Data

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| MarshalJSON | 561.6 | 8003 | 14.3x | 208 | 592 | 1 | 5 |
| UnmarshalJSON | 296.6 | 7020 | 23.7x | 96 | 354 | 1 | 4 |

### Hash

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| String | 71.12 | 131.8 | 1.9x | 48 | 48 | 1 | 1 |
| FromBase58 | 43.62 | 84.83 | 1.9x | 0 | 0 | 0 | 0 |

### Message

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| MarshalBinary | 195.5 | 222.2 | 1.1x | 288 | 288 | 1 | 1 |
| UnmarshalBinary | 315.8 | 486.7 | 1.5x | 320 | 368 | 4 | 5 |
| MarshalJSON | 1061 | 3768 | 3.6x | 640 | 1177 | 1 | 15 |
| UnmarshalJSON | 3120 | 5917 | 1.9x | 1408 | 2309 | 13 | 50 |

### PrivateKey

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| Sign | 11410 | 21685 | 1.9x | 64 | 64 | 1 | 1 |
| PublicKey | 2.73 | 10856 | 3969.3x | 0 | 0 | 0 | 0 |
| FromKeygenFile | 10298 | 11752 | 1.1x | 64 | 360 | 1 | 67 |
| FromMnemonic | 1022195 | 1016129 | 1.0x | 6160 | 6480 | 55 | 76 |

### PublicKey

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| String | 71.40 | 121.0 | 1.7x | 48 | 48 | 1 | 1 |
| FromBase58 | 42.30 | 74.44 | 1.8x | 0 | 0 | 0 | 0 |
| MarshalJSON | 80.61 | 115.7 | 1.4x | 48 | 48 | 1 | 1 |
| UnmarshalJSON | 70.39 | 206.0 | 2.9x | 0 | 64 | 0 | 2 |

### Signature

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| String | 138.3 | 467.4 | 3.4x | 96 | 96 | 1 | 1 |
| FromBase58 | 68.69 | 333.7 | 4.9x | 0 | 0 | 0 | 0 |
| MarshalJSON | 143.2 | 483.0 | 3.4x | 96 | 96 | 1 | 1 |
| UnmarshalJSON | 119.4 | 518.4 | 4.3x | 0 | 112 | 0 | 2 |
| Verify | 31705 | 32995 | 1.0x | 0 | 0 | 0 | 0 |

### Transaction

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| MarshalBinary | 251.0 | 306.5 | 1.2x | 640 | 648 | 2 | 3 |
| FromBytes | 343.1 | 687.2 | 2.0x | 528 | 632 | 6 | 9 |
| MarshalJSON | 5213 | 7751 | 1.5x | 2178 | 2045 | 3 | 17 |
| UnmarshalJSON | 8579 | 9855 | 1.1x | 2557 | 2762 | 20 | 59 |
| Build | 1149 | 1647 | 1.4x | 864 | 1152 | 8 | 15 |
| BuildV0 | 2048 | 2316 | 1.1x | 1472 | 1200 | 14 | 22 |

### RpcTransactionMeta

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| UnmarshalJSON | 4322 | 7736 | 1.8x | 1635 | 2893 | 4 | 44 |
| MarshalJSON | 3432 | 6224 | 1.8x | 2118 | 2143 | 15 | 21 |

### RpcAccount

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| UnmarshalJSON | 1044 | 1399 | 1.3x | 277 | 528 | 4 | 13 |
| MarshalJSON | 694.6 | 1577 | 2.3x | 307 | 464 | 5 | 11 |

### RpcGetBlockResult

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| UnmarshalJSON | 6303 | 10071 | 1.6x | 4801 | 5556 | 16 | 48 |
| MarshalJSON | 4342 | 11517 | 2.7x | 4114 | 6754 | 15 | 40 |

### RpcGetTransactionResult

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| UnmarshalJSON | 2253 | 4023 | 1.8x | 1365 | 2446 | 3 | 20 |

### RpcParsedTransaction

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| UnmarshalJSON | 4112 | 9247 | 2.2x | 1889 | 3328 | 15 | 49 |

### Pda

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| FindProgramAddress | 3361 | 3746 | 1.1x | 0 | 625 | 0 | 6 |
| CreateProgramAddress | 3396 | 3510 | 1.0x | 0 | 168 | 0 | 4 |
| FindAssociatedTokenAddress | 3649 | 5396 | 1.5x | 0 | 721 | 0 | 9 |
| IsOnCurve | 3899 | 3724 | 1.0x | 0 | 0 | 0 | 0 |

### Wallet

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| New | 11664 | 11402 | 1.0x | 152 | 152 | 4 | 4 |

### ns/op comparison

```text
Base58Data_MarshalJSON
  flux  ███                                      561.6 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 8003 ns/op

Base58Data_UnmarshalJSON
  flux  ██                                       296.6 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 7020 ns/op

Hash_String
  flux  ██████████████████████                   71.12 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 131.8 ns/op

Hash_FromBase58
  flux  █████████████████████                    43.62 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 84.83 ns/op

Message_MarshalBinary
  flux  ███████████████████████████████████      195.5 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 222.2 ns/op

Message_UnmarshalBinary
  flux  ██████████████████████████               315.8 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 486.7 ns/op

Message_MarshalJSON
  flux  ███████████                              1061 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 3768 ns/op

Message_UnmarshalJSON
  flux  █████████████████████                    3120 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 5917 ns/op

PrivateKey_Sign
  flux  █████████████████████                    11410 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 21685 ns/op

PrivateKey_PublicKey
  flux  █                                        2.73 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 10856 ns/op

PrivateKey_FromKeygenFile
  flux  ███████████████████████████████████      10298 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 11752 ns/op

PrivateKey_FromMnemonic
  flux  ████████████████████████████████████████ 1022195 ns/op
  gagl  ████████████████████████████████████████ 1016129 ns/op  <-- faster

PublicKey_String
  flux  ████████████████████████                 71.40 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 121.0 ns/op

PublicKey_FromBase58
  flux  ███████████████████████                  42.30 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 74.44 ns/op

PublicKey_MarshalJSON
  flux  ████████████████████████████             80.61 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 115.7 ns/op

PublicKey_UnmarshalJSON
  flux  ██████████████                           70.39 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 206.0 ns/op

Signature_String
  flux  ████████████                             138.3 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 467.4 ns/op

Signature_FromBase58
  flux  ████████                                 68.69 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 333.7 ns/op

Signature_MarshalJSON
  flux  ████████████                             143.2 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 483.0 ns/op

Signature_UnmarshalJSON
  flux  █████████                                119.4 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 518.4 ns/op

Signature_Verify
  flux  ██████████████████████████████████████   31705 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 32995 ns/op

Transaction_MarshalBinary
  flux  █████████████████████████████████        251.0 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 306.5 ns/op

Transaction_FromBytes
  flux  ████████████████████                     343.1 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 687.2 ns/op

Transaction_MarshalJSON
  flux  ███████████████████████████              5213 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 7751 ns/op

Transaction_UnmarshalJSON
  flux  ███████████████████████████████████      8579 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 9855 ns/op

Transaction_Build
  flux  ████████████████████████████             1149 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 1647 ns/op

Transaction_BuildV0
  flux  ███████████████████████████████████      2048 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 2316 ns/op

RpcTransactionMeta_UnmarshalJSON
  flux  ██████████████████████                   4322 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 7736 ns/op

RpcTransactionMeta_MarshalJSON
  flux  ██████████████████████                   3432 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 6224 ns/op

RpcAccount_UnmarshalJSON
  flux  ██████████████████████████████           1044 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 1399 ns/op

RpcAccount_MarshalJSON
  flux  ██████████████████                       694.6 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 1577 ns/op

RpcGetBlockResult_UnmarshalJSON
  flux  █████████████████████████                6303 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 10071 ns/op

RpcGetBlockResult_MarshalJSON
  flux  ███████████████                          4342 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 11517 ns/op

RpcGetTransactionResult_UnmarshalJSON
  flux  ██████████████████████                   2253 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 4023 ns/op

RpcParsedTransaction_UnmarshalJSON
  flux  ██████████████████                       4112 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 9247 ns/op

Pda_FindProgramAddress
  flux  ████████████████████████████████████     3361 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 3746 ns/op

Pda_CreateProgramAddress
  flux  ███████████████████████████████████████  3396 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 3510 ns/op

Pda_FindAssociatedTokenAddress
  flux  ███████████████████████████              3649 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 5396 ns/op

Pda_IsOnCurve
  flux  ████████████████████████████████████████ 3899 ns/op
  gagl  ██████████████████████████████████████   3724 ns/op  <-- faster

Wallet_New
  flux  ████████████████████████████████████████ 11664 ns/op
  gagl  ███████████████████████████████████████  11402 ns/op  <-- faster
```
<!-- BENCHMARKS:END -->

JSON strategy: encoding is hand-rolled straight into one buffer (measured ~2.5x faster than any reflection-based encoder walking our structs, including sonic); whole-struct decoding uses [bytedance/sonic](https://github.com/bytedance/sonic), which beats the `goccy/go-json` decoder used by gagliardetto's client — that pairing is what the tables above measure. Callers get these paths regardless of which JSON package they invoke, since they are wired in via the `MarshalJSON`/`UnmarshalJSON` methods.

## Testing

Unit tests run without any network access:

```bash
go test ./...
cd yellowstone && go test ./...
```

Live conformance suites exercise every method against a real endpoint; the RPC suite additionally re-marshals each response type and diffs it against the raw JSON, failing if the server sent a populated field the DTO would drop:

```bash
RPC_URL=https://... go test ./rpc/ -run TestLiveRPC
WS_URL=wss://...   go test ./ws/  -run TestLive
YELLOWSTONE_ENDPOINT=... YELLOWSTONE_TOKEN=... go test ./yellowstone/ -run TestLive
```

## Documentation

Full API documentation is godoc-first: every exported symbol is documented, each package carries an architectural overview in its `doc.go`, and runnable examples appear under Examples on [pkg.go.dev](https://pkg.go.dev/github.com/fluxrpc/solana-go). Browse locally with:

```bash
go doc github.com/fluxrpc/solana-go            # or any symbol, e.g. go doc rpc.Client
go run golang.org/x/pkgsite/cmd/pkgsite@latest -open .
```

## License

MIT
