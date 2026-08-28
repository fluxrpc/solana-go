package token2022

import (
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

func (service ConfidentialTransferService) ScaleElGamalCiphertext(ciphertext ElGamalCiphertext, multiplier uint64) (ElGamalCiphertext, error) {
	commitment, handle, err := service.elGamalCiphertextPoints(ciphertext)
	if err != nil {
		return ElGamalCiphertext{}, err
	}
	factor := scalar.New().SetUint64(multiplier)
	return service.elGamalCiphertext(
		curve.NewRistrettoPoint().Mul(commitment, factor),
		curve.NewRistrettoPoint().Mul(handle, factor),
	), nil
}

func (service ConfidentialTransferService) CombineElGamalAmountCiphertexts(low, high ElGamalCiphertext) (ElGamalCiphertext, error) {
	scaled, err := service.ScaleElGamalCiphertext(high, 1<<16)
	if err != nil {
		return ElGamalCiphertext{}, err
	}
	return service.AddElGamalCiphertexts(low, scaled)
}

func (service ConfidentialTransferService) AddPedersenCommitments(left, right PedersenCommitment) (PedersenCommitment, error) {
	return service.combinePedersenCommitments(left, right, false)
}

func (service ConfidentialTransferService) SubtractPedersenCommitments(left, right PedersenCommitment) (PedersenCommitment, error) {
	return service.combinePedersenCommitments(left, right, true)
}

func (service ConfidentialTransferService) ScalePedersenCommitment(commitment PedersenCommitment, multiplier uint64) (PedersenCommitment, error) {
	point, err := service.ristrettoPoint(commitment[:], "Pedersen commitment")
	if err != nil {
		return PedersenCommitment{}, err
	}
	return service.pedersenCommitment(curve.NewRistrettoPoint().Mul(point, scalar.New().SetUint64(multiplier))), nil
}

func (service ConfidentialTransferService) AddPedersenOpenings(left, right PedersenOpening) (PedersenOpening, error) {
	return service.combinePedersenOpenings(left, right, false)
}

func (service ConfidentialTransferService) SubtractPedersenOpenings(left, right PedersenOpening) (PedersenOpening, error) {
	return service.combinePedersenOpenings(left, right, true)
}

func (service ConfidentialTransferService) ScalePedersenOpening(opening PedersenOpening, multiplier uint64) (PedersenOpening, error) {
	value, err := service.scalarFromBytes(opening[:], "Pedersen opening")
	if err != nil {
		return PedersenOpening{}, err
	}
	return service.pedersenOpening(scalar.New().Mul(value, scalar.New().SetUint64(multiplier)))
}

func (service ConfidentialTransferService) combinePedersenCommitments(left, right PedersenCommitment, subtract bool) (PedersenCommitment, error) {
	leftPoint, err := service.ristrettoPoint(left[:], "left Pedersen commitment")
	if err != nil {
		return PedersenCommitment{}, err
	}
	rightPoint, err := service.ristrettoPoint(right[:], "right Pedersen commitment")
	if err != nil {
		return PedersenCommitment{}, err
	}
	if subtract {
		return service.pedersenCommitment(curve.NewRistrettoPoint().Sub(leftPoint, rightPoint)), nil
	}
	return service.pedersenCommitment(curve.NewRistrettoPoint().Add(leftPoint, rightPoint)), nil
}

func (service ConfidentialTransferService) combinePedersenOpenings(left, right PedersenOpening, subtract bool) (PedersenOpening, error) {
	leftScalar, err := service.scalarFromBytes(left[:], "left Pedersen opening")
	if err != nil {
		return PedersenOpening{}, err
	}
	rightScalar, err := service.scalarFromBytes(right[:], "right Pedersen opening")
	if err != nil {
		return PedersenOpening{}, err
	}
	if subtract {
		return service.pedersenOpening(scalar.New().Sub(leftScalar, rightScalar))
	}
	return service.pedersenOpening(scalar.New().Add(leftScalar, rightScalar))
}

func (service ConfidentialTransferService) CombinePedersenAmount(commitmentLow, commitmentHigh PedersenCommitment, openingLow, openingHigh PedersenOpening) (PedersenCommitment, PedersenOpening, error) {
	scaledCommitment, err := service.ScalePedersenCommitment(commitmentHigh, 1<<16)
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, fmt.Errorf("combine Pedersen amount: %w", err)
	}
	commitment, err := service.AddPedersenCommitments(commitmentLow, scaledCommitment)
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, fmt.Errorf("combine Pedersen amount: %w", err)
	}
	scaledOpening, err := service.ScalePedersenOpening(openingHigh, 1<<16)
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, fmt.Errorf("combine Pedersen amount: %w", err)
	}
	opening, err := service.AddPedersenOpenings(openingLow, scaledOpening)
	if err != nil {
		return PedersenCommitment{}, PedersenOpening{}, fmt.Errorf("combine Pedersen amount: %w", err)
	}
	return commitment, opening, nil
}
