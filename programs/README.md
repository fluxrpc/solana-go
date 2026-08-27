# Program codecs

## Supported programs

| Package | Instruction variants |
| --- | ---: |
| `address-lookup-table` | 5 |
| `compute-budget` | 5 current + 1 historical compatibility form |
| `loader-v2` | 2 |
| `loader-v3` | 8 |
| `loader-v4` | 7 |
| `stake` | 18 |
| `system` | 14 |
| `vote` | 20 |
| `token` | 21 |
| `token-2022` | 26 base variants + extension codecs |
| `token-metadata` | 25 Metaplex discriminators |

This directory contains opt-in, handwritten codecs for Solana programs. Each
program package owns its instruction layouts, account ordering, and decoding
API. The packages are part of the root module and add no program-specific
third-party dependencies; Go only compiles packages an application imports.

Program packages are deterministic wire codecs. They do not perform RPC calls,
discover accounts, calculate prices, or prepare protocol-specific operations.
Those concerns belong in a higher layer which supplies resolved accounts and
arguments to these constructors.

## System Program

The System package implements all 14 variants in `solana-system-interface`
3.3.0.

Constructors return concrete instruction types which implement
`solana.Instruction` directly:

```go
transfer := system.NewTransferInstruction(
	1_000_000,
	fundingAccount,
	recipientAccount,
)

instructions := []solana.Instruction{transfer}
```

No builder or additional build step is required. Instruction data is encoded
when `Data` is requested:

```go
data, err := transfer.Data()
if err != nil {
	return err
}
```

Decoding returns a typed package-specific envelope. Select the discriminator
and access the corresponding concrete field directly; no type assertion is
needed:

```go
decoded, err := system.DecodeInstruction(accounts, data)
if err != nil {
	return err
}

switch decoded.Type {
case system.TransferInstruction:
	fmt.Println(decoded.Transfer.Lamports)
case system.CreateAccountInstruction:
	fmt.Println(decoded.CreateAccount.Space)
}
```

Exactly one instruction field is populated for a successful decode.

## Adding a program

New program packages should follow the System package's public shape while
keeping their wire rules local:

1. Define the program ID and concrete instruction structs. Store required
   parameters as values and ordered accounts as `solana.AccountMetaSlice`.
2. Provide `New<Name>Instruction` constructors which return concrete pointers,
   set canonical signer/writable roles, and reject only genuine semantic
   constraints such as a protocol-defined maximum seed length. Keep every
   `solana.NewAccountMeta` call on its own line so account order and roles are
   immediately readable.
3. Embed a package-local instruction type which provides the shared
   `ProgramID` and `Accounts` methods. Implement `Data` on each concrete
   instruction so it satisfies `solana.Instruction` without repeated methods.
4. Encode fields explicitly with `github.com/fluxrpc/solana-go/binary`. Allocate
   the result once at its known or calculated size and write fields in wire
   order.
5. Define a package-specific discriminator type and typed decoded envelope.
   `DecodeInstruction` should dispatch with a direct switch and populate one
   concrete field. Each `Data` method writes its own discriminator; do not hide
   encoding behind discriminator helpers.
6. Keep `DecodeInstruction` as a direct typed switch over a local
   `binary.Decoder`. Do not add decoder wrappers, registries, interfaces, or
   package-private decode functions. Reusable payload encoding and decoding
   belongs on the payload type.

Keep each instruction's type, constructors, account layout, and `Data` method
in its own file. Shared public types belong in `types.go`, wire constants in
`constants.go`, errors in `errors.go`, the program ID in `program.go`, and the
typed dispatch in `decode.go`.

Do not introduce a shared discriminator format. A package may use a one-byte
tag, a little-endian integer, an eight-byte prefix, raw bytes, or another custom
layout. Its `Data` and `DecodeInstruction` implementations must express that
format explicitly. Likewise, select the actual string and collection encoding:
Borsh and bincode layouts are not interchangeable.

The root `binary` package is the shared source of explicit Solana wire
primitives. Add a primitive there only when it represents a reusable encoding,
with focused encoder and decoder tests. Program-specific layouts remain in the
program package.

Program packages must not use reflection, runtime codec registries, package
`init` registration, RPC clients, or add program-specific runtime third-party
dependencies. Importing a package is sufficient to use its constructors and
decoder.

## Required verification

Every instruction must include:

- Exact encoded-byte fixtures produced independently by a canonical program
  implementation or client.
- Constructor tests for account order and signer/writable roles.
- Typed decode and encode-decode-encode round-trip tests.
- Tests for every truncated input boundary, unknown discriminators, the
  program's trailing-data behavior, malformed lengths, and instruction-specific
  limits.
- Fuzz coverage for the package dispatcher and variable-length layouts.
- Benchmarks of the normal constructor, `Data`, and typed decode APIs against a
  pinned reference revision, including time, bytes, and allocations.

Parity and malformed-input tests are correctness gates. Allocation regressions
should be enforced independently of timing, while timing comparisons should use
multiple samples and statistical analysis rather than a single benchmark run.
