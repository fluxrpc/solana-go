package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetVoteAccountsOpts is the optional configuration object for the
// getVoteAccounts method.
type GetVoteAccountsOpts struct {
	Commitment CommitmentType `json:"commitment,omitempty"`

	// (optional) Only return results for this validator vote address.
	VotePubkey *solana.PublicKey `json:"votePubkey,omitempty"`

	// (optional) Do not filter out delinquent validators with no stake.
	KeepUnstakedDelinquents *bool `json:"keepUnstakedDelinquents,omitempty"`

	// (optional) Specify the number of slots behind the tip that
	// a validator must fall to be considered delinquent.
	// NOTE: For the sake of consistency between ecosystem products,
	// it is not recommended that this argument be specified.
	DelinquentSlotDistance *uint64 `json:"delinquentSlotDistance,omitempty"`
}

// GetVoteAccountsResult is the result of the getVoteAccounts method.
type GetVoteAccountsResult struct {
	Current    []VoteAccountsResult `json:"current"`
	Delinquent []VoteAccountsResult `json:"delinquent"`
}

// VoteAccountsResult describes a single vote account and its associated stake.
type VoteAccountsResult struct {
	// Vote account address.
	VotePubkey solana.PublicKey `json:"votePubkey,omitempty"`

	// Validator identity.
	NodePubkey solana.PublicKey `json:"nodePubkey,omitempty"`

	// The stake, in lamports, delegated to this vote account and active in this epoch.
	ActivatedStake uint64 `json:"activatedStake,omitempty"`

	// Whether the vote account is staked for this epoch.
	EpochVoteAccount bool `json:"epochVoteAccount,omitempty"`

	// Percentage (0-100) of rewards payout owed to the vote account.
	Commission uint8 `json:"commission,omitempty"`

	// Commission in basis points for inflation rewards.
	InflationRewardsCommissionBps *uint16 `json:"inflationRewardsCommissionBps,omitempty"`

	// Most recent slot voted on by this vote account.
	LastVote uint64 `json:"lastVote,omitempty"`

	RootSlot uint64 `json:"rootSlot,omitempty"` //

	// History of how many credits earned by the end of each epoch,
	// as an array of arrays containing: [epoch, credits, previousCredits]
	EpochCredits [][]int64 `json:"epochCredits,omitempty"`
}
