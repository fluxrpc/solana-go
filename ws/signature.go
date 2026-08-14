package ws

import (
	"context"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

// SignatureResult is a signatureNotification payload. The subscription
// auto-cancels server-side after delivering one processed notification.
type SignatureResult struct {
	Context rpc.Context          `json:"context"`
	Value   SignatureResultValue `json:"value"`
}

// SignatureResultValue is either the processed-transaction status object or,
// when received-notifications are enabled, the "receivedSignature" marker.
type SignatureResultValue struct {
	// Error if the transaction failed, nil if it succeeded.
	Err any `json:"err"`
	// Received is true for a receivedSignature notification (the
	// transaction reached the node but is not yet processed).
	Received bool `json:"-"`
}

func (v *SignatureResultValue) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		// "receivedSignature"
		v.Received = true
		return nil
	}
	type alias SignatureResultValue
	return sonic.Unmarshal(data, (*alias)(v))
}

// SignatureSubscribe subscribes to a notification when the transaction with
// the given signature reaches the given commitment.
func (c *Client) SignatureSubscribe(ctx context.Context, signature solana.Signature, commitment rpc.CommitmentType) (*Subscription[SignatureResult], error) {
	return c.SignatureSubscribeWithOpts(ctx, signature, commitment, false)
}

func (c *Client) SignatureSubscribeWithOpts(ctx context.Context, signature solana.Signature, commitment rpc.CommitmentType, enableReceivedNotification bool) (*Subscription[SignatureResult], error) {
	opts := rpc.M{}
	if commitment != "" {
		opts["commitment"] = commitment
	}
	if enableReceivedNotification {
		opts["enableReceivedNotification"] = true
	}
	params := []any{signature}
	if len(opts) > 0 {
		params = append(params, opts)
	}
	return subscribe[SignatureResult](ctx, c, "signatureSubscribe", "signatureUnsubscribe", params)
}
