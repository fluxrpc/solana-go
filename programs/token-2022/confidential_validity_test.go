package token2022

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestGroupedCiphertextValidityProofRustInteroperability(t *testing.T) {
	service := ConfidentialTransferService{}
	secretValues := [3]byte{1, 3, 4}
	publicKeys := [3]ElGamalPubkey{}
	for index, value := range secretValues {
		secretBytes := make([]byte, 32)
		secretBytes[0] = value
		secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
		if err != nil {
			t.Fatal(err)
		}
		publicKeys[index], err = service.DeriveElGamalPubkey(secret)
		if err != nil {
			t.Fatal(err)
		}
	}
	openingBytes := make([]byte, 32)
	openingBytes[0] = 5
	opening, err := service.PedersenOpeningFromBytes(openingBytes)
	if err != nil {
		t.Fatal(err)
	}
	highOpeningBytes := make([]byte, 32)
	highOpeningBytes[0] = 6
	highOpening, err := service.PedersenOpeningFromBytes(highOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	grouped2, err := service.EncryptGroupedElGamal2WithOpening([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	grouped2High, err := service.EncryptGroupedElGamal2WithOpening([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, 7, highOpening)
	if err != nil {
		t.Fatal(err)
	}
	grouped3, err := service.EncryptGroupedElGamal3WithOpening(publicKeys, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	grouped3High, err := service.EncryptGroupedElGamal3WithOpening(publicKeys, 7, highOpening)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 128)
	random[0] = 8
	random[64] = 9
	grouped2Data, err := service.generateGroupedCiphertext2HandlesValidityProofWithReader([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, grouped2, 42, opening, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	batched2Data, err := service.generateBatchedGroupedCiphertext2HandlesValidityProofWithReader(
		[2]ElGamalPubkey{publicKeys[0], publicKeys[1]},
		[2]GroupedElGamalCiphertext2Handles{grouped2, grouped2High},
		[2]uint64{42, 7},
		[2]PedersenOpening{opening, highOpening},
		bytes.NewReader(random),
	)
	if err != nil {
		t.Fatal(err)
	}
	grouped3Data, err := service.generateGroupedCiphertext3HandlesValidityProofWithReader(publicKeys, grouped3, 42, opening, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	batched3Data, err := service.generateBatchedGroupedCiphertext3HandlesValidityProofWithReader(
		publicKeys,
		[2]GroupedElGamalCiphertext3Handles{grouped3, grouped3High},
		[2]uint64{42, 7},
		[2]PedersenOpening{opening, highOpening},
		bytes.NewReader(random),
	)
	if err != nil {
		t.Fatal(err)
	}
	vectors := []struct {
		Name       string
		Data       ZKProofData
		ContextHex string
		ProofHex   string
	}{
		{
			Name:       "grouped-2",
			Data:       grouped2Data,
			ContextHex: "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134c29d170ab8a5b42a3520878501a87a27f9b5653fca8b0c59fc2786cf26e37824422d138fc131895156f91e3e7e11de68e4c370da80e54cbc2bd1b5e3285c3e2184dffada0b0ea52ecd8ad35f8e608eb6b1b5da75901afd5e378c9184bb6b92713e34d08d94993e15383c5e26aa984b3096bdd41012b951fdaf74796b71fb8930",
			ProofHex:   "06fdc74ba753cccf69b196d50a83d576cc8d95a18a62efced6ed911855bf8a750c9022fbdb3f718a13d10f169e72dbc4b646f8341cc634f5f888517f7a36ec4bac1292bfb063e1a2d55a3a5730c8159d39d32199a07220335114e6880301ca6f9d99892ab5dd5d4f87cb09818f47a692eee150e40e846e9d05caa3a0142f370fb42ed329240e85b8e87f60deec88a9c639d00db116226d2a2fd45f45adbe0203",
		},
		{
			Name:       "batched-2",
			Data:       batched2Data,
			ContextHex: "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134c29d170ab8a5b42a3520878501a87a27f9b5653fca8b0c59fc2786cf26e37824422d138fc131895156f91e3e7e11de68e4c370da80e54cbc2bd1b5e3285c3e2184dffada0b0ea52ecd8ad35f8e608eb6b1b5da75901afd5e378c9184bb6b92713e34d08d94993e15383c5e26aa984b3096bdd41012b951fdaf74796b71fb893032fcf6eb56fe829ee0c4b418820557ec9f7b0bdc2b8252e93851bd56370f2a0e844c0f39d5b92254a3cffd1089761a2f12e01e9f0b6f899fc4d041c9e0d6e54710f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207",
			ProofHex:   "06fdc74ba753cccf69b196d50a83d576cc8d95a18a62efced6ed911855bf8a750c9022fbdb3f718a13d10f169e72dbc4b646f8341cc634f5f888517f7a36ec4bac1292bfb063e1a2d55a3a5730c8159d39d32199a07220335114e6880301ca6f2c6f936992a41b276cd40bcda0bee94e4c905a0bc657b48bd0b632acbf39800f8902fb830d03305d995b7b29aa17182760cd7680df062f1cdc0a36fc13e70a05",
		},
		{
			Name:       "grouped-3",
			Data:       grouped3Data,
			ContextHex: "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134c29d170ab8a5b42a3520878501a87a27f9b5653fca8b0c59fc2786cf26e378247ea555bf91bfb985561a91afcd669a79c0cc115ce03baf687cb8dd7e1e996e7b422d138fc131895156f91e3e7e11de68e4c370da80e54cbc2bd1b5e3285c3e2184dffada0b0ea52ecd8ad35f8e608eb6b1b5da75901afd5e378c9184bb6b92713e34d08d94993e15383c5e26aa984b3096bdd41012b951fdaf74796b71fb893036bb1095d1b2e3119abc4e6e5e13cfcce694c6554f455dae113c017a7395572e",
			ProofHex:   "06fdc74ba753cccf69b196d50a83d576cc8d95a18a62efced6ed911855bf8a750c9022fbdb3f718a13d10f169e72dbc4b646f8341cc634f5f888517f7a36ec4bac1292bfb063e1a2d55a3a5730c8159d39d32199a07220335114e6880301ca6f10f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207404172d44654b6529d598dc77d789f322112407a69f164f6733c10cb3e6d120a0f095b48305a0220028e68d69aad4dd67ccbb3cf4252b615cefb2110a9956701",
		},
		{
			Name:       "batched-3",
			Data:       batched3Data,
			ContextHex: "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134c29d170ab8a5b42a3520878501a87a27f9b5653fca8b0c59fc2786cf26e378247ea555bf91bfb985561a91afcd669a79c0cc115ce03baf687cb8dd7e1e996e7b422d138fc131895156f91e3e7e11de68e4c370da80e54cbc2bd1b5e3285c3e2184dffada0b0ea52ecd8ad35f8e608eb6b1b5da75901afd5e378c9184bb6b92713e34d08d94993e15383c5e26aa984b3096bdd41012b951fdaf74796b71fb893036bb1095d1b2e3119abc4e6e5e13cfcce694c6554f455dae113c017a7395572e32fcf6eb56fe829ee0c4b418820557ec9f7b0bdc2b8252e93851bd56370f2a0e844c0f39d5b92254a3cffd1089761a2f12e01e9f0b6f899fc4d041c9e0d6e54710f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207bc306d229cefdc72f0c54f6c456162c3c8b97562f9cb99baa1e2fe62512f1a04",
			ProofHex:   "06fdc74ba753cccf69b196d50a83d576cc8d95a18a62efced6ed911855bf8a750c9022fbdb3f718a13d10f169e72dbc4b646f8341cc634f5f888517f7a36ec4bac1292bfb063e1a2d55a3a5730c8159d39d32199a07220335114e6880301ca6f10f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207979e29237a4460a51db894d49316c699de0e0c920c288e4af93046bbe0fc7a0d14bf7a39e2048dd4853d67fda7e6692201fb5582529dc2148d55d8af4e57b403",
		},
	}
	for _, vector := range vectors {
		if got := hex.EncodeToString(vector.Data.Context); got != vector.ContextHex {
			t.Fatalf("%s context = %s, want %s", vector.Name, got, vector.ContextHex)
		}
		if got := hex.EncodeToString(vector.Data.Proof); got != vector.ProofHex {
			t.Fatalf("%s proof = %s, want %s", vector.Name, got, vector.ProofHex)
		}
		rustContext, err := hex.DecodeString(vector.ContextHex)
		if err != nil {
			t.Fatal(err)
		}
		rustProof, err := hex.DecodeString(vector.ProofHex)
		if err != nil {
			t.Fatal(err)
		}
		rustData := ZKProofData{Discriminator: vector.Data.Discriminator, Context: rustContext, Proof: rustProof}
		var verifyErr error
		switch vector.Data.Discriminator {
		case 9:
			verifyErr = service.VerifyGroupedCiphertext2HandlesValidityProof(rustData)
		case 10:
			verifyErr = service.VerifyBatchedGroupedCiphertext2HandlesValidityProof(rustData)
		case 11:
			verifyErr = service.VerifyGroupedCiphertext3HandlesValidityProof(rustData)
		case 12:
			verifyErr = service.VerifyBatchedGroupedCiphertext3HandlesValidityProof(rustData)
		}
		if verifyErr != nil {
			t.Fatalf("verify Rust %s proof: %v", vector.Name, verifyErr)
		}
		tampered := ZKProofData{Discriminator: rustData.Discriminator, Context: append([]byte(nil), rustData.Context...), Proof: append([]byte(nil), rustData.Proof...)}
		tampered.Proof[len(tampered.Proof)-1] ^= 1
		switch vector.Data.Discriminator {
		case 9:
			verifyErr = service.VerifyGroupedCiphertext2HandlesValidityProof(tampered)
		case 10:
			verifyErr = service.VerifyBatchedGroupedCiphertext2HandlesValidityProof(tampered)
		case 11:
			verifyErr = service.VerifyGroupedCiphertext3HandlesValidityProof(tampered)
		case 12:
			verifyErr = service.VerifyBatchedGroupedCiphertext3HandlesValidityProof(tampered)
		}
		if verifyErr == nil {
			t.Fatalf("expected tampered %s proof error", vector.Name)
		}
	}
	if _, err := service.generateGroupedCiphertext2HandlesValidityProofWithReader([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, grouped2, 43, opening, bytes.NewReader(random)); err == nil {
		t.Fatal("expected grouped ciphertext witness mismatch")
	}
	if _, err := service.generateBatchedGroupedCiphertext3HandlesValidityProofWithReader(publicKeys, [2]GroupedElGamalCiphertext3Handles{grouped3High, grouped3}, [2]uint64{42, 7}, [2]PedersenOpening{opening, highOpening}, bytes.NewReader(random)); err == nil {
		t.Fatal("expected batched grouped ciphertext witness mismatch")
	}
	if _, err := service.generateGroupedCiphertext3HandlesValidityProofWithReader(publicKeys, grouped3, 42, opening, bytes.NewReader(random[:64])); err == nil {
		t.Fatal("expected random source error")
	}
}

func TestGroupedCiphertextValidityProofIdentityRules(t *testing.T) {
	service := ConfidentialTransferService{}
	secretValues := [2]byte{1, 3}
	publicKeys := [2]ElGamalPubkey{}
	for index, value := range secretValues {
		secretBytes := make([]byte, 32)
		secretBytes[0] = value
		secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
		if err != nil {
			t.Fatal(err)
		}
		publicKeys[index], err = service.DeriveElGamalPubkey(secret)
		if err != nil {
			t.Fatal(err)
		}
	}
	openingBytes := make([]byte, 32)
	openingBytes[0] = 5
	opening, err := service.PedersenOpeningFromBytes(openingBytes)
	if err != nil {
		t.Fatal(err)
	}
	random := make([]byte, 128)
	random[0] = 8
	random[64] = 9
	publicKeys2 := [2]ElGamalPubkey{publicKeys[0], ElGamalPubkey{}}
	grouped2, err := service.EncryptGroupedElGamal2WithOpening(publicKeys2, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	data2, err := service.generateGroupedCiphertext2HandlesValidityProofWithReader(publicKeys2, grouped2, 42, opening, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyGroupedCiphertext2HandlesValidityProof(data2); err != nil {
		t.Fatalf("verify identity auditor proof: %v", err)
	}
	invalidProof := ZKProofData{Discriminator: data2.Discriminator, Context: append([]byte(nil), data2.Context...), Proof: append([]byte(nil), data2.Proof...)}
	clear(invalidProof.Proof[:32])
	if err := service.VerifyGroupedCiphertext2HandlesValidityProof(invalidProof); err == nil {
		t.Fatal("expected identity proof commitment error")
	}
	invalidCommitment := ZKProofData{Discriminator: data2.Discriminator, Context: append([]byte(nil), data2.Context...), Proof: append([]byte(nil), data2.Proof...)}
	clear(invalidCommitment.Context[64:96])
	if err := service.VerifyGroupedCiphertext2HandlesValidityProof(invalidCommitment); err == nil {
		t.Fatal("expected identity ciphertext commitment error")
	}
	invalidScalar := ZKProofData{Discriminator: data2.Discriminator, Context: append([]byte(nil), data2.Context...), Proof: append([]byte(nil), data2.Proof...)}
	for index := len(invalidScalar.Proof) - 32; index < len(invalidScalar.Proof); index++ {
		invalidScalar.Proof[index] = 0xff
	}
	if err := service.VerifyGroupedCiphertext2HandlesValidityProof(invalidScalar); err == nil {
		t.Fatal("expected non-canonical response error")
	}
	publicKeys3 := [3]ElGamalPubkey{publicKeys[0], publicKeys[1], ElGamalPubkey{}}
	grouped3, err := service.EncryptGroupedElGamal3WithOpening(publicKeys3, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	data3, err := service.generateGroupedCiphertext3HandlesValidityProofWithReader(publicKeys3, grouped3, 42, opening, bytes.NewReader(random))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyGroupedCiphertext3HandlesValidityProof(data3); err != nil {
		t.Fatalf("verify identity auditor proof: %v", err)
	}
	invalid := ZKProofData{Discriminator: data3.Discriminator, Context: append([]byte(nil), data3.Context...), Proof: append([]byte(nil), data3.Proof...)}
	copy(invalid.Context[:32], make([]byte, 32))
	if err := service.VerifyGroupedCiphertext3HandlesValidityProof(invalid); err == nil {
		t.Fatal("expected identity first public key error")
	}
	copy(invalid.Context[:32], data3.Context[:32])
	copy(invalid.Context[32:64], make([]byte, 32))
	if err := service.VerifyGroupedCiphertext3HandlesValidityProof(invalid); err == nil {
		t.Fatal("expected identity second public key error")
	}
}

func TestGroupedCiphertextValidityProofRandomAndValidation(t *testing.T) {
	service := ConfidentialTransferService{}
	secretValues := [3]byte{1, 3, 4}
	publicKeys := [3]ElGamalPubkey{}
	for index, value := range secretValues {
		secretBytes := make([]byte, 32)
		secretBytes[0] = value
		secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
		if err != nil {
			t.Fatal(err)
		}
		publicKeys[index], err = service.DeriveElGamalPubkey(secret)
		if err != nil {
			t.Fatal(err)
		}
	}
	opening, err := service.GeneratePedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	highOpening, err := service.GeneratePedersenOpening()
	if err != nil {
		t.Fatal(err)
	}
	grouped2, err := service.EncryptGroupedElGamal2WithOpening([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	grouped2High, err := service.EncryptGroupedElGamal2WithOpening([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, 7, highOpening)
	if err != nil {
		t.Fatal(err)
	}
	data2, err := service.GenerateGroupedCiphertext2HandlesValidityProof([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, grouped2, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyGroupedCiphertext2HandlesValidityProof(data2); err != nil {
		t.Fatal(err)
	}
	batched2, err := service.GenerateBatchedGroupedCiphertext2HandlesValidityProof([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, [2]GroupedElGamalCiphertext2Handles{grouped2, grouped2High}, [2]uint64{42, 7}, [2]PedersenOpening{opening, highOpening})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyBatchedGroupedCiphertext2HandlesValidityProof(batched2); err != nil {
		t.Fatal(err)
	}
	grouped3, err := service.EncryptGroupedElGamal3WithOpening(publicKeys, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	grouped3High, err := service.EncryptGroupedElGamal3WithOpening(publicKeys, 7, highOpening)
	if err != nil {
		t.Fatal(err)
	}
	data3, err := service.GenerateGroupedCiphertext3HandlesValidityProof(publicKeys, grouped3, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyGroupedCiphertext3HandlesValidityProof(data3); err != nil {
		t.Fatal(err)
	}
	batched3, err := service.GenerateBatchedGroupedCiphertext3HandlesValidityProof(publicKeys, [2]GroupedElGamalCiphertext3Handles{grouped3, grouped3High}, [2]uint64{42, 7}, [2]PedersenOpening{opening, highOpening})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyBatchedGroupedCiphertext3HandlesValidityProof(batched3); err != nil {
		t.Fatal(err)
	}
	if err := service.VerifyGroupedCiphertext2HandlesValidityProof(ZKProofData{Discriminator: 9}); err == nil {
		t.Fatal("expected grouped ciphertext proof size error")
	}
	if err := service.VerifyBatchedGroupedCiphertext3HandlesValidityProof(ZKProofData{Discriminator: 12}); err == nil {
		t.Fatal("expected batched grouped ciphertext proof size error")
	}
}
