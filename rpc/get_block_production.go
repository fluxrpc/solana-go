package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
)

// GetBlockProductionResult is the response of the getBlockProduction RPC
// method.
type GetBlockProductionResult struct {
	RPCContext
	Value BlockProductionResult `json:"value"`
}

// GetBlockProductionOpts is the optional configuration object for the
// getBlockProduction RPC method.
type GetBlockProductionOpts struct {
	//
	// This parameter is optional.
	Commitment CommitmentType `json:"commitment,omitempty"`

	// Slot range to return block production for.
	// If parameter not provided, defaults to current epoch.
	//
	// This parameter is optional.
	Range *SlotRangeRequest `json:"range,omitempty"`

	// Only return results for this validator identity.
	//
	// This parameter is optional.
	Identity *solana.PublicKey `json:"identity,omitempty"`
}

// SlotRangeRequest is the slot range parameter of getBlockProduction.
type SlotRangeRequest struct {
	// First slot to return block production information for (inclusive)
	FirstSlot uint64 `json:"firstSlot"`

	// Last slot to return block production information for (inclusive).
	// If parameter not provided, defaults to the highest slot
	//
	// This parameter is optional.
	LastSlot *uint64 `json:"lastSlot,omitempty"`
}

// BlockProductionResult is the value of a getBlockProduction response.
type BlockProductionResult struct {
	ByIdentity IdentityToSlotsBlocks `json:"byIdentity"`

	Range SlotRangeResponse `json:"range"`
}

// IdentityToSlotsBlocks is a dictionary of validator identities.
// Value is a two element array containing the number
// of leader slots and the number of blocks produced.
type IdentityToSlotsBlocks map[solana.PublicKey][2]int64

// MarshalJSON encodes the map with base58 identity strings as keys.
// (solana.PublicKey is not a TextMarshaler, so the JSON key conversion is
// done here instead.)
func (m IdentityToSlotsBlocks) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	out := make(map[string][2]int64, len(m))
	for identity, slotsBlocks := range m {
		out[identity.String()] = slotsBlocks
	}
	return json.Marshal(out)
}

// UnmarshalJSON decodes an object keyed by base58 identity strings.
func (m *IdentityToSlotsBlocks) UnmarshalJSON(data []byte) error {
	var raw map[string][2]int64
	if err := sonic.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == nil {
		*m = nil
		return nil
	}
	out := make(IdentityToSlotsBlocks, len(raw))
	for identity, slotsBlocks := range raw {
		key, err := solana.PublicKeyFromBase58(identity)
		if err != nil {
			return fmt.Errorf("invalid identity key %q: %w", identity, err)
		}
		out[key] = slotsBlocks
	}
	*m = out
	return nil
}

// SlotRangeResponse is the slot range a getBlockProduction response covers.
type SlotRangeResponse struct {
	// First slot of the block production information (inclusive)
	FirstSlot uint64 `json:"firstSlot"`

	// Last slot of block production information (inclusive)
	LastSlot uint64 `json:"lastSlot"`
}
