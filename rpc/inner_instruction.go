package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// InnerInstruction groups the cross-program instructions invoked during one
// transaction instruction.
type InnerInstruction struct {
	// Index of the transaction instruction from which the inner instruction(s) originated.
	Index uint16 `json:"index"`

	// Ordered list of inner program instructions that were invoked during a single transaction instruction.
	Instructions []CompiledInstruction `json:"instructions"`
}

// CompiledInstruction is an instruction as returned by the RPC in
// non-parsed encodings, referencing accounts by index into the message's
// account keys.
type CompiledInstruction struct {
	// Index into the message.accountKeys array indicating the program account that executes this instruction.
	// NOTE: it is actually a uint8, but using a uint16 because uint8 is treated as a byte everywhere,
	// and that can be an issue.
	ProgramIDIndex uint16 `json:"programIdIndex"`

	// List of ordered indices into the message.accountKeys array indicating which accounts to pass to the program.
	// NOTE: it is actually a []uint8, but using a uint16 because []uint8 is treated as a []byte everywhere,
	// and that can be an issue.
	Accounts []uint16 `json:"accounts"`

	// The program input data encoded in a base-58 string.
	Data solana.Base58 `json:"data"`

	StackHeight uint16 `json:"stackHeight"`
}
