package ws

import (
	"context"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

// BlockResult is a blockNotification payload.
type BlockResult struct {
	Context rpc.Context `json:"context"`
	Value   struct {
		// The slot of the block.
		Slot uint64 `json:"slot"`
		// Error if something went wrong publishing the notification.
		Err any `json:"err,omitempty"`
		// The block itself, as a getBlock result.
		Block *rpc.GetBlockResult `json:"block,omitempty"`
	} `json:"value"`
}

// BlockSubscribe subscribes to all new finalized blocks with base64-encoded
// transactions and maxSupportedTransactionVersion 0.
// NOTE: the node must run with --rpc-pubsub-enable-block-subscription.
func (c *Client) BlockSubscribe(ctx context.Context, commitment rpc.CommitmentType) (*Subscription[BlockResult], error) {
	return c.blockSubscribe(ctx, "all", commitment, nil)
}

// BlockSubscribeMentions subscribes to new blocks containing transactions
// that mention the given account or program.
func (c *Client) BlockSubscribeMentions(ctx context.Context, mentions solana.PublicKey, commitment rpc.CommitmentType) (*Subscription[BlockResult], error) {
	return c.blockSubscribe(ctx, rpc.M{"mentionsAccountOrProgram": mentions}, commitment, nil)
}

// BlockSubscribeWithOpts subscribes with full control of the getBlock-style
// options object.
func (c *Client) BlockSubscribeWithOpts(ctx context.Context, filter any, opts rpc.M) (*Subscription[BlockResult], error) {
	return subscribe[BlockResult](ctx, c, "blockSubscribe", "blockUnsubscribe", []any{filter, opts})
}

func (c *Client) blockSubscribe(ctx context.Context, filter any, commitment rpc.CommitmentType, extra rpc.M) (*Subscription[BlockResult], error) {
	version := uint64(0)
	opts := rpc.M{
		"encoding":                       solana.EncodingBase64,
		"maxSupportedTransactionVersion": version,
	}
	if commitment != "" {
		opts["commitment"] = commitment
	}
	for k, v := range extra {
		opts[k] = v
	}
	return subscribe[BlockResult](ctx, c, "blockSubscribe", "blockUnsubscribe", []any{filter, opts})
}
