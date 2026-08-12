package benchcmp

import (
	"encoding/json"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// Uses the legacyTxBase64 / legacyTxBytes fixture and the decoded
// gaglFixtureTx / fluxFixtureTx from message_bench_test.go.
//
// NOTE: neither implementation defines explicit (Un)MarshalJSON methods on
// Transaction; both rely on encoding/json over struct tags. The JSON
// benchmarks therefore go through json.Marshal / json.Unmarshal, which also
// picks up custom Marshaler/Unmarshaler implementations if either side adds
// them later.

var fixtureTxJSON = func() []byte {
	data, err := json.Marshal(gaglFixtureTx)
	if err != nil {
		panic(err)
	}
	return data
}()

var (
	sinkFluxTx *flux.Transaction
	sinkGaglTx *gagl.Transaction
)

func BenchmarkTransaction_MarshalBinary(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = fluxFixtureTx.MarshalBinary()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = gaglFixtureTx.MarshalBinary()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkTransaction_FromBytes(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxTx, sinkErr = flux.TransactionFromBytes(legacyTxBytes)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglTx, sinkErr = gagl.TransactionFromBytes(legacyTxBytes)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkTransaction_MarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = json.Marshal(fluxFixtureTx)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = json.Marshal(gaglFixtureTx)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkTransaction_UnmarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		var tx flux.Transaction
		for b.Loop() {
			sinkErr = json.Unmarshal(fixtureTxJSON, &tx)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkFluxTx = &tx
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		var tx gagl.Transaction
		for b.Loop() {
			sinkErr = json.Unmarshal(fixtureTxJSON, &tx)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkGaglTx = &tx
	})
}
