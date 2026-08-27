package token2022

import "testing"

func TestConfidentialArithmetic(t *testing.T) {
	service := ConfidentialTransferService{}
	keypair, err := service.ElGamalKeypairFromSecret(ElGamalSecretKey{1})
	if err != nil {
		t.Fatal(err)
	}
	lowOpening, highOpening := PedersenOpening{2}, PedersenOpening{3}
	low, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 7, lowOpening)
	if err != nil {
		t.Fatal(err)
	}
	high, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 11, highOpening)
	if err != nil {
		t.Fatal(err)
	}
	combined, err := service.CombineElGamalAmountCiphertexts(low, high)
	if err != nil {
		t.Fatal(err)
	}
	wantOpening, err := service.ScalePedersenOpening(highOpening, 1<<16)
	if err != nil {
		t.Fatal(err)
	}
	wantOpening, err = service.AddPedersenOpenings(lowOpening, wantOpening)
	if err != nil {
		t.Fatal(err)
	}
	want, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 7+11<<16, wantOpening)
	if err != nil {
		t.Fatal(err)
	}
	if combined != want {
		t.Fatalf("combined = %x, want %x", combined, want)
	}
	lowCommitment, err := service.CommitPedersen(7, lowOpening)
	if err != nil {
		t.Fatal(err)
	}
	highCommitment, err := service.CommitPedersen(11, highOpening)
	if err != nil {
		t.Fatal(err)
	}
	commitment, opening, err := service.CombinePedersenAmount(lowCommitment, highCommitment, lowOpening, highOpening)
	if err != nil {
		t.Fatal(err)
	}
	if opening != wantOpening {
		t.Fatalf("opening = %x, want %x", opening, wantOpening)
	}
	wantCommitment, err := service.CommitPedersen(7+11<<16, opening)
	if err != nil {
		t.Fatal(err)
	}
	if commitment != wantCommitment {
		t.Fatalf("commitment = %x, want %x", commitment, wantCommitment)
	}
}
