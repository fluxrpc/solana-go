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
	ata, bump, err := FindAssociatedTokenAddress(pdaWallet, pdaMint)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := ata.String(), "Gkrr6Cr5bPLQhxSfJMBaU6BPuCWmwDyMpCRf5up41mXR"; got != want || bump != 255 {
		t.Fatalf("ATA = %s bump %d, want %s bump 255", got, bump, want)
	}
}

func TestFindProgramAddress(t *testing.T) {
	pda, bump, err := FindProgramAddress([][]byte{[]byte("metadata"), []byte("test-seed")}, TokenProgramID)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := pda.String(), "6Dh3itYoie51kVVad3M9vZ5TvP5L9oJvRBRwSFFXBJCd"; got != want || bump != 255 {
		t.Fatalf("PDA = %s bump %d, want %s bump 255", got, bump, want)
	}

	// The result must be reproducible through CreateProgramAddress with the
	// found bump, and must be off the curve.
	direct, err := CreateProgramAddress([][]byte{[]byte("metadata"), []byte("test-seed"), {bump}}, TokenProgramID)
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
	got, err := CreateProgramAddress([][]byte{[]byte("seed-one"), {255}}, TokenProgramID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "3cfmR2vaNq2i6x2T12ZXBDy8ueKQsyYsHWCagBsXvmbk"; got.String() != want {
		t.Fatalf("CreateProgramAddress = %s, want %s", got, want)
	}
}

func TestCreateWithSeed(t *testing.T) {
	got, err := CreateWithSeed(pdaWallet, "my-seed", SystemProgramID)
	if err != nil {
		t.Fatal(err)
	}
	if want := "2eQS3o8qKfPrPfz895WqKMLyY3B4NDKsproXbfnSmWxq"; got.String() != want {
		t.Fatalf("CreateWithSeed = %s, want %s", got, want)
	}

	if _, err := CreateWithSeed(pdaWallet, strings.Repeat("s", MaxSeedLength+1), SystemProgramID); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("oversized seed: %v", err)
	}
}

func TestFindTokenMetadataAddress(t *testing.T) {
	got, bump, err := FindTokenMetadataAddress(pdaMint)
	if err != nil {
		t.Fatal(err)
	}
	if want := "5x38Kp4hvdomTCnCrAny4UtMUt5rQBdB6px2K1Ui45Wq"; got.String() != want || bump != 255 {
		t.Fatalf("metadata = %s bump %d, want %s bump 255", got, bump, want)
	}
}

func TestPDASeedValidation(t *testing.T) {
	long := bytes.Repeat([]byte{1}, MaxSeedLength+1)
	if _, err := CreateProgramAddress([][]byte{long}, TokenProgramID); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("oversized seed: %v", err)
	}
	if _, _, err := FindProgramAddress([][]byte{long}, TokenProgramID); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("oversized seed: %v", err)
	}

	many := make([][]byte, MaxSeeds+1)
	for i := range many {
		many[i] = []byte{byte(i)}
	}
	if _, err := CreateProgramAddress(many, TokenProgramID); err != ErrMaxSeedLengthExceeded {
		t.Fatalf("too many seeds: %v", err)
	}
	// FindProgramAddress needs headroom for the bump seed.
	if _, _, err := FindProgramAddress(many[:MaxSeeds], TokenProgramID); err != ErrMaxSeedLengthExceeded {
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
	ata, _, err := FindAssociatedTokenAddress(pdaWallet, pdaMint)
	if err != nil {
		t.Fatal(err)
	}
	if ata.IsOnCurve() {
		t.Fatal("PDA reported on-curve")
	}
	if IsOnCurve([]byte{1, 2, 3}) {
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
		benchmarkPDA, benchmarkPDABump, err = FindProgramAddress(seeds, SPLAssociatedTokenAccountProgramID)
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
		benchmarkPDA, err = CreateProgramAddress(seeds, TokenProgramID)
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
