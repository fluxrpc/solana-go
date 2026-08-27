package token2022

import (
	"bytes"
	"os"
	"testing"
)

func TestBatchedRangeProofFamilies(t *testing.T) {
	service := ConfidentialTransferService{}
	tests := []struct {
		bitLengths []uint8
		amounts    []uint64
		discrim    uint8
		proofSize  int
	}{
		{bitLengths: []uint8{8, 16, 40}, amounts: []uint64{255, 65535, 1<<40 - 1}, discrim: 6, proofSize: 672},
		{bitLengths: []uint8{64, 32, 16, 16}, amounts: []uint64{^uint64(0), 1<<32 - 1, 65535, 7}, discrim: 7, proofSize: 736},
		{bitLengths: []uint8{64, 64, 64, 64}, amounts: []uint64{0, 1, 1 << 63, ^uint64(0)}, discrim: 8, proofSize: 800},
	}
	for _, test := range tests {
		commitments := make([]PedersenCommitment, len(test.amounts))
		openings := make([]PedersenOpening, len(test.amounts))
		for index, amount := range test.amounts {
			openingBytes := make([]byte, 32)
			openingBytes[0] = byte(index + 1)
			opening, err := service.PedersenOpeningFromBytes(openingBytes)
			if err != nil {
				t.Fatal(err)
			}
			openings[index] = opening
			commitments[index], err = service.CommitPedersen(amount, opening)
			if err != nil {
				t.Fatal(err)
			}
		}
		proof, err := service.generateBatchedRangeProofWithReader(commitments, test.amounts, test.bitLengths, openings, bytes.NewReader(bytes.Repeat([]byte{0x42}, 65536)))
		if err != nil {
			t.Fatal(err)
		}
		if proof.Discriminator != test.discrim || len(proof.Context) != 264 || len(proof.Proof) != test.proofSize {
			t.Fatalf("proof shape = %d/%d/%d", proof.Discriminator, len(proof.Context), len(proof.Proof))
		}
		if err := service.VerifyBatchedRangeProof(proof); err != nil {
			t.Fatalf("verify discriminator %d: %v", proof.Discriminator, err)
		}
	}
}

func TestBatchedRangeProofRustFixtures(t *testing.T) {
	service := ConfidentialTransferService{}
	fixtures := []struct {
		path          string
		discriminator uint8
	}{
		{path: "testdata/range_u64.bin", discriminator: 6},
		{path: "testdata/range_u128.bin", discriminator: 7},
		{path: "testdata/range_u256.bin", discriminator: 8},
	}
	for _, fixture := range fixtures {
		data, err := os.ReadFile(fixture.path)
		if err != nil {
			t.Fatal(err)
		}
		proof := ZKProofData{
			Discriminator: fixture.discriminator,
			Context:       data[:264],
			Proof:         data[264:],
		}
		if err := service.VerifyBatchedRangeProof(proof); err != nil {
			t.Fatalf("verify Rust fixture %d: %v", fixture.discriminator, err)
		}
	}
}

func TestBatchedRangeProofBoundariesAndMalformedData(t *testing.T) {
	service := ConfidentialTransferService{}
	amounts := []uint64{255, 65535, 1<<40 - 1}
	bitLengths := []uint8{8, 16, 40}
	commitments := make([]PedersenCommitment, len(amounts))
	openings := make([]PedersenOpening, len(amounts))
	for index, amount := range amounts {
		openingBytes := make([]byte, 32)
		openingBytes[0] = byte(index + 1)
		opening, err := service.PedersenOpeningFromBytes(openingBytes)
		if err != nil {
			t.Fatal(err)
		}
		openings[index] = opening
		commitments[index], err = service.CommitPedersen(amount, opening)
		if err != nil {
			t.Fatal(err)
		}
	}
	proof, err := service.generateBatchedRangeProofWithReader(commitments, amounts, bitLengths, openings, bytes.NewReader(bytes.Repeat([]byte{0x24}, 32768)))
	if err != nil {
		t.Fatal(err)
	}

	tampered := ZKProofData{Discriminator: proof.Discriminator, Context: append([]byte{}, proof.Context...), Proof: append([]byte{}, proof.Proof...)}
	tampered.Proof[32] ^= 1
	if err := service.VerifyBatchedRangeProof(tampered); err == nil {
		t.Fatal("expected tampered proof error")
	}
	identity := ZKProofData{Discriminator: proof.Discriminator, Context: append([]byte{}, proof.Context...), Proof: append([]byte{}, proof.Proof...)}
	clear(identity.Proof[:32])
	if err := service.VerifyBatchedRangeProof(identity); err == nil {
		t.Fatal("expected identity proof point error")
	}
	nonCanonical := ZKProofData{Discriminator: proof.Discriminator, Context: append([]byte{}, proof.Context...), Proof: append([]byte{}, proof.Proof...)}
	for index := 128; index < 160; index++ {
		nonCanonical.Proof[index] = 0xff
	}
	if err := service.VerifyBatchedRangeProof(nonCanonical); err == nil {
		t.Fatal("expected non-canonical scalar error")
	}
	padding := ZKProofData{Discriminator: proof.Discriminator, Context: append([]byte{}, proof.Context...), Proof: append([]byte{}, proof.Proof...)}
	padding.Context[259] = 1
	if err := service.VerifyBatchedRangeProof(padding); err == nil {
		t.Fatal("expected context padding error")
	}
	short := ZKProofData{Discriminator: proof.Discriminator, Context: proof.Context, Proof: proof.Proof[:len(proof.Proof)-1]}
	if err := service.VerifyBatchedRangeProof(short); err == nil {
		t.Fatal("expected proof length error")
	}

	outOfRangeAmounts := append([]uint64{}, amounts...)
	outOfRangeAmounts[0] = 256
	outOfRangeCommitments := append([]PedersenCommitment{}, commitments...)
	outOfRangeCommitments[0], err = service.CommitPedersen(256, openings[0])
	if err != nil {
		t.Fatal(err)
	}
	outOfRange, err := service.generateBatchedRangeProofWithReader(outOfRangeCommitments, outOfRangeAmounts, bitLengths, openings, bytes.NewReader(bytes.Repeat([]byte{0x24}, 32768)))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyBatchedRangeProof(outOfRange); err == nil {
		t.Fatal("expected out-of-range amount verification error")
	}
	if _, err := service.generateBatchedRangeProofWithReader(commitments, amounts, bitLengths, openings, bytes.NewReader([]byte{1})); err == nil {
		t.Fatal("expected random reader error")
	}
	if _, err := service.GenerateBatchedRangeProof(commitments, amounts, []uint8{0, 24, 40}, openings); err == nil {
		t.Fatal("expected zero bit-length error")
	}
}
