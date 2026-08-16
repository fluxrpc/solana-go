package benchcmp

import (
	"bytes"
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

// Shared deterministic fixture for the transaction-builder comparison: two
// instructions over two programs with overlapping accounts (one account
// upgraded readonly->writable across instructions), plus a lookup-table
// variant that compresses the non-signer accounts into a v0 message.

func buildKey(tag byte) [32]byte {
	var b [32]byte
	for i := range b {
		b[i] = tag ^ byte(i*7)
	}
	return b
}

var (
	buildAlice   = buildKey(0x01) // fee payer, signer+writable
	buildBob     = buildKey(0x02) // writable in ix1, readonly in ix2
	buildCarol   = buildKey(0x03) // readonly
	buildDave    = buildKey(0x04) // readonly
	buildProg1   = buildKey(0xE1)
	buildProg2   = buildKey(0xE2)
	buildTable   = buildKey(0xF0)
	buildHashRaw = buildKey(0xAA)
)

var buildIxData1 = []byte{1, 2, 3, 4}
var buildIxData2 = []byte{9, 8}

func fluxBuildInstructions() []flux.Instruction {
	return []flux.Instruction{
		flux.NewInstruction(
			flux.PublicKey(buildProg1),
			flux.AccountMetaSlice{
				{PublicKey: flux.PublicKey(buildAlice), IsSigner: true, IsWritable: true},
				{PublicKey: flux.PublicKey(buildBob), IsWritable: true},
				{PublicKey: flux.PublicKey(buildCarol)},
			},
			buildIxData1,
		),
		flux.NewInstruction(
			flux.PublicKey(buildProg2),
			flux.AccountMetaSlice{
				{PublicKey: flux.PublicKey(buildAlice), IsSigner: true, IsWritable: true},
				{PublicKey: flux.PublicKey(buildDave)},
				{PublicKey: flux.PublicKey(buildBob)},
			},
			buildIxData2,
		),
	}
}

func gaglBuildInstructions() []gagl.Instruction {
	return []gagl.Instruction{
		gagl.NewInstruction(
			gagl.PublicKey(buildProg1),
			gagl.AccountMetaSlice{
				{PublicKey: gagl.PublicKey(buildAlice), IsSigner: true, IsWritable: true},
				{PublicKey: gagl.PublicKey(buildBob), IsWritable: true},
				{PublicKey: gagl.PublicKey(buildCarol)},
			},
			buildIxData1,
		),
		gagl.NewInstruction(
			gagl.PublicKey(buildProg2),
			gagl.AccountMetaSlice{
				{PublicKey: gagl.PublicKey(buildAlice), IsSigner: true, IsWritable: true},
				{PublicKey: gagl.PublicKey(buildDave)},
				{PublicKey: gagl.PublicKey(buildBob)},
			},
			buildIxData2,
		),
	}
}

func fluxBuildTables() map[flux.PublicKey][]flux.PublicKey {
	return map[flux.PublicKey][]flux.PublicKey{
		flux.PublicKey(buildTable): {
			flux.PublicKey(buildKey(0x77)), // unrelated entry before the used ones
			flux.PublicKey(buildBob),
			flux.PublicKey(buildCarol),
			flux.PublicKey(buildDave),
		},
	}
}

func gaglBuildTables() map[gagl.PublicKey]gagl.PublicKeySlice {
	return map[gagl.PublicKey]gagl.PublicKeySlice{
		gagl.PublicKey(buildTable): {
			gagl.PublicKey(buildKey(0x77)),
			gagl.PublicKey(buildBob),
			gagl.PublicKey(buildCarol),
			gagl.PublicKey(buildDave),
		},
	}
}

// TestTransactionBuild_Parity asserts that flux's NewTransaction produces a
// byte-identical message to upstream's for the same inputs, for both the
// legacy and the lookup-table (v0) paths.
func TestTransactionBuild_Parity(t *testing.T) {
	// Legacy.
	fluxTx, err := flux.NewTransaction(fluxBuildInstructions(), flux.Hash(buildHashRaw))
	if err != nil {
		t.Fatal(err)
	}
	gaglTx, err := gagl.NewTransaction(gaglBuildInstructions(), gagl.Hash(buildHashRaw))
	if err != nil {
		t.Fatal(err)
	}
	fluxBytes, err := fluxTx.Message.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	gaglBytes, err := gaglTx.Message.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fluxBytes, gaglBytes) {
		t.Errorf("legacy build mismatch:\nflux %x\ngagl %x", fluxBytes, gaglBytes)
	}
	t.Logf("legacy message base64: %s", fluxTx.Message.ToBase64())

	// V0 with lookup tables.
	fluxTx, err = flux.NewTransaction(
		fluxBuildInstructions(), flux.Hash(buildHashRaw),
		flux.TransactionAddressTables(fluxBuildTables()),
	)
	if err != nil {
		t.Fatal(err)
	}
	gaglTx, err = gagl.NewTransaction(
		gaglBuildInstructions(), gagl.Hash(buildHashRaw),
		gagl.TransactionAddressTables(gaglBuildTables()),
	)
	if err != nil {
		t.Fatal(err)
	}
	fluxBytes, err = fluxTx.Message.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	gaglBytes, err = gaglTx.Message.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fluxBytes, gaglBytes) {
		t.Errorf("v0 build mismatch:\nflux %x\ngagl %x", fluxBytes, gaglBytes)
	}
	t.Logf("v0 message base64: %s", fluxTx.Message.ToBase64())
}

func BenchmarkTransaction_Build(b *testing.B) {
	fluxIxs := fluxBuildInstructions()
	gaglIxs := gaglBuildInstructions()
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxTx, sinkErr = flux.NewTransaction(fluxIxs, flux.Hash(buildHashRaw))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglTx, sinkErr = gagl.NewTransaction(gaglIxs, gagl.Hash(buildHashRaw))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}

func BenchmarkTransaction_BuildV0(b *testing.B) {
	fluxIxs := fluxBuildInstructions()
	gaglIxs := gaglBuildInstructions()
	fluxTables := fluxBuildTables()
	gaglTables := gaglBuildTables()
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxTx, sinkErr = flux.NewTransaction(fluxIxs, flux.Hash(buildHashRaw), flux.TransactionAddressTables(fluxTables))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglTx, sinkErr = gagl.NewTransaction(gaglIxs, gagl.Hash(buildHashRaw), gagl.TransactionAddressTables(gaglTables))
		}
		if sinkErr != nil {
			b.Fatal(sinkErr)
		}
	})
}
