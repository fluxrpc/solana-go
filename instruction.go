package solana_go

// Instruction is a program instruction that can be compiled into a
// transaction message.
type Instruction interface {
	// ProgramID is the program the instruction acts on.
	ProgramID() PublicKey
	// Accounts returns the list of accounts the instruction requires.
	Accounts() []*AccountMeta
	// Data is the binary encoded instruction data.
	Data() ([]byte, error)
}

// NewInstruction creates a GenericInstruction from a program ID, account list
// and already-encoded instruction data.
func NewInstruction(programID PublicKey, accounts AccountMetaSlice, data []byte) *GenericInstruction {
	return &GenericInstruction{
		AccountValues: accounts,
		ProgID:        programID,
		DataBytes:     data,
	}
}

// GenericInstruction is a basic Instruction implementation.
type GenericInstruction struct {
	AccountValues AccountMetaSlice
	ProgID        PublicKey
	DataBytes     []byte
}

var _ Instruction = (*GenericInstruction)(nil)

// ProgramID returns the program the instruction acts on.
func (in *GenericInstruction) ProgramID() PublicKey { return in.ProgID }

// Accounts returns the accounts the instruction requires.
func (in *GenericInstruction) Accounts() []*AccountMeta { return in.AccountValues }

// Data returns the binary encoded instruction data.
func (in *GenericInstruction) Data() ([]byte, error) { return in.DataBytes, nil }

// CompiledInstruction is an instruction inside a compiled transaction
// message, referencing accounts by their index in the message account list.
type CompiledInstruction struct {
	// Index into the message.accountKeys array indicating the program account
	// that executes this instruction.
	// NOTE: it is actually a uint8, but using a uint16 because uint8 is
	// treated as a byte everywhere, and that can be an issue.
	ProgramIDIndex uint16 `json:"programIdIndex"`

	// List of ordered indices into the message.accountKeys array indicating
	// which accounts to pass to the program.
	// NOTE: it is actually a []uint8, but using a uint16 because []uint8 is
	// treated as a []byte everywhere, and that can be an issue.
	Accounts []uint16 `json:"accounts"`

	// The program input data encoded in a base-58 string.
	Data Base58 `json:"data"`
}
