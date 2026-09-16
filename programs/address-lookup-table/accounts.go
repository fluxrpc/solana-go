package addresslookuptable

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

// lookupTableMetaSize is the offset at which addresses begin. The meta
// section before it is padded, not packed.
const lookupTableMetaSize = 56

// LookupTable is the on-chain state of an address lookup table account.
type LookupTable struct {
	TypeIndex                  uint32
	DeactivationSlot           uint64
	LastExtendedSlot           uint64
	LastExtendedSlotStartIndex uint8
	// Authority is nil once the table has been frozen.
	Authority *solana.PublicKey
	Addresses []solana.PublicKey
}

// DecodeLookupTable decodes an address lookup table account.
func DecodeLookupTable(data []byte) (LookupTable, error) {
	if len(data) < lookupTableMetaSize {
		return LookupTable{}, fmt.Errorf("%w: lookup table: %d", ErrInvalidAccountSize, len(data))
	}

	dec := bin.NewDecoder(data)
	table := LookupTable{
		TypeIndex:                  dec.ReadUint32(),
		DeactivationSlot:           dec.ReadUint64(),
		LastExtendedSlot:           dec.ReadUint64(),
		LastExtendedSlotStartIndex: dec.ReadUint8(),
	}
	// The Option's payload is always present, so the key is read either way.
	hasAuthority := dec.ReadOption()
	authority := dec.ReadPublicKey()
	if hasAuthority {
		table.Authority = &authority
	}
	if err := dec.Err(); err != nil {
		return LookupTable{}, fmt.Errorf("decode lookup table meta: %w", err)
	}

	// Trailing bytes that do not form a whole address are ignored, matching the
	// runtime's deserialization.
	count := (len(data) - lookupTableMetaSize) / solana.PublicKeyLength
	if count > LookupTableMaxAddresses {
		return LookupTable{}, fmt.Errorf("%w: %d > %d", ErrTooManyAddresses, count, LookupTableMaxAddresses)
	}

	addresses := data[lookupTableMetaSize:]
	table.Addresses = make([]solana.PublicKey, count)
	for index := range table.Addresses {
		table.Addresses[index] = solana.PublicKey(addresses[index*solana.PublicKeyLength:])
	}
	return table, nil
}
