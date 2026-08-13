package rpc

import (
	solana "github.com/fluxrpc/solana-go"
)

// LoadedAddresses lists the addresses a transaction loaded from address
// lookup tables, split by writability.
type LoadedAddresses struct {
	ReadOnly []solana.PublicKey `json:"readonly"`
	Writable []solana.PublicKey `json:"writable"`
}
