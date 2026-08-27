package addresslookuptable

import "errors"

var (
	ErrUnknownInstruction = errors.New("addresslookup: unknown instruction")
	ErrTooManyAddresses   = errors.New("addresslookup: too many addresses")
)
