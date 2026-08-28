package token2022

import (
	"bytes"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestConfidentialTransferFeeTypedFixtures(t *testing.T) {
	service := ConfidentialTransferService{}
	mint, destination, authority := token2022Key(1), token2022Key(2), token2022Key(3)
	withdrawKey := ElGamalPubkey{}
	decryptable := AeCiphertext{}
	for i := range withdrawKey {
		withdrawKey[i] = 0x80 + byte(i)
	}
	for i := range decryptable {
		decryptable[i] = 0x40 + byte(i)
	}
	initialize := service.InitializeConfidentialTransferFee(mint, &authority, withdrawKey)
	data, err := initialize.Data()
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{byte(ConfidentialTransferFeeExtensionInstruction), 0}
	want = append(want, authority[:]...)
	want = append(want, withdrawKey[:]...)
	if !bytes.Equal(data, want) {
		t.Fatalf("initialize data = %x, want %x", data, want)
	}
	decoded, err := DecodeInstruction(initialize.Accounts(), data)
	if err != nil || decoded.ConfidentialTransferFee.Decoded.InitializeConfig == nil {
		t.Fatalf("decode initialize = %+v, %v", decoded, err)
	}

	proof := ProofLocation{ContextState: token2022Key(4)}
	withdraw := service.WithdrawConfidentialWithheldTokensFromAccounts(
		mint, destination, authority, []solana.PublicKey{token2022Key(5)},
		[]solana.PublicKey{token2022Key(6), token2022Key(7)}, decryptable, proof,
	)
	data, err = withdraw.Data()
	if err != nil {
		t.Fatal(err)
	}
	want = []byte{byte(ConfidentialTransferFeeExtensionInstruction), 2, 2, 0}
	want = append(want, decryptable[:]...)
	if !bytes.Equal(data, want) {
		t.Fatalf("withdraw data = %x, want %x", data, want)
	}
	decoded, err = DecodeInstruction(withdraw.Accounts(), data)
	if err != nil || decoded.ConfidentialTransferFee.Decoded.WithdrawWithheldTokensFromAccounts == nil {
		t.Fatalf("decode withdraw = %+v, %v", decoded, err)
	}
	decodedWithdraw := decoded.ConfidentialTransferFee.Decoded.WithdrawWithheldTokensFromAccounts
	if decodedWithdraw.NumTokenAccounts != 2 || decodedWithdraw.ProofInstructionOffset != 0 || decodedWithdraw.NewDecryptableAvailableBalance != decryptable {
		t.Fatalf("decoded withdraw = %+v", decodedWithdraw)
	}
	wantAccounts := []solana.AccountMeta{
		{PublicKey: mint, IsWritable: true},
		{PublicKey: destination, IsWritable: true},
		{PublicKey: token2022Key(4)},
		{PublicKey: authority},
		{PublicKey: token2022Key(5), IsSigner: true},
		{PublicKey: token2022Key(6), IsWritable: true},
		{PublicKey: token2022Key(7), IsWritable: true},
	}
	if len(withdraw.Accounts()) != len(wantAccounts) {
		t.Fatalf("withdraw accounts = %+v", withdraw.Accounts())
	}
	for i, account := range withdraw.Accounts() {
		if *account != wantAccounts[i] {
			t.Fatalf("withdraw account %d = %+v, want %+v", i, account, wantAccounts[i])
		}
	}
}

func TestConfidentialTransferFeeTypedVariants(t *testing.T) {
	service := ConfidentialTransferService{}
	k1, k2, k3 := token2022Key(1), token2022Key(2), token2022Key(3)
	decryptable := AeCiphertext{}
	proof := ProofLocation{InstructionOffset: 1}
	tests := []struct {
		name        string
		instruction *ConfidentialTransferFeeExtension
		sub         byte
		payloadLen  int
	}{
		{"withdraw mint", service.WithdrawConfidentialWithheldTokensFromMint(k1, k2, k3, nil, decryptable, proof), 1, 37},
		{"harvest", service.HarvestConfidentialWithheldTokensToMint(k1, []solana.PublicKey{k2}), 3, 0},
		{"enable", service.EnableConfidentialHarvestToMint(k1, k3, nil), 4, 0},
		{"disable", service.DisableConfidentialHarvestToMint(k1, k3, nil), 5, 0},
	}
	for _, test := range tests {
		data, err := test.instruction.Data()
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		if len(data) != test.payloadLen+2 || data[0] != byte(ConfidentialTransferFeeExtensionInstruction) || data[1] != test.sub {
			t.Fatalf("%s data = %x", test.name, data)
		}
		decoded, err := DecodeInstruction(test.instruction.Accounts(), data)
		if err != nil || decoded.ConfidentialTransferFee == nil || decoded.ConfidentialTransferFee.Decoded == nil {
			t.Fatalf("%s decode = %+v, %v", test.name, decoded, err)
		}
	}
	accounts := tests[0].instruction.Accounts()
	wantAccounts := []solana.AccountMeta{
		{PublicKey: k1, IsWritable: true},
		{PublicKey: k2, IsWritable: true},
		{PublicKey: solana.SysVarInstructionsPubkey},
		{PublicKey: k3, IsSigner: true},
	}
	if len(accounts) != len(wantAccounts) {
		t.Fatalf("withdraw mint accounts = %+v", accounts)
	}
	for i, account := range accounts {
		if *account != wantAccounts[i] {
			t.Fatalf("withdraw mint account %d = %+v, want %+v", i, account, wantAccounts[i])
		}
	}
}
