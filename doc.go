// Package solana_go is a performance-focused Solana SDK: the core chain
// types (keys, signatures, transactions, messages) with fast base58, JSON
// and binary wire-format codecs, ed25519 signing/verification, and
// program-derived address (PDA) utilities.
//
// The SDK is organized in four packages:
//
//   - This root package: chain types and cryptography. Every type supports
//     String, Bytes and JSON; Transaction and Message additionally
//     round-trip the Solana binary wire format.
//   - [github.com/fluxrpc/solana-go/rpc]: the JSON-RPC HTTP client and
//     every request/response type of the Solana RPC API, including
//     incremental streaming for getProgramAccounts.
//   - [github.com/fluxrpc/solana-go/ws]: WebSocket pubsub subscriptions.
//   - [github.com/fluxrpc/solana-go/yellowstone]: a gRPC Geyser
//     (Yellowstone) client, shipped as a separate nested module so its
//     gRPC dependency tree stays out of consumers of the core SDK.
//
// # Performance
//
// Base58 uses github.com/fluxrpc/base58 (SIMD-accelerated with fixed-size
// 32/64-byte fast paths). JSON encoding is hand-written into single exact-
// or upper-bounded buffers; JSON decoding routes through bytedance/sonic.
// Binary encoding computes its exact output size first and allocates once;
// PDA derivation and fixed-size JSON decodes are allocation-free. The
// benchcmp submodule tracks every claim against upstream
// (github.com/gagliardetto/solana-go); results live in the README.
//
// # Buffer aliasing
//
// Two deliberate zero-copy contracts exist, both documented at their
// functions: Message.UnmarshalBinary / TransactionFromBytes alias the input
// buffer for instruction data (do not mutate or reuse the input while the
// decoded value is alive), and the yellowstone account converter aliases
// the protobuf buffer. Everything else copies.
//
// # Quick start
//
//	key := solana_go.MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111")
//
//	// Decode a transaction from the wire.
//	decoded, err := solana_go.TransactionFromBase64("...")
//
//	// Derive an associated token account.
//	ata, bump, err := wallet.FindAssociatedTokenAddress(mint)
//
//	// Build, sign and serialize.
//	privateKey, err := solana_go.NewRandomPrivateKey()
//	signed := &solana_go.Transaction{Message: msg}
//	signed.Sign(func(pub solana_go.PublicKey) *solana_go.PrivateKey {
//		if pub == privateKey.PublicKey() {
//			return &privateKey
//		}
//		return nil
//	})
//	wire, err := signed.MarshalBinary()
package solana_go
