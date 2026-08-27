package solana_go

// Native programs:
var (
	// Create new accounts, allocate account data, assign accounts to owning
	// programs, transfer lamports and pay transaction fees.
	SystemProgramID = MustPublicKeyFromBase58("11111111111111111111111111111111")

	// Add configuration data to the chain and the list of public keys that
	// are permitted to modify it.
	ConfigProgramID = MustPublicKeyFromBase58("Config1111111111111111111111111111111111111")

	// Create and manage accounts representing stake and rewards for
	// delegations to validators.
	StakeProgramID = MustPublicKeyFromBase58("Stake11111111111111111111111111111111111111")

	// Create and manage accounts that track validator voting state and rewards.
	VoteProgramID = MustPublicKeyFromBase58("Vote111111111111111111111111111111111111111")

	BPFLoaderDeprecatedProgramID = MustPublicKeyFromBase58("BPFLoader1111111111111111111111111111111111")

	// Deploys, upgrades, and executes programs on the chain.
	BPFLoaderProgramID            = MustPublicKeyFromBase58("BPFLoader2111111111111111111111111111111111")
	BPFLoaderUpgradeableProgramID = MustPublicKeyFromBase58("BPFLoaderUpgradeab1e11111111111111111111111")
	LoaderV4ProgramID             = MustPublicKeyFromBase58("LoaderV411111111111111111111111111111111111")
	NativeLoaderID                = MustPublicKeyFromBase58("NativeLoader1111111111111111111111111111111")

	// Verify secp256k1 public key recovery operations (ecrecover).
	Secp256k1ProgramID = MustPublicKeyFromBase58("KeccakSecp256k11111111111111111111111111111")

	FeatureProgramID = MustPublicKeyFromBase58("Feature111111111111111111111111111111111111")

	ComputeBudgetProgramID = MustPublicKeyFromBase58("ComputeBudget111111111111111111111111111111")
	// ComputeBudget is retained as a compatibility alias.
	ComputeBudget = ComputeBudgetProgramID

	// Create and manage address lookup tables.
	AddressLookupTableProgramID = MustPublicKeyFromBase58("AddressLookupTab1e1111111111111111111111111")
)

// SPL programs:
var (
	// Common implementation for fungible and non-fungible tokens.
	TokenProgramID = MustPublicKeyFromBase58("TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA")

	Token2022ProgramID = MustPublicKeyFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb")

	TokenSwapProgramID = MustPublicKeyFromBase58("SwaPpA9LAaLfeLi3a68M4DjnLqgtticKg6CnyNwgAC8")
	TokenSwapFeeOwner  = MustPublicKeyFromBase58("HfoTxFR1Tm6kGmWgYWD6J7YHVy1UwqSULUGVLXkJqaKN")

	TokenLendingProgramID = MustPublicKeyFromBase58("LendZqTs8gn5CTSJU1jWKhKuVpjJGom45nnwPb2AMTi")

	// Maps a wallet address to its associated token accounts.
	SPLAssociatedTokenAccountProgramID = MustPublicKeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL")

	// Validates a UTF-8 memo and verifies provided accounts signed the
	// transaction, logging the memo to the transaction log.
	MemoProgramID = MustPublicKeyFromBase58("MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr")

	// MemoProgramIDV1 is the deprecated v1 Memo program; some legacy
	// transactions still reference it.
	MemoProgramIDV1 = MustPublicKeyFromBase58("Memo1UhkJRfHyvLMcVucJwxXeuD728EqVDDwQDxFMNo")

	TokenMetadataProgramID = MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
)

var (
	// SolMint is the native SOL placeholder mint.
	SolMint = MustPublicKeyFromBase58("So11111111111111111111111111111111111111111")
	// WrappedSol is the wrapped SOL (wSOL) SPL mint.
	WrappedSol = MustPublicKeyFromBase58("So11111111111111111111111111111111111111112")
)

// Sysvars:
var (
	SysVarPubkey                  = MustPublicKeyFromBase58("Sysvar1111111111111111111111111111111111111")
	SysVarClockPubkey             = MustPublicKeyFromBase58("SysvarC1ock11111111111111111111111111111111")
	SysVarEpochRewardsPubkey      = MustPublicKeyFromBase58("SysvarEpochRewards1111111111111111111111111")
	SysVarEpochSchedulePubkey     = MustPublicKeyFromBase58("SysvarEpochSchedu1e111111111111111111111111")
	SysVarFeesPubkey              = MustPublicKeyFromBase58("SysvarFees111111111111111111111111111111111")
	SysVarInstructionsPubkey      = MustPublicKeyFromBase58("Sysvar1nstructions1111111111111111111111111")
	SysVarLastRestartSlotPubkey   = MustPublicKeyFromBase58("SysvarLastRestartS1ot1111111111111111111111")
	SysVarRecentBlockHashesPubkey = MustPublicKeyFromBase58("SysvarRecentB1ockHashes11111111111111111111")
	SysVarRentPubkey              = MustPublicKeyFromBase58("SysvarRent111111111111111111111111111111111")
	SysVarRewardsPubkey           = MustPublicKeyFromBase58("SysvarRewards111111111111111111111111111111")
	SysVarSlotHashesPubkey        = MustPublicKeyFromBase58("SysvarS1otHashes111111111111111111111111111")
	SysVarSlotHistoryPubkey       = MustPublicKeyFromBase58("SysvarS1otHistory11111111111111111111111111")
	SysVarStakeHistoryPubkey      = MustPublicKeyFromBase58("SysvarStakeHistory1111111111111111111111111")
	SysVarStakeConfigPubkey       = MustPublicKeyFromBase58("StakeConfig11111111111111111111111111111111")
)
