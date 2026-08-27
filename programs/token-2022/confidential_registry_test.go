package token2022

import (
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestElGamalRegistryInstructions(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	owner := token2022Key(1)
	context := token2022Key(2)
	registry, bump, err := service.ElGamalRegistryAddress(owner)
	if err != nil {
		t.Fatal(err)
	}
	if registry.String() != "5Ybk5Qnx5vDkeCtqExCMJBUUE3JHE9x1fLE7hgeaLGJ6" || bump != 255 {
		t.Fatalf("registry = %s/%d", registry, bump)
	}
	proof := &ZKProofData{Discriminator: 4, Context: make([]byte, 32), Proof: make([]byte, 64)}
	create, err := service.CreateElGamalRegistry(owner, ProofLocation{InstructionOffset: 1, Proof: proof})
	if err != nil {
		t.Fatal(err)
	}
	if len(create) != 2 || create[0].ProgramID() != service.RegistryProgramID || create[1].ProgramID() != service.ProofProgramID {
		t.Fatalf("create = %#v", create)
	}
	data, _ := create[0].Data()
	if len(data) != 2 || data[0] != 0 || data[1] != 1 {
		t.Fatalf("create data = %x", data)
	}
	accounts := create[0].Accounts()
	if len(accounts) != 4 || accounts[0].PublicKey != registry || !accounts[0].IsWritable || accounts[0].IsSigner || accounts[1].PublicKey != owner || accounts[1].IsWritable || !accounts[1].IsSigner || accounts[2].PublicKey != solana.SystemProgramID || accounts[2].IsWritable || accounts[2].IsSigner || accounts[3].PublicKey != solana.SysVarInstructionsPubkey || accounts[3].IsWritable || accounts[3].IsSigner {
		t.Fatalf("create accounts = %#v", accounts)
	}
	proofData, _ := create[1].Data()
	if len(proofData) != 97 || proofData[0] != 4 {
		t.Fatalf("proof data = %x", proofData)
	}
	update, err := service.UpdateElGamalRegistry(owner, ProofLocation{ContextState: context})
	if err != nil {
		t.Fatal(err)
	}
	if len(update) != 1 {
		t.Fatalf("update instructions = %#v", update)
	}
	accounts = update[0].Accounts()
	if len(accounts) != 3 || accounts[0].PublicKey != registry || !accounts[0].IsWritable || accounts[0].IsSigner || accounts[1].PublicKey != context || accounts[1].IsWritable || accounts[1].IsSigner || accounts[2].PublicKey != owner || accounts[2].IsWritable || !accounts[2].IsSigner {
		t.Fatalf("update accounts = %#v", accounts)
	}
	if _, err := service.CreateElGamalRegistry(owner, ProofLocation{InstructionOffset: 2, Proof: proof}); err == nil {
		t.Fatal("expected inline offset error")
	}
	invalidProof := &ZKProofData{Discriminator: 1}
	if _, err := service.CreateElGamalRegistry(owner, ProofLocation{InstructionOffset: 1, Proof: invalidProof}); err == nil {
		t.Fatal("expected proof type error")
	}
}

func TestDecodeElGamalRegistry(t *testing.T) {
	service := ConfidentialTransferService{}
	data := make([]byte, 64)
	owner := token2022Key(3)
	copy(data, owner[:])
	for i := 32; i < len(data); i++ {
		data[i] = byte(i)
	}
	state, err := service.DecodeElGamalRegistry(data)
	if err != nil {
		t.Fatal(err)
	}
	if state.Owner != owner || state.ElGamalPubkey[0] != 32 || state.ElGamalPubkey[31] != 63 {
		t.Fatalf("state = %#v", state)
	}
	if _, err := service.DecodeElGamalRegistry(data[:63]); err == nil {
		t.Fatal("expected size error")
	}
}
