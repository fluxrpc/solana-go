package rpc

// BlocksResult is the response of the getBlocks RPC method: an array of u64
// integers listing confirmed blocks in the requested slot range.
type BlocksResult []uint64
