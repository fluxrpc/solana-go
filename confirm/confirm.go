// Package confirm sends transactions and waits for cluster confirmation.
//
// The WebSocket path subscribes to the transaction signature BEFORE
// sending, so a confirmation can never race the subscription: the
// notification pipeline is armed by the time the cluster first sees the
// transaction. Without a WebSocket client, a polling path over
// getSignatureStatuses is used instead.
//
// A confirmed-but-failed transaction (an instruction errored on chain)
// returns *ExecutionError, which carries the raw error payload; the wait is
// still considered successful in the sense that the signature reached the
// requested commitment.
package confirm

import (
	"context"
	"errors"
	"fmt"
	"time"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
	"github.com/fluxrpc/solana-go/ws"
)

// ErrTimeout is returned when the confirmation wait exceeds Opts.Timeout.
var ErrTimeout = errors.New("timeout waiting for transaction confirmation")

// ExecutionError is returned when the transaction reached the requested
// commitment but failed while executing.
type ExecutionError struct {
	// Err is the raw error payload from the cluster (the JSON "err" value).
	Err any
}

// Error implements the error interface.
func (e *ExecutionError) Error() string {
	return fmt.Sprintf("transaction confirmed with execution error: %v", e.Err)
}

// Opts configures SendAndConfirmWithOpts and the wait helpers.
type Opts struct {
	// Commitment the transaction must reach. Empty defaults to finalized.
	Commitment rpc.CommitmentType

	// SendOpts are passed to sendTransaction. A zero PreflightCommitment
	// defaults to the confirmation commitment, keeping preflight and
	// confirmation consistent.
	SendOpts rpc.TransactionOpts

	// Timeout bounds the confirmation wait. Zero or negative defaults to
	// 2 minutes.
	Timeout time.Duration

	// PollInterval is the getSignatureStatuses cadence of the polling
	// path. Zero or negative defaults to 500ms.
	PollInterval time.Duration
}

const (
	defaultTimeout      = 2 * time.Minute
	defaultPollInterval = 500 * time.Millisecond
)

// SendAndConfirm sends the signed transaction and waits until it reaches
// finalized commitment. wsClient may be nil, in which case confirmation is
// polled over RPC.
func SendAndConfirm(ctx context.Context, rpcClient *rpc.Client, wsClient *ws.Client, tx *solana.Transaction) (solana.Signature, error) {
	return SendAndConfirmWithOpts(ctx, rpcClient, wsClient, tx, Opts{})
}

// SendAndConfirmWithOpts sends the signed transaction and waits until it
// reaches the configured commitment. wsClient may be nil, in which case
// confirmation is polled over RPC.
//
// The transaction must already be signed: its first signature is the
// subscription key, and with a WebSocket client the subscription is
// registered before the transaction is sent.
func SendAndConfirmWithOpts(ctx context.Context, rpcClient *rpc.Client, wsClient *ws.Client, tx *solana.Transaction, opts Opts) (solana.Signature, error) {
	if len(tx.Signatures) == 0 {
		return solana.Signature{}, errors.New("transaction is not signed")
	}
	sig := tx.Signatures[0]

	commitment := opts.Commitment
	if commitment == "" {
		commitment = rpc.CommitmentFinalized
	}
	sendOpts := opts.SendOpts
	if sendOpts.PreflightCommitment == "" {
		sendOpts.PreflightCommitment = commitment
	}

	if wsClient == nil {
		if _, err := rpcClient.SendTransactionWithOpts(ctx, tx, sendOpts); err != nil {
			return sig, err
		}
		return sig, waitViaPolling(ctx, rpcClient, sig, commitment, opts)
	}

	// Arm the notification pipeline before the cluster can see the
	// transaction, then send.
	sub, err := wsClient.SignatureSubscribe(ctx, sig, commitment)
	if err != nil {
		return sig, fmt.Errorf("signature subscribe: %w", err)
	}
	if _, err := rpcClient.SendTransactionWithOpts(ctx, tx, sendOpts); err != nil {
		unsubscribe(ctx, sub)
		return sig, err
	}
	err = waitOnSubscription(ctx, sub, opts)
	cleanupSubscription(ctx, sub, err)
	return sig, err
}

// WaitForConfirmation waits until the given signature reaches the
// commitment configured in opts, using a signature subscription. It is the
// standalone wait used when the transaction was already sent elsewhere;
// note that a transaction confirmed before the subscription registers will
// only be caught at finalized commitment or by the polling variant.
func WaitForConfirmation(ctx context.Context, wsClient *ws.Client, sig solana.Signature, opts Opts) error {
	commitment := opts.Commitment
	if commitment == "" {
		commitment = rpc.CommitmentFinalized
	}
	sub, err := wsClient.SignatureSubscribe(ctx, sig, commitment)
	if err != nil {
		return fmt.Errorf("signature subscribe: %w", err)
	}
	err = waitOnSubscription(ctx, sub, opts)
	cleanupSubscription(ctx, sub, err)
	return err
}

// WaitForConfirmationViaPolling waits until the given signature reaches
// the commitment configured in opts by polling getSignatureStatuses.
func WaitForConfirmationViaPolling(ctx context.Context, rpcClient *rpc.Client, sig solana.Signature, opts Opts) error {
	commitment := opts.Commitment
	if commitment == "" {
		commitment = rpc.CommitmentFinalized
	}
	return waitViaPolling(ctx, rpcClient, sig, commitment, opts)
}

func waitOnSubscription(ctx context.Context, sub *ws.Subscription[ws.SignatureResult], opts Opts) error {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		got, err := sub.Recv(waitCtx)
		if err != nil {
			if waitCtx.Err() == context.DeadlineExceeded && ctx.Err() == nil {
				return ErrTimeout
			}
			return err
		}
		if got.Value.Received {
			// receivedSignature notification: the node has the
			// transaction; keep waiting for the commitment.
			continue
		}
		if got.Value.Err != nil {
			return &ExecutionError{Err: got.Value.Err}
		}
		return nil
	}
}

func waitViaPolling(ctx context.Context, rpcClient *rpc.Client, sig solana.Signature, commitment rpc.CommitmentType, opts Opts) error {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	interval := opts.PollInterval
	if interval <= 0 {
		interval = defaultPollInterval
	}
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		res, err := rpcClient.GetSignatureStatuses(waitCtx, false, sig)
		if err != nil && waitCtx.Err() == nil {
			return err
		}
		if err == nil && len(res.Value) > 0 && res.Value[0] != nil {
			status := res.Value[0]
			if statusReaches(status, commitment) {
				if status.Err != nil {
					return &ExecutionError{Err: status.Err}
				}
				return nil
			}
		}

		select {
		case <-waitCtx.Done():
			if ctx.Err() == nil {
				return ErrTimeout
			}
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// statusReaches reports whether a signature status satisfies the target
// commitment. Finality is also recognized via Confirmations == null (rooted)
// for nodes that omit confirmationStatus.
func statusReaches(status *rpc.SignatureStatusesResult, target rpc.CommitmentType) bool {
	rank := func(c rpc.ConfirmationStatusType) int {
		switch c {
		case rpc.ConfirmationStatusProcessed:
			return 0
		case rpc.ConfirmationStatusConfirmed:
			return 1
		case rpc.ConfirmationStatusFinalized:
			return 2
		}
		return -1
	}
	want := 2 // finalized
	switch target {
	case rpc.CommitmentProcessed:
		want = 0
	case rpc.CommitmentConfirmed:
		want = 1
	}
	if got := rank(status.ConfirmationStatus); got >= 0 {
		return got >= want
	}
	return status.Confirmations == nil // rooted
}

// cleanupSubscription releases the subscription. After a terminal
// notification (success or execution error) the server has already
// auto-canceled the single-shot signature subscription, so only the local
// resources are dropped; otherwise a real unsubscribe is sent.
func cleanupSubscription(ctx context.Context, sub *ws.Subscription[ws.SignatureResult], waitErr error) {
	var execErr *ExecutionError
	if waitErr == nil || errors.As(waitErr, &execErr) {
		sub.Release()
		return
	}
	unsubscribe(ctx, sub)
}

// unsubscribe releases the server-side subscription with a short deadline
// detached from the (possibly already canceled) parent context. Errors are
// ignored: the subscription may already be gone.
func unsubscribe(ctx context.Context, sub *ws.Subscription[ws.SignatureResult]) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	_ = sub.Unsubscribe(cleanupCtx)
}
