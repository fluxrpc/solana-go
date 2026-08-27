package addresslookuptable

import solana "github.com/fluxrpc/solana-go"

// CloseLookupTable closes a deactivated table and drains its lamports.
type CloseLookupTable struct{ instruction }

func NewCloseLookupTableInstruction(lookupTable, authority, recipient solana.PublicKey) *CloseLookupTable {
	return &CloseLookupTable{instruction: instruction{AccountMetaSlice: solana.AccountMetaSlice{
		solana.NewAccountMeta(lookupTable, true, false),
		solana.NewAccountMeta(authority, false, true),
		solana.NewAccountMeta(recipient, true, false),
	}}}
}

func (*CloseLookupTable) Data() ([]byte, error) {
	return []byte{byte(CloseLookupTableInstruction), 0, 0, 0}, nil
}
