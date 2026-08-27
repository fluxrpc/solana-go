package addresslookuptable

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

// ExtendLookupTable adds addresses to a lookup table. Payer and System Program
// accounts are present only when additional rent funding may be required.
type ExtendLookupTable struct {
	Addresses []solana.PublicKey
	instruction
}

func NewExtendLookupTableInstruction(
	lookupTable, authority, payer solana.PublicKey,
	addresses []solana.PublicKey,
) *ExtendLookupTable {
	return &ExtendLookupTable{
		Addresses: addresses,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(lookupTable, true, false),
			solana.NewAccountMeta(authority, false, true),
			solana.NewAccountMeta(payer, true, true),
			solana.NewAccountMeta(solana.SystemProgramID, false, false),
		}},
	}
}

// NewExtendLookupTableInstructionWithoutPayer creates the two-account form
// used when the table already has enough lamports for its new size.
func NewExtendLookupTableInstructionWithoutPayer(
	lookupTable, authority solana.PublicKey,
	addresses []solana.PublicKey,
) *ExtendLookupTable {
	return &ExtendLookupTable{
		Addresses: addresses,
		instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
			solana.NewAccountMeta(lookupTable, true, false),
			solana.NewAccountMeta(authority, false, true),
		}},
	}
}

func (inst *ExtendLookupTable) Data() ([]byte, error) {
	if len(inst.Addresses) > LookupTableMaxAddresses {
		return nil, fmt.Errorf("%w: %d > %d", ErrTooManyAddresses, len(inst.Addresses), LookupTableMaxAddresses)
	}
	enc := bin.NewEncoder(make([]byte, 0, 12+len(inst.Addresses)*solana.PublicKeyLength))
	enc.WriteUint32(uint32(ExtendLookupTableInstruction))
	enc.WriteUint64(uint64(len(inst.Addresses)))
	for _, address := range inst.Addresses {
		enc.WritePublicKey(address)
	}
	return enc.Bytes(), enc.Err()
}
