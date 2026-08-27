package solana_go

import (
	"encoding"
	"encoding/json"
	"testing"

	"github.com/bytedance/sonic"
)

var (
	_ encoding.TextMarshaler   = Hash{}
	_ encoding.TextUnmarshaler = (*Hash)(nil)
	_ encoding.TextMarshaler   = Signature{}
	_ encoding.TextUnmarshaler = (*Signature)(nil)
)

func TestHashTextRoundTrip(t *testing.T) {
	want := testHash()
	text, err := want.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != want.String() {
		t.Fatalf("MarshalText() = %q, want %q", text, want.String())
	}

	var got Hash
	if err := got.UnmarshalText(text); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("text round trip mismatch: got %x, want %x", got, want)
	}
	testJSONTextMapKey(t, want)
}

func TestSignatureTextRoundTrip(t *testing.T) {
	want := testSignature()
	text, err := want.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != want.String() {
		t.Fatalf("MarshalText() = %q, want %q", text, want.String())
	}

	var got Signature
	if err := got.UnmarshalText(text); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("text round trip mismatch: got %x, want %x", got, want)
	}
	testJSONTextMapKey(t, want)
}

func TestTextUnmarshalPreservesValueOnError(t *testing.T) {
	hash := testHash()
	wantHash := hash
	if err := hash.UnmarshalText([]byte("invalid")); err == nil {
		t.Fatal("Hash.UnmarshalText unexpectedly succeeded")
	}
	if hash != wantHash {
		t.Fatalf("Hash.UnmarshalText changed receiver to %x", hash)
	}

	signature := testSignature()
	wantSignature := signature
	if err := signature.UnmarshalText([]byte("invalid")); err == nil {
		t.Fatal("Signature.UnmarshalText unexpectedly succeeded")
	}
	if signature != wantSignature {
		t.Fatalf("Signature.UnmarshalText changed receiver to %x", signature)
	}
}

func testJSONTextMapKey[K comparable](t *testing.T, key K) {
	t.Helper()
	for _, codec := range []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{name: "encoding/json", marshal: json.Marshal, unmarshal: json.Unmarshal},
		{name: "sonic", marshal: sonic.Marshal, unmarshal: sonic.Unmarshal},
	} {
		t.Run(codec.name, func(t *testing.T) {
			data, err := codec.marshal(map[K]uint64{key: 42})
			if err != nil {
				t.Fatal(err)
			}
			var got map[K]uint64
			if err := codec.unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || got[key] != 42 {
				t.Fatalf("JSON map round trip = %v", got)
			}
		})
	}
}
