package addresslookuptable

import (
	"encoding/binary"
	"errors"
	"math"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func buildLookupTable(authority *solana.PublicKey, addresses []solana.PublicKey) []byte {
	data := make([]byte, lookupTableMetaSize+len(addresses)*solana.PublicKeyLength)
	binary.LittleEndian.PutUint32(data[0:], 1)
	binary.LittleEndian.PutUint64(data[4:], math.MaxUint64)
	binary.LittleEndian.PutUint64(data[12:], 42)
	data[20] = 7
	if authority != nil {
		data[21] = 1
		copy(data[22:], authority[:])
	}
	for index, address := range addresses {
		copy(data[lookupTableMetaSize+index*solana.PublicKeyLength:], address[:])
	}
	return data
}

func TestDecodeLookupTable(t *testing.T) {
	authority := lookupKey(0xAA)
	addresses := []solana.PublicKey{lookupKey(0x01), lookupKey(0x02), lookupKey(0x03)}

	table, err := DecodeLookupTable(buildLookupTable(&authority, addresses))
	if err != nil {
		t.Fatal(err)
	}

	if table.TypeIndex != 1 {
		t.Errorf("TypeIndex = %d, want 1", table.TypeIndex)
	}
	if table.DeactivationSlot != math.MaxUint64 {
		t.Errorf("DeactivationSlot = %d, want %d", table.DeactivationSlot, uint64(math.MaxUint64))
	}
	if table.LastExtendedSlot != 42 {
		t.Errorf("LastExtendedSlot = %d, want 42", table.LastExtendedSlot)
	}
	if table.LastExtendedSlotStartIndex != 7 {
		t.Errorf("LastExtendedSlotStartIndex = %d, want 7", table.LastExtendedSlotStartIndex)
	}
	if table.Authority == nil || *table.Authority != authority {
		t.Errorf("Authority = %v, want %s", table.Authority, authority)
	}
	if len(table.Addresses) != len(addresses) {
		t.Fatalf("address count = %d, want %d", len(table.Addresses), len(addresses))
	}
	for index, want := range addresses {
		if table.Addresses[index] != want {
			t.Errorf("address %d = %s, want %s", index, table.Addresses[index], want)
		}
	}
}

func TestDecodeLookupTableFrozen(t *testing.T) {
	table, err := DecodeLookupTable(buildLookupTable(nil, []solana.PublicKey{lookupKey(0x09)}))
	if err != nil {
		t.Fatal(err)
	}
	if table.Authority != nil {
		t.Errorf("Authority = %s, want nil", table.Authority)
	}
}

func TestDecodeLookupTableEmpty(t *testing.T) {
	authority := lookupKey(0xBB)
	table, err := DecodeLookupTable(buildLookupTable(&authority, nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Addresses) != 0 {
		t.Errorf("address count = %d, want 0", len(table.Addresses))
	}
}

func TestDecodeLookupTableTrailingBytes(t *testing.T) {
	authority := lookupKey(0xCC)
	data := append(buildLookupTable(&authority, []solana.PublicKey{lookupKey(0x04)}), 0xFF, 0xFE)

	table, err := DecodeLookupTable(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Addresses) != 1 {
		t.Errorf("address count = %d, want 1", len(table.Addresses))
	}
}

func TestDecodeLookupTableTooShort(t *testing.T) {
	_, err := DecodeLookupTable(make([]byte, lookupTableMetaSize-1))
	if !errors.Is(err, ErrInvalidAccountSize) {
		t.Fatalf("err = %v, want %v", err, ErrInvalidAccountSize)
	}
}

func TestDecodeLookupTableTooManyAddresses(t *testing.T) {
	data := make([]byte, lookupTableMetaSize+(LookupTableMaxAddresses+1)*solana.PublicKeyLength)
	_, err := DecodeLookupTable(data)
	if !errors.Is(err, ErrTooManyAddresses) {
		t.Fatalf("err = %v, want %v", err, ErrTooManyAddresses)
	}
}
