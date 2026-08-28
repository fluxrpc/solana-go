package token2022

import (
	"bytes"
	"testing"
)

func TestZKProofInstructionFixtures(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	contextState, contextAuthority := token2022Key(1), token2022Key(2)
	proof := service.VerifyProof(ZKProofData{Discriminator: 4, Context: []byte{3, 5}, Proof: []byte{7, 11}}, &contextState, &contextAuthority)
	data, err := proof.Data()
	if err != nil {
		t.Fatal(err)
	}
	if proof.ProgramID() != service.ProofProgramID || !bytes.Equal(data, []byte{4, 3, 5, 7, 11}) {
		t.Fatalf("proof = %s %x", proof.ProgramID(), data)
	}
	if accounts := proof.Accounts(); len(accounts) != 2 || !accounts[0].IsWritable || accounts[1].IsWritable || accounts[1].IsSigner {
		t.Fatalf("proof accounts = %+v", accounts)
	}

	proofAccount := token2022Key(3)
	fromAccount := service.VerifyProofFromAccount(8, proofAccount, 0x12345678, &contextState, &contextAuthority)
	data, _ = fromAccount.Data()
	if !bytes.Equal(data, []byte{8, 0x78, 0x56, 0x34, 0x12}) || len(fromAccount.Accounts()) != 3 || fromAccount.Accounts()[0].PublicKey != proofAccount {
		t.Fatalf("from account = %x, %+v", data, fromAccount.Accounts())
	}

	closeInstruction := service.CloseProofContext(contextState, token2022Key(4), contextAuthority)
	data, _ = closeInstruction.Data()
	if !bytes.Equal(data, []byte{0}) || len(closeInstruction.Accounts()) != 3 || !closeInstruction.Accounts()[2].IsSigner {
		t.Fatalf("close = %x, %+v", data, closeInstruction.Accounts())
	}
}
