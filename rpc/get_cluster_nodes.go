package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetClusterNodesResult describes one node participating in the cluster, as
// returned by the getClusterNodes RPC method.
type GetClusterNodesResult struct {
	// Node public key.
	Pubkey solana.PublicKey `json:"pubkey"`

	// Gossip network address for the node.
	Gossip *string `json:"gossip,omitempty"`

	// TPU network address for the node.
	TPU *string `json:"tpu,omitempty"`

	// TPU QUIC network address for the node.
	TPUQUIC *string `json:"tpuQuic,omitempty"`

	// TPU forwards network address for the node.
	TPUForwards *string `json:"tpuForwards,omitempty"`

	// TPU forwards QUIC network address for the node.
	TPUForwardsQUIC *string `json:"tpuForwardsQuic,omitempty"`

	// TPU vote network address for the node.
	TPUVote *string `json:"tpuVote,omitempty"`

	// Serve repair network address for the node.
	ServeRepair *string `json:"serveRepair,omitempty"`

	// RPC WebSocket network address for the node, or empty if the WebSocket RPC service is not enabled.
	PubSub *string `json:"pubsub,omitempty"`

	// JSON RPC network address for the node, or empty if the JSON RPC service is not enabled.
	RPC *string `json:"rpc,omitempty"`

	// The software version of the node, or empty if the version information is not available.
	Version *string `json:"version,omitempty"`

	// The unique identifier of the node's feature set.
	FeatureSet *uint32 `json:"featureSet,omitempty"`

	// The shred version the node has been configured to use.
	ShredVersion uint16 `json:"shredVersion,omitempty"`

	// The client identifier for the node.
	ClientID *string `json:"clientId,omitempty"`
}
