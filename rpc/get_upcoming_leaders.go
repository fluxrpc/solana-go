package rpc

// GetUpcomingLeadersResult is the response from FluxRPC's custom
// getUpcomingLeaders method.
type GetUpcomingLeadersResult struct {
	Slot    uint64                `json:"slot"`
	Leaders []*UpcomingLeaderInfo `json:"leaders"`
}

// UpcomingLeaderInfo groups a validator's upcoming slots with its cluster
// contact information.
type UpcomingLeaderInfo struct {
	Slots       []uint64               `json:"slots"`
	ClusterInfo *GetClusterNodesResult `json:"clusterInfo"`
}
