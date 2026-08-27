package stake

const (
	InitializeInstruction InstructionType = iota
	AuthorizeInstruction
	DelegateStakeInstruction
	SplitInstruction
	WithdrawInstruction
	DeactivateInstruction
	SetLockupInstruction
	MergeInstruction
	AuthorizeWithSeedInstruction
	InitializeCheckedInstruction
	AuthorizeCheckedInstruction
	AuthorizeCheckedWithSeedInstruction
	SetLockupCheckedInstruction
	GetMinimumDelegationInstruction
	DeactivateDelinquentInstruction
	RedelegateInstruction
	MoveStakeInstruction
	MoveLamportsInstruction
)
const (
	StakeAuthorizeStaker StakeAuthorize = iota
	StakeAuthorizeWithdrawer
)
