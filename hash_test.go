package solana_go

import (
	"encoding/json"
	"strings"
	"testing"
)

func testHash() Hash {
	var h Hash
	for i := range h {
		h[i] = byte(i + 1)
	}
	return h
}

func TestHashRoundTrip(t *testing.T) {
	tests := []Hash{
		{},
		testHash(),
		HashFromBytes(bytesOf(0xff)),
	}

	for _, want := range tests {
		encoded := want.String()
		got, err := HashFromBase58(encoded)
		if err != nil {
			t.Fatalf("HashFromBase58(%q): %v", encoded, err)
		}
		if got != want {
			t.Fatalf("round trip mismatch: got %x, want %x", got, want)
		}
	}
}

func TestHashFromBase58RejectsInvalidInput(t *testing.T) {
	tests := []string{
		"",
		"abc",
		strings.Repeat("0", 32),
		strings.Repeat("z", 45),
	}

	for _, input := range tests {
		got, err := HashFromBase58(input)
		if err == nil {
			t.Errorf("HashFromBase58(%q) unexpectedly succeeded", input)
		}
		if !got.IsZero() {
			t.Errorf("HashFromBase58(%q) returned data on error: %x", input, got)
		}
	}
}

func TestMustHashFromBase58Panics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustHashFromBase58 accepted invalid input")
		}
	}()
	MustHashFromBase58("not-base58")
}

func TestHashJSONRoundTrip(t *testing.T) {
	want := testHash()
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"`+want.String()+`"` {
		t.Fatalf("MarshalJSON() = %s, want quoted %q", data, want.String())
	}

	var got Hash
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("JSON round trip mismatch: got %x, want %x", got, want)
	}
}

func TestHashUnmarshalJSONPreservesValueOnError(t *testing.T) {
	want := testHash()
	got := want
	if err := json.Unmarshal([]byte(`"invalid"`), &got); err == nil {
		t.Fatal("json.Unmarshal unexpectedly succeeded")
	}
	if got != want {
		t.Fatalf("json.Unmarshal changed receiver to %x", got)
	}
}

func TestHashZero(t *testing.T) {
	var zero Hash
	if !zero.IsZero() {
		t.Fatal("zero Hash is not zero")
	}
	if testHash().IsZero() {
		t.Fatal("non-zero Hash is zero")
	}
	if !zero.Equals(Hash{}) || zero.Equals(testHash()) {
		t.Fatal("Equals mismatch")
	}
}

var (
	benchmarkHash       = testHash()
	benchmarkHashString string
)

func BenchmarkHashString(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkHashString = benchmarkHash.String()
	}
}

func BenchmarkHashFromBase58(b *testing.B) {
	encoded := benchmarkHash.String()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := HashFromBase58(encoded); err != nil {
			b.Fatal(err)
		}
	}
}
