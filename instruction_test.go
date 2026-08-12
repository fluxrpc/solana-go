package solana_go

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestGenericInstruction(t *testing.T) {
	program := PublicKey{9}
	accounts := AccountMetaSlice{
		NewAccountMeta(testPublicKey(), true, true),
		NewAccountMeta(PublicKey{2}, false, false),
	}
	data := []byte{1, 2, 3}

	in := NewInstruction(program, accounts, data)
	if in.ProgramID() != program {
		t.Fatalf("ProgramID() = %v", in.ProgramID())
	}
	if got := in.Accounts(); len(got) != 2 || got[0] != accounts[0] {
		t.Fatalf("Accounts() = %v", got)
	}
	got, err := in.Data()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("Data() = %x, want %x", got, data)
	}
}

func TestCompiledInstructionJSONRoundTrip(t *testing.T) {
	want := CompiledInstruction{
		ProgramIDIndex: 4,
		Accounts:       []uint16{1, 2, 3},
		Data:           Base58{0xde, 0xad, 0xbe, 0xef},
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"programIdIndex":4,"accounts":[1,2,3],"data":"` + want.Data.String() + `"}`
	if string(data) != expected {
		t.Fatalf("MarshalJSON() = %s, want %s", data, expected)
	}

	var got CompiledInstruction
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON round trip mismatch: got %+v, want %+v", got, want)
	}
}

var benchmarkInstructionJSON []byte

func BenchmarkCompiledInstructionMarshalJSON(b *testing.B) {
	in := CompiledInstruction{
		ProgramIDIndex: 4,
		Accounts:       []uint16{1, 2, 3, 4, 5},
		Data:           Base58(testPayload()),
	}
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkInstructionJSON, err = json.Marshal(in)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompiledInstructionUnmarshalJSON(b *testing.B) {
	in := CompiledInstruction{
		ProgramIDIndex: 4,
		Accounts:       []uint16{1, 2, 3, 4, 5},
		Data:           Base58(testPayload()),
	}
	data, err := json.Marshal(in)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var out CompiledInstruction
		if err := json.Unmarshal(data, &out); err != nil {
			b.Fatal(err)
		}
	}
}
