package rpc

import (
	"encoding/json"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

// Fixture: getClusterNodes response from upstream gagliardetto/solana-go
// rpc/client_test.go (TestClient_GetClusterNodes_AllFields).
const getClusterNodesFixture = `[{"pubkey":"hyp3Eo67t6FgeuWg5Qxbeme8NPXJPXXdKT4iJ4DsLf2","gossip":"127.0.0.1:8000","tpu":null,"tpuQuic":"127.0.0.1:8009","tpuForwards":null,"tpuForwardsQuic":"127.0.0.1:8010","tpuVote":"127.0.0.1:8005","serveRepair":"127.0.0.1:8008","rpc":"127.0.0.1:8899","pubsub":"127.0.0.1:8900","version":"2.2.1","featureSet":3580551090,"shredVersion":50093,"clientId":"Agave"}]`

func TestGetClusterNodesResultJSON(t *testing.T) {
	nodes := jsonRoundTrip[[]*GetClusterNodesResult](t, []byte(getClusterNodesFixture))

	if len(nodes) != 1 {
		t.Fatalf("len(nodes) = %d", len(nodes))
	}
	node := nodes[0]
	if node.Pubkey != solana.MustPublicKeyFromBase58("hyp3Eo67t6FgeuWg5Qxbeme8NPXJPXXdKT4iJ4DsLf2") {
		t.Fatalf("Pubkey = %s", node.Pubkey)
	}
	if node.Gossip == nil || *node.Gossip != "127.0.0.1:8000" {
		t.Fatalf("Gossip = %v", node.Gossip)
	}
	if node.TPU != nil || node.TPUForwards != nil {
		t.Fatalf("TPU = %v, TPUForwards = %v", node.TPU, node.TPUForwards)
	}
	if node.TPUQUIC == nil || *node.TPUQUIC != "127.0.0.1:8009" {
		t.Fatalf("TPUQUIC = %v", node.TPUQUIC)
	}
	if node.FeatureSet == nil || *node.FeatureSet != 3580551090 {
		t.Fatalf("FeatureSet = %v", node.FeatureSet)
	}
	if node.ShredVersion != 50093 {
		t.Fatalf("ShredVersion = %d", node.ShredVersion)
	}
	if node.ClientID == nil || *node.ClientID != "Agave" {
		t.Fatalf("ClientID = %v", node.ClientID)
	}
}

func TestGetClusterNodesResultBackwardCompatible(t *testing.T) {
	// Old response without the newer fields, from upstream
	// gagliardetto/solana-go rpc/client_test.go
	// (TestClient_GetClusterNodes_BackwardCompatible).
	in := `[{"pubkey":"hyp3Eo67t6FgeuWg5Qxbeme8NPXJPXXdKT4iJ4DsLf2","gossip":"127.0.0.1:8000","tpu":"127.0.0.1:8003","tpuQuic":"127.0.0.1:8009","rpc":"127.0.0.1:8899","version":"1.17.22","featureSet":3580551090,"shredVersion":50093}]`
	var nodes []*GetClusterNodesResult
	if err := json.Unmarshal([]byte(in), &nodes); err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("len(nodes) = %d", len(nodes))
	}
	node := nodes[0]
	if node.TPUForwards != nil || node.TPUForwardsQUIC != nil || node.TPUVote != nil ||
		node.ServeRepair != nil || node.ClientID != nil {
		t.Fatalf("newer fields not nil: %+v", node)
	}
	if node.TPU == nil || *node.TPU != "127.0.0.1:8003" {
		t.Fatalf("TPU = %v", node.TPU)
	}
}
