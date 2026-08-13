package rpc

import solana "github.com/fluxrpc/solana-go"

// DataSlice limits the returned account data to the given range.
type DataSlice struct {
	Offset *uint64 `json:"offset,omitempty"`
	Length *uint64 `json:"length,omitempty"`
}

// RPCFilter is a single account filter; exactly one of the fields is set.
type RPCFilter struct {
	Memcmp   *RPCFilterMemcmp `json:"memcmp,omitempty"`
	DataSize uint64           `json:"dataSize,omitempty"`
}

// RPCFilterMemcmp matches accounts whose data equals the provided bytes at
// the provided offset.
type RPCFilterMemcmp struct {
	Offset uint64        `json:"offset"`
	Bytes  solana.Base58 `json:"bytes"`
}
