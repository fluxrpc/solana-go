package solana_go

import (
	"bytes"
	"errors"
	"slices"
	"testing"
)

func TestResolveLookupsPreservesWireAndSignatures(t *testing.T) {
	message, tables := lookupMessage()
	message.SetAddressTableLookups(message.AddressTableLookups)
	message.Header = MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1}
	key, err := PrivateKeyFromSeed(make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}
	message.AccountKeys[0] = key.PublicKey()
	if err := message.SetAddressTables(tables); err != nil {
		t.Fatal(err)
	}
	tx := Transaction{Message: *message}
	if _, err := tx.Sign(func(PublicKey) *PrivateKey { return &key }); err != nil {
		t.Fatal(err)
	}
	if err := tx.VerifySignatures(); err != nil {
		t.Fatal(err)
	}
	beforeBinary, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	beforeJSON, err := tx.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Message.ResolveLookups(); err != nil {
		t.Fatal(err)
	}
	afterBinary, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	afterJSON, err := tx.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(beforeBinary, afterBinary) {
		t.Error("resolution changed transaction wire bytes")
	}
	if !bytes.Equal(beforeJSON, afterJSON) {
		t.Error("resolution changed transaction JSON")
	}
	if err := tx.VerifySignatures(); err != nil {
		t.Fatalf("signature after resolution: %v", err)
	}
}

func TestResolvedAccountRoles(t *testing.T) {
	message, tables := lookupMessage()
	message.SetAddressTableLookups(message.AddressTableLookups)
	message.AccountKeys = []PublicKey{tableKey(1), tableKey(2), tableKey(3), tableKey(4)}
	message.Header = MessageHeader{NumRequiredSignatures: 2, NumReadonlySignedAccounts: 1, NumReadonlyUnsignedAccounts: 1}
	if err := message.SetAddressTables(tables); err != nil {
		t.Fatal(err)
	}
	if err := message.ResolveLookups(); err != nil {
		t.Fatal(err)
	}
	metas, err := message.AccountMetaList()
	if err != nil {
		t.Fatal(err)
	}
	writable := []bool{true, false, true, false, true, true, true, false, false, false}
	if len(metas) != len(writable) {
		t.Fatalf("got %d metas, want %d", len(metas), len(writable))
	}
	for i, meta := range metas {
		if meta.PublicKey != message.AccountKeys[i] || meta.IsSigner != (i < 2) || meta.IsWritable != writable[i] {
			t.Errorf("meta %d has wrong key or roles: %+v", i, meta)
		}
		wantStaticWritable := i < 4 && writable[i]
		if got := message.IsWritableStatic(meta.PublicKey); got != wantStaticWritable {
			t.Errorf("IsWritableStatic(%d) = %v, want %v", i, got, wantStaticWritable)
		}
		if got := message.IsSigner(meta.PublicKey); got != (i < 2) {
			t.Errorf("IsSigner(%d) = %v", i, got)
		}
		if got, err := message.Account(uint16(i)); err != nil || got != meta.PublicKey {
			t.Errorf("Account(%d) = %v, %v", i, got, err)
		}
	}
	if got := message.Signers(); !slices.Equal(got, message.AccountKeys[:2]) {
		t.Errorf("wrong signers: %v", got)
	}
}

func TestDecodeClearsAddressTableState(t *testing.T) {
	for _, format := range []string{"binary", "JSON"} {
		for _, legacy := range []bool{false, true} {
			name := format + "/v0"
			if legacy {
				name = format + "/legacy"
			}
			t.Run(name, func(t *testing.T) {
				message, tables := lookupMessage()
				if err := message.SetAddressTables(tables); err != nil {
					t.Fatal(err)
				}
				if err := message.ResolveLookups(); err != nil {
					t.Fatal(err)
				}
				fresh, freshTables := lookupMessage()
				fresh.SetAddressTableLookups(fresh.AddressTableLookups)
				for _, table := range freshTables {
					for i := range table {
						table[i][0] = 0xff
					}
				}
				if legacy {
					fresh = &Message{AccountKeys: []PublicKey{tableKey(9)}}
				}
				var data []byte
				var err error
				if format == "binary" {
					data, err = fresh.MarshalBinary()
				} else {
					data, err = fresh.MarshalJSON()
				}
				if err != nil {
					t.Fatal(err)
				}
				if format == "binary" {
					err = message.UnmarshalBinary(data)
				} else {
					err = message.UnmarshalJSON(data)
				}
				if err != nil {
					t.Fatal(err)
				}
				if message.IsResolved() || message.GetAddressTables() != nil || message.numStaticAccounts != 0 {
					t.Fatal("decoder retained address table state")
				}
				if err := message.SetAddressTables(freshTables); err != nil {
					t.Fatal(err)
				}
				if err := fresh.SetAddressTables(freshTables); err != nil {
					t.Fatal(err)
				}
				want, err := fresh.GetAllKeys()
				if err != nil {
					t.Fatal(err)
				}
				if err := message.ResolveLookups(); err != nil {
					t.Fatal(err)
				}
				if !slices.Equal(message.AccountKeys, want) {
					t.Fatal("resolution used stale keys")
				}
			})
		}
	}
}

func TestLookupSettersInvalidateResolution(t *testing.T) {
	for _, add := range []bool{false, true} {
		name := "Set"
		if add {
			name = "Add"
		}
		t.Run(name, func(t *testing.T) {
			message, tables := lookupMessage()
			if err := message.SetAddressTables(tables); err != nil {
				t.Fatal(err)
			}
			if err := message.ResolveLookups(); err != nil {
				t.Fatal(err)
			}
			lookup := MessageAddressTableLookup{AccountKey: tableKey(0xa1), ReadonlyIndexes: []uint8{0}}
			if add {
				message.AddAddressTableLookup(lookup)
			} else {
				message.SetAddressTableLookups([]MessageAddressTableLookup{lookup})
			}
			if message.IsResolved() || len(message.AccountKeys) != 2 {
				t.Fatal("setter retained loaded accounts")
			}
			want, err := message.GetAllKeys()
			if err != nil {
				t.Fatal(err)
			}
			if err := message.ResolveLookups(); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(message.AccountKeys, want) {
				t.Fatal("resolution did not use new lookups")
			}
			if add && len(want) != 9 || !add && len(want) != 3 {
				t.Fatalf("wrong key count: %d", len(want))
			}
		})
	}
}

func TestResolveLookupsErrorDoesNotMutateMessage(t *testing.T) {
	for _, missing := range []bool{false, true} {
		name := "ReadonlyIndex"
		if missing {
			name = "MissingTable"
		}
		t.Run(name, func(t *testing.T) {
			message, tables := lookupMessage()
			second := message.AddressTableLookups[1].AccountKey
			saved := tables[second]
			wantErr := ErrAddressTableIndexRange
			if missing {
				delete(tables, second)
				wantErr = ErrAddressTableNotFound
			} else {
				tables[second] = saved[:1]
			}
			if err := message.SetAddressTables(tables); err != nil {
				t.Fatal(err)
			}
			before := slices.Clone(message.AccountKeys)
			if err := message.ResolveLookups(); !errors.Is(err, wantErr) {
				t.Fatalf("got %v, want %v", err, wantErr)
			}
			if message.IsResolved() || !slices.Equal(message.AccountKeys, before) {
				t.Fatal("failed resolution mutated message")
			}
			tables[second] = saved
			if err := message.ResolveLookups(); err != nil {
				t.Fatal(err)
			}
			if len(message.AccountKeys) != 8 {
				t.Fatal("retry did not resolve accounts")
			}
		})
	}
}

func TestGetAllKeysOwnsUnresolvedResult(t *testing.T) {
	for _, lookups := range []bool{false, true} {
		message, tables := lookupMessage()
		if !lookups {
			message.AddressTableLookups = nil
		}
		if err := message.SetAddressTables(tables); err != nil {
			t.Fatal(err)
		}
		keys, err := message.GetAllKeys()
		if err != nil {
			t.Fatal(err)
		}
		keys[0] = tableKey(0xee)
		if message.AccountKeys[0] == keys[0] {
			t.Fatal("unresolved result aliases static keys")
		}
		if lookups {
			keys[2] = tableKey(0xff)
			if tables[tableKey(0xa1)][1] == keys[2] {
				t.Fatal("result aliases table contents")
			}
		}
	}
}

func TestLookupAccountRegions(t *testing.T) {
	for _, kind := range []string{"empty", "writable", "readonly", "mixed"} {
		t.Run(kind, func(t *testing.T) {
			message, tables := lookupMessage()
			var want PublicKeySlice
			switch kind {
			case "empty":
				message.AddressTableLookups = nil
			case "writable":
				for i := range message.AddressTableLookups {
					message.AddressTableLookups[i].ReadonlyIndexes = nil
				}
				want = PublicKeySlice{tableKey(0x11), tableKey(0x10), tableKey(0x20)}
			case "readonly":
				for i := range message.AddressTableLookups {
					message.AddressTableLookups[i].WritableIndexes = nil
				}
				want = PublicKeySlice{tableKey(0x12), tableKey(0x22), tableKey(0x21)}
			case "mixed":
				want = PublicKeySlice{tableKey(0x11), tableKey(0x10), tableKey(0x20), tableKey(0x12), tableKey(0x22), tableKey(0x21)}
			}
			if err := message.SetAddressTables(tables); err != nil {
				t.Fatal(err)
			}
			got, err := message.GetAddressTableLookupAccounts()
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, want) {
				t.Fatalf("got %v, want %v", got, want)
			}
			all, err := message.GetAllKeys()
			if err != nil {
				t.Fatal(err)
			}
			expected := append(slices.Clone(message.AccountKeys), want...)
			if !slices.Equal(all, expected) {
				t.Fatal("wrong full account order")
			}
			if err := message.ResolveLookups(); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(message.AccountKeys, expected) {
				t.Fatal("wrong resolved account order")
			}
		})
	}
}
