package token2022

import (
	"encoding/hex"
	"testing"
)

func TestConfidentialEncryptionRustFixtures(t *testing.T) {
	service := ConfidentialTransferService{}
	secretBytes, err := hex.DecodeString("0100000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	secret, err := service.ElGamalSecretKeyFromBytes(secretBytes)
	if err != nil {
		t.Fatal(err)
	}
	keypair, err := service.ElGamalKeypairFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(keypair.PublicKey[:]), "8c9240b456a9e6dc65c377a1048d745f94a08cdb7f44cbcd7b46f34048871134"; got != want {
		t.Fatalf("public key = %s, want %s", got, want)
	}
	openingBytes, err := hex.DecodeString("0200000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	opening, err := service.PedersenOpeningFromBytes(openingBytes)
	if err != nil {
		t.Fatal(err)
	}
	commitment, err := service.CommitPedersen(42, opening)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(commitment[:]), "f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e874"; got != want {
		t.Fatalf("commitment = %s, want %s", got, want)
	}
	ciphertext, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(ciphertext[:]), "f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207"; got != want {
		t.Fatalf("ciphertext = %s, want %s", got, want)
	}
	target, err := service.DecryptElGamalTarget(secret, ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(target[:]), "e00af9c74d9edb8ebcc160ceec97d531cbd6e2956f9e9162b8e9eda260e82e43"; got != want {
		t.Fatalf("decryption target = %s, want %s", got, want)
	}
	secondOpeningBytes, err := hex.DecodeString("0500000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	secondOpening, err := service.PedersenOpeningFromBytes(secondOpeningBytes)
	if err != nil {
		t.Fatal(err)
	}
	secondCiphertext, err := service.EncryptElGamalWithOpening(keypair.PublicKey, 7, secondOpening)
	if err != nil {
		t.Fatal(err)
	}
	sum, err := service.AddElGamalCiphertexts(ciphertext, secondCiphertext)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(sum[:]), "9a1d57880a0a958c94a61a164bef365149d8eedfa9d828e4ac2de4433e3f6505ae8f4180fd4eed5b16bcec7f462ca9d6707a79069191767bfc5196b3c519c476"; got != want {
		t.Fatalf("ciphertext sum = %s, want %s", got, want)
	}
	difference, err := service.SubtractElGamalCiphertexts(sum, secondCiphertext)
	if err != nil {
		t.Fatal(err)
	}
	if difference != ciphertext {
		t.Fatalf("ciphertext difference = %x, want %x", difference, ciphertext)
	}
	addedAmount, err := service.AddElGamalAmount(ciphertext, 7)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(addedAmount[:]), "081d947cbf132bde05676596e9cf6df3ae36e2bf152e1463c347c8c44d375b5210f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207"; got != want {
		t.Fatalf("ciphertext plus amount = %s, want %s", got, want)
	}
	subtractedAmount, err := service.SubtractElGamalAmount(ciphertext, 7)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(subtractedAmount[:]), "4cc879e22a7906a07b4e0f168729677bb9786b91df2f7f1bb3256808720f145c10f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207"; got != want {
		t.Fatalf("ciphertext minus amount = %s, want %s", got, want)
	}
}

func TestConfidentialEncryptionGroupedRustFixtures(t *testing.T) {
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
	openingBytes[0] = 2
	opening, err := service.PedersenOpeningFromBytes(openingBytes)
	if err != nil {
		t.Fatal(err)
	}
	grouped2, err := service.EncryptGroupedElGamal2WithOpening([2]ElGamalPubkey{publicKeys[0], publicKeys[1]}, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(grouped2[:]), "f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207a0aca2678c338c9023b57c6b65b3bc8bf43a4c542c11484c3495c2d21d4c6455"; got != want {
		t.Fatalf("grouped ciphertext with 2 handles = %s, want %s", got, want)
	}
	grouped3, err := service.EncryptGroupedElGamal3WithOpening(publicKeys, 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(grouped3[:]), "f06bddcc8c6fc05e28e513c93c8745048b70923dae7cd67e02376100b909e87410f21a8723d8943e2b37207a4815638fcc0b5efc9dc3346445ca985c6d5c2207a0aca2678c338c9023b57c6b65b3bc8bf43a4c542c11484c3495c2d21d4c6455f05bc1df2831717c2992d85b57e0cf3d123fd6c254257de5f784be369747b249"; got != want {
		t.Fatalf("grouped ciphertext with 3 handles = %s, want %s", got, want)
	}
	first, err := service.GroupedElGamalCiphertext3Handle(grouped3, 0)
	if err != nil {
		t.Fatal(err)
	}
	wantFirst, err := service.EncryptElGamalWithOpening(publicKeys[0], 42, opening)
	if err != nil {
		t.Fatal(err)
	}
	if first != wantFirst {
		t.Fatalf("first extracted ciphertext = %x, want %x", first, wantFirst)
	}
	if _, err := service.GroupedElGamalCiphertext2Handle(grouped2, 2); err == nil {
		t.Fatal("expected grouped ciphertext index error")
	}
	if _, err := service.GroupedElGamalCiphertext3FromBytes(grouped3[:127]); err == nil {
		t.Fatal("expected grouped ciphertext length error")
	}
	decoded, err := service.GroupedElGamalCiphertext3FromBytes(grouped3[:])
	if err != nil || decoded != grouped3 {
		t.Fatalf("decoded grouped ciphertext = %x, %v", decoded, err)
	}
}

func TestConfidentialEncryptionRoundTripAndValidation(t *testing.T) {
	service := ConfidentialTransferService{}
	keypair, err := service.GenerateElGamalKeypair()
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, _, err := service.EncryptElGamal(keypair.PublicKey, 12345)
	if err != nil {
		t.Fatal(err)
	}
	target, err := service.DecryptElGamalTarget(keypair.SecretKey, ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	encodedTarget, err := service.CommitPedersen(12345, PedersenOpening{})
	if err != nil {
		t.Fatal(err)
	}
	if target != encodedTarget {
		t.Fatalf("decryption target = %x, want %x", target, encodedTarget)
	}
	low, high, err := service.SplitAmount((uint64(1) << 48) - 1)
	if err != nil {
		t.Fatal(err)
	}
	if low != ^uint16(0) || high != ^uint32(0) {
		t.Fatalf("split amount = %d, %d", low, high)
	}
	if amount := service.CombineAmount(low, high); amount != (uint64(1)<<48)-1 {
		t.Fatalf("combined amount = %d", amount)
	}
	if _, _, err := service.SplitAmount(uint64(1) << 48); err == nil {
		t.Fatal("expected 48-bit amount error")
	}
	if _, err := service.ElGamalSecretKeyFromBytes(make([]byte, 32)); err == nil {
		t.Fatal("expected zero secret key error")
	}
	invalid := make([]byte, 32)
	for index := range invalid {
		invalid[index] = 0xff
	}
	if _, err := service.PedersenOpeningFromBytes(invalid); err == nil {
		t.Fatal("expected non-canonical opening error")
	}
	if _, err := service.ElGamalPubkeyFromBytes(invalid); err == nil {
		t.Fatal("expected invalid public key error")
	}
	invalidCiphertext := ElGamalCiphertext{}
	copy(invalidCiphertext[:32], invalid)
	if _, err := service.AddElGamalAmount(invalidCiphertext, 1); err == nil {
		t.Fatal("expected invalid ciphertext error")
	}
	if len(ElGamalCiphertext{}) != 64 || len(GroupedElGamalCiphertext2Handles{}) != 96 || len(GroupedElGamalCiphertext3Handles{}) != 128 {
		t.Fatal("unexpected confidential encryption wire size")
	}
}
