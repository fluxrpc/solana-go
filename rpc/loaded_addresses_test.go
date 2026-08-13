package rpc

import (
	"testing"
)

// Fixture: loadedAddresses object of a transaction meta, from upstream
// gagliardetto/solana-go rpc/types_test.go (TestParsedTransactionMeta_Decode).
const loadedAddressesFixture = `{"writable":["4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg"],"readonly":["11111111111111111111111111111111"]}`

func TestLoadedAddressesJSON(t *testing.T) {
	loaded := jsonRoundTrip[LoadedAddresses](t, []byte(loadedAddressesFixture))

	if len(loaded.Writable) != 1 || loaded.Writable[0].String() != "4ejjNYBbaETZyqaiK8aDj2BWER8LKHgDcCnRrPC22YGg" {
		t.Fatalf("Writable = %v", loaded.Writable)
	}
	if len(loaded.ReadOnly) != 1 || loaded.ReadOnly[0].String() != "11111111111111111111111111111111" {
		t.Fatalf("ReadOnly = %v", loaded.ReadOnly)
	}
}

func TestLoadedAddressesEmpty(t *testing.T) {
	// Legacy transactions report empty lookup lists.
	loaded := jsonRoundTrip[LoadedAddresses](t, []byte(`{"readonly":[],"writable":[]}`))
	if len(loaded.ReadOnly) != 0 || len(loaded.Writable) != 0 {
		t.Fatalf("got %+v", loaded)
	}
	if loaded.ReadOnly == nil || loaded.Writable == nil {
		t.Fatalf("empty arrays decoded as nil: %+v", loaded)
	}
}
