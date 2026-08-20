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
	sub, err := c.subscribe(ctx, "blockSubscribe", "blockUnsubscribe", []any{filter, opts})
	return (*Subscription[BlockResult])(sub), err
}

func (c *Client) blockSubscribe(ctx context.Context, filter any, commitment rpc.CommitmentType, extra rpc.M) (*Subscription[BlockResult], error) {
	sub, err := c.subscribe(ctx, "blockSubscribe", "blockUnsubscribe",
		[]any{filter, c.blockSubscribeOpts(solana.EncodingBase64, commitment, extra)})
	return (*Subscription[BlockResult])(sub), err
}

func (c *Client) blockSubscribeOpts(encoding solana.EncodingType, commitment rpc.CommitmentType, extra rpc.M) rpc.M {
	opts := rpc.M{
		"encoding":                       encoding,
		"maxSupportedTransactionVersion": uint64(0),
	}
	if commitment != "" {
		opts["commitment"] = commitment
	}
	for k, v := range extra {
		opts[k] = v
	}
	return opts
}

// ParsedBlockResult is a blockNotification payload with jsonParsed
// transaction encoding.
type ParsedBlockResult struct {
	Context rpc.Context `json:"context"`
	Value   struct {
		// The slot of the block.
		Slot uint64 `json:"slot"`
		// Error if something went wrong publishing the notification.
		Err any `json:"err,omitempty"`
		// The block itself, as a getBlock jsonParsed result.
		Block *rpc.GetParsedBlockResult `json:"block,omitempty"`
	} `json:"value"`
}

// ParsedBlockSubscribe subscribes to all new finalized blocks with
// jsonParsed transactions and maxSupportedTransactionVersion 0.
// NOTE: the node must run with --rpc-pubsub-enable-block-subscription.
func (c *Client) ParsedBlockSubscribe(ctx context.Context, commitment rpc.CommitmentType) (*Subscription[ParsedBlockResult], error) {
	return c.parsedBlockSubscribe(ctx, "all", commitment, nil)
}

// ParsedBlockSubscribeMentions subscribes to new blocks containing
// transactions that mention the given account or program, with jsonParsed
// transaction encoding.
func (c *Client) ParsedBlockSubscribeMentions(ctx context.Context, mentions solana.PublicKey, commitment rpc.CommitmentType) (*Subscription[ParsedBlockResult], error) {
	return c.parsedBlockSubscribe(ctx, rpc.M{"mentionsAccountOrProgram": mentions}, commitment, nil)
}

// ParsedBlockSubscribeWithOpts subscribes with full control of the
// getBlock-style options object; the encoding is forced to jsonParsed to
// match the ParsedBlockResult payload type.
func (c *Client) ParsedBlockSubscribeWithOpts(ctx context.Context, filter any, opts rpc.M) (*Subscription[ParsedBlockResult], error) {
	merged := rpc.M{}
	for k, v := range opts {
		merged[k] = v
	}
	merged["encoding"] = solana.EncodingJSONParsed
	sub, err := c.subscribe(ctx, "blockSubscribe", "blockUnsubscribe", []any{filter, merged})
	return (*Subscription[ParsedBlockResult])(sub), err
}

func (c *Client) parsedBlockSubscribe(ctx context.Context, filter any, commitment rpc.CommitmentType, extra rpc.M) (*Subscription[ParsedBlockResult], error) {
	sub, err := c.subscribe(ctx, "blockSubscribe", "blockUnsubscribe",
		[]any{filter, c.blockSubscribeOpts(solana.EncodingJSONParsed, commitment, extra)})
	return (*Subscription[ParsedBlockResult])(sub), err
}
