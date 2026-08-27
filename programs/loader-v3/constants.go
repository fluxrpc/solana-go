package loaderv3

const (
	InitializeBufferInstruction InstructionType = iota
	WriteInstruction
	DeployWithMaxDataLenInstruction
	UpgradeInstruction
	SetAuthorityInstruction
	CloseInstruction
	ExtendProgramInstruction
	SetAuthorityCheckedInstruction
)
