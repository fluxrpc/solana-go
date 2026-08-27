package token2022

import (
	solana "github.com/fluxrpc/solana-go"
	"github.com/oasisprotocol/curve25519-voi/curve"
)

// ConfidentialTransferService owns Token-2022 confidential-transfer logic.
type ConfidentialTransferService struct {
	ProofProgramID    solana.PublicKey
	RegistryProgramID solana.PublicKey
	RecordProgramID   solana.PublicKey
	discreteLog       map[[32]byte]uint16
}

func (service *ConfidentialTransferService) Start() {
	service.ProofProgramID = solana.MustPublicKeyFromBase58("ZkE1Gama1Proof11111111111111111111111111111")
	service.RegistryProgramID = solana.MustPublicKeyFromBase58("regVYJW7tcT8zipN5YiBvHsvR5jXW1uLFxaHSbugABg")
	service.RecordProgramID = solana.MustPublicKeyFromBase58("recr1L3PCGKLbckBqMNcJhuuyU1zgo8nBhfLVsJNwr5")
	service.discreteLog = service.buildDiscreteLogTable()
}

func (service ConfidentialTransferService) buildDiscreteLogTable() map[[32]byte]uint16 {
	table := make(map[[32]byte]uint16, 65536)
	point := curve.NewRistrettoPoint().Identity()
	for value := 0; value < 65536; value++ {
		table[service.compressedRistrettoKey(point)] = uint16(value)
		point.Add(point, curve.RISTRETTO_BASEPOINT_POINT)
	}
	return table
}
