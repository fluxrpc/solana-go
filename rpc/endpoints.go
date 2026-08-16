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
	MainnetBeta = Cluster{
		Name: "mainnet-beta",
		RPC:  MainnetBetaRPC,
		WS:   MainnetBetaWS,
	}
	Testnet = Cluster{
		Name: "testnet",
		RPC:  TestnetRPC,
		WS:   TestnetWS,
	}
	Devnet = Cluster{
		Name: "devnet",
		RPC:  DevnetRPC,
		WS:   DevnetWS,
	}
	Localnet = Cluster{
		Name: "localnet",
		RPC:  LocalnetRPC,
		WS:   LocalnetWS,
	}
)

// Public cluster HTTP RPC endpoints.
const (
	MainnetBetaRPC = "https://api.mainnet-beta.solana.com"
	TestnetRPC     = "https://api.testnet.solana.com"
	DevnetRPC      = "https://api.devnet.solana.com"
	LocalnetRPC    = "http://127.0.0.1:8899"
)

// Public cluster WebSocket pubsub endpoints.
const (
	MainnetBetaWS = "wss://api.mainnet-beta.solana.com"
	TestnetWS     = "wss://api.testnet.solana.com"
	DevnetWS      = "wss://api.devnet.solana.com"
	LocalnetWS    = "ws://127.0.0.1:8900"
)
