// Package system provides handwritten instruction encoders and decoders for
// Solana's native System Program.
//
// Constructors return concrete instruction types which can be placed directly
// in a []solana.Instruction. DecodeInstruction returns a concrete, typed
// envelope: its Type identifies the one populated instruction field.
package system
