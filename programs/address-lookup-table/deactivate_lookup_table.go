package addresslookuptable

import solana "github.com/fluxrpc/solana-go"

// DeactivateLookupTable starts the cooldown after which a table may be closed.
type DeactivateLookupTable struct{ instruction }

func NewDeactivateLookupTableInstruction(lookupTable, authority solana.PublicKey) *DeactivateLookupTable {
	return &DeactivateLookupTable{instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(lookupTable, true, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func (*DeactivateLookupTable) Data() ([]byte, error) {
	return []byte{byte(DeactivateLookupTableInstruction), 0, 0, 0}, nil
}
