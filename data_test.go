package solana_go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/fluxrpc/base58"
)

func testPayload() []byte {
	out := make([]byte, 96)
	for i := range out {
		out[i] = byte(i * 7)
	}
	return out
}

func TestBase58JSONRoundTrip(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		{0, 0, 0},
		testPayload(),
	}

	for _, payload := range tests {
		want := Base58(payload)
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+base58.Encode(payload)+`"` {
			t.Fatalf("MarshalJSON() = %s, want quoted %q", data, base58.Encode(payload))
		}

		var got Base58
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("JSON round trip mismatch: got %x, want %x", got, want)
		}
	}
}

func TestBase58UnmarshalJSONRejectsInvalidInput(t *testing.T) {
	for _, data := range []string{`"0OIl"`, `{}`, `123`} {
		var got Base58
		if err := json.Unmarshal([]byte(data), &got); err == nil {
			t.Errorf("json.Unmarshal(%s) unexpectedly succeeded", data)
		}
	}
}

func TestBase58String(t *testing.T) {
	payload := testPayload()
	if got, want := Base58(payload).String(), base58.Encode(payload); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestBase64JSONRoundTrip(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		{0, 0, 0},
		testPayload(),
	}

	for _, payload := range tests {
		want := Base64(payload)
		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `"`+base64.StdEncoding.EncodeToString(payload)+`"` {
			t.Fatalf("MarshalJSON() = %s, want base64 of %x", data, payload)
		}

		var got Base64
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("JSON round trip mismatch: got %x, want %x", got, want)
		}
	}
}

func TestBase64UnmarshalJSONRejectsInvalidInput(t *testing.T) {
	for _, data := range []string{`"!!!"`, `{}`, `123`} {
		var got Base64
		if err := json.Unmarshal([]byte(data), &got); err == nil {
			t.Errorf("json.Unmarshal(%s) unexpectedly succeeded", data)
		}
	}
}

func TestBase64String(t *testing.T) {
	payload := testPayload()
	if got, want := Base64(payload).String(), base64.StdEncoding.EncodeToString(payload); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

var (
	benchmarkPayload  = Base58(testPayload())
	benchmarkDataJSON []byte
)

func BenchmarkBase58MarshalJSON(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkDataJSON, err = benchmarkPayload.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBase58UnmarshalJSON(b *testing.B) {
	data, err := benchmarkPayload.MarshalJSON()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var out Base58
		if err := out.UnmarshalJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBase64MarshalJSON(b *testing.B) {
	payload := Base64(testPayload())
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkDataJSON, err = payload.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBase64UnmarshalJSON(b *testing.B) {
	data, err := Base64(testPayload()).MarshalJSON()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var out Base64
		if err := out.UnmarshalJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}
