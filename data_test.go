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

func TestDataZstdKnownAnswer(t *testing.T) {
	// Fixture from upstream gagliardetto/solana-go rpc/types_test.go:
	// zstd-compressed, base64-encoded "hello-world".
	var got Data
	if err := json.Unmarshal([]byte(`["KLUv/QQAWQAAaGVsbG8td29ybGTcLcaB","base64+zstd"]`), &got); err != nil {
		t.Fatal(err)
	}
	if string(got.Content) != "hello-world" || got.Encoding != EncodingBase64Zstd {
		t.Fatalf("got %q (%s)", got.Content, got.Encoding)
	}

	// And content survives a full marshal/unmarshal cycle through our own
	// compressor.
	back := jsonRoundTripData(t, got)
	if string(back.Content) != "hello-world" {
		t.Fatalf("round trip = %q", back.Content)
	}
}

func jsonRoundTripData(t *testing.T, want Data) Data {
	t.Helper()
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got Data
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestDataJSONRoundTrip(t *testing.T) {
	for _, encoding := range []EncodingType{EncodingBase58, EncodingBase64, EncodingBase64Zstd} {
		want := Data{Content: testPayload(), Encoding: encoding}

		data, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		if expected := `["` + want.String() + `","` + string(encoding) + `"]`; string(data) != expected {
			t.Fatalf("MarshalJSON() = %s, want %s", data, expected)
		}

		var got Data
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatal(err)
		}
		if got.Encoding != want.Encoding || !bytes.Equal(got.Content, want.Content) {
			t.Fatalf("round trip mismatch: got %+v, want %+v", got, want)
		}
	}
}

func TestDataUnmarshalJSONEmptyContent(t *testing.T) {
	var got Data
	if err := json.Unmarshal([]byte(`["","base64"]`), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Content) != 0 || got.Encoding != EncodingBase64 {
		t.Fatalf("got %+v", got)
	}
}

func TestDataMarshalJSONRejectsUnsupportedEncoding(t *testing.T) {
	// Non-empty content with an encoding String() cannot render must error
	// instead of silently emitting empty content.
	for _, encoding := range []EncodingType{EncodingJSONParsed, ""} {
		if _, err := json.Marshal(Data{Content: []byte{1}, Encoding: encoding}); err == nil {
			t.Errorf("Marshal with encoding %q unexpectedly succeeded", encoding)
		}
	}

	// Empty content is representable under the known encodings, and the
	// zero value round-trips as ["",""].
	if _, err := json.Marshal(Data{Encoding: EncodingBase64Zstd}); err != nil {
		t.Fatal(err)
	}
	zero, err := json.Marshal(Data{})
	if err != nil {
		t.Fatal(err)
	}
	if string(zero) != `["",""]` {
		t.Fatalf("zero value marshals as %s", zero)
	}
	var back Data
	if err := json.Unmarshal(zero, &back); err != nil {
		t.Fatal(err)
	}
}

func TestDataUnmarshalJSONRejectsInvalidInput(t *testing.T) {
	tests := []string{
		`["AQID"]`,                  // wrong tuple length
		`["AQID","base64","extra"]`, // wrong tuple length
		`["AQID","base64+zstd"]`,    // valid base64 but not a zstd frame
		`["AQID","what"]`,           // unknown encoding
		`["","what"]`,               // unknown encoding, even with empty content
		`["","evil\"],[\"x"]`,       // encoding must never be stored unvalidated
		`["!!!","base64"]`,          // bad base64
		`["0OIl","base58"]`,         // bad base58
		`{"content":"x"}`,           // not a tuple
	}

	for _, input := range tests {
		var got Data
		if err := json.Unmarshal([]byte(input), &got); err == nil {
			t.Errorf("json.Unmarshal(%s) unexpectedly succeeded", input)
		}
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
