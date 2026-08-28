# Confidential transfer

This example builds and locally verifies a complete Token-2022 confidential
transfer. It derives the source encryption keys from a wallet signature,
constructs representative encrypted account state, and compiles the transfer
and its three proof-verification instructions into a signed SIMD-0385 v1
transaction.

```sh
go run .
```

The output confirms that four instructions were prepared, reports the signed
transaction's wire and base64 sizes, and shows that the new decryptable source
balance is `990000` after transferring `10000` from `1000000`.

A confidential transfer does not fit in the 1232-byte legacy/v0 transaction
limit because the proof instruction data itself is larger than that limit.
Address lookup tables only compress account addresses, so they cannot make this
payload fit. `NewTransactionV1` uses the SIMD-0385 4096-byte format and carries
the compute-unit limit in `TransactionConfig`; do not add a Compute Budget
program instruction.

The example is self-contained and does not submit a transaction. In an
application, replace the generated addresses and representative state with:

- the mint and token-account addresses used by the transaction;
- the decoded `ConfidentialTransferAccount` extensions fetched from RPC;
- the destination account's ElGamal public key;
- the mint's auditor ElGamal public key, when configured;
- the latest blockhash returned by `rpc.Client.GetLatestBlockhash`.

V1 transactions must be submitted as base64. Once the example's representative
state and blockhash have been replaced with live values, send the encoded value
directly:

```go
encoded := transaction.ToBase64()
signature, err := client.SendEncodedTransaction(ctx, encoded)
```

The proof instructions must remain directly after the confidential-transfer
instruction because their offsets are encoded into it.

This directory is a separate Go module so applications that do not use
Token-2022 confidential transfers do not inherit its cryptographic dependencies.
