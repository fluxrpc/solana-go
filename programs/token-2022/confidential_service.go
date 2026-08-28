package token2022

import (
	"runtime"
	"sync"

	solana "github.com/fluxrpc/solana-go"
	"github.com/oasisprotocol/curve25519-voi/curve"
	"github.com/oasisprotocol/curve25519-voi/curve/scalar"
)

// ConfidentialTransferService owns Token-2022 confidential-transfer logic.
type ConfidentialTransferService struct {
	ProofProgramID       solana.PublicKey
	RegistryProgramID    solana.PublicKey
	RecordProgramID      solana.PublicKey
	discreteLog          map[[32]byte]uint16
	rangeProofGenerators rangeProofGenerators
	pedersenOpeningPoint *curve.RistrettoPoint
}

type discreteLogEntry struct {
	key   [32]byte
	value uint16
}

type discreteLogBuilder struct {
	entries []discreteLogEntry
	service ConfidentialTransferService
	wait    sync.WaitGroup
}

func (service *ConfidentialTransferService) Start() {
	service.ProofProgramID = solana.MustPublicKeyFromBase58("ZkE1Gama1Proof11111111111111111111111111111")
	service.RegistryProgramID = solana.MustPublicKeyFromBase58("regVYJW7tcT8zipN5YiBvHsvR5jXW1uLFxaHSbugABg")
	service.RecordProgramID = solana.MustPublicKeyFromBase58("recr1L3PCGKLbckBqMNcJhuuyU1zgo8nBhfLVsJNwr5")
	service.discreteLog = service.buildDiscreteLogTable()
	var err error
	service.rangeProofGenerators, err = service.newRangeProofGenerators(256)
	if err != nil {
		panic(err)
	}
	service.pedersenOpeningPoint, err = service.pedersenOpeningBasepoint()
	if err != nil {
		panic(err)
	}
}

func (service ConfidentialTransferService) buildDiscreteLogTable() map[[32]byte]uint16 {
	builder := discreteLogBuilder{entries: make([]discreteLogEntry, 65536), service: service}
	workers := min(runtime.GOMAXPROCS(0), len(builder.entries))
	chunk := (len(builder.entries) + workers - 1) / workers
	for start := 0; start < len(builder.entries); start += chunk {
		builder.wait.Add(1)
		go builder.build(start, min(start+chunk, len(builder.entries)))
	}
	builder.wait.Wait()
	table := make(map[[32]byte]uint16, len(builder.entries))
	for _, entry := range builder.entries {
		table[entry.key] = entry.value
	}
	return table
}

func (builder *discreteLogBuilder) build(start, end int) {
	defer builder.wait.Done()
	point := curve.NewRistrettoPoint().MulBasepoint(curve.RISTRETTO_BASEPOINT_TABLE, scalar.New().SetUint64(uint64(start)))
	for value := start; value < end; value++ {
		builder.entries[value] = discreteLogEntry{key: builder.service.compressedRistrettoKey(point), value: uint16(value)}
		point.Add(point, curve.RISTRETTO_BASEPOINT_POINT)
	}
}
