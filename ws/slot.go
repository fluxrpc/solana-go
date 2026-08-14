package ws

import (
	"context"

	solana "github.com/fluxrpc/solana-go"
)

// SlotResult is a slotNotification payload.
type SlotResult struct {
	Parent uint64 `json:"parent"`
	Root   uint64 `json:"root"`
	Slot   uint64 `json:"slot"`
}

// SlotSubscribe subscribes to slot processing notifications.
func (c *Client) SlotSubscribe(ctx context.Context) (*Subscription[SlotResult], error) {
	return subscribe[SlotResult](ctx, c, "slotSubscribe", "slotUnsubscribe", nil)
}

// RootResult is a rootNotification payload: the latest root slot.
type RootResult uint64

// RootSubscribe subscribes to root changes.
func (c *Client) RootSubscribe(ctx context.Context) (*Subscription[RootResult], error) {
	return subscribe[RootResult](ctx, c, "rootSubscribe", "rootUnsubscribe", nil)
}

// SlotsUpdatesResult is a slotsUpdatesNotification payload.
type SlotsUpdatesResult struct {
	// The parent slot (only for createdBank updates).
	Parent uint64 `json:"parent,omitempty"`
	// The updated slot.
	Slot uint64 `json:"slot"`
	// Unix timestamp of the update, in milliseconds.
	Timestamp int64 `json:"timestamp"`
	// The update type: firstShredReceived, completed, createdBank, frozen,
	// dead, optimisticConfirmation or root.
	Type string `json:"type"`
	// Error detail (only for dead updates).
	Err string `json:"err,omitempty"`
	// Bank statistics (only for frozen updates).
	Stats *SlotsUpdatesStats `json:"stats,omitempty"`
}

type SlotsUpdatesStats struct {
	MaxTransactionsPerEntry   uint64 `json:"maxTransactionsPerEntry"`
	NumFailedTransactions     uint64 `json:"numFailedTransactions"`
	NumSuccessfulTransactions uint64 `json:"numSuccessfulTransactions"`
	NumTransactionEntries     uint64 `json:"numTransactionEntries"`
}

// SlotsUpdatesSubscribe subscribes to detailed per-slot lifecycle updates.
func (c *Client) SlotsUpdatesSubscribe(ctx context.Context) (*Subscription[SlotsUpdatesResult], error) {
	return subscribe[SlotsUpdatesResult](ctx, c, "slotsUpdatesSubscribe", "slotsUpdatesUnsubscribe", nil)
}

// VoteResult is a voteNotification payload.
type VoteResult struct {
	// The vote hash.
	Hash solana.Hash `json:"hash"`
	// The slots covered by the vote.
	Slots []uint64 `json:"slots"`
	// The timestamp of the vote, if any.
	Timestamp *solana.UnixTimeSeconds `json:"timestamp"`
	// The signature of the transaction containing this vote.
	Signature solana.Signature `json:"signature"`
	// The public key of the vote account.
	VotePubkey solana.PublicKey `json:"votePubkey"`
}

// VoteSubscribe subscribes to vote notifications (requires the node to run
// with --rpc-pubsub-enable-vote-subscription).
func (c *Client) VoteSubscribe(ctx context.Context) (*Subscription[VoteResult], error) {
	return subscribe[VoteResult](ctx, c, "voteSubscribe", "voteUnsubscribe", nil)
}
