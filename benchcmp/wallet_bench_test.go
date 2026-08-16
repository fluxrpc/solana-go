package benchcmp

import (
	"testing"

	flux "github.com/fluxrpc/solana-go"
	gagl "github.com/gagliardetto/solana-go"
)

var (
	sinkFluxWallet *flux.Wallet
	sinkGaglWallet *gagl.Wallet
)

// TestWallet_Parity asserts a base58 key round-trips into equal wallets on
// both implementations.
func TestWallet_Parity(t *testing.T) {
	base := flux.NewWallet()
	fluxWallet, err := flux.WalletFromPrivateKeyBase58(base.PrivateKey.String())
	if err != nil {
		t.Fatal(err)
	}
	gaglWallet, err := gagl.WalletFromPrivateKeyBase58(base.PrivateKey.String())
	if err != nil {
		t.Fatal(err)
	}
	if fluxWallet.PublicKey().String() != gaglWallet.PublicKey().String() {
		t.Fatalf("wallet mismatch: flux %s gagl %s", fluxWallet.PublicKey(), gaglWallet.PublicKey())
	}
}

func BenchmarkWallet_New(b *testing.B) {
	b.Run("flux", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkFluxWallet = flux.NewWallet()
		}
	})
	b.Run("gagl", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sinkGaglWallet = gagl.NewWallet()
		}
	})
}
