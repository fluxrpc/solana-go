package addresslookuptable

// LookupTableMaxAddresses is the maximum number of addresses one table can
// hold. It also bounds instruction decode allocation.
const LookupTableMaxAddresses = 256

const (
	CreateLookupTableInstruction InstructionType = iota
	FreezeLookupTableInstruction
	ExtendLookupTableInstruction
	DeactivateLookupTableInstruction
	CloseLookupTableInstruction
)
