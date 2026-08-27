package tokenmetadata

import "testing"

func TestFindMasterEditionAddress(t *testing.T) {
	mint := metadataKey(1)
	got, bump, err := FindMasterEditionAddress(mint)
	if err != nil {
		t.Fatal(err)
	}
	want, wantBump, err := ProgramID.FindProgramAddress(
		[][]byte{[]byte("metadata"), ProgramID[:], mint[:], []byte("edition")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got != want || bump != wantBump {
		t.Fatalf("address = %s/%d, want %s/%d", got, bump, want, wantBump)
	}
}
