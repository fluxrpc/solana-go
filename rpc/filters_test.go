package rpc

import (
	"encoding/json"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestRPCFilterMarshalShape(t *testing.T) {
	memcmp := RPCFilter{
		Memcmp: &RPCFilterMemcmp{
			Offset: 4,
			Bytes:  solana.Base58{1, 2, 3, 4},
		},
	}
	data, err := json.Marshal(memcmp)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"memcmp":{"offset":4,"bytes":"2VfUX"}}`; string(data) != want {
		t.Fatalf("MarshalJSON() = %s, want %s", data, want)
	}

	dataSize := RPCFilter{DataSize: 165}
	data, err = json.Marshal(dataSize)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"dataSize":165}`; string(data) != want {
		t.Fatalf("MarshalJSON() = %s, want %s", data, want)
	}
}

func TestDataSliceOmitEmpty(t *testing.T) {
	data, err := json.Marshal(DataSlice{})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{}` {
		t.Fatalf("MarshalJSON() = %s, want {}", data)
	}

	offset, length := uint64(129), uint64(64)
	got := jsonRoundTrip[DataSlice](t, []byte(`{"offset":129,"length":64}`))
	if *got.Offset != offset || *got.Length != length {
		t.Fatalf("got %+v", got)
	}
}
