package associatedtokenaccount

const (
	CreateInstruction InstructionType = iota
	CreateIdempotentInstruction
	RecoverNestedInstruction
)
