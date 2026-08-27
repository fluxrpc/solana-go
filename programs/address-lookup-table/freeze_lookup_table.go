package addresslookuptable

import solana "github.com/fluxrpc/solana-go"

// FreezeLookupTable permanently makes a populated lookup table immutable.
type FreezeLookupTable struct{ instruction }

func NewFreezeLookupTableInstruction(lookupTable, authority solana.PublicKey) *FreezeLookupTable {
	return &FreezeLookupTable{instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(lookupTable, true, false),
		solana.NewAccountMeta(authority, false, true),
	}}}
}

func (*FreezeLookupTable) Data() ([]byte, error) {
	return []byte{byte(FreezeLookupTableInstruction), 0, 0, 0}, nil
}
