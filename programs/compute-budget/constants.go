package computebudget

const (
	UnusedInstruction InstructionType = iota
	RequestHeapFrameInstruction
	SetComputeUnitLimitInstruction
	SetComputeUnitPriceInstruction
	SetLoadedAccountsDataSizeLimitInstruction
)

// RequestUnitsDeprecatedInstruction is the historical Foundation name for
// tag zero. Current Rust interfaces reserve the same tag as Unused.
const RequestUnitsDeprecatedInstruction = UnusedInstruction
