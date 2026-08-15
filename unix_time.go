package solana_go

import "time"

// UnixTimeSeconds represents a UNIX second-resolution timestamp,
// JSON-encoded as a plain number.
type UnixTimeSeconds int64

// Time converts the timestamp to a time.Time.
func (res UnixTimeSeconds) Time() time.Time {
	return time.Unix(int64(res), 0)
}

// String formats the timestamp in the default time.Time format.
func (res UnixTimeSeconds) String() string {
	return res.Time().String()
}
