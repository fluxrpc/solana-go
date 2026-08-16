package rpc

import "testing"

func TestClusterEndpoints(t *testing.T) {
	// Values match the upstream solana-go constants and the official
	// cluster documentation.
	cases := []struct {
		cluster Cluster
		name    string
		rpc     string
		ws      string
	}{
		{MainnetBeta, "mainnet-beta", "https://api.mainnet-beta.solana.com", "wss://api.mainnet-beta.solana.com"},
		{Testnet, "testnet", "https://api.testnet.solana.com", "wss://api.testnet.solana.com"},
		{Devnet, "devnet", "https://api.devnet.solana.com", "wss://api.devnet.solana.com"},
		{Localnet, "localnet", "http://127.0.0.1:8899", "ws://127.0.0.1:8900"},
	}
	for _, tc := range cases {
		if tc.cluster.Name != tc.name || tc.cluster.RPC != tc.rpc || tc.cluster.WS != tc.ws {
			t.Errorf("cluster %s = %+v, want {%s %s %s}", tc.name, tc.cluster, tc.name, tc.rpc, tc.ws)
		}
	}
}
