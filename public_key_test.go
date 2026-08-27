package solana_go

import (
	"encoding"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/bytedance/sonic"
)

var (
	_ encoding.TextMarshaler   = PublicKey{}
	_ encoding.TextUnmarshaler = (*PublicKey)(nil)
)

func testPublicKey() PublicKey {
	var key PublicKey
	for i := range key {
		key[i] = byte(i + 1)
	}
	return key
}

func TestPublicKeyRoundTrip(t *testing.T) {
	tests := []PublicKey{
		{},
		testPublicKey(),
		PublicKeyFromBytes(bytesOf(0xff)),
	}

	for _, want := range tests {
		encoded := want.String()
		got, err := PublicKeyFromBase58(encoded)
		if err != nil {
			t.Fatalf("PublicKeyFromBase58(%q): %v", encoded, err)
		}
		if got != want {
			t.Fatalf("round trip mismatch: got %x, want %x", got, want)
		}
	}
}

func TestPublicKeyFromBytesCopiesInput(t *testing.T) {
	in := bytesOf(1)
	key := PublicKeyFromBytes(in)
	in[0] = 2
	if key[0] != 1 {
		t.Fatal("PublicKeyFromBytes retained the input slice")
	}

	out := key.Bytes()
	out[0] = 3
	if key[0] != 1 {
		t.Fatal("Bytes exposed mutable PublicKey storage")
	}
}

func TestPublicKeyFromBytesPanicsOnInvalidLength(t *testing.T) {
	for _, size := range []int{0, PublicKeyLength - 1, PublicKeyLength + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("PublicKeyFromBytes accepted %d bytes", size)
				}
			}()
			PublicKeyFromBytes(make([]byte, size))
		})
	}
}

func TestPublicKeyFromBase58RejectsInvalidInput(t *testing.T) {
	tests := []string{
		"",
		strings.Repeat("1", publicKeyMinEncodedLength-1),
		strings.Repeat("1", base58EncodedMaxLen32ForTest+1),
		strings.Repeat("0", publicKeyMinEncodedLength),
		strings.Repeat("2", publicKeyMinEncodedLength),
		strings.Repeat("z", base58EncodedMaxLen32ForTest),
	}

	for _, input := range tests {
		got, err := PublicKeyFromBase58(input)
		if err == nil {
			t.Errorf("PublicKeyFromBase58(%q) unexpectedly succeeded", input)
		}
		if !got.IsZero() {
			t.Errorf("PublicKeyFromBase58(%q) returned data on error: %x", input, got)
		}
	}
}

func TestPublicKeyJSONRoundTrip(t *testing.T) {
	want := testPublicKey()
	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"`+want.String()+`"` {
		t.Fatalf("MarshalJSON() = %s, want quoted %q", data, want.String())
	}

	var got PublicKey
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("JSON round trip mismatch: got %x, want %x", got, want)
	}
}

func TestPublicKeyTextRoundTrip(t *testing.T) {
	tests := []PublicKey{
		{},
		testPublicKey(),
		PublicKeyFromBytes(bytesOf(0xff)),
	}

	for _, want := range tests {
		data, err := want.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != want.String() {
			t.Fatalf("MarshalText() = %q, want %q", data, want.String())
		}

		var got PublicKey
		if err := got.UnmarshalText(data); err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("text round trip mismatch: got %x, want %x", got, want)
		}
	}
}

func TestPublicKeyJSONMapRoundTrip(t *testing.T) {
	want := map[PublicKey][]uint64{
		testPublicKey(): {1, 2, 3},
	}
	tests := []struct {
		name      string
		marshal   func(any) ([]byte, error)
		unmarshal func([]byte, any) error
	}{
		{name: "encoding/json", marshal: json.Marshal, unmarshal: json.Unmarshal},
		{name: "sonic", marshal: sonic.Marshal, unmarshal: sonic.Unmarshal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.marshal(want)
			if err != nil {
				t.Fatal(err)
			}

			var got map[PublicKey][]uint64
			if err := tt.unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 {
				t.Fatalf("decoded map has %d entries, want 1", len(got))
			}
			for key, slots := range want {
				gotSlots, ok := got[key]
				if !ok || len(gotSlots) != len(slots) {
					t.Fatalf("decoded map entry = %v, %v; want %v, true", gotSlots, ok, slots)
				}
				for i := range slots {
					if gotSlots[i] != slots[i] {
						t.Fatalf("decoded slots = %v, want %v", gotSlots, slots)
					}
				}
			}
		})
	}
}

func TestPublicKeyUnmarshalJSONSupportsEscapes(t *testing.T) {
	want := PublicKey{}
	data := []byte(`"\u0031` + strings.Repeat("1", PublicKeyLength-1) + `"`)

	var got PublicKey
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("UnmarshalJSON() = %x, want %x", got, want)
	}
}

func TestPublicKeyUnmarshalJSONPreservesValueOnError(t *testing.T) {
	want := testPublicKey()
	tests := [][]byte{
		[]byte(`"invalid"`),
		[]byte(`{"not":"a string"}`),
		[]byte(`"unterminated`),
	}

	for _, data := range tests {
		got := want
		if err := json.Unmarshal(data, &got); err == nil {
			t.Errorf("json.Unmarshal(%s) unexpectedly succeeded", data)
		}
		if got != want {
			t.Errorf("json.Unmarshal(%s) changed receiver to %x", data, got)
		}
	}
}

func TestPublicKeyUnmarshalTextPreservesValueOnError(t *testing.T) {
	want := testPublicKey()
	tests := [][]byte{
		nil,
		[]byte("invalid"),
		[]byte(strings.Repeat("0", PublicKeyLength)),
	}

	for _, data := range tests {
		got := want
		if err := got.UnmarshalText(data); err == nil {
			t.Errorf("UnmarshalText(%q) unexpectedly succeeded", data)
		}
		if got != want {
			t.Errorf("UnmarshalText(%q) changed receiver to %x", data, got)
		}
	}
}

func TestPublicKeyZero(t *testing.T) {
	var zero PublicKey
	if !zero.IsZero() {
		t.Fatal("zero PublicKey is not zero")
	}
	if got, want := zero.String(), strings.Repeat("1", PublicKeyLength); got != want {
		t.Fatalf("zero PublicKey String() = %q, want %q", got, want)
	}
	if testPublicKey().IsZero() {
		t.Fatal("non-zero PublicKey is zero")
	}
}

func bytesOf(value byte) []byte {
	out := make([]byte, PublicKeyLength)
	for i := range out {
		out[i] = value
	}
	return out
}

// Keep this local so the tests independently exercise the package's private
// input bound without importing an implementation dependency.
const base58EncodedMaxLen32ForTest = 44

var (
	benchmarkPublicKey       = testPublicKey()
	benchmarkPublicKeyString string
	benchmarkPublicKeyJSON   []byte
	benchmarkPublicKeyText   []byte
)

func BenchmarkPublicKeyString(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkPublicKeyString = benchmarkPublicKey.String()
	}
}

func BenchmarkPublicKeyFromBase58(b *testing.B) {
	encoded := benchmarkPublicKey.String()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := PublicKeyFromBase58(encoded); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPublicKeyMarshalJSON(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkPublicKeyJSON, err = benchmarkPublicKey.MarshalJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPublicKeyUnmarshalJSON(b *testing.B) {
	data, err := benchmarkPublicKey.MarshalJSON()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var key PublicKey
		if err := key.UnmarshalJSON(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPublicKeyMarshalText(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkPublicKeyText, err = benchmarkPublicKey.MarshalText()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPublicKeyUnmarshalText(b *testing.B) {
	data, err := benchmarkPublicKey.MarshalText()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var key PublicKey
		if err := key.UnmarshalText(data); err != nil {
			b.Fatal(err)
		}
	}
}
