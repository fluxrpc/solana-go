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

// Opts configures Client.SendAndConfirmWithOpts and confirmation waits.
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

// Client owns the RPC and optional WebSocket clients used by a confirmation
// workflow. It is safe for concurrent use when the supplied clients are.
type Client struct {
	rpc *rpc.Client
	ws  *ws.Client
}

// New creates a confirmation client. wsClient may be nil, in which case all
// confirmation waits use RPC polling.
func New(rpcClient *rpc.Client, wsClient *ws.Client) *Client {
	return &Client{rpc: rpcClient, ws: wsClient}
}

// SendAndConfirm sends the signed transaction and waits until finalized.
func (c *Client) SendAndConfirm(ctx context.Context, tx *solana.Transaction) (solana.Signature, error) {
	return c.SendAndConfirmWithOpts(ctx, tx, Opts{})
}

// SendAndConfirmWithOpts sends the signed transaction and waits until it
// reaches the configured commitment. When no WebSocket client was supplied,
// confirmation is polled over RPC.
//
// The transaction must already be signed: its first signature is the
// subscription key, and with a WebSocket client the subscription is
// registered before the transaction is sent.
func (c *Client) SendAndConfirmWithOpts(ctx context.Context, tx *solana.Transaction, opts Opts) (solana.Signature, error) {
	if tx == nil || len(tx.Signatures) == 0 {
		return solana.Signature{}, errors.New("transaction is not signed")
	}
	if c == nil || c.rpc == nil {
		return solana.Signature{}, errors.New("RPC client is required")
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

	if c.ws == nil {
		if _, err := c.rpc.SendTransactionWithOpts(ctx, tx, sendOpts); err != nil {
			return sig, err
		}
		return sig, c.waitViaPolling(ctx, sig, confirmationTarget(commitment), opts)
	}

	// Arm the notification pipeline before the cluster can see the
	// transaction, then send.
	sub, err := c.ws.SignatureSubscribe(ctx, sig, commitment)
	if err != nil {
		return sig, fmt.Errorf("signature subscribe: %w", err)
	}
	if _, err := c.rpc.SendTransactionWithOpts(ctx, tx, sendOpts); err != nil {
		c.unsubscribe(ctx, sub)
		return sig, err
	}
	err = c.waitOnSubscription(ctx, sub, opts)
	c.cleanupSubscription(ctx, sub, err)
	return sig, err
}

// WaitForConfirmation waits for an already-sent signature. It uses the
// configured WebSocket client when available and otherwise polls RPC.
func (c *Client) WaitForConfirmation(ctx context.Context, sig solana.Signature, opts Opts) error {
	commitment := opts.Commitment
	if commitment == "" {
		commitment = rpc.CommitmentFinalized
	}
	if c == nil || c.rpc == nil {
		return errors.New("RPC client is required")
	}
	if c.ws == nil {
		return c.waitViaPolling(ctx, sig, confirmationTarget(commitment), opts)
	}
	sub, err := c.ws.SignatureSubscribe(ctx, sig, commitment)
	if err != nil {
		return fmt.Errorf("signature subscribe: %w", err)
	}
	err = c.waitOnSubscription(ctx, sub, opts)
	c.cleanupSubscription(ctx, sub, err)
	return err
}

func (c *Client) waitOnSubscription(ctx context.Context, sub *ws.Subscription[ws.SignatureResult], opts Opts) error {
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

func (c *Client) waitViaPolling(ctx context.Context, sig solana.Signature, target confirmationTarget, opts Opts) error {
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
		res, err := c.rpc.GetSignatureStatuses(waitCtx, false, sig)
		if err != nil && waitCtx.Err() == nil {
			return err
		}
		if err == nil && len(res.Value) > 0 && res.Value[0] != nil {
			status := res.Value[0]
			if target.reached(status) {
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

// confirmationTarget is a normalized commitment target for a wait.
type confirmationTarget rpc.CommitmentType

// reached reports whether a signature status satisfies the target
// commitment. Finality is also recognized via Confirmations == null (rooted)
// for nodes that omit confirmationStatus.
func (target confirmationTarget) reached(status *rpc.SignatureStatusesResult) bool {
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
	switch rpc.CommitmentType(target) {
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
func (c *Client) cleanupSubscription(ctx context.Context, sub *ws.Subscription[ws.SignatureResult], waitErr error) {
	var execErr *ExecutionError
	if waitErr == nil || errors.As(waitErr, &execErr) {
		sub.Release()
		return
	}
	c.unsubscribe(ctx, sub)
}

// unsubscribe releases the server-side subscription with a short deadline
// detached from the (possibly already canceled) parent context. Errors are
// ignored: the subscription may already be gone.
func (c *Client) unsubscribe(ctx context.Context, sub *ws.Subscription[ws.SignatureResult]) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer cancel()
	_ = sub.Unsubscribe(cleanupCtx)
}
