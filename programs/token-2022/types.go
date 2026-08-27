package token2022

import token "github.com/fluxrpc/solana-go/programs/token"

type (
	InstructionType     = token.InstructionType
	AuthorityType       = token.AuthorityType
	AccountState        = token.AccountState
	InitializeMint      = token.InitializeMint
	InitializeAccount   = token.InitializeAccount
	InitializeMultisig  = token.InitializeMultisig
	Transfer            = token.Transfer
	Approve             = token.Approve
	Revoke              = token.Revoke
	SetAuthority        = token.SetAuthority
	MintTo              = token.MintTo
	Burn                = token.Burn
	CloseAccount        = token.CloseAccount
	FreezeAccount       = token.FreezeAccount
	ThawAccount         = token.ThawAccount
	TransferChecked     = token.TransferChecked
	ApproveChecked      = token.ApproveChecked
	MintToChecked       = token.MintToChecked
	BurnChecked         = token.BurnChecked
	InitializeAccount2  = token.InitializeAccount2
	SyncNative          = token.SyncNative
	InitializeAccount3  = token.InitializeAccount3
	InitializeMultisig2 = token.InitializeMultisig2
	InitializeMint2     = token.InitializeMint2
	Authority           = token.AuthorityType
)

const (
	MaxSigners             = token.MaxSigners
	AuthorityMintTokens    = token.AuthorityMintTokens
	AuthorityFreezeAccount = token.AuthorityFreezeAccount
	AuthorityAccountOwner  = token.AuthorityAccountOwner
	AuthorityCloseAccount  = token.AuthorityCloseAccount
	Uninitialized          = token.Uninitialized
	Initialized            = token.Initialized
	Frozen                 = token.Frozen
)

const (
	InitializeMintInstruction      = token.InitializeMintInstruction
	InitializeAccountInstruction   = token.InitializeAccountInstruction
	InitializeMultisigInstruction  = token.InitializeMultisigInstruction
	TransferInstruction            = token.TransferInstruction
	ApproveInstruction             = token.ApproveInstruction
	RevokeInstruction              = token.RevokeInstruction
	SetAuthorityInstruction        = token.SetAuthorityInstruction
	MintToInstruction              = token.MintToInstruction
	BurnInstruction                = token.BurnInstruction
	CloseAccountInstruction        = token.CloseAccountInstruction
	FreezeAccountInstruction       = token.FreezeAccountInstruction
	ThawAccountInstruction         = token.ThawAccountInstruction
	TransferCheckedInstruction     = token.TransferCheckedInstruction
	ApproveCheckedInstruction      = token.ApproveCheckedInstruction
	MintToCheckedInstruction       = token.MintToCheckedInstruction
	BurnCheckedInstruction         = token.BurnCheckedInstruction
	InitializeAccount2Instruction  = token.InitializeAccount2Instruction
	SyncNativeInstruction          = token.SyncNativeInstruction
	InitializeAccount3Instruction  = token.InitializeAccount3Instruction
	InitializeMultisig2Instruction = token.InitializeMultisig2Instruction
	InitializeMint2Instruction     = token.InitializeMint2Instruction

	GetAccountDataSizeInstruction               InstructionType = 21
	InitializeImmutableOwnerInstruction         InstructionType = 22
	AmountToUIAmountInstruction                 InstructionType = 23
	UIAmountToAmountInstruction                 InstructionType = 24
	InitializeMintCloseAuthorityInstruction     InstructionType = 25
	TransferFeeExtensionInstruction             InstructionType = 26
	ConfidentialTransferExtensionInstruction    InstructionType = 27
	DefaultAccountStateExtensionInstruction     InstructionType = 28
	ReallocateInstruction                       InstructionType = 29
	MemoTransferExtensionInstruction            InstructionType = 30
	CreateNativeMintInstruction                 InstructionType = 31
	InitializeNonTransferableMintInstruction    InstructionType = 32
	InterestBearingMintExtensionInstruction     InstructionType = 33
	CPIGuardExtensionInstruction                InstructionType = 34
	InitializePermanentDelegateInstruction      InstructionType = 35
	TransferHookExtensionInstruction            InstructionType = 36
	ConfidentialTransferFeeExtensionInstruction InstructionType = 37
	WithdrawExcessLamportsInstruction           InstructionType = 38
	MetadataPointerExtensionInstruction         InstructionType = 39
	GroupPointerExtensionInstruction            InstructionType = 40
	GroupMemberPointerExtensionInstruction      InstructionType = 41
	ConfidentialMintBurnExtensionInstruction    InstructionType = 42
	ScaledUIAmountExtensionInstruction          InstructionType = 43
	PausableExtensionInstruction                InstructionType = 44
	UnwrapLamportsInstruction                   InstructionType = 45
	PermissionedBurnExtensionInstruction        InstructionType = 46
)

type DecodedInstruction struct {
	Base                               *token.DecodedInstruction
	GetAccountDataSize                 *GetAccountDataSize
	InitializeImmutableOwner           *InitializeImmutableOwner
	AmountToUIAmount                   *AmountToUIAmount
	UIAmountToAmount                   *UIAmountToAmount
	InitializeMintCloseAuthority       *InitializeMintCloseAuthority
	InitializeTransferFeeConfig        *InitializeTransferFeeConfig
	TransferCheckedWithFee             *TransferCheckedWithFee
	WithdrawWithheldTokensFromMint     *WithdrawWithheldTokensFromMint
	WithdrawWithheldTokensFromAccounts *WithdrawWithheldTokensFromAccounts
	HarvestWithheldTokensToMint        *HarvestWithheldTokensToMint
	SetTransferFee                     *SetTransferFee
	InitializeMetadataPointer          *InitializeMetadataPointer
	InitializeMetadata                 *InitializeMetadata
}
