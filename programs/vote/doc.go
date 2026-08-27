// Package vote implements the native Solana Vote Program instruction codec.
//
// The codec is handwritten and reflection-free. DecodeInstruction returns a
// typed envelope, so callers never need a type assertion.
package vote
