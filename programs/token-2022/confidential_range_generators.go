package token2022

import (
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
	"golang.org/x/crypto/sha3"
)

type rangeProofGenerators struct {
	g []*curve.RistrettoPoint
	h []*curve.RistrettoPoint
}

type rangeVectorPolynomial struct {
	constant []*scalar.Scalar
	linear   []*scalar.Scalar
}

func (service ConfidentialTransferService) rangeGenerators(size int) (rangeProofGenerators, error) {
	g, err := service.rangeGeneratorChain("G", size)
	if err != nil {
		return rangeProofGenerators{}, err
	}
	h, err := service.rangeGeneratorChain("H", size)
	if err != nil {
		return rangeProofGenerators{}, err
	}
	return rangeProofGenerators{g: g, h: h}, nil
}

func (ConfidentialTransferService) rangeGeneratorChain(label string, size int) ([]*curve.RistrettoPoint, error) {
	shake := sha3.NewShake256()
	if _, err := shake.Write([]byte("GeneratorsChain")); err != nil {
		return nil, fmt.Errorf("range proof generators: %w", err)
	}
	if _, err := shake.Write([]byte(label)); err != nil {
		return nil, fmt.Errorf("range proof generators: %w", err)
	}
	points := make([]*curve.RistrettoPoint, size)
	for index := range points {
		uniform := make([]byte, 64)
		if _, err := shake.Read(uniform); err != nil {
			return nil, fmt.Errorf("range proof generators: %w", err)
		}
		point, err := curve.NewRistrettoPoint().SetUniformBytes(uniform)
		if err != nil {
			return nil, fmt.Errorf("range proof generators: %w", err)
		}
		points[index] = point
	}
	return points, nil
}

func (ConfidentialTransferService) newRangeVectorPolynomial(size int) rangeVectorPolynomial {
	polynomial := rangeVectorPolynomial{
		constant: make([]*scalar.Scalar, size),
		linear:   make([]*scalar.Scalar, size),
	}
	for index := 0; index < size; index++ {
		polynomial.constant[index] = scalar.New()
		polynomial.linear[index] = scalar.New()
	}
	return polynomial
}

func (service ConfidentialTransferService) rangeVectorPolynomialInnerProduct(left, right rangeVectorPolynomial) (*scalar.Scalar, *scalar.Scalar, *scalar.Scalar, error) {
	if len(left.constant) != len(right.constant) || len(left.linear) != len(right.linear) || len(left.constant) != len(left.linear) {
		return nil, nil, nil, fmt.Errorf("range proof polynomial length mismatch")
	}
	t0, err := service.scalarInnerProduct(left.constant, right.constant)
	if err != nil {
		return nil, nil, nil, err
	}
	t2, err := service.scalarInnerProduct(left.linear, right.linear)
	if err != nil {
		return nil, nil, nil, err
	}
	leftSum := make([]*scalar.Scalar, len(left.constant))
	rightSum := make([]*scalar.Scalar, len(right.constant))
	for index := range leftSum {
		leftSum[index] = scalar.New().Add(left.constant[index], left.linear[index])
		rightSum[index] = scalar.New().Add(right.constant[index], right.linear[index])
	}
	t1, err := service.scalarInnerProduct(leftSum, rightSum)
	if err != nil {
		return nil, nil, nil, err
	}
	t1.Sub(t1, t0)
	t1.Sub(t1, t2)
	return t0, t1, t2, nil
}

func (ConfidentialTransferService) evaluateRangeVectorPolynomial(polynomial rangeVectorPolynomial, x *scalar.Scalar) []*scalar.Scalar {
	values := make([]*scalar.Scalar, len(polynomial.constant))
	for index := range values {
		values[index] = scalar.New().Add(polynomial.constant[index], scalar.New().Mul(polynomial.linear[index], x))
	}
	return values
}

func (ConfidentialTransferService) scalarInnerProduct(left, right []*scalar.Scalar) (*scalar.Scalar, error) {
	if len(left) != len(right) {
		return nil, fmt.Errorf("range proof inner product length mismatch")
	}
	result := scalar.New()
	for index := range left {
		result.Add(result, scalar.New().Mul(left[index], right[index]))
	}
	return result, nil
}

func (ConfidentialTransferService) scalarPowers(value *scalar.Scalar, size int) []*scalar.Scalar {
	powers := make([]*scalar.Scalar, size)
	current := scalar.New().One()
	for index := range powers {
		powers[index] = scalar.New().Set(current)
		current.Mul(current, value)
	}
	return powers
}

func (service ConfidentialTransferService) scalarPowerSum(value *scalar.Scalar, size int) *scalar.Scalar {
	powers := service.scalarPowers(value, size)
	result := scalar.New()
	for _, power := range powers {
		result.Add(result, power)
	}
	return result
}

func (ConfidentialTransferService) cloneScalars(values []*scalar.Scalar) []*scalar.Scalar {
	cloned := make([]*scalar.Scalar, len(values))
	for index := range values {
		cloned[index] = scalar.New().Set(values[index])
	}
	return cloned
}

func (ConfidentialTransferService) clonePoints(values []*curve.RistrettoPoint) []*curve.RistrettoPoint {
	cloned := make([]*curve.RistrettoPoint, len(values))
	for index := range values {
		cloned[index] = curve.NewRistrettoPoint().Set(values[index])
	}
	return cloned
}
