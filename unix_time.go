package solana_go

import "time"

// UnixTimeSeconds represents a UNIX second-resolution timestamp,
// JSON-encoded as a plain number.
type UnixTimeSeconds int64

func (res UnixTimeSeconds) Time() time.Time {
	return time.Unix(int64(res), 0)
}

func (res UnixTimeSeconds) String() string {
	return res.Time().String()
}
