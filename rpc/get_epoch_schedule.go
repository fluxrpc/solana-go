package rpc

// GetEpochScheduleResult is the response of the getEpochSchedule RPC method.
type GetEpochScheduleResult struct {
	// The maximum number of slots in each epoch.
	SlotsPerEpoch uint64 `json:"slotsPerEpoch"`

	// The number of slots before beginning of an epoch to calculate a leader schedule for that epoch.
	LeaderScheduleSlotOffset uint64 `json:"leaderScheduleSlotOffset"`

	// Whether epochs start short and grow.
	Warmup bool `json:"warmup"`

	// First normal-length epoch, log2(slotsPerEpoch) - log2(MINIMUM_SLOTS_PER_EPOCH)
	FirstNormalEpoch uint64 `json:"firstNormalEpoch"`

	// MINIMUM_SLOTS_PER_EPOCH * (2.pow(firstNormalEpoch) - 1)
	FirstNormalSlot uint64 `json:"firstNormalSlot"`
}
