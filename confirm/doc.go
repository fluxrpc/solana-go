// Package confirm provides a client that sends transactions and waits for
// cluster confirmation.
//
// The WebSocket path subscribes to the transaction signature before sending,
// so confirmation cannot race subscription setup. Without a WebSocket client,
// the same [Client] polls getSignatureStatuses instead:
//
//	confirmer := confirm.New(rpcClient, wsClient) // wsClient may be nil
//	signature, err := confirmer.SendAndConfirm(ctx, signedTransaction)
//
// [Client.WaitForConfirmation] also waits for a transaction sent elsewhere.
// A confirmed-but-failed transaction returns [ExecutionError], which retains
// the raw on-chain error payload.
package confirm
