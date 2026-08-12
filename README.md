# solana-go

Optimized Golang SDK for Solana.

A lean port of the core [solana-go](https://github.com/gagliardetto/solana-go) types, rebuilt for performance:

- Base58 encoding/decoding via [fluxrpc/base58](https://github.com/fluxrpc/base58) (SIMD-accelerated, fixed-size fast paths for 32/64-byte values).
- Zero-allocation-conscious JSON marshaling (direct quoted-buffer writes, no `json.Marshal` round trips for base58/base64 strings).
- Single-allocation binary (wire format) encoding with exact size precomputation.
- Only supported serializations: `String`, `Bytes` (binary wire format) and JSON. No BSON, no text marshalers, no kitchen sink.
- Dependency-lean: Go stdlib + `fluxrpc/base58`, nothing else.

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

## Install

```bash
go get github.com/fluxrpc/solana-go
```

## Usage

```go
key := solana_go.MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111")
fmt.Println(key.String())

tx, err := solana_go.TransactionFromBase64("AfjEs3XhTc3hrxEvlnMPkm/cocvAUbFNbCl00qKnrFue6J53AhEqIFmcJJlJW3EDP5RmcMz+cNTTcZHW/WJYwAcBAAEDO8hh4VddzfcO5jbCt95jryl6y8ff65UcgukHNLWH+UQGgxCGGpgyfQVQV02EQYqm4QwzUt2qf9f1gVLM7rI4hwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA6ANIF55zOZWROWRkeh+lExxZBnKFqbvIxZDLE7EijjoBAgIAAQwCAAAAOTAAAAAAAAA=")
fmt.Println(len(tx.Signatures), tx.Message.RecentBlockhash)
```

<!-- BENCHMARKS:BEGIN -->
## Benchmarks

Comparing `github.com/fluxrpc/solana-go` (with `github.com/fluxrpc/base58`) against upstream `github.com/gagliardetto/solana-go` (with its bundled base58) for identical operations.

Machine: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz. Go: go1.26.4 (linux/amd64).

### Base58Data

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| MarshalJSON | 538.5 | 8196 | 15.2x | 208 | 592 | 1 | 5 |
| UnmarshalJSON | 357.9 | 7602 | 21.2x | 240 | 354 | 2 | 4 |

### Hash

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| String | 64.39 | 113.2 | 1.8x | 48 | 48 | 1 | 1 |
| FromBase58 | 42.28 | 87.20 | 2.1x | 0 | 0 | 0 | 0 |

### PrivateKey

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| Sign | 22527 | 21112 | 0.9x | 0 | 64 | 0 | 1 |
| PublicKey | 2.65 | 9995 | 3766.0x | 0 | 0 | 0 | 0 |

### PublicKey

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| String | 72.22 | 120.3 | 1.7x | 48 | 48 | 1 | 1 |
| FromBase58 | 41.04 | 76.16 | 1.9x | 0 | 0 | 0 | 0 |
| MarshalJSON | 69.19 | 123.1 | 1.8x | 48 | 48 | 1 | 1 |
| UnmarshalJSON | 101.2 | 207.4 | 2.0x | 48 | 64 | 1 | 2 |

### Signature

| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |
|---|---:|---:|---:|---:|---:|---:|---:|
| String | 140.8 | 478.3 | 3.4x | 96 | 96 | 1 | 1 |
| FromBase58 | 71.19 | 361.8 | 5.1x | 0 | 0 | 0 | 0 |
| MarshalJSON | 135.6 | 482.5 | 3.6x | 96 | 96 | 1 | 1 |
| UnmarshalJSON | 167.0 | 556.1 | 3.3x | 96 | 112 | 1 | 2 |

### ns/op comparison

```text
Base58Data_MarshalJSON
  flux  ███                                      538.5 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 8196 ns/op

Base58Data_UnmarshalJSON
  flux  ██                                       357.9 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 7602 ns/op

Hash_String
  flux  ███████████████████████                  64.39 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 113.2 ns/op

Hash_FromBase58
  flux  ███████████████████                      42.28 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 87.20 ns/op

PrivateKey_Sign
  flux  ████████████████████████████████████████ 22527 ns/op
  gagl  █████████████████████████████████████    21112 ns/op  <-- faster

PrivateKey_PublicKey
  flux  █                                        2.65 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 9995 ns/op

PublicKey_String
  flux  ████████████████████████                 72.22 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 120.3 ns/op

PublicKey_FromBase58
  flux  ██████████████████████                   41.04 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 76.16 ns/op

PublicKey_MarshalJSON
  flux  ██████████████████████                   69.19 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 123.1 ns/op

PublicKey_UnmarshalJSON
  flux  ████████████████████                     101.2 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 207.4 ns/op

Signature_String
  flux  ████████████                             140.8 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 478.3 ns/op

Signature_FromBase58
  flux  ████████                                 71.19 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 361.8 ns/op

Signature_MarshalJSON
  flux  ███████████                              135.6 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 482.5 ns/op

Signature_UnmarshalJSON
  flux  ████████████                             167.0 ns/op  <-- faster
  gagl  ████████████████████████████████████████ 556.1 ns/op
```
<!-- BENCHMARKS:END -->

## License

MIT
