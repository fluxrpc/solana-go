# Confidential transfer

This example builds and locally verifies a complete Token-2022 confidential
transfer. It derives the source encryption keys from a wallet signature,
constructs representative encrypted account state, and returns the transfer
instruction followed by its three proof-verification instructions.

```sh
go run .
```

The output confirms that four instructions were prepared and that the new
decryptable source balance is `990000` after transferring `10000` from
`1000000`.

The example is self-contained and does not submit a transaction. In an
application, replace the generated addresses and representative state with:

- the mint and token-account addresses used by the transaction;
- the decoded `ConfidentialTransferAccount` extensions fetched from RPC;
- the destination account's ElGamal public key;
- the mint's auditor ElGamal public key, when configured.

Add `plan.Instructions` to one transaction in the returned order, then set its
blockhash, sign it with the authority, and submit it through the RPC client. The
proof instructions must remain directly after the confidential-transfer
instruction because their offsets are encoded into it.

This directory is a separate Go module so applications that do not use
Token-2022 confidential transfers do not inherit its cryptographic dependencies.
