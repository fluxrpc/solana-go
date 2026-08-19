package solana_go

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"
)

func TestTransactionFromBase64Legacy(t *testing.T) {
	tx, err := TransactionFromBase64(legacyTxBase64)
	if err != nil {
		t.Fatal(err)
	}

	if len(tx.Signatures) != 1 {
		t.Fatalf("len(Signatures) = %d, want 1", len(tx.Signatures))
	}
	wantSig := "5yUSwqQqeZLEEYKxnG4JC4XhaaBpV3RS4nQbK8bQTyjLX5btVq9A1Ja5nuJzV7Z3Zq8G6EVKFvN4DKUL6PSAxmTk"
	if got := tx.Signatures[0].String(); got != wantSig {
		t.Fatalf("Signatures[0] = %s, want %s", got, wantSig)
	}
	if tx.Message.GetVersion() != MessageVersionLegacy {
		t.Fatalf("message version = %d, want legacy", tx.Message.GetVersion())
	}
	if len(tx.Message.AccountKeys) != 3 || len(tx.Message.Instructions) != 1 {
		t.Fatalf("message = %+v", tx.Message)
	}
}

func TestTransactionBinaryRoundTrip(t *testing.T) {
	for name, fixture := range map[string]string{"legacy": legacyTxBase64, "v0": v0TxBase64} {
		t.Run(name, func(t *testing.T) {
			raw, err := base64.StdEncoding.DecodeString(fixture)
			if err != nil {
				t.Fatal(err)
			}

			tx, err := TransactionFromBytes(raw)
			if err != nil {
				t.Fatal(err)
			}
			out, err := tx.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(out, raw) {
				t.Fatalf("binary round trip mismatch:\ngot  %x\nwant %x", out, raw)
			}
			if tx.ToBase64() != fixture {
				t.Fatalf("ToBase64() mismatch")
			}
		})
	}
}

func TestTransactionJSONRoundTrip(t *testing.T) {
	for name, fixture := range map[string]string{"legacy": legacyTxBase64, "v0": v0TxBase64} {
		t.Run(name, func(t *testing.T) {
			want, err := TransactionFromBase64(fixture)
			if err != nil {
				t.Fatal(err)
			}

			data, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			var got Transaction
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&got, want) {
				t.Fatalf("JSON round trip mismatch:\ngot  %+v\nwant %+v", &got, want)
			}
		})
	}
}

func TestTransactionFromBytesRejectsBadInput(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(legacyTxBase64)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string][]byte{
		"empty":                nil,
		"only signature count": {1},
		"truncated signature":  raw[:32],
		"absurd sig count":     {0xff, 0xff, 0x03},
		"missing message":      raw[:1+SignatureLength],
		"truncated message":    raw[:len(raw)-1],
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := TransactionFromBytes(data); err == nil {
				t.Fatal("TransactionFromBytes unexpectedly succeeded")
			}
		})
	}

	if _, err := TransactionFromBase64("not@base64!"); err == nil {
		t.Fatal("TransactionFromBase64 accepted invalid base64")
	}

	defer func() {
		if recover() == nil {
			t.Fatal("MustTransactionFromBytes accepted invalid input")
		}
	}()
	MustTransactionFromBytes(nil)
}

// testUnsignedTransaction builds a single-signer legacy transaction for the
// given fee payer.
func testUnsignedTransaction(payer PublicKey) *Transaction {
	receiver := PublicKey{7}
	return &Transaction{
		Message: Message{
			AccountKeys: []PublicKey{payer, receiver, {}},
			Header: MessageHeader{
				NumRequiredSignatures:       1,
				NumReadonlyUnsignedAccounts: 1,
			},
			RecentBlockhash: Hash{1, 2, 3},
			Instructions: []CompiledInstruction{{
				ProgramIDIndex: 2,
				Accounts:       []uint16{0, 1},
				Data:           Base58{2, 0, 0, 0, 1},
			}},
		},
	}
}

func TestTransactionSignAndVerify(t *testing.T) {
	key := testPrivateKey(t)
	tx := testUnsignedTransaction(key.PublicKey())

	sigs, err := tx.Sign(func(pub PublicKey) *PrivateKey {
		if pub == key.PublicKey() {
			return &key
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 1 || sigs[0].IsZero() {
		t.Fatalf("Sign() = %v", sigs)
	}
	if err := tx.VerifySignatures(); err != nil {
		t.Fatal(err)
	}

	// Tampering with the message must invalidate the signature.
	tx.Message.RecentBlockhash = Hash{9}
	if err := tx.VerifySignatures(); err == nil {
		t.Fatal("VerifySignatures accepted a tampered message")
	}

	// The signed transaction survives a binary round trip.
	tx.Message.RecentBlockhash = Hash{1, 2, 3}
	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := TransactionFromBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := decoded.VerifySignatures(); err != nil {
		t.Fatal(err)
	}
}

func TestTransactionSignMissingKey(t *testing.T) {
	key := testPrivateKey(t)
	tx := testUnsignedTransaction(key.PublicKey())

	if _, err := tx.Sign(func(PublicKey) *PrivateKey { return nil }); err == nil {
		t.Fatal("Sign succeeded without the signer key")
	}

	// PartialSign tolerates the missing key and leaves a zero signature.
	sigs, err := tx.PartialSign(func(PublicKey) *PrivateKey { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(sigs) != 1 || !sigs[0].IsZero() {
		t.Fatalf("PartialSign() = %v", sigs)
	}
}

func TestTransactionMarshalBinaryPadsMissingSignatures(t *testing.T) {
	key := testPrivateKey(t)
	tx := testUnsignedTransaction(key.PublicKey())

	raw, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if raw[0] != 1 {
		t.Fatalf("signature count = %d, want 1", raw[0])
	}
	if !bytes.Equal(raw[1:1+SignatureLength], make([]byte, SignatureLength)) {
		t.Fatal("missing signature was not encoded as zeros")
	}

	decoded, err := TransactionFromBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Message, tx.Message) {
		t.Fatal("message did not survive the round trip")
	}
}

var (
	benchmarkTransaction       *Transaction
	benchmarkTransactionBinary []byte
	benchmarkTransactionJSON   []byte
)

func benchmarkLegacyTransaction(b *testing.B) *Transaction {
	b.Helper()
	tx, err := TransactionFromBase64(legacyTxBase64)
	if err != nil {
		b.Fatal(err)
	}
	return tx
}

func BenchmarkTransactionMarshalBinary(b *testing.B) {
	tx := benchmarkLegacyTransaction(b)
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkTransactionBinary, err = tx.MarshalBinary()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransactionFromBytes(b *testing.B) {
	raw, err := base64.StdEncoding.DecodeString(legacyTxBase64)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		benchmarkTransaction, err = TransactionFromBytes(raw)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransactionMarshalJSON(b *testing.B) {
	tx := benchmarkLegacyTransaction(b)
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkTransactionJSON, err = json.Marshal(tx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTransactionUnmarshalJSON(b *testing.B) {
	data, err := json.Marshal(benchmarkLegacyTransaction(b))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		var tx Transaction
		if err := json.Unmarshal(data, &tx); err != nil {
			b.Fatal(err)
		}
	}
}

func TestTransactionFromBytesConsumed(t *testing.T) {
	for name, fixture := range map[string]string{"legacy": legacyTxBase64, "v0": v0TxBase64} {
		t.Run(name, func(t *testing.T) {
			raw, err := base64.StdEncoding.DecodeString(fixture)
			if err != nil {
				t.Fatal(err)
			}

			// Exact-length input consumes everything.
			tx, consumed, err := TransactionFromBytesConsumed(raw)
			if err != nil {
				t.Fatal(err)
			}
			if consumed != len(raw) {
				t.Fatalf("consumed = %d, want %d", consumed, len(raw))
			}

			// Trailing bytes are ignored and not counted.
			padded := append(append([]byte{}, raw...), 0xDE, 0xAD, 0xBE, 0xEF)
			tx2, consumed2, err := TransactionFromBytesConsumed(padded)
			if err != nil {
				t.Fatal(err)
			}
			if consumed2 != len(raw) {
				t.Fatalf("padded consumed = %d, want %d", consumed2, len(raw))
			}
			if tx.ToBase64() != tx2.ToBase64() {
				t.Fatal("trailing bytes changed the decoded transaction")
			}

			// Walking a concatenated buffer yields each transaction in turn.
			joined := append(append([]byte{}, raw...), raw...)
			off := 0
			for i := 0; i < 2; i++ {
				txi, n, err := TransactionFromBytesConsumed(joined[off:])
				if err != nil {
					t.Fatalf("tx %d at offset %d: %v", i, off, err)
				}
				if txi.ToBase64() != fixture {
					t.Fatalf("tx %d round-trip mismatch", i)
				}
				off += n
			}
			if off != len(joined) {
				t.Fatalf("walk consumed %d of %d bytes", off, len(joined))
			}
		})
	}
}
