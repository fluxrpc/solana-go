package loaderv4

const (
	WriteInstruction InstructionType = iota
	CopyInstruction
	SetProgramLengthInstruction
	DeployInstruction
	RetractInstruction
	TransferAuthorityInstruction
	FinalizeInstruction
)
