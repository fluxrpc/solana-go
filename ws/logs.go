package ws

import (
	"context"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

// LogsSubscribeFilter selects which transactions to receive logs for.
type LogsSubscribeFilter string

const (
	// LogsAll subscribes to all transactions except simple vote transactions.
	LogsAll LogsSubscribeFilter = "all"
	// LogsAllWithVotes subscribes to all transactions including votes.
	LogsAllWithVotes LogsSubscribeFilter = "allWithVotes"
)

// LogResult is a logsNotification payload.
type LogResult struct {
	Context rpc.Context `json:"context"`
	Value   struct {
		// The transaction signature.
		Signature solana.Signature `json:"signature"`
		// Error if the transaction failed, nil if it succeeded.
		Err any `json:"err"`
		// Log messages the transaction instructions output.
		Logs []string `json:"logs"`
	} `json:"value"`
}

// LogsSubscribe subscribes to transaction logging.
func (c *Client) LogsSubscribe(ctx context.Context, filter LogsSubscribeFilter, commitment rpc.CommitmentType) (*Subscription[LogResult], error) {
	return c.logsSubscribe(ctx, string(filter), commitment)
}

// LogsSubscribeMentions subscribes to logs of transactions mentioning the
// given account.
func (c *Client) LogsSubscribeMentions(ctx context.Context, mentions solana.PublicKey, commitment rpc.CommitmentType) (*Subscription[LogResult], error) {
	return c.logsSubscribe(ctx, rpc.M{"mentions": []solana.PublicKey{mentions}}, commitment)
}

func (c *Client) logsSubscribe(ctx context.Context, filter any, commitment rpc.CommitmentType) (*Subscription[LogResult], error) {
	params := []any{filter}
	if commitment != "" {
		params = append(params, rpc.M{"commitment": commitment})
	}
	return subscribe[LogResult](ctx, c, "logsSubscribe", "logsUnsubscribe", params)
}
