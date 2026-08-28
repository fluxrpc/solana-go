package token2022

import (
	"testing"

	system "github.com/fluxrpc/solana-go/programs/system"
)

func TestProofContextPlan(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	payer, contextState, authority, destination := token2022Key(1), token2022Key(2), token2022Key(3), token2022Key(4)
	plan, err := service.ProofContextInstructions(payer, contextState, authority, destination, 123, ZKProofData{Discriminator: 12, Context: make([]byte, 352), Proof: make([]byte, 192)})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.CreateAndVerify) != 2 || plan.Close == nil {
		t.Fatalf("plan = %+v", plan)
	}
	create, ok := plan.CreateAndVerify[0].(*system.CreateAccount)
	if !ok || create.Space != 385 || create.Owner != service.ProofProgramID || create.Lamports != 123 {
		t.Fatalf("create = %+v", plan.CreateAndVerify[0])
	}
	verifyData, _ := plan.CreateAndVerify[1].Data()
	if len(verifyData) != 545 || verifyData[0] != 12 {
		t.Fatalf("verify = %x", verifyData)
	}
	if _, err := service.ProofContextInstructions(payer, contextState, authority, destination, 123, ZKProofData{Discriminator: 13}); err == nil {
		t.Fatal("expected discriminator error")
	}
	fromRecord, err := service.ProofContextFromRecordInstructions(payer, contextState, authority, destination, token2022Key(5), 321, 8)
	if err != nil {
		t.Fatal(err)
	}
	verifyData, _ = fromRecord.CreateAndVerify[1].Data()
	if len(fromRecord.CreateAndVerify) != 2 || verifyData[0] != 8 || verifyData[1] != 33 {
		t.Fatalf("record context plan = %+v / %x", fromRecord, verifyData)
	}
}
