package rpc

// Cluster describes one Solana cluster: its name and public RPC/WebSocket
// endpoints. See https://solana.com/docs/references/clusters.
//
// The public endpoints are heavily rate-limited and intended for
// development only; production workloads should use a dedicated RPC
// provider.
type Cluster struct {
	Name string
	RPC  string
	WS   string
}

// The well-known public clusters.
var (
	MainNetBeta = Cluster{
		Name: "mainnet-beta",
		RPC:  MainNetBeta_RPC,
		WS:   MainNetBeta_WS,
	}
	TestNet = Cluster{
		Name: "testnet",
		RPC:  TestNet_RPC,
		WS:   TestNet_WS,
	}
	DevNet = Cluster{
		Name: "devnet",
		RPC:  DevNet_RPC,
		WS:   DevNet_WS,
	}
	LocalNet = Cluster{
		Name: "localnet",
		RPC:  LocalNet_RPC,
		WS:   LocalNet_WS,
	}
)

// Public cluster HTTP RPC endpoints.
const (
	MainNetBeta_RPC = "https://api.mainnet-beta.solana.com"
	TestNet_RPC     = "https://api.testnet.solana.com"
	DevNet_RPC      = "https://api.devnet.solana.com"
	LocalNet_RPC    = "http://127.0.0.1:8899"
)

// Public cluster WebSocket pubsub endpoints.
const (
	MainNetBeta_WS = "wss://api.mainnet-beta.solana.com"
	TestNet_WS     = "wss://api.testnet.solana.com"
	DevNet_WS      = "wss://api.devnet.solana.com"
	LocalNet_WS    = "ws://127.0.0.1:8900"
)
