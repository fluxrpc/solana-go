package benchcmp

import (
	"encoding/base64"
	"testing"

	bin "github.com/gagliardetto/binary"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// legacyTxBase64 is a real mainnet legacy transaction taken from upstream's
// transaction_test.go (TestTransactionVerifySignatures): 1 signature, 6 account
// keys, 2 instructions. Shared with transaction_bench_test.go.
const legacyTxBase64 = "AVBFwRrn4wroV9+NVQfgg/GbjFtQFodLnNI5oTpDMQiQ4HfZNyFzcFamHSSFW4p5wc3efeEKvykbmk8jzf2LCQwBAAIGjYddInd/DSl2KJCP18GhEDlaJyPKVrgBGGsr3TF6jSYPgr3AdITNKr2UQVQ5I+Wh5StQv/a5XdLr6VN4Y21My1M/Y1FNK5wQLKJa1LYfN/HAudufFVtc0fRPR6AMUJ9UrkRI7sjY/PnpcXLF7A7SBvJrWu+o8+7QIaD8sL9aXkGFDy1uAqR6+CTQmradxC1wyyjL+iSft+5XudJWwSdi7wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAi+i1vCST+HNO0DEchpEJImMHhZ1BReuf7poRqmXpeA8CBAUBAgMCAgcAAwAAAAEABQIAAAwCAAAA6w0AAAAAAAA="

var legacyTxBytes = func() []byte {
	data, err := base64.StdEncoding.DecodeString(legacyTxBase64)
	if err != nil {
		panic(err)
	}
	return data
}()

// Decoded fixtures for both implementations, plus serialized forms used as
// unmarshal inputs. Message binary/JSON bytes are produced by upstream so both
// implementations parse identical input.
var (
	gaglFixtureTx = func() *gagl.Transaction {
		tx, err := gagl.TransactionFromBytes(legacyTxBytes)
		if err != nil {
			panic(err)
		}
		return tx
	}()
	fluxFixtureTx = func() *flux.Transaction {
		tx, err := flux.TransactionFromBytes(legacyTxBytes)
		if err != nil {
			panic(err)
		}
		return tx
	}()

	gaglFixtureMsg = gaglFixtureTx.Message
	fluxFixtureMsg = fluxFixtureTx.Message

	fixtureMsgBytes = func() []byte {
		data, err := gaglFixtureMsg.MarshalBinary()
		if err != nil {
			panic(err)
		}
		return data
	}()
	fixtureMsgJSON = func() []byte {
		data, err := gaglFixtureMsg.MarshalJSON()
		if err != nil {
			panic(err)
		}
		return data
	}()
)

var (
	sinkFluxMsg flux.Message
	sinkGaglMsg gagl.Message
)

func BenchmarkMessage_MarshalBinary(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = fluxFixtureMsg.MarshalBinary()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = gaglFixtureMsg.MarshalBinary()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkMessage_UnmarshalBinary(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		var m flux.Message
		for b.Loop() {
			sinkErr = m.UnmarshalBinary(fixtureMsgBytes)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkFluxMsg = m
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		// Upstream has no UnmarshalBinary; its equivalent is
		// UnmarshalWithDecoder over a fresh decoder.
		var m gagl.Message
		for b.Loop() {
			sinkErr = m.UnmarshalWithDecoder(bin.NewBinDecoder(fixtureMsgBytes))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkGaglMsg = m
	})
}

func BenchmarkMessage_MarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = fluxFixtureMsg.MarshalJSON()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkBytes, sinkErr = gaglFixtureMsg.MarshalJSON()
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkMessage_UnmarshalJSON(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		var m flux.Message
		for b.Loop() {
			sinkErr = m.UnmarshalJSON(fixtureMsgJSON)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkFluxMsg = m
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		var m gagl.Message
		for b.Loop() {
			sinkErr = m.UnmarshalJSON(fixtureMsgJSON)
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
		sinkGaglMsg = m
	})
}
