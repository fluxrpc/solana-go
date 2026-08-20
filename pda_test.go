package solana_go

import (
	"bytes"
	"strings"
	"testing"
)

// Ground-truth vectors generated with gagliardetto/solana-go v1.22.0.
var (
	pdaWallet = MustPublicKeyFromBase58("G7Hf2J55BAkHtbbXPh94UTGRCQioKPpnb5oKQMBteXo")
	pdaMint   = MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v") // USDC
)

func TestFindAssociatedTokenAddress(t *testing.T) {
	ata, bump, err := pdaWallet.FindAssociatedTokenAddress(pdaMint)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ata.String(), "Gkrr6Cr5bPLQhxSfJMBaU6BPuCWmwDyMpCRf5up41mXR"; got != want || bump != 255 {
		t.Fatalf("ATA = %s bump %d, want %s bump 255", got, bump, want)
	}
}

func TestFindProgramAddress(t *testing.T) {
	pda, bump, err := TokenProgramID.FindProgramAddress([][]byte{[]byte("metadata"), []byte("test-seed")})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := pda.String(), "6Dh3itYoie51kVVad3M9vZ5TvP5L9oJvRBRwSFFXBJCd"; got != want || bump != 255 {
		t.Fatalf("PDA = %s bump %d, want %s bump 255", got, bump, want)
	}

	// The result must be reproducible through CreateProgramAddress with the
	// found bump, and must be off the curve.
	direct, err := TokenProgramID.CreateProgramAddress([][]byte{[]byte("metadata"), []byte("test-seed"), {bump}})
	if err != nil {
		t.Fatal(err)
	}
	if direct != pda {
		t.Fatalf("CreateProgramAddress(bump) = %s, want %s", direct, pda)
	}
	if pda.IsOnCurve() {
		t.Fatal("derived address is on the curve")
	}
}

func TestCreateProgramAddress(t *testing.T) {
	got, err := TokenProgramID.CreateProgramAddress([][]byte{[]byte("seed-one"), {255}})
	if err != nil {
		t.Fatal(err)
	}
	if want := "3cfmR2vaNq2i6x2T12ZXBDy8ueKQsyYsHWCagBsXvmbk"; got.String() != want {
		t.Fatalf("CreateProgramAddress = %s, want %s", got, want)
	}
}

func TestCreateWithSeed(t *testing.T) {
	got, err := pdaWallet.CreateWithSeed("my-seed", SystemProgramID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "2eQS3o8qKfPrPfz895WqKMLyY3B4NDKsproXbfnSmWxq"; got.String() != want {
		t.Fatalf("CreateWithSeed = %s, want %s", got, want)
	}

	if _, err := pdaWallet.CreateWithSeed(strings.Repeat("s", MaxSeedLength+1), SystemProgramID); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("oversized seed: %v", err)
	}
}

func TestFindTokenMetadataAddress(t *testing.T) {
	got, bump, err := pdaMint.FindTokenMetadataAddress()
	if err != nil {
		t.Fatal(err)
	}
	if want := "5x38Kp4hvdomTCnCrAny4UtMUt5rQBdB6px2K1Ui45Wq"; got.String() != want || bump != 255 {
		t.Fatalf("metadata = %s bump %d, want %s bump 255", got, bump, want)
	}
}

func TestPDASeedValidation(t *testing.T) {
	long := bytes.Repeat([]byte{1}, MaxSeedLength+1)
	if _, err := TokenProgramID.CreateProgramAddress([][]byte{long}); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("oversized seed: %v", err)
	}
	if _, _, err := TokenProgramID.FindProgramAddress([][]byte{long}); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("oversized seed: %v", err)
	}

	many := make([][]byte, MaxSeeds+1)
	for i := range many {
		many[i] = []byte{byte(i)}
	}
	if _, err := TokenProgramID.CreateProgramAddress(many); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("too many seeds: %v", err)
	}
	// FindProgramAddress needs headroom for the bump seed.
	if _, _, err := TokenProgramID.FindProgramAddress(many[:MaxSeeds]); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("too many seeds incl. bump: %v", err)
	}
}

func TestIsOnCurve(t *testing.T) {
	// Real ed25519 public keys are on the curve.
	key := testPrivateKey(t)
	if !key.PublicKey().IsOnCurve() {
		t.Fatal("ed25519 public key reported off-curve")
	}
	// PDAs are off the curve by construction.
	ata, _, err := pdaWallet.FindAssociatedTokenAddress(pdaMint)
	if err != nil {
		t.Fatal(err)
	}
	if ata.IsOnCurve() {
		t.Fatal("PDA reported on-curve")
	}
	if isOnCurve([]byte{1, 2, 3}) {
		t.Fatal("short input reported on-curve")
	}
}

var (
	benchmarkPDA     PublicKey
	benchmarkPDABump uint8
)

func BenchmarkFindProgramAddress(b *testing.B) {
	seeds := [][]byte{pdaWallet[:], TokenProgramID[:], pdaMint[:]}
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkPDA, benchmarkPDABump, err = SPLAssociatedTokenAccountProgramID.FindProgramAddress(seeds)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCreateProgramAddress(b *testing.B) {
	seeds := [][]byte{[]byte("seed-one"), {255}}
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkPDA, err = TokenProgramID.CreateProgramAddress(seeds)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIsOnCurve(b *testing.B) {
	key := pdaWallet
	b.ReportAllocs()
	for b.Loop() {
		if !key.IsOnCurve() {
			b.Fatal("wallet key off-curve")
		}
	}
}
