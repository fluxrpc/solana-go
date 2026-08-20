package ws

import (
	"context"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

// AccountResult is an accountNotification payload.
type AccountResult struct {
	Context rpc.Context `json:"context"`
	Value   rpc.Account `json:"value"`
}

// AccountSubscribe subscribes to changes of the given account, with base64
// data encoding.
func (c *Client) AccountSubscribe(ctx context.Context, account solana.PublicKey, commitment rpc.CommitmentType) (*Subscription[AccountResult], error) {
	return c.AccountSubscribeWithOpts(ctx, account, commitment, solana.EncodingBase64)
}

// AccountSubscribeWithOpts is AccountSubscribe with an explicit data
// encoding.
func (c *Client) AccountSubscribeWithOpts(ctx context.Context, account solana.PublicKey, commitment rpc.CommitmentType, encoding solana.EncodingType) (*Subscription[AccountResult], error) {
	opts := rpc.M{"encoding": encoding}
	if commitment != "" {
		opts["commitment"] = commitment
	}
	sub, err := c.subscribe(ctx, "accountSubscribe", "accountUnsubscribe", []any{account, opts})
	return (*Subscription[AccountResult])(sub), err
}

// ProgramResult is a programNotification payload.
type ProgramResult struct {
	Context rpc.Context      `json:"context"`
	Value   rpc.KeyedAccount `json:"value"`
}

// ProgramSubscribe subscribes to changes of all accounts owned by program.
func (c *Client) ProgramSubscribe(ctx context.Context, program solana.PublicKey, commitment rpc.CommitmentType) (*Subscription[ProgramResult], error) {
	return c.ProgramSubscribeWithOpts(ctx, program, commitment, solana.EncodingBase64, nil)
}

// ProgramSubscribeWithOpts is ProgramSubscribe with an explicit data
// encoding and account filters.
func (c *Client) ProgramSubscribeWithOpts(ctx context.Context, program solana.PublicKey, commitment rpc.CommitmentType, encoding solana.EncodingType, filters []rpc.RPCFilter) (*Subscription[ProgramResult], error) {
	opts := rpc.M{"encoding": encoding}
	if commitment != "" {
		opts["commitment"] = commitment
	}
	if len(filters) > 0 {
		opts["filters"] = filters
	}
	sub, err := c.subscribe(ctx, "programSubscribe", "programUnsubscribe", []any{program, opts})
	return (*Subscription[ProgramResult])(sub), err
}
