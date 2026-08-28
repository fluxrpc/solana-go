package token2022

import (
	"fmt"
	"math/bits"
)

func (service ConfidentialTransferService) CalculateConfidentialTransferFee(amount uint64, basisPoints uint16, maximumFee uint64) (uint64, uint64, error) {
	if basisPoints > 10_000 {
		return 0, 0, service.invalidInputError(fmt.Errorf("calculate confidential transfer fee: basis points exceed 10000"))
	}
	hi, lo := bits.Mul64(amount, uint64(basisPoints))
	lo, carry := bits.Add64(lo, 9_999, 0)
	hi += carry
	rawFee, _ := bits.Div64(hi, lo, 10_000)
	if maximumFee < rawFee {
		return maximumFee, 0, nil
	}
	feeHi, feeLo := bits.Mul64(rawFee, 10_000)
	numeratorHi, numeratorLo := bits.Mul64(amount, uint64(basisPoints))
	deltaLo, borrow := bits.Sub64(feeLo, numeratorLo, 0)
	deltaHi, _ := bits.Sub64(feeHi, numeratorHi, borrow)
	if deltaHi != 0 {
		return 0, 0, service.invalidInputError(fmt.Errorf("calculate confidential transfer fee: delta overflow"))
	}
	return rawFee, deltaLo, nil
}
