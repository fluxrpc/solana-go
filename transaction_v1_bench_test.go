package solana_go

import "testing"

var (
	benchV1Sink    []byte
	benchV1TxSink  *TransactionV1
	benchV1ErrSink error
)

func benchV1Tx(b *testing.B) *TransactionV1 {
	b.Helper()
	tx, err := NewTransactionV1(builderInstructions(), Hash(builderKey(0xAB)), TransactionConfig{
		PriorityFeeLamports: u64ptr(5000),
		ComputeUnitLimit:    u32ptr(200_000),
	})
	if err != nil {
		b.Fatal(err)
	}
	tx.Signatures = make([]Signature, tx.Header.NumRequiredSignatures)
	return tx
}

func BenchmarkTransactionV1_MarshalBinary(b *testing.B) {
	tx := benchV1Tx(b)
	b.ReportAllocs()
	for b.Loop() {
		benchV1Sink, benchV1ErrSink = tx.MarshalBinary()
	}
	if benchV1ErrSink != nil {
		b.Fatal(benchV1ErrSink)
	}
}

func BenchmarkTransactionV1_FromBytes(b *testing.B) {
	raw, err := benchV1Tx(b).MarshalBinary()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		benchV1TxSink, benchV1ErrSink = TransactionV1FromBytes(raw)
	}
	if benchV1ErrSink != nil {
		b.Fatal(benchV1ErrSink)
	}
}

func BenchmarkTransactionV1_Build(b *testing.B) {
	ixs := builderInstructions()
	config := TransactionConfig{PriorityFeeLamports: u64ptr(5000)}
	b.ReportAllocs()
	for b.Loop() {
		benchV1TxSink, benchV1ErrSink = NewTransactionV1(ixs, Hash(builderKey(0xAB)), config)
	}
	if benchV1ErrSink != nil {
		b.Fatal(benchV1ErrSink)
	}
}
