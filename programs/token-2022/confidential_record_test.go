package token2022

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestConfidentialProofRecordInstructions(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	proof := []byte{1, 2, 3, 4, 5}
	plan, err := service.ConfidentialProofRecordInstructions(token2022Key(1), token2022Key(2), token2022Key(3), token2022Key(4), 50, proof, 2)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Space != 38 || len(plan.CreateAndWrite) != 5 {
		t.Fatalf("plan = %#v", plan)
	}
	for i, expected := range [][]byte{{1, 2}, {3, 4}, {5}} {
		instruction := plan.CreateAndWrite[i+2]
		data, err := instruction.Data()
		if err != nil {
			t.Fatal(err)
		}
		if data[0] != 1 || binary.LittleEndian.Uint64(data[1:]) != uint64(i*2) || binary.LittleEndian.Uint32(data[9:]) != uint32(len(expected)) || !bytes.Equal(data[13:], expected) {
			t.Fatalf("write %d = %x", i, data)
		}
		accounts := instruction.Accounts()
		if len(accounts) != 2 || accounts[0].PublicKey != token2022Key(2) || !accounts[0].IsWritable || accounts[0].IsSigner || accounts[1].PublicKey != token2022Key(3) || accounts[1].IsWritable || !accounts[1].IsSigner {
			t.Fatalf("write %d accounts = %#v", i, accounts)
		}
	}
	largeChunkPlan, err := service.ConfidentialProofRecordInstructions(token2022Key(1), token2022Key(2), token2022Key(3), token2022Key(4), 50, proof, int(^uint(0)>>1))
	if err != nil || len(largeChunkPlan.CreateAndWrite) != 3 {
		t.Fatalf("large chunk plan = %#v, %v", largeChunkPlan, err)
	}
	verify := service.VerifyProofFromRecord(7, token2022Key(2), nil, nil)
	data, _ := verify.Data()
	if data[0] != 7 || binary.LittleEndian.Uint32(data[1:]) != 33 {
		t.Fatalf("verify data = %x", data)
	}
	typed := ZKProofData{Discriminator: 4, Context: make([]byte, 32), Proof: make([]byte, 64)}
	typedPlan, err := service.ConfidentialProofRecordInstructionsForProof(token2022Key(1), token2022Key(2), token2022Key(3), token2022Key(4), 50, typed, 128)
	if err != nil || typedPlan.Space != 129 || len(typedPlan.CreateAndWrite) != 3 {
		t.Fatalf("typed proof plan = %#v, %v", typedPlan, err)
	}
}

func TestConfidentialRecordInstructionsAndState(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	record, authority, destination := token2022Key(1), token2022Key(2), token2022Key(3)
	tests := []struct {
		name string
		inst *ConfidentialAuxiliaryInstruction
		tag  byte
	}{
		{name: "initialize", inst: service.InitializeConfidentialRecord(record, authority), tag: 0},
		{name: "authority", inst: service.SetConfidentialRecordAuthority(record, authority, destination), tag: 2},
		{name: "close", inst: service.CloseConfidentialRecord(record, authority, destination), tag: 3},
		{name: "reallocate", inst: service.ReallocateConfidentialRecord(record, authority, 99), tag: 4},
	}
	for _, test := range tests {
		data, err := test.inst.Data()
		if err != nil {
			t.Fatal(err)
		}
		if test.inst.ProgramID() != service.RecordProgramID || data[0] != test.tag {
			t.Fatalf("%s = %s/%x", test.name, test.inst.ProgramID(), data)
		}
	}
	initializeAccounts := tests[0].inst.Accounts()
	if len(initializeAccounts) != 2 || !initializeAccounts[0].IsWritable || initializeAccounts[0].IsSigner || initializeAccounts[1].IsWritable || initializeAccounts[1].IsSigner {
		t.Fatalf("initialize accounts = %#v", initializeAccounts)
	}
	authorityAccounts := tests[1].inst.Accounts()
	if len(authorityAccounts) != 3 || !authorityAccounts[0].IsWritable || !authorityAccounts[1].IsSigner || authorityAccounts[1].IsWritable || authorityAccounts[2].IsWritable || authorityAccounts[2].IsSigner {
		t.Fatalf("authority accounts = %#v", authorityAccounts)
	}
	closeAccounts := tests[2].inst.Accounts()
	if len(closeAccounts) != 3 || !closeAccounts[0].IsWritable || !closeAccounts[1].IsSigner || closeAccounts[1].IsWritable || !closeAccounts[2].IsWritable || closeAccounts[2].IsSigner {
		t.Fatalf("close accounts = %#v", closeAccounts)
	}
	reallocateAccounts := tests[3].inst.Accounts()
	if len(reallocateAccounts) != 2 || !reallocateAccounts[0].IsWritable || !reallocateAccounts[1].IsSigner || reallocateAccounts[1].IsWritable {
		t.Fatalf("reallocate accounts = %#v", reallocateAccounts)
	}
	data := make([]byte, 36)
	data[0] = 1
	copy(data[1:], authority[:])
	copy(data[33:], []byte{8, 9, 10})
	state, err := service.DecodeConfidentialRecord(data)
	if err != nil {
		t.Fatal(err)
	}
	if state.Version != 1 || state.Authority != authority || !bytes.Equal(state.Data, []byte{8, 9, 10}) {
		t.Fatalf("state = %#v", state)
	}
	if _, err := service.DecodeConfidentialRecord(data[:32]); err == nil {
		t.Fatal("expected short state error")
	}
}
