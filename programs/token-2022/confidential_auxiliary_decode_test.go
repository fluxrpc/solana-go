package token2022

import (
	"bytes"
	"testing"
)

func TestConfidentialAuxiliaryInstructionDecoders(t *testing.T) {
	service := ConfidentialTransferService{}
	service.Start()
	registry, decodedRegistry, err := service.DecodeElGamalRegistryInstruction(nil, []byte{1, 0xff})
	if err != nil {
		t.Fatal(err)
	}
	if registry.ProgramID() != service.RegistryProgramID || decodedRegistry.InstructionOffset != -1 {
		t.Fatalf("registry = %#v / %#v", registry, decodedRegistry)
	}

	write := service.WriteConfidentialRecord(token2022Key(1), token2022Key(2), 7, []byte{3, 4})
	data, _ := write.Data()
	record, decodedRecord, err := service.DecodeConfidentialRecordInstruction(write.Accounts(), data)
	if err != nil {
		t.Fatal(err)
	}
	if record.ProgramID() != service.RecordProgramID || decodedRecord.Offset != 7 || !bytes.Equal(decodedRecord.Data, []byte{3, 4}) {
		t.Fatalf("record = %#v / %#v", record, decodedRecord)
	}
	if _, _, err := service.DecodeConfidentialRecordInstruction(nil, append(data, 5)); err == nil {
		t.Fatal("decoded mismatched record data length")
	}
}
