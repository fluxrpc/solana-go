package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	solana "github.com/fluxrpc/solana-go"
)

// GetLeaderScheduleOpts is the optional configuration object for the
// getLeaderSchedule RPC method.
type GetLeaderScheduleOpts struct {
	Commitment CommitmentType

	// Fetch the leader schedule for the epoch that corresponds
	// to the provided slot.
	// If unspecified, the leader schedule for the current epoch is fetched
	Epoch *uint64

	Identity *solana.PublicKey // Only return results for this validator identity
}

// GetLeaderScheduleResult is the response of the getLeaderSchedule RPC
// method: a dictionary of validator identities and their corresponding
// leader slot indices as values (indices are relative to the first slot in
// the requested epoch).
type GetLeaderScheduleResult map[solana.PublicKey][]uint64

// MarshalJSON encodes the map with base58 identity strings as keys.
// (solana.PublicKey is not a TextMarshaler, so the JSON key conversion is
// done here instead.)
func (m GetLeaderScheduleResult) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	out := make(map[string][]uint64, len(m))
	for identity, slots := range m {
		out[identity.String()] = slots
	}
	return json.Marshal(out)
}

// UnmarshalJSON decodes an object keyed by base58 identity strings.
func (m *GetLeaderScheduleResult) UnmarshalJSON(data []byte) error {
	var raw map[string][]uint64
	if err := sonic.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == nil {
		*m = nil
		return nil
	}
	out := make(GetLeaderScheduleResult, len(raw))
	for identity, slots := range raw {
		key, err := solana.PublicKeyFromBase58(identity)
		if err != nil {
			return fmt.Errorf("invalid identity key %q: %w", identity, err)
		}
		out[key] = slots
	}
	*m = out
	return nil
}
