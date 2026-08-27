package vote

import "errors"

var (
	ErrUnknownInstruction       = errors.New("vote: unknown instruction")
	ErrInvalidVoteAuthorize     = errors.New("vote: invalid authorization kind")
	ErrInvalidCommissionKind    = errors.New("vote: invalid commission kind")
	ErrTooManyLockouts          = errors.New("vote: too many lockouts")
	ErrInvalidConfirmationCount = errors.New("vote: confirmation count exceeds compact u8")
	ErrSlotOverflow             = errors.New("vote: compact slot delta overflow")
	ErrInvalidLockoutOrder      = errors.New("vote: lockout slots must not decrease")
)
