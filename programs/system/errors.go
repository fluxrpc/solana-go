package system

import "errors"

var (
	// ErrUnknownInstruction is returned for a discriminator not implemented by
	// the System Program.
	ErrUnknownInstruction = errors.New("system: unknown instruction")
)
