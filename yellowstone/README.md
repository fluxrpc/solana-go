# Yellowstone gRPC Geyser client for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/fluxrpc/solana-go/yellowstone.svg)](https://pkg.go.dev/github.com/fluxrpc/solana-go/yellowstone)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE)

`github.com/fluxrpc/solana-go/yellowstone` is a complete Go client for Solana
Yellowstone gRPC, the Geyser streaming protocol also known as Dragon's Mouth.
It provides a compact API for authenticated connections, subscriptions,
runtime filter updates, unary Geyser RPCs, and conversion from Yellowstone
protobuf messages into the native types in
[`github.com/fluxrpc/solana-go`](https://github.com/fluxrpc/solana-go).

The Yellowstone client is a separate Go module. Its gRPC and protobuf
dependencies are only installed when an application uses it.

## Install

```bash
go get github.com/fluxrpc/solana-go/yellowstone
```

## Quick start

```go
package main

import (
	"context"
	"log"

	"github.com/fluxrpc/solana-go/yellowstone"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

func main() {
	ctx := context.Background()
	client, err := yellowstone.Connect(ctx, "https://your-geyser:443",
		yellowstone.WithToken("your-x-token"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	req := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED).
		AccountsByOwner("token-accounts", "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA").
		AllSlots("slots")

	stream, err := client.Subscribe(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	for {
		update, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		if update.GetAccount() != nil {
			converted := update.Account()
			log.Printf("account=%s slot=%d lamports=%d",
				converted.Pubkey, converted.Slot, converted.Lamports)
		}
	}
}
```

`Connect` accepts `host:port`, `http://`, and `https://` endpoints. TLS is
enabled automatically for HTTPS and port 443. Tokens are sent as `x-token`
metadata on unary and streaming calls.

## Subscription filters

`NewRequest` initializes every Yellowstone filter map. Fluent methods keep
filter construction attached to the request they mutate:

```go
req := yellowstone.NewRequest(pb.CommitmentLevel_PROCESSED).
	AccountsByKey("wallets", walletA.String(), walletB.String()).
	AccountsByOwner("program", programID.String()).
	TransactionsByAccount("transactions", walletA.String()).
	TransactionStatusesByAccount("transaction-status", walletA.String()).
	AllSlots("slots").
	BlocksIncluding("blocks", programID.String()).
	AllBlocksMeta("block-meta").
	AllEntries("entries")
```

The helper API covers accounts, slots, transactions, blocks, block metadata,
and ledger entries. The returned protobuf request remains fully accessible for
advanced Yellowstone filters such as memcmp/data-size account filters,
transaction vote/failed/signature filters, and transaction-status streams.

Replace filters on an active bidirectional stream without reconnecting:

```go
next := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED).
	AccountsByOwner("new-program", newProgramID.String())
if err := stream.Update(next); err != nil {
	return err
}
```

## Convert Yellowstone updates

Converters bridge Geyser protobuf updates to the core SDK:

```go
switch {
case update.GetAccount() != nil:
	account := update.Account()
	// account.Pubkey, Owner, Lamports, Data, Slot, WriteVersion, ...

case update.GetTransaction() != nil:
	tx, meta, err := update.Transaction()
	if err != nil {
		return err
	}
	_ = tx
	_ = meta
}
```

Converted transactions use `solana.Transaction` and re-serialize
byte-identically to the on-chain wire transaction. `Update.Transaction` also
returns the full `rpc.TransactionMeta`, including balances, logs, inner
instructions, token balances, rewards, loaded addresses, return data, compute
units, and cost units.

For throughput, `Update.Account` aliases the protobuf account-data buffer. Copy
the data if it must be retained after receiving subsequent stream messages.
The cache pipe described below performs that copy automatically.

## Feed the JSON-RPC cache

The Yellowstone stream can keep the root SDK's built-in account and chain-head
cache current. Once the cache is enabled, tracked account reads and fresh slot,
block-height, and blockhash reads are served locally:

```go
rpcClient := rpc.New("https://your-json-rpc")
rpcClient.EnableCache()

req := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED).
	AccountsByOwner("watched", programID.String()).
	AllSlots("slots").
	AllBlocksMeta("blocks")

stream, err := geyserClient.Subscribe(ctx, req)
if err != nil {
	return err
}
go func() {
	if err := stream.Pipe(rpc.CommitmentConfirmed, rpcClient); err != nil {
		log.Printf("Yellowstone stream ended: %v", err)
	}
}()
```

Use `PipeAccounts` when only account updates should be mirrored.

## Unary Geyser calls

The client includes typed wrappers for the Yellowstone unary API:

```go
pong, err := client.Ping(ctx, 1)
version, err := client.GetVersion(ctx)
slot, err := client.GetSlot(ctx, pb.CommitmentLevel_FINALIZED)
blockhash, err := client.GetLatestBlockhash(ctx, pb.CommitmentLevel_CONFIRMED)
height, err := client.GetBlockHeight(ctx, pb.CommitmentLevel_CONFIRMED)
valid, err := client.IsBlockhashValid(ctx, hash, pb.CommitmentLevel_CONFIRMED)
```

## Connection options

- `WithToken(token)` sends authentication as `x-token` metadata.
- `WithInsecure()` forces plaintext transport, including for port 443.
- `WithMaxRecvSize(bytes)` changes the default 1 GiB receive limit. The large
  default prevents full block subscriptions from hitting gRPC's 4 MiB limit.
- `WithKeepalive(duration)` changes the default 30-second HTTP/2 keepalive.
- `WithGRPCOptions(...)` appends raw `grpc.DialOption` values for custom
  credentials, interceptors, proxies, or transport behavior.

`Client` is safe for concurrent use. One goroutine may call `Stream.Recv` while
another calls `Stream.Update` or `Stream.Close`, following the normal gRPC
stream contract.

## Why this client

- Full Yellowstone subscription coverage with thin helpers and direct
  protobuf escape hatches.
- Automatic TLS and `x-token` authentication.
- A 1 GiB default receive limit suitable for large Solana block updates.
- Live filter changes without reconnecting.
- Allocation-light account and transaction conversion.
- Direct integration with the `solana-go/rpc` realtime cache.
- A nested module that keeps gRPC out of non-Yellowstone applications.

Development is sponsored and maintained by [FluxRPC](https://fluxrpc.com),
Solana and Fogo RPC infrastructure.
