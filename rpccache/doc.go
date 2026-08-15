// Package rpccache serves Solana account lookups from an in-memory cache,
// falling back to an rpc.Client for misses — and, crucially, lets realtime
// account updates (e.g. a Yellowstone geyser stream) keep cached entries
// current, so reads for locally-tracked accounts never leave the process.
//
// # Model
//
// [New] wraps an [rpc.Client]. [Client.GetAccountInfo] and
// [Client.GetMultipleAccounts] consult the cache first;
// getMultipleAccounts fetches only the keys it misses (deduplicated, in
// one call) and fills every requested position. An entry is served when
// any of the following holds:
//
//   - it is streamed: a realtime feed keeps it current ([Client.StoreStreamed],
//     wired up by yellowstone.PipeAccounts),
//   - it is immutable: the caller declared it never changes
//     ([Client.GetAccountInfoImmutable], [Client.StoreImmutable]),
//   - it was fetched within the freshness window (Options.FreshFor,
//     default 800ms).
//
// Writes are slot-ordered: a late RPC response can never overwrite a newer
// streamed update, and the streamed/immutable flags are sticky — a plain
// refetch does not demote an entry ([Client.Clear] does).
//
// A background janitor evicts entries not read within Options.TTL;
// [Client.Close] stops it.
//
// # Sharing
//
// Cache hits return the same *rpc.Account to every caller. Treat returned
// accounts as read-only.
//
//	client := rpccache.New(rpc.New(endpoint), nil)
//	defer client.Close()
//
//	// Reads hit the network at most once per freshness window...
//	account, err := client.GetAccountInfo(ctx, key)
//
//	// ...and never, once a stream keeps the account current:
//	go yellowstone.PipeAccounts(stream, client)
package rpccache
