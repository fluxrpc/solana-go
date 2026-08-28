package token2022

import (
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

func (service ConfidentialTransferService) DecodePedersenCommitmentU32(target PedersenCommitment) (uint32, error) {
	return service.decodePedersenCommitmentU32(target, service.discreteLog)
}

func (service ConfidentialTransferService) DecryptElGamalU32(secret ElGamalSecretKey, ciphertext ElGamalCiphertext) (uint32, error) {
	target, err := service.DecryptElGamalTarget(secret, ciphertext)
	if err != nil {
		return 0, service.decryptionError(err)
	}
	amount, err := service.DecodePedersenCommitmentU32(target)
	if err != nil {
		return 0, service.decryptionError(err)
	}
	return amount, nil
}

func (service ConfidentialTransferService) DecryptGroupedElGamal2U32(secret ElGamalSecretKey, ciphertext GroupedElGamalCiphertext2Handles, index int) (uint32, error) {
	elGamalCiphertext, err := service.GroupedElGamalCiphertext2Handle(ciphertext, index)
	if err != nil {
		return 0, err
	}
	return service.DecryptElGamalU32(secret, elGamalCiphertext)
}

func (service ConfidentialTransferService) DecryptGroupedElGamal3U32(secret ElGamalSecretKey, ciphertext GroupedElGamalCiphertext3Handles, index int) (uint32, error) {
	elGamalCiphertext, err := service.GroupedElGamalCiphertext3Handle(ciphertext, index)
	if err != nil {
		return 0, err
	}
	return service.DecryptElGamalU32(secret, elGamalCiphertext)
}

func (service ConfidentialTransferService) DecryptPendingBalance(secret ElGamalSecretKey, low, high ElGamalCiphertext) (uint64, error) {
	lowAmount, err := service.decryptElGamalU32(secret, low, service.discreteLog)
	if err != nil {
		return 0, service.decryptionError(err)
	}
	if lowAmount > 65535 {
		return 0, service.decryptionError(fmt.Errorf("decrypt pending balance: low amount exceeds 16 bits"))
	}
	highAmount, err := service.decryptElGamalU32(secret, high, service.discreteLog)
	if err != nil {
		return 0, service.decryptionError(err)
	}
	return service.CombineAmount(uint16(lowAmount), highAmount), nil
}

func (service ConfidentialTransferService) decryptElGamalU32(secret ElGamalSecretKey, ciphertext ElGamalCiphertext, table map[[32]byte]uint16) (uint32, error) {
	target, err := service.DecryptElGamalTarget(secret, ciphertext)
	if err != nil {
		return 0, err
	}
	return service.decodePedersenCommitmentU32(target, table)
}

func (service ConfidentialTransferService) decodePedersenCommitmentU32(target PedersenCommitment, table map[[32]byte]uint16) (uint32, error) {
	point, err := service.ristrettoPoint(target[:], "Pedersen decrypt target")
	if err != nil {
		return 0, err
	}
	step := curve.NewRistrettoPoint().Mul(curve.RISTRETTO_BASEPOINT_POINT, scalar.New().SetUint64(65536))
	for high := uint32(0); high < 65536; high++ {
		low, found := table[service.compressedRistrettoKey(point)]
		if found {
			return high<<16 | uint32(low), nil
		}
		point.Sub(point, step)
	}
	return 0, fmt.Errorf("decode Pedersen decrypt target: discrete log outside u32 range")
}

func (ConfidentialTransferService) compressedRistrettoKey(point *curve.RistrettoPoint) [32]byte {
	return [32]byte(*curve.NewCompressedRistretto().SetRistrettoPoint(point))
}
