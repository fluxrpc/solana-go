package solana_go

import (
	"errors"
	"testing"
)

func tableKey(fill byte) (key PublicKey) {
	for index := range key {
		key[index] = fill
	}
	return key
}

// lookupMessage uses two lookups so the ordering spans more than one table.
func lookupMessage() (*Message, map[PublicKey]PublicKeySlice) {
	first, second := tableKey(0xA1), tableKey(0xA2)
	message := &Message{
		AccountKeys: []PublicKey{tableKey(0x01), tableKey(0x02)},
		AddressTableLookups: MessageAddressTableLookupSlice{
			{AccountKey: first, WritableIndexes: []uint8{1, 0}, ReadonlyIndexes: []uint8{2}},
			{AccountKey: second, WritableIndexes: []uint8{0}, ReadonlyIndexes: []uint8{2, 1}},
		},
	}
	tables := map[PublicKey]PublicKeySlice{
		first:  {tableKey(0x10), tableKey(0x11), tableKey(0x12)},
		second: {tableKey(0x20), tableKey(0x21), tableKey(0x22)},
	}
	return message, tables
}

func TestGetAddressTableLookupAccountsOrder(t *testing.T) {
	message, tables := lookupMessage()
	if err := message.SetAddressTables(tables); err != nil {
		t.Fatal(err)
	}

	got, err := message.GetAddressTableLookupAccounts()
	if err != nil {
		t.Fatal(err)
	}
	want := PublicKeySlice{
		tableKey(0x11), tableKey(0x10), tableKey(0x20),
		tableKey(0x12), tableKey(0x22), tableKey(0x21),
	}
	if len(got) != len(want) {
		t.Fatalf("account count = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("account %d = %s, want %s", index, got[index], want[index])
		}
	}
}

func TestSetAddressTablesTwice(t *testing.T) {
	message, tables := lookupMessage()
	if err := message.SetAddressTables(tables); err != nil {
		t.Fatal(err)
	}
	if err := message.SetAddressTables(tables); !errors.Is(err, ErrAddressTablesAlreadySet) {
		t.Fatalf("err = %v, want %v", err, ErrAddressTablesAlreadySet)
	}
}

func TestAddressTableLookupErrors(t *testing.T) {
	t.Run("TablesNotSet", func(t *testing.T) {
		message, _ := lookupMessage()
		if _, err := message.GetAddressTableLookupAccounts(); !errors.Is(err, ErrAddressTablesNotSet) {
			t.Fatalf("err = %v, want %v", err, ErrAddressTablesNotSet)
		}
	})

	t.Run("TableNotFound", func(t *testing.T) {
		message, tables := lookupMessage()
		delete(tables, message.AddressTableLookups[1].AccountKey)
		if err := message.SetAddressTables(tables); err != nil {
			t.Fatal(err)
		}
		if _, err := message.GetAddressTableLookupAccounts(); !errors.Is(err, ErrAddressTableNotFound) {
			t.Fatalf("err = %v, want %v", err, ErrAddressTableNotFound)
		}
	})

	t.Run("IndexOutOfRange", func(t *testing.T) {
		message, tables := lookupMessage()
		tables[message.AddressTableLookups[0].AccountKey] = PublicKeySlice{tableKey(0x10)}
		if err := message.SetAddressTables(tables); err != nil {
			t.Fatal(err)
		}
		if _, err := message.GetAddressTableLookupAccounts(); !errors.Is(err, ErrAddressTableIndexRange) {
			t.Fatalf("err = %v, want %v", err, ErrAddressTableIndexRange)
		}
	})
}

func TestResolveLookups(t *testing.T) {
	message, tables := lookupMessage()
	if err := message.SetAddressTables(tables); err != nil {
		t.Fatal(err)
	}
	static := len(message.AccountKeys)

	if message.IsResolved() {
		t.Fatal("IsResolved() = true before resolving")
	}
	if err := message.ResolveLookups(); err != nil {
		t.Fatal(err)
	}
	if !message.IsResolved() {
		t.Error("IsResolved() = false after resolving")
	}
	if len(message.AccountKeys) != static+6 {
		t.Fatalf("AccountKeys = %d, want %d", len(message.AccountKeys), static+6)
	}

	if err := message.ResolveLookups(); err != nil {
		t.Fatal(err)
	}
	if len(message.AccountKeys) != static+6 {
		t.Errorf("AccountKeys after second resolve = %d, want %d", len(message.AccountKeys), static+6)
	}
}

func TestGetAllKeys(t *testing.T) {
	message, tables := lookupMessage()
	if err := message.SetAddressTables(tables); err != nil {
		t.Fatal(err)
	}

	before, err := message.GetAllKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != 8 {
		t.Fatalf("key count = %d, want 8", len(before))
	}
	if len(message.AccountKeys) != 2 {
		t.Errorf("AccountKeys = %d, want 2 (GetAllKeys mutated the message)", len(message.AccountKeys))
	}

	if err := message.ResolveLookups(); err != nil {
		t.Fatal(err)
	}
	after, err := message.GetAllKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("key count after resolve = %d, want %d", len(after), len(before))
	}
	for index := range before {
		if before[index] != after[index] {
			t.Errorf("key %d changed across resolve: %s vs %s", index, before[index], after[index])
		}
	}
}

func TestGetAllKeysWithoutLookups(t *testing.T) {
	message := &Message{AccountKeys: []PublicKey{tableKey(0x01), tableKey(0x02)}}
	keys, err := message.GetAllKeys()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("key count = %d, want 2", len(keys))
	}
}
