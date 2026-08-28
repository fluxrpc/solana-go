package token2022

import (
	"bytes"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestConfidentialMintBurnInstructionFixtures(t *testing.T) {
	service := ConfidentialTransferService{}
	mint, tokenAccount, authority := token2022Key(1), token2022Key(2), token2022Key(3)
	signer, context := token2022Key(4), token2022Key(5)
	var pubkey ElGamalPubkey
	var decryptable AeCiphertext
	var ciphertextLo, ciphertextHi ElGamalCiphertext
	for i := range pubkey {
		pubkey[i] = byte(i + 1)
	}
	for i := range decryptable {
		decryptable[i] = byte(i + 33)
	}
	for i := range ciphertextLo {
		ciphertextLo[i] = byte(i + 69)
		ciphertextHi[i] = byte(i + 133)
	}

	initialize := service.InitializeConfidentialMintBurn(mint, pubkey, decryptable)
	data, err := initialize.Data()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 70 || data[0] != byte(ConfidentialMintBurnExtensionInstruction) || data[1] != 0 || !bytes.Equal(data[2:34], pubkey[:]) || !bytes.Equal(data[34:], decryptable[:]) {
		t.Fatalf("initialize = %x", data)
	}

	rotate := service.RotateConfidentialSupplyKey(mint, authority, []solana.PublicKey{signer}, pubkey, ProofLocation{ContextState: context})
	data, _ = rotate.Data()
	if len(data) != 35 || data[1] != 1 || data[34] != 0 {
		t.Fatalf("rotate = %x", data)
	}
	if accounts := rotate.Accounts(); len(accounts) != 4 || accounts[1].PublicKey != context || accounts[2].IsSigner || !accounts[3].IsSigner {
		t.Fatalf("rotate accounts = %+v", accounts)
	}

	update := service.UpdateConfidentialSupply(mint, authority, nil, decryptable)
	data, _ = update.Data()
	if len(data) != 38 || data[1] != 2 || !bytes.Equal(data[2:], decryptable[:]) || !update.Accounts()[1].IsSigner {
		t.Fatalf("update = %x, %+v", data, update.Accounts())
	}

	equality := ProofLocation{InstructionOffset: 1}
	validity := ProofLocation{ContextState: context}
	rangeProof := ProofLocation{InstructionOffset: 2}
	mintInstruction := service.ConfidentialMint(tokenAccount, mint, authority, nil, decryptable, ciphertextLo, ciphertextHi, equality, validity, rangeProof)
	data, _ = mintInstruction.Data()
	if len(data) != 169 || data[1] != 3 || !bytes.Equal(data[2:38], decryptable[:]) || data[166] != 1 || data[167] != 0 || data[168] != 2 {
		t.Fatalf("mint = %x", data)
	}
	if accounts := mintInstruction.Accounts(); len(accounts) != 5 || accounts[2].PublicKey != solana.SysVarInstructionsPubkey || accounts[3].PublicKey != context || !accounts[4].IsSigner {
		t.Fatalf("mint accounts = %+v", accounts)
	}

	burn := service.ConfidentialBurn(tokenAccount, mint, authority, nil, decryptable, ciphertextLo, ciphertextHi, equality, validity, rangeProof)
	data, _ = burn.Data()
	if len(data) != 169 || data[1] != 4 {
		t.Fatalf("burn = %x", data)
	}

	apply := service.ApplyPendingConfidentialBurn(mint, authority, nil)
	data, _ = apply.Data()
	if !bytes.Equal(data, []byte{byte(ConfidentialMintBurnExtensionInstruction), 5}) || len(apply.Accounts()) != 2 || !apply.Accounts()[1].IsSigner {
		t.Fatalf("apply = %x, %+v", data, apply.Accounts())
	}

	instructions := []*ConfidentialMintBurnExtension{initialize, rotate, update, mintInstruction, burn, apply}
	for subInstruction, instruction := range instructions {
		data, _ := instruction.Data()
		decoded, err := DecodeInstruction(instruction.Accounts(), data)
		if err != nil {
			t.Fatalf("decode %d: %v", subInstruction, err)
		}
		if decoded.ConfidentialMintBurn == nil || decoded.ConfidentialMintBurn.Decoded == nil {
			t.Fatalf("decode %d = %+v", subInstruction, decoded)
		}
	}
	mintData, _ := mintInstruction.Data()
	decoded, _ := DecodeInstruction(mintInstruction.Accounts(), mintData)
	if payload := decoded.ConfidentialMintBurn.Decoded.Mint; payload == nil || payload.NewDecryptableSupply != decryptable || payload.MintAmountAuditorCiphertextLo != ciphertextLo || payload.MintAmountAuditorCiphertextHi != ciphertextHi || payload.EqualityProofInstructionOffset != 1 || payload.CiphertextValidityProofInstructionOffset != 0 || payload.RangeProofInstructionOffset != 2 {
		t.Fatalf("decoded mint = %+v", payload)
	}
}
