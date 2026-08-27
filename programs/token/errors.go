package token

import "errors"

var (
	ErrUnknownInstruction = errors.New("token: unknown instruction")
	ErrInvalidAccountSize = errors.New("token: invalid account size")
)
