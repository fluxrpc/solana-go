package token

const MaxSigners = 11

const (
	InitializeMintInstruction InstructionType = iota
	InitializeAccountInstruction
	InitializeMultisigInstruction
	TransferInstruction
	ApproveInstruction
	RevokeInstruction
	SetAuthorityInstruction
	MintToInstruction
	BurnInstruction
	CloseAccountInstruction
	FreezeAccountInstruction
	ThawAccountInstruction
	TransferCheckedInstruction
	ApproveCheckedInstruction
	MintToCheckedInstruction
	BurnCheckedInstruction
	InitializeAccount2Instruction
	SyncNativeInstruction
	InitializeAccount3Instruction
	InitializeMultisig2Instruction
	InitializeMint2Instruction
)

const (
	MintSize     = 82
	AccountSize  = 165
	MultisigSize = 355
)
