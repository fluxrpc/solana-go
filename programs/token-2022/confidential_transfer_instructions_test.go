package token2022

import (
	"bytes"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestConfidentialTransferTypedDataFixtures(t *testing.T) {
	service := ConfidentialTransferService{}
	mint, authority := token2022Key(1), token2022Key(2)
	auditor := ElGamalPubkey{}
	for i := range auditor {
		auditor[i] = byte(i + 1)
	}
	data, err := service.InitializeConfidentialTransferMint(mint, &authority, true, &auditor).Data()
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{byte(ConfidentialTransferExtensionInstruction), 0}
	want = append(want, authority[:]...)
	want = append(want, 1)
	want = append(want, auditor[:]...)
	if !bytes.Equal(data, want) {
		t.Fatalf("initialize data = %x, want %x", data, want)
	}
	decoded, err := DecodeInstruction(service.InitializeConfidentialTransferMint(mint, &authority, true, &auditor).Accounts(), data)
	if err != nil || decoded.ConfidentialTransfer.Decoded.InitializeMint == nil {
		t.Fatalf("decode initialize = %+v, %v", decoded, err)
	}

	decryptable := AeCiphertext{}
	lo, hi := ElGamalCiphertext{}, ElGamalCiphertext{}
	for i := range decryptable {
		decryptable[i] = 0xa0 + byte(i)
	}
	for i := range lo {
		lo[i], hi[i] = byte(i), 0xff-byte(i)
	}
	transfer := service.ConfidentialTransfer(
		token2022Key(3), mint, token2022Key(4), authority, []solana.PublicKey{token2022Key(5)},
		decryptable, lo, hi,
		ProofLocation{InstructionOffset: 1},
		ProofLocation{ContextState: token2022Key(6)},
		ProofLocation{InstructionOffset: 2},
	)
	data, err = transfer.Data()
	if err != nil {
		t.Fatal(err)
	}
	want = []byte{byte(ConfidentialTransferExtensionInstruction), 7}
	want = append(want, decryptable[:]...)
	want = append(want, lo[:]...)
	want = append(want, hi[:]...)
	want = append(want, 1, 0, 2)
	if !bytes.Equal(data, want) {
		t.Fatalf("transfer data = %x, want %x", data, want)
	}
	decoded, err = DecodeInstruction(transfer.Accounts(), data)
	if err != nil || decoded.ConfidentialTransfer.Decoded.Transfer == nil {
		t.Fatalf("decode transfer = %+v, %v", decoded, err)
	}
	decodedTransfer := decoded.ConfidentialTransfer.Decoded.Transfer
	if decodedTransfer.NewSourceDecryptableAvailableBalance != decryptable || decodedTransfer.TransferAmountAuditorCiphertextLo != lo || decodedTransfer.TransferAmountAuditorCiphertextHi != hi || decodedTransfer.EqualityProofInstructionOffset != 1 || decodedTransfer.CiphertextValidityProofInstructionOffset != 0 || decodedTransfer.RangeProofInstructionOffset != 2 {
		t.Fatalf("decoded transfer = %+v", decodedTransfer)
	}
	wantAccounts := []solana.AccountMeta{
		{PublicKey: token2022Key(3), IsWritable: true},
		{PublicKey: mint},
		{PublicKey: token2022Key(4), IsWritable: true},
		{PublicKey: solana.SysVarInstructionsPubkey},
		{PublicKey: token2022Key(6)},
		{PublicKey: authority},
		{PublicKey: token2022Key(5), IsSigner: true},
	}
	if len(transfer.Accounts()) != len(wantAccounts) {
		t.Fatalf("transfer accounts = %+v", transfer.Accounts())
	}
	for i, account := range transfer.Accounts() {
		if *account != wantAccounts[i] {
			t.Fatalf("transfer account %d = %+v, want %+v", i, account, wantAccounts[i])
		}
	}

	transferWithFee := service.ConfidentialTransferWithFee(
		token2022Key(3), mint, token2022Key(4), authority, nil, decryptable, lo, hi,
		ProofLocation{InstructionOffset: 1},
		ProofLocation{ContextState: token2022Key(7)},
		ProofLocation{InstructionOffset: 2},
		ProofLocation{ContextState: token2022Key(8)},
		ProofLocation{InstructionOffset: 3},
	)
	data, err = transferWithFee.Data()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 171 || !bytes.Equal(data[len(data)-5:], []byte{1, 0, 2, 0, 3}) {
		t.Fatalf("transfer with fee data = %x", data)
	}
	wantAccounts = []solana.AccountMeta{
		{PublicKey: token2022Key(3), IsWritable: true},
		{PublicKey: mint},
		{PublicKey: token2022Key(4), IsWritable: true},
		{PublicKey: solana.SysVarInstructionsPubkey},
		{PublicKey: token2022Key(7)},
		{PublicKey: token2022Key(8)},
		{PublicKey: authority, IsSigner: true},
	}
	for i, account := range transferWithFee.Accounts() {
		if *account != wantAccounts[i] {
			t.Fatalf("transfer with fee account %d = %+v, want %+v", i, account, wantAccounts[i])
		}
	}
}

func TestConfidentialTransferTypedVariantFixtures(t *testing.T) {
	service := ConfidentialTransferService{}
	k1, k2, k3, k4 := token2022Key(1), token2022Key(2), token2022Key(3), token2022Key(4)
	decryptable := AeCiphertext{}
	lo, hi := ElGamalCiphertext{}, ElGamalCiphertext{}
	proof := ProofLocation{ContextState: k4}
	payer := token2022Key(5)
	tests := []struct {
		name        string
		instruction *ConfidentialTransferExtension
		sub         byte
		payloadLen  int
	}{
		{"update", service.UpdateConfidentialTransferMint(k1, k2, nil, false, nil), 1, 33},
		{"configure", service.ConfigureConfidentialTransferAccount(k1, k2, k3, nil, decryptable, 7, proof), 2, 45},
		{"approve", service.ApproveConfidentialTransferAccount(k1, k2, k3, nil), 3, 0},
		{"empty", service.EmptyConfidentialTransferAccount(k1, k2, nil, proof), 4, 1},
		{"deposit", service.ConfidentialDeposit(k1, k2, k3, nil, 9, 2), 5, 9},
		{"withdraw", service.ConfidentialWithdraw(k1, k2, k3, nil, 9, 2, decryptable, proof, proof), 6, 47},
		{"apply", service.ApplyConfidentialPendingBalance(k1, k2, nil, 7, decryptable), 8, 44},
		{"enable confidential", service.EnableConfidentialCredits(k1, k2, nil), 9, 0},
		{"disable confidential", service.DisableConfidentialCredits(k1, k2, nil), 10, 0},
		{"enable non-confidential", service.EnableNonConfidentialCredits(k1, k2, nil), 11, 0},
		{"disable non-confidential", service.DisableNonConfidentialCredits(k1, k2, nil), 12, 0},
		{"with fee", service.ConfidentialTransferWithFee(k1, k2, k3, k4, nil, decryptable, lo, hi, proof, proof, proof, proof, proof), 13, 169},
		{"registry", service.ConfigureConfidentialTransferAccountWithRegistry(k1, k2, k3, &payer), 14, 0},
	}
	for _, test := range tests {
		data, err := test.instruction.Data()
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		if len(data) != test.payloadLen+2 || data[0] != byte(ConfidentialTransferExtensionInstruction) || data[1] != test.sub {
			t.Fatalf("%s data = %x", test.name, data)
		}
		decoded, err := DecodeInstruction(test.instruction.Accounts(), data)
		if err != nil || decoded.ConfidentialTransfer == nil || decoded.ConfidentialTransfer.Decoded == nil {
			t.Fatalf("%s decode = %+v, %v", test.name, decoded, err)
		}
	}

	registry := tests[len(tests)-1].instruction
	wantAccounts := []solana.AccountMeta{
		{PublicKey: k1, IsWritable: true},
		{PublicKey: k2},
		{PublicKey: k3},
		{PublicKey: payer, IsWritable: true, IsSigner: true},
		{PublicKey: solana.SystemProgramID},
	}
	for i, account := range registry.Accounts() {
		if *account != wantAccounts[i] {
			t.Fatalf("registry account %d = %+v, want %+v", i, account, wantAccounts[i])
		}
	}
}
