package stake

import "errors"

// ErrUnknownInstruction is returned for a discriminator not implemented by
// the current Stake Program interface.
var ErrUnknownInstruction = errors.New("stake: unknown instruction")

// ErrInvalidStakeAuthorize is returned for a StakeAuthorize value other than
// Staker or Withdrawer.
var ErrInvalidStakeAuthorize = errors.New("stake: invalid authority type")
