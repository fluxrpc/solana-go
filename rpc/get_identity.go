package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// GetIdentityResult is the response of the getIdentity RPC method.
type GetIdentityResult struct {
	// The identity pubkey of the current node.
	Identity solana.PublicKey `json:"identity"`
}
