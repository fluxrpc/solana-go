// Package rpc provides the Solana JSON-RPC HTTP client and every
// request/response type of the Solana RPC API.
//
// # Client
//
// [New] creates a [Client] with a transport tuned for high request
// concurrency against a single endpoint. Each call decodes the entire
// response envelope in a single sonic pass into pooled buffers:
//
//	client := rpc.New("https://your-endpoint")
//	account, err := client.GetAccountInfo(ctx, pubkey)
//	slot, err := client.GetSlot(ctx, rpc.CommitmentFinalized)
//
// Methods that look up a single item (GetAccountInfo, GetTransaction,
// GetBlock, GetParsedTransaction) return [ErrNotFound] when the RPC result
// is null. Every method has a WithOpts variant exposing the full option
// set; the plain variants use pragmatic defaults, notably
// maxSupportedTransactionVersion=0 so versioned transactions decode out of
// the box. Use [Client.SetHeader] for authenticated endpoints.
//
// # Account cache
//
// [Client.EnableCache] puts an in-memory sharded account cache in front of
// GetAccountInfo and GetMultipleAccounts (the WithOpts variants always
// bypass it). An entry is served when it is streamed (kept current by a
// realtime feed such as the yellowstone package's PipeAccounts),
// immutable ([Client.GetAccountInfoImmutable]), or fetched within the
// freshness window. Writes are slot-ordered, so a late RPC response can
// never overwrite a newer streamed update. Cache hits return the same
// *Account to every caller: treat cached accounts as read-only.
//
//	client.EnableCache(nil)
//	go yellowstone.PipeAccounts(stream, client) // reads for streamed accounts never hit the network
//
// # Streaming getProgramAccounts
//
// getProgramAccounts responses can be arbitrarily large.
// [Client.GetProgramAccountsStream] (or [StreamProgramAccounts] over any
// io.Reader) decodes accounts incrementally while the response body
// downloads: memory stays bounded by the largest single account and
// decoding overlaps the transfer, which is built for providers that
// deliver responses as a constant stream.
//
// # Using the types without the client
//
// All response types are plain JSON-tagged structs that work with any JSON
// library; decode with sonic for the fastest path. Custom codecs exist
// only where the wire format demands them (DataBytesOrJSON,
// TransactionVersion, the transaction envelopes, and public-key-keyed
// maps).
//
// # Conformance
//
// A live test suite exercises every method against a real endpoint and
// fails if a response carries a populated field its DTO drops:
//
//	RPC_URL=https://... go test ./rpc/ -run TestLiveRPC
package rpc
