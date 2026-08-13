package rpc

// GetHighestSnapshotSlotResult is the response of the getHighestSnapshotSlot
// RPC method.
type GetHighestSnapshotSlotResult struct {
	Full        uint64  `json:"full"`                  // Highest full snapshot slot.
	Incremental *uint64 `json:"incremental,omitempty"` // Highest incremental snapshot slot based on full.
}
