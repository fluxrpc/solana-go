package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// RewardType is the category of a block or transaction reward.
type RewardType string

const (
	// RewardTypeFee is a transaction fee reward.
	RewardTypeFee RewardType = "Fee"
	// RewardTypeRent is a rent collection reward.
	RewardTypeRent RewardType = "Rent"
	// RewardTypeVoting is a voting reward.
	RewardTypeVoting RewardType = "Voting"
	// RewardTypeStaking is a staking reward.
	RewardTypeStaking RewardType = "Staking"
	// RewardTypeDeactivatedStake is a final deactivating stake reward.
	RewardTypeDeactivatedStake RewardType = "DeactivatedStake"
)

// BlockReward describes a reward credited or debited to an account.
type BlockReward struct {
	// The public key of the account that received the reward.
	Pubkey solana.PublicKey `json:"pubkey"`

	// Number of reward lamports credited or debited by the account, as a i64.
	Lamports int64 `json:"lamports"`

	// Account balance in lamports after the reward was applied.
	PostBalance uint64 `json:"postBalance"`

	// Type of reward: "Fee", "Rent", "Voting", "Staking", "DeactivatedStake".
	RewardType RewardType `json:"rewardType"`

	// Vote account commission when the reward was credited,
	// only present for voting and staking rewards.
	Commission *uint8 `json:"commission,omitempty"`

	// Vote account commission in basis points.
	CommissionBps *uint16 `json:"commissionBps,omitempty"`
}
