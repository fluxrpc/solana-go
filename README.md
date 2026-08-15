# solana-go

Optimized Golang SDK for Solana.

A lean port of the core [solana-go](https://github.com/gagliardetto/solana-go) types, rebuilt for performance:

- Base58 encoding/decoding via [fluxrpc/base58](https://github.com/fluxrpc/base58) (SIMD-accelerated, fixed-size fast paths for 32/64-byte values).
- Zero-allocation-conscious JSON marshaling (direct quoted-buffer writes, no `json.Marshal` round trips for base58/base64 strings).
- Single-allocation binary (wire format) encoding with exact size precomputation.
- Only supported serializations: `String`, `Bytes` (binary wire format) and JSON. No BSON, no text marshalers, no kitchen sink.
- Five dependencies, each earning its keep: `fluxrpc/base58` (base58), `bytedance/sonic` (JSON decoding), `oasisprotocol/curve25519-voi` (ed25519 sign/verify), `klauspost/compress` (base64+zstd account data), `gobwas/ws` (raw WebSocket frames). The gRPC stack lives only in the nested `yellowstone` module.

## Documentation

Full API documentation is godoc-first: every exported symbol is documented, each package carries an architectural overview in its `doc.go`, and runnable examples appear under Examples on [pkg.go.dev](https://pkg.go.dev/github.com/fluxrpc/solana-go). Browse locally with:

```bash
go doc github.com/fluxrpc/solana-go            # or any symbol, e.g. go doc rpc.Client
go run golang.org/x/pkgsite/cmd/pkgsite@latest -open .
```

| Package | Purpose |
|---|---|
| [`solana-go`](https://pkg.go.dev/github.com/fluxrpc/solana-go) | Core chain types, codecs, signing, PDA derivation |
| [`solana-go/rpc`](https://pkg.go.dev/github.com/fluxrpc/solana-go/rpc) | JSON-RPC HTTP client + every RPC request/response type, streaming gPA |
| [`solana-go/ws`](https://pkg.go.dev/github.com/fluxrpc/solana-go/ws) | WebSocket pubsub subscriptions |
| [`solana-go/yellowstone`](https://pkg.go.dev/github.com/fluxrpc/solana-go/yellowstone) | gRPC Geyser client (separate nested module) |

## Types

| Type | File | Notes |
|---|---|---|
| `PublicKey` | `public_key.go` | 32-byte account key |
| `Signature` | `signature.go` | 64-byte ed25519 signature, `Verify` |
| `Hash` | `hash.go` | 32-byte blockhash |
| `Base58` / `Base64` | `data.go` | byte slices with base58/base64 JSON encoding |
| `AccountMeta` | `account_meta.go` | account role in an instruction |
| `Instruction` / `CompiledInstruction` | `instruction.go` | instruction interface + compiled form |
| `PrivateKey` | `private_key.go` | 64-byte ed25519 keypair, `Sign` |
| `Message` | `message.go` | legacy + v0 messages, binary & JSON |
| `Transaction` | `transaction.go` | signatures + message, binary & JSON |
| `EncodingType` / `Data` | `encoding.go` / `data.go` | RPC data encodings and the `["<content>","<encoding>"]` tuple |

## RPC types

The [`rpc`](rpc/) package ports every request/response type of the upstream `rpc` package — one file per method (`rpc/get_block.go`, `rpc/get_transaction.go`, …) plus the shared response types (`TransactionMeta`, `Account`, `DataBytesOrJSON`, the `Parsed*` family). Client call machinery, BSON and binary-codec baggage are cut; the types are plain JSON-tagged structs with custom codecs only where the wire format demands it (`DataBytesOrJSON`, `TransactionVersion`, envelopes, pubkey-keyed maps).

They work with any JSON library; decode responses with `sonic.Unmarshal` to get the numbers below (upstream's client decodes with `goccy/go-json` — that's what it is benchmarked with).

For `getProgramAccounts`, `rpc.StreamProgramAccounts(body, fn)` decodes accounts incrementally off the response body as it downloads — memory stays bounded by the largest account instead of the whole response, and decoding overlaps the transfer (built for constant-stream delivery such as fluxrpc's). With a paced-delivery benchmark (2000 accounts arriving as 32KB chunks every 250µs, `BenchmarkStreamProgramAccountsNetwork` vs `BenchmarkBufferedProgramAccountsNetwork`):

| | streamed | buffered (read all, then decode) | |
|---|---:|---:|---|
| total wall clock | 4.0 ms | 6.2 ms | 1.5x faster |
| time to first account | 0.09 ms | 6.2 ms | ~68x faster |
| memory per response | 686 KB | 2.1 MB | 3.0x less |
| allocations | 4.3 k | 14.1 k | 3.3x fewer |

Live-endpoint conformance tests for every method run with `RPC_URL=... go test ./rpc/ -run TestLiveRPC`.

## WebSocket subscriptions

The [`ws`](ws/) package covers all nine pubsub subscriptions (`accountSubscribe`, `programSubscribe`, `logsSubscribe`, `signatureSubscribe`, `slotSubscribe`, `slotsUpdatesSubscribe`, `rootSubscribe`, `voteSubscribe`, `blockSubscribe`) over [gobwas/ws](https://github.com/gobwas/ws) raw frames. One read loop reuses a single message buffer and routes exact-size payload copies to buffered per-subscription channels; typed decoding happens on the consumer's goroutine via generic `Subscription[T]`, so a slow consumer drops (and counts) its own notifications instead of stalling the socket. Subscription channels are registered inside the read loop's ack handling, so notifications arriving immediately behind the subscribe ack are never lost.

Notification pipeline throughput (20k account notifications flooded over a local socket, `BenchmarkWs_AccountNotifications`):

| per notification | fluxrpc `ws` | upstream `rpc/ws` (gorilla) | |
|---|---:|---:|---|
| latency | 4.97 µs | 7.23 µs | 1.45x faster |
| memory | 1.66 KB | 2.20 KB | 1.3x less |
| allocations | 15.2 | 30.0 | 2.0x fewer |

Live test: `WS_URL=wss://... go test ./ws/ -run TestLive`.

## Yellowstone (gRPC Geyser)

The [`yellowstone`](yellowstone/) package is a separate nested Go module (`go get github.com/fluxrpc/solana-go/yellowstone`), so its gRPC/protobuf dependency tree never touches the core SDK. It wraps the Geyser protocol: `Connect` with `x-token` auth, TLS auto-detection, tuned keepalive and a 1GB receive limit; `Subscribe` with live filter updates and thin filter builders; unary `Ping`/`GetVersion`/`GetSlot`/`GetLatestBlockhash`/`GetBlockHeight`/`IsBlockhashValid`; and allocation-light converters from geyser protobuf into this SDK's types — converted transactions re-serialize byte-identical and pass signature verification.

| benchmark (bufconn, i7-9700K) | result |
|---|---:|
| subscribe throughput (incl. account conversion) | ~600k updates/sec |
| `ConvertTransaction` | 375 ns/op, 7 allocs |
| `ConvertAccount` | 55 ns/op, 1 alloc |

Live test: `YELLOWSTONE_ENDPOINT=... [YELLOWSTONE_TOKEN=...] go test ./yellowstone/ -run TestLive`.

## Install

```bash
go get github.com/fluxrpc/solana-go
```

## Usage

```go
key := solana_go.MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111")
fmt.Println(key.String())

// Zero-allocation PDA derivation:
ata, bump, _ := solana_go.FindAssociatedTokenAddress(wallet, mint)

// HTTP JSON-RPC client (single-pass sonic decode, pooled buffers):
client := rpc.New("https://your-endpoint")
account, _ := client.GetAccountInfo(ctx, ata)
slot, _ := client.GetSlot(ctx, rpc.CommitmentFinalized)

// getProgramAccounts, decoded incrementally while the response downloads:
client.GetProgramAccountsStream(ctx, program, nil, func(ka *rpc.KeyedAccount) error {
	process(ka)
	return nil
})
```

The tables and graph below are generated by the [`benchcmp`](benchcmp/) submodule (its own Go module, so upstream's dependency tree never touches this one):

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
| Sign | 12832 | 23765 | 1.9x | 64 | 64 | 1 | 1 |
| PublicKey | 2.80 | 11020 | 3932.9x | 0 | 0 | 0 | 0 |

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
| MarshalBinary | 278.8 | 361.8 | 1.3x | 640 | 648 | 2 | 3 |
| FromBytes | 358.7 | 592.2 | 1.7x | 528 | 632 | 6 | 9 |
| MarshalJSON | 5298 | 7895 | 1.5x | 2177 | 2044 | 3 | 17 |
| UnmarshalJSON | 8470 | 10241 | 1.2x | 2554 | 2758 | 20 | 59 |

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
  flux  ██████████████████████                   12832 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 23765 ns/op

PrivateKey_PublicKey
  flux  █                                        2.80 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 11020 ns/op

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
  flux  ███████████████████████████████          278.8 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 361.8 ns/op

Transaction_FromBytes
  flux  ████████████████████████                 358.7 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 592.2 ns/op

Transaction_MarshalJSON
  flux  ███████████████████████████              5298 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 7895 ns/op

Transaction_UnmarshalJSON
  flux  █████████████████████████████████        8470 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 10241 ns/op

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
```
<!-- BENCHMARKS:END -->

JSON strategy: encoding is hand-rolled straight into one buffer (measured ~2.5x faster than any reflection-based encoder walking our structs, including sonic); whole-struct decoding uses [bytedance/sonic](https://github.com/bytedance/sonic), which beats upstream's `goccy/go-json`. Callers get these paths regardless of which JSON package they invoke, since they are wired in via the `MarshalJSON`/`UnmarshalJSON` methods.

## License

MIT
