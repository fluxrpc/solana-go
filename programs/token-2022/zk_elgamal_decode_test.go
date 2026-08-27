package token2022

import (
	"encoding/binary"
	"testing"

	solana "github.com/fluxrpc/solana-go"
)

func TestZKProofInstructionDecoding(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	fixtures := []struct {
		discriminator uint8
		contextSize   int
		proofSize     int
	}{
		{1, 96, 96},
		{2, 192, 224},
		{3, 128, 192},
		{4, 32, 64},
		{5, 104, 256},
		{6, 264, 672},
		{7, 264, 736},
		{8, 264, 800},
		{9, 160, 160},
		{10, 256, 160},
		{11, 224, 192},
		{12, 352, 192},
	}
	for _, fixture := range fixtures {
		direct := make([]byte, 1+fixture.contextSize+fixture.proofSize)
		direct[0] = fixture.discriminator
		instruction, err := service.DecodeZKProofInstruction(nil, direct)
		if err != nil {
			t.Fatalf("decode direct %d: %v", fixture.discriminator, err)
		}
		if instruction.Decoded == nil || instruction.Decoded.Discriminator != fixture.discriminator || instruction.Decoded.Context == nil || len(instruction.Decoded.RawContext) != fixture.contextSize || len(instruction.Decoded.Proof) != fixture.proofSize {
			t.Fatalf("decoded direct %d = %+v", fixture.discriminator, instruction.Decoded)
		}
		if _, err := service.DecodeZKProofInstruction(nil, direct[:len(direct)-1]); err == nil {
			t.Fatalf("decode short direct %d succeeded", fixture.discriminator)
		}

		fromAccount := []byte{fixture.discriminator, 0x78, 0x56, 0x34, 0x12}
		instruction, err = service.DecodeZKProofInstruction(solana.AccountMetaSlice{}, fromAccount)
		if err != nil {
			t.Fatalf("decode account %d: %v", fixture.discriminator, err)
		}
		if instruction.Decoded.ProofAccountOffset == nil || *instruction.Decoded.ProofAccountOffset != 0x12345678 || instruction.Decoded.Context != nil || len(instruction.Decoded.Proof) != 0 {
			t.Fatalf("decoded account %d = %+v", fixture.discriminator, instruction.Decoded)
		}

		stateData := make([]byte, 33+fixture.contextSize)
		stateData[32] = fixture.discriminator
		state, err := service.DecodeProofContextState(stateData)
		if err != nil {
			t.Fatalf("decode state %d: %v", fixture.discriminator, err)
		}
		if state.ProofType != fixture.discriminator || state.Context == nil || len(state.RawContext) != fixture.contextSize {
			t.Fatalf("decoded state %d = %+v", fixture.discriminator, state)
		}
		if _, err := service.DecodeProofContextState(stateData[:len(stateData)-1]); err == nil {
			t.Fatalf("decode short state %d succeeded", fixture.discriminator)
		}
	}

	closeInstruction, err := service.DecodeZKProofInstruction(nil, []byte{0})
	if err != nil || closeInstruction.Decoded == nil || closeInstruction.Decoded.Discriminator != 0 {
		t.Fatalf("decode close = %+v, %v", closeInstruction, err)
	}
	if _, err := service.DecodeZKProofInstruction(nil, []byte{0, 0}); err == nil {
		t.Fatal("decode oversized close succeeded")
	}
	if _, err := service.DecodeZKProofInstruction(nil, []byte{13}); err == nil {
		t.Fatal("decode unknown proof succeeded")
	}
	if _, err := service.DecodeProofContextState(make([]byte, 32)); err == nil {
		t.Fatal("decode short state succeeded")
	}
	uninitialized, err := service.DecodeProofContextState(make([]byte, 33))
	if err != nil || uninitialized.ProofType != 0 || uninitialized.Context != nil {
		t.Fatalf("decode uninitialized = %+v, %v", uninitialized, err)
	}
	uninitialized, err = service.DecodeProofContextState(make([]byte, 385))
	if err != nil || len(uninitialized.RawContext) != 352 {
		t.Fatalf("decode allocated uninitialized = %+v, %v", uninitialized, err)
	}
}

func TestZKProofTypedContextFixtures(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	percentage := make([]byte, 1+104+256)
	percentage[0] = 5
	for i := 0; i < 96; i++ {
		percentage[1+i] = byte(i + 1)
	}
	binary.LittleEndian.PutUint64(percentage[97:105], 0x123456789abcdef0)
	instruction, err := service.DecodeZKProofInstruction(nil, percentage)
	if err != nil {
		t.Fatal(err)
	}
	context := instruction.Decoded.Context.PercentageWithCap
	if context == nil || context.PercentageCommitment[0] != 1 || context.DeltaCommitment[0] != 33 || context.ClaimedCommitment[0] != 65 || context.MaxValue != 0x123456789abcdef0 {
		t.Fatalf("percentage context = %+v", context)
	}

	rangeProof := make([]byte, 1+264+672)
	rangeProof[0] = 6
	copy(rangeProof[257:265], []byte{1, 2, 3, 4, 5, 6, 7, 8})
	instruction, err = service.DecodeZKProofInstruction(nil, rangeProof)
	if err != nil {
		t.Fatal(err)
	}
	if context := instruction.Decoded.Context.BatchedRange; context == nil || context.BitLengths != [8]uint8{1, 2, 3, 4, 5, 6, 7, 8} {
		t.Fatalf("range context = %+v", context)
	}
}
