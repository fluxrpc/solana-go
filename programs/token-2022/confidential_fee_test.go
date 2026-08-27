package token2022

import (
	"math"
	"testing"
)

func TestCalculateConfidentialTransferFee(t *testing.T) {
	service := ConfidentialTransferService{}
	tests := []struct {
		amount       uint64
		basisPoints  uint16
		maximumFee   uint64
		fee          uint64
		claimedDelta uint64
	}{
		{amount: 0, basisPoints: 100, maximumFee: 10, fee: 0, claimedDelta: 0},
		{amount: 1, basisPoints: 1, maximumFee: 10, fee: 1, claimedDelta: 9_999},
		{amount: 10_001, basisPoints: 100, maximumFee: 1_000, fee: 101, claimedDelta: 9_900},
		{amount: 10_001, basisPoints: 100, maximumFee: 50, fee: 50, claimedDelta: 0},
		{amount: math.MaxUint64, basisPoints: 10_000, maximumFee: math.MaxUint64, fee: math.MaxUint64, claimedDelta: 0},
	}
	for _, test := range tests {
		fee, delta, err := service.CalculateConfidentialTransferFee(test.amount, test.basisPoints, test.maximumFee)
		if err != nil {
			t.Fatal(err)
		}
		if fee != test.fee || delta != test.claimedDelta {
			t.Fatalf("fee(%d,%d,%d) = %d/%d, want %d/%d", test.amount, test.basisPoints, test.maximumFee, fee, delta, test.fee, test.claimedDelta)
		}
	}
	if _, _, err := service.CalculateConfidentialTransferFee(1, 10_001, 1); err == nil {
		t.Fatal("expected basis-point error")
	}
}
