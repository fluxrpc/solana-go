package benchcmp

import "testing"

func TestComputeInstructionParity(t *testing.T) {
	fixtures := newComputeInstructionBenchmarks(t)
	if len(fixtures) != 5 {
		t.Fatalf("parity fixture count = %d, want 5", len(fixtures))
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			stakeParityInstruction(t, fixture.flux, fixture.foundation)
		})
	}
}
