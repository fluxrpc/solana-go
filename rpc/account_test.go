package rpc

import (
	"encoding/json"
	"github.com/bytedance/sonic"
	"math/big"
	"testing"
)

// accountFixture mirrors a real mainnet getAccountInfo value (shape from
// upstream gagliardetto/solana-go rpc/client_test.go). rentEpoch is the
// u64-max sentinel used since rent collection was disabled, which does not
// fit in an int64 — exercising the *big.Int field.
var accountFixture = []byte(`{"lamports":88849814690250,"owner":"11111111111111111111111111111111","data":["dGVzdCBkYXRh","base64"],"executable":false,"rentEpoch":18446744073709551615,"space":9}`)

func TestAccountJSONRoundTrip(t *testing.T) {
	got := jsonRoundTrip[Account](t, accountFixture)

	if got.Lamports != 88849814690250 || got.Space != 9 || got.Executable {
		t.Fatalf("got %+v", got)
	}
	if !got.Owner.IsZero() {
		t.Fatalf("Owner = %s, want system program", got.Owner)
	}
	if string(got.Data.GetBinary()) != "test data" {
		t.Fatalf("Data = %q", got.Data.GetBinary())
	}
	if want := new(big.Int).SetUint64(1<<64 - 1); got.RentEpoch.Cmp(want) != 0 {
		t.Fatalf("RentEpoch = %s, want %s", got.RentEpoch, want)
	}
}

func TestKeyedAccountJSONRoundTrip(t *testing.T) {
	fixture := []byte(`{"pubkey":"SysvarC1ock11111111111111111111111111111111","account":` + string(accountFixture) + `}`)
	got := jsonRoundTrip[KeyedAccount](t, fixture)
	if got.Pubkey.String() != "SysvarC1ock11111111111111111111111111111111" {
		t.Fatalf("Pubkey = %s", got.Pubkey)
	}
	if got.Account == nil || got.Account.Lamports != 88849814690250 {
		t.Fatalf("Account = %+v", got.Account)
	}
}

var (
	benchmarkAccount     Account
	benchmarkAccountJSON []byte
)

func BenchmarkAccountUnmarshalJSON(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if err := sonic.Unmarshal(accountFixture, &benchmarkAccount); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAccountMarshalJSON(b *testing.B) {
	if err := sonic.Unmarshal(accountFixture, &benchmarkAccount); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkAccountJSON, err = json.Marshal(&benchmarkAccount)
		if err != nil {
			b.Fatal(err)
		}
	}
}
