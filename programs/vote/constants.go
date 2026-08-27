package vote

const (
	InitializeAccountInstruction InstructionType = iota
	AuthorizeInstruction
	VoteInstruction
	WithdrawInstruction
	UpdateValidatorIdentityInstruction
	UpdateCommissionInstruction
	VoteSwitchInstruction
	AuthorizeCheckedInstruction
	UpdateVoteStateInstruction
	UpdateVoteStateSwitchInstruction
	AuthorizeWithSeedInstruction
	AuthorizeCheckedWithSeedInstruction
	CompactUpdateVoteStateInstruction
	CompactUpdateVoteStateSwitchInstruction
	TowerSyncInstruction
	TowerSyncSwitchInstruction
	InitializeAccountV2Instruction
	UpdateCommissionCollectorInstruction
	UpdateCommissionBpsInstruction
	DepositDelegatorRewardsInstruction
)
const (
	MaxLockoutHistory                  = 31
	BLSPublicKeyCompressedSize         = 48
	BLSProofOfPossessionCompressedSize = 96
)

const (
	VoteAuthorizeVoter VoteAuthorizeKind = iota
	VoteAuthorizeWithdrawer
	VoteAuthorizeVoterWithBLS
)

const (
	CommissionKindInflationRewards CommissionKind = iota
	CommissionKindBlockRevenue
)
