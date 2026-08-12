package solana_go

import (
	"bytes"
	"encoding/json"
	"testing"
)

func testPrivateKey(t testing.TB) PrivateKey {
	t.Helper()
	key, err := NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func TestPrivateKeyRoundTrip(t *testing.T) {
	want := testPrivateKey(t)

	got, err := PrivateKeyFromBase58(want.String())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("round trip mismatch: got %x, want %x", got, want)
	}
}

func TestPrivateKeyFromBase58RejectsInvalidInput(t *testing.T) {
	tests := []string{
		"",
		"0OIl",
		"abc", // valid base58, wrong size
	}

	for _, input := range tests {
		if _, err := PrivateKeyFromBase58(input); err == nil {
			t.Errorf("PrivateKeyFromBase58(%q) unexpectedly succeeded", input)
		}
	}

	defer func() {
		if recover() == nil {
			t.Fatal("MustPrivateKeyFromBase58 accepted invalid input")
		}
	}()
	MustPrivateKeyFromBase58("abc")
}

func TestPrivateKeySignAndVerify(t *testing.T) {
	key := testPrivateKey(t)
	msg := []byte("solana test message")

	sig, err := key.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !sig.Verify(key.PublicKey(), msg) {
		t.Fatal("signature does not verify against the keypair's public key")
	}
	if sig.Verify(key.PublicKey(), []byte("other message")) {
		t.Fatal("signature verified against a different message")
	}
}

func TestPrivateKeyInvalidSize(t *testing.T) {
	bad := PrivateKey{1, 2, 3}
	if bad.IsValid() {
		t.Fatal("IsValid() accepted a 3-byte key")
	}
	if _, err := bad.Sign([]byte("msg")); err == nil {
		t.Fatal("Sign accepted an invalid key")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("PublicKey() accepted an invalid key")
		}
	}()
	bad.PublicKey()
}

func TestPrivateKeyJSONRoundTrip(t *testing.T) {
	want := testPrivateKey(t)

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"`+want.String()+`"` {
		t.Fatalf("MarshalJSON() = %s, want quoted %q", data, want.String())
	}

	var got PrivateKey
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("JSON round trip mismatch: got %x, want %x", got, want)
	}

	if err := json.Unmarshal([]byte(`"abc"`), &got); err == nil {
		t.Fatal("UnmarshalJSON accepted a wrong-size key")
	}
}

var (
	benchmarkSignResult   Signature
	benchmarkPubKeyResult PublicKey
)

func BenchmarkPrivateKeySign(b *testing.B) {
	key := testPrivateKey(b)
	msg := make([]byte, 200)
	for i := range msg {
		msg[i] = byte(i)
	}
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkSignResult, err = key.Sign(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrivateKeyPublicKey(b *testing.B) {
	key := testPrivateKey(b)
	b.ReportAllocs()
	for b.Loop() {
		benchmarkPubKeyResult = key.PublicKey()
	}
}
