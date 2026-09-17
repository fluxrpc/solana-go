package solana_go

import "testing"

var sinkLookupKeys PublicKeySlice

func benchmarkLookupMessage() *Message {
	message := &Message{AccountKeys: []PublicKey{tableKey(1), tableKey(2), tableKey(3)}}
	tables := make(map[PublicKey]PublicKeySlice)
	for table := byte(0); table < 4; table++ {
		key := PublicKey{0xf0, table}
		addresses := make(PublicKeySlice, 256)
		for i := range addresses {
			addresses[i] = PublicKey{table + 1, byte(i)}
		}
		tables[key] = addresses
		message.AddAddressTableLookup(MessageAddressTableLookup{
			AccountKey:      key,
			WritableIndexes: []uint8{0, 1, 2, 3, 4, 5, 6, 7},
			ReadonlyIndexes: []uint8{8, 9, 10, 11, 12, 13, 14, 15},
		})
	}
	if err := message.SetAddressTables(tables); err != nil {
		panic(err)
	}
	return message
}

func BenchmarkAddressTableResolution(b *testing.B) {
	message := benchmarkLookupMessage()
	for _, operation := range []string{"LookupAccounts", "GetAllKeys", "ResolveLookups"} {
		b.Run(operation, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				var err error
				switch operation {
				case "LookupAccounts":
					sinkLookupKeys, err = message.GetAddressTableLookupAccounts()
				case "GetAllKeys":
					sinkLookupKeys, err = message.GetAllKeys()
				case "ResolveLookups":
					copyMessage := *message
					err = copyMessage.ResolveLookups()
					sinkLookupKeys = copyMessage.AccountKeys
				}
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func TestAddressTableResolutionAllocations(t *testing.T) {
	message := benchmarkLookupMessage()
	for _, operation := range []string{"LookupAccounts", "GetAllKeys", "ResolveLookups"} {
		t.Run(operation, func(t *testing.T) {
			allocs := testing.AllocsPerRun(100, func() {
				var err error
				switch operation {
				case "LookupAccounts":
					sinkLookupKeys, err = message.GetAddressTableLookupAccounts()
				case "GetAllKeys":
					sinkLookupKeys, err = message.GetAllKeys()
				case "ResolveLookups":
					copyMessage := *message
					err = copyMessage.ResolveLookups()
					sinkLookupKeys = copyMessage.AccountKeys
				}
				if err != nil {
					t.Fatal(err)
				}
			})
			if allocs > 1 {
				t.Fatalf("got %g allocations, want at most 1", allocs)
			}
		})
	}
}
