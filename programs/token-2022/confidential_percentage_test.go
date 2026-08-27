package token2022

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestPercentageWithCapProofBranches(t *testing.T) {
	service := ConfidentialTransferService{}
	percentageOpeningBytes := make([]byte, 32)
	percentageOpeningBytes[0] = 1
	percentageOpening, err := service.PedersenOpeningFromBytes(percentageOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	deltaOpeningBytes := make([]byte, 32)
	deltaOpeningBytes[0] = 2
	deltaOpening, err := service.PedersenOpeningFromBytes(deltaOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	claimedOpeningBytes := make([]byte, 32)
	claimedOpeningBytes[0] = 3
	claimedOpening, err := service.PedersenOpeningFromBytes(claimedOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 640)
	for i := range random {
		random[i] = byte(i + 1)
	}

	percentageCommitment, err := service.CommitPedersen(1, percentageOpening)
	if err != nil {
		t.Fatal(err)
	}
	deltaCommitment, err := service.CommitPedersen(9600, deltaOpening)
	if err != nil {
		t.Fatal(err)
	}
	claimedCommitment, err := service.CommitPedersen(9600, claimedOpening)
	if err != nil {
		t.Fatal(err)
	}
	below, err := service.generatePercentageWithCapProofWithReader(percentageCommitment, percentageOpening, 1, deltaCommitment, deltaOpening, 9600, claimedCommitment, claimedOpening, 3, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	if below.Discriminator != 5 || len(below.Context) != 104 || len(below.Proof) != 256 {
		t.Fatalf("below proof sizes = %d/%d/%d", below.Discriminator, len(below.Context), len(below.Proof))
	}
	if err := service.VerifyPercentageWithCapProof(below); err != nil {
		t.Fatalf("verify below: %v", err)
	}
	if got, want := hex.EncodeToString(below.Context), "b8180a6778aba0f7bd121a403e09146d274edf702241a67c67689dc9bd87dd10d02622bb311a55ae806d54d797e8628b242e541af431aa9195ea8ec619d0403e7ee045a4808a362e4246d35d6140f115840005dc55bb5b832563c1a4351b40550300000000000000"; got != want {
		t.Fatalf("below context = %s", got)
	}
	if got, want := hex.EncodeToString(below.Proof), "640d69d82dfd9254011949f74405bdb94b0138fe677fed1e05424c749daf6b5b2b7ca133ae48e2e38e0fefd8894f81823cecf9d3ca70d936e5c13f76eb4f08017aad2fbda5dfd99b32ce11bea88d2f3f7e94320f80317d1f5aa7fe28bfdc44076e494962bd65b967d799bb1d3813cfc332034b608f82b1ae889dfef21151e5012a97ac8b2ca45917c209d58b2f3d09d6d8511b08f91dee186cf8ded7bf2df0499ceda15fc08bc7a2897db2e47d788fb124d1f695e39f35089f40547ac65a0e011748bce2ef19e4c6548aba48aba0945a4a4e8254a3614ededa164fc4d2a9f7080d0e24da76334f204297d8067b2934d7b37b9b6d1f7bfe0e851975f7032a4a0e"; got != want {
		t.Fatalf("below proof = %s", got)
	}

	percentageCommitment, err = service.CommitPedersen(3, percentageOpening)
	if err != nil {
		t.Fatal(err)
	}
	deltaCommitment, err = service.CommitPedersen(77, deltaOpening)
	if err != nil {
		t.Fatal(err)
	}
	claimedCommitment, err = service.CommitPedersen(0, claimedOpening)
	if err != nil {
		t.Fatal(err)
	}
	above, err := service.generatePercentageWithCapProofWithReader(percentageCommitment, percentageOpening, 3, deltaCommitment, deltaOpening, 0, claimedCommitment, claimedOpening, 3, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyPercentageWithCapProof(above); err != nil {
		t.Fatalf("verify capped: %v", err)
	}

	tampered := ZKProofData{Discriminator: below.Discriminator, Context: append([]byte(nil), below.Context...), Proof: append([]byte(nil), below.Proof...)}
	tampered.Proof[255] ^= 1
	if err := service.VerifyPercentageWithCapProof(tampered); err == nil {
		t.Fatal("expected tampered proof error")
	}
	tampered.Proof = append([]byte(nil), below.Proof...)
	tampered.Context[0] ^= 1
	if err := service.VerifyPercentageWithCapProof(tampered); err == nil {
		t.Fatal("expected tampered context error")
	}
	if err := service.VerifyPercentageWithCapProof(ZKProofData{Discriminator: 5}); err == nil {
		t.Fatal("expected invalid proof data error")
	}
}

func TestPercentageWithCapRustProofFixture(t *testing.T) {
	service := ConfidentialTransferService{}
	encoded, err := hex.DecodeString("b8180a6778aba0f7bd121a403e09146d274edf702241a67c67689dc9bd87dd10d02622bb311a55ae806d54d797e8628b242e541af431aa9195ea8ec619d0403e7ee045a4808a362e4246d35d6140f115840005dc55bb5b832563c1a4351b4055030000000000000044e193659e531d195ea2bd6504b46349f5663188e0ffbcb408ca115f88491f582afa65071e26356388598d8cac153c7d82039b7c62d6af14dffa3e9846dbcb047a40bd4f5e29073eef1f1e420c17c0044af396d0899c36622d87af86c743ce0aca7ea01c5a0d0694aec936b9f7b2f21121feea68abe5f0af8cb4110e8ef3004df24e2925fb571853c93f876e4474b153011a857f54db4f3eb89c8319c8bf4827dc7002ee5652571763cac9b96646beb5ff675a238be119d0fa76ff8389888303b7855378eddef0be14c4421c5325868e53d7257e6fecfc3e94c33e18322494007e7f647d00a2b173b310c3a2fa3ae9f1cbf7744d6d5ee9072bcfd05c25346602")
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 360 {
		t.Fatalf("Rust fixture length = %d", len(encoded))
	}
	data := ZKProofData{Discriminator: 5, Context: encoded[:104], Proof: encoded[104:]}
	if err := service.VerifyPercentageWithCapProof(data); err != nil {
		t.Fatalf("verify Rust proof: %v", err)
	}
}

func TestPercentageWithCapProofInputValidation(t *testing.T) {
	service := ConfidentialTransferService{}
	percentageOpeningBytes := make([]byte, 32)
	percentageOpeningBytes[0] = 1
	percentageOpening, _ := service.PedersenOpeningFromBytes(percentageOpeningBytes)
	deltaOpeningBytes := make([]byte, 32)
	deltaOpeningBytes[0] = 2
	deltaOpening, _ := service.PedersenOpeningFromBytes(deltaOpeningBytes)
	claimedOpeningBytes := make([]byte, 32)
	claimedOpeningBytes[0] = 3
	claimedOpening, _ := service.PedersenOpeningFromBytes(claimedOpeningBytes)
	percentageCommitment, _ := service.CommitPedersen(1, percentageOpening)
	deltaCommitment, _ := service.CommitPedersen(7, deltaOpening)
	claimedCommitment, _ := service.CommitPedersen(9, claimedOpening)
	random := bytes.NewReader(make([]byte, 640))
	if _, err := service.generatePercentageWithCapProofWithReader(percentageCommitment, percentageOpening, 2, deltaCommitment, deltaOpening, 9, claimedCommitment, claimedOpening, 3, random); err == nil {
		t.Fatal("expected percentage commitment error")
	}
	if _, err := service.generatePercentageWithCapProofWithReader(percentageCommitment, percentageOpening, 1, deltaCommitment, deltaOpening, 8, claimedCommitment, claimedOpening, 3, random); err == nil {
		t.Fatal("expected claimed commitment error")
	}
	if _, err := service.generatePercentageWithCapProofWithReader(percentageCommitment, percentageOpening, 1, deltaCommitment, deltaOpening, 9, claimedCommitment, claimedOpening, 3, random); err == nil {
		t.Fatal("expected delta commitment error")
	}
	overCapCommitment, _ := service.CommitPedersen(4, percentageOpening)
	if _, err := service.generatePercentageWithCapProofWithReader(overCapCommitment, percentageOpening, 4, deltaCommitment, deltaOpening, 9, claimedCommitment, claimedOpening, 3, random); err == nil {
		t.Fatal("expected capped amount error")
	}
	validDelta, _ := service.CommitPedersen(9, deltaOpening)
	if _, err := service.generatePercentageWithCapProofWithReader(percentageCommitment, percentageOpening, 1, validDelta, deltaOpening, 9, claimedCommitment, claimedOpening, 3, bytes.NewReader(nil)); err == nil {
		t.Fatal("expected random source error")
	}
}
