package token2022

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type ZeroCiphertextProofContext struct {
	Pubkey     ElGamalPubkey
	Ciphertext ElGamalCiphertext
}

type CiphertextCiphertextEqualityProofContext struct {
	FirstPubkey      ElGamalPubkey
	SecondPubkey     ElGamalPubkey
	FirstCiphertext  ElGamalCiphertext
	SecondCiphertext ElGamalCiphertext
}

type CiphertextCommitmentEqualityProofContext struct {
	Pubkey     ElGamalPubkey
	Ciphertext ElGamalCiphertext
	Commitment PedersenCommitment
}

type PubkeyValidityProofContext struct {
	Pubkey ElGamalPubkey
}

type PercentageWithCapProofContext struct {
	PercentageCommitment PedersenCommitment
	DeltaCommitment      PedersenCommitment
	ClaimedCommitment    PedersenCommitment
	MaxValue             uint64
}

type BatchedRangeProofContext struct {
	Commitments [8]PedersenCommitment
	BitLengths  [8]uint8
}

type GroupedCiphertext2HandlesValidityProofContext struct {
	FirstPubkey       ElGamalPubkey
	SecondPubkey      ElGamalPubkey
	GroupedCiphertext GroupedElGamalCiphertext2Handles
}

type BatchedGroupedCiphertext2HandlesValidityProofContext struct {
	FirstPubkey         ElGamalPubkey
	SecondPubkey        ElGamalPubkey
	GroupedCiphertextLo GroupedElGamalCiphertext2Handles
	GroupedCiphertextHi GroupedElGamalCiphertext2Handles
}

type GroupedCiphertext3HandlesValidityProofContext struct {
	FirstPubkey       ElGamalPubkey
	SecondPubkey      ElGamalPubkey
	ThirdPubkey       ElGamalPubkey
	GroupedCiphertext GroupedElGamalCiphertext3Handles
}

type BatchedGroupedCiphertext3HandlesValidityProofContext struct {
	FirstPubkey         ElGamalPubkey
	SecondPubkey        ElGamalPubkey
	ThirdPubkey         ElGamalPubkey
	GroupedCiphertextLo GroupedElGamalCiphertext3Handles
	GroupedCiphertextHi GroupedElGamalCiphertext3Handles
}

type ZKProofContextData struct {
	ZeroCiphertext                           *ZeroCiphertextProofContext
	CiphertextCiphertextEquality             *CiphertextCiphertextEqualityProofContext
	CiphertextCommitmentEquality             *CiphertextCommitmentEqualityProofContext
	PubkeyValidity                           *PubkeyValidityProofContext
	PercentageWithCap                        *PercentageWithCapProofContext
	BatchedRange                             *BatchedRangeProofContext
	GroupedCiphertext2HandlesValidity        *GroupedCiphertext2HandlesValidityProofContext
	BatchedGroupedCiphertext2HandlesValidity *BatchedGroupedCiphertext2HandlesValidityProofContext
	GroupedCiphertext3HandlesValidity        *GroupedCiphertext3HandlesValidityProofContext
	BatchedGroupedCiphertext3HandlesValidity *BatchedGroupedCiphertext3HandlesValidityProofContext
}

type ZKProofInstructionData struct {
	Discriminator      uint8
	ProofAccountOffset *uint32
	Context            *ZKProofContextData
	RawContext         []byte
	Proof              []byte
}

type ZKProofContextState struct {
	Authority  solana.PublicKey
	ProofType  uint8
	Context    *ZKProofContextData
	RawContext []byte
}

func (service ConfidentialTransferService) DecodeZKProofInstruction(accounts solana.AccountMetaSlice, data []byte) (*ZKProofInstruction, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("decode zk proof instruction: expected discriminator")
	}
	decoded := &ZKProofInstructionData{Discriminator: data[0]}
	if data[0] == 0 {
		if len(data) != 1 {
			return nil, fmt.Errorf("decode zk proof close context: expected 1 byte, got %d", len(data))
		}
		return &ZKProofInstruction{programID: service.ProofProgramID, AccountMetaSlice: accounts, data: data, Decoded: decoded}, nil
	}
	contextSize, proofSize, ok := service.zkProofSizes(data[0])
	if !ok {
		return nil, fmt.Errorf("decode zk proof instruction: unknown discriminator: %d", data[0])
	}
	if len(data) == 5 {
		offset := bin.NewDecoder(data[1:]).ReadUint32()
		decoded.ProofAccountOffset = &offset
		return &ZKProofInstruction{programID: service.ProofProgramID, AccountMetaSlice: accounts, data: data, Decoded: decoded}, nil
	}
	expected := 1 + contextSize + proofSize
	if len(data) != expected {
		return nil, fmt.Errorf("decode zk proof instruction %d: expected %d bytes, got %d", data[0], expected, len(data))
	}
	decoded.RawContext = data[1 : 1+contextSize]
	decoded.Proof = data[1+contextSize:]
	context, err := service.decodeZKProofContext(data[0], decoded.RawContext)
	if err != nil {
		return nil, err
	}
	decoded.Context = context
	return &ZKProofInstruction{programID: service.ProofProgramID, AccountMetaSlice: accounts, data: data, Decoded: decoded}, nil
}

func (service ConfidentialTransferService) DecodeProofContextState(data []byte) (*ZKProofContextState, error) {
	if len(data) < 33 {
		return nil, fmt.Errorf("decode zk proof context state: expected at least 33 bytes, got %d", len(data))
	}
	state := &ZKProofContextState{ProofType: data[32], RawContext: data[33:]}
	copy(state.Authority[:], data[:32])
	if state.ProofType == 0 {
		return state, nil
	}
	contextSize, _, ok := service.zkProofSizes(state.ProofType)
	if !ok {
		return nil, fmt.Errorf("decode zk proof context state: unknown proof type: %d", state.ProofType)
	}
	if len(state.RawContext) != contextSize {
		return nil, fmt.Errorf("decode zk proof context state %d: expected %d bytes, got %d", state.ProofType, contextSize+33, len(data))
	}
	context, err := service.decodeZKProofContext(state.ProofType, state.RawContext)
	if err != nil {
		return nil, err
	}
	state.Context = context
	return state, nil
}

func (ConfidentialTransferService) zkProofSizes(discriminator uint8) (int, int, bool) {
	switch discriminator {
	case 1:
		return 96, 96, true
	case 2:
		return 192, 224, true
	case 3:
		return 128, 192, true
	case 4:
		return 32, 64, true
	case 5:
		return 104, 256, true
	case 6:
		return 264, 672, true
	case 7:
		return 264, 736, true
	case 8:
		return 264, 800, true
	case 9:
		return 160, 160, true
	case 10:
		return 256, 160, true
	case 11:
		return 224, 192, true
	case 12:
		return 352, 192, true
	default:
		return 0, 0, false
	}
}

func (service ConfidentialTransferService) decodeZKProofContext(proofType uint8, data []byte) (*ZKProofContextData, error) {
	dec := bin.NewDecoder(data)
	context := &ZKProofContextData{}
	switch proofType {
	case 1:
		context.ZeroCiphertext = &ZeroCiphertextProofContext{Pubkey: service.readElGamalPubkey(dec), Ciphertext: service.readElGamalCiphertext(dec)}
	case 2:
		context.CiphertextCiphertextEquality = &CiphertextCiphertextEqualityProofContext{
			FirstPubkey:      service.readElGamalPubkey(dec),
			SecondPubkey:     service.readElGamalPubkey(dec),
			FirstCiphertext:  service.readElGamalCiphertext(dec),
			SecondCiphertext: service.readElGamalCiphertext(dec),
		}
	case 3:
		context.CiphertextCommitmentEquality = &CiphertextCommitmentEqualityProofContext{
			Pubkey:     service.readElGamalPubkey(dec),
			Ciphertext: service.readElGamalCiphertext(dec),
			Commitment: service.readPedersenCommitment(dec),
		}
	case 4:
		context.PubkeyValidity = &PubkeyValidityProofContext{Pubkey: service.readElGamalPubkey(dec)}
	case 5:
		context.PercentageWithCap = &PercentageWithCapProofContext{
			PercentageCommitment: service.readPedersenCommitment(dec),
			DeltaCommitment:      service.readPedersenCommitment(dec),
			ClaimedCommitment:    service.readPedersenCommitment(dec),
			MaxValue:             dec.ReadUint64(),
		}
	case 6, 7, 8:
		context.BatchedRange = service.decodeBatchedRangeProofContext(dec)
	case 9:
		context.GroupedCiphertext2HandlesValidity = &GroupedCiphertext2HandlesValidityProofContext{
			FirstPubkey:       service.readElGamalPubkey(dec),
			SecondPubkey:      service.readElGamalPubkey(dec),
			GroupedCiphertext: service.readGroupedElGamalCiphertext2Handles(dec),
		}
	case 10:
		context.BatchedGroupedCiphertext2HandlesValidity = &BatchedGroupedCiphertext2HandlesValidityProofContext{
			FirstPubkey:         service.readElGamalPubkey(dec),
			SecondPubkey:        service.readElGamalPubkey(dec),
			GroupedCiphertextLo: service.readGroupedElGamalCiphertext2Handles(dec),
			GroupedCiphertextHi: service.readGroupedElGamalCiphertext2Handles(dec),
		}
	case 11:
		context.GroupedCiphertext3HandlesValidity = &GroupedCiphertext3HandlesValidityProofContext{
			FirstPubkey:       service.readElGamalPubkey(dec),
			SecondPubkey:      service.readElGamalPubkey(dec),
			ThirdPubkey:       service.readElGamalPubkey(dec),
			GroupedCiphertext: service.readGroupedElGamalCiphertext3Handles(dec),
		}
	case 12:
		context.BatchedGroupedCiphertext3HandlesValidity = &BatchedGroupedCiphertext3HandlesValidityProofContext{
			FirstPubkey:         service.readElGamalPubkey(dec),
			SecondPubkey:        service.readElGamalPubkey(dec),
			ThirdPubkey:         service.readElGamalPubkey(dec),
			GroupedCiphertextLo: service.readGroupedElGamalCiphertext3Handles(dec),
			GroupedCiphertextHi: service.readGroupedElGamalCiphertext3Handles(dec),
		}
	default:
		return nil, fmt.Errorf("decode zk proof context: unknown proof type: %d", proofType)
	}
	if err := dec.Err(); err != nil {
		return nil, fmt.Errorf("decode zk proof context %d: %w", proofType, err)
	}
	return context, nil
}

func (service ConfidentialTransferService) decodeBatchedRangeProofContext(dec *bin.Decoder) *BatchedRangeProofContext {
	context := &BatchedRangeProofContext{}
	for i := range context.Commitments {
		context.Commitments[i] = service.readPedersenCommitment(dec)
	}
	for i := range context.BitLengths {
		context.BitLengths[i] = dec.ReadUint8()
	}
	return context
}

func (ConfidentialTransferService) readPedersenCommitment(dec *bin.Decoder) PedersenCommitment {
	value := PedersenCommitment{}
	copy(value[:], dec.ReadBytes(len(value)))
	return value
}

func (ConfidentialTransferService) readGroupedElGamalCiphertext2Handles(dec *bin.Decoder) GroupedElGamalCiphertext2Handles {
	value := GroupedElGamalCiphertext2Handles{}
	copy(value[:], dec.ReadBytes(len(value)))
	return value
}

func (ConfidentialTransferService) readGroupedElGamalCiphertext3Handles(dec *bin.Decoder) GroupedElGamalCiphertext3Handles {
	value := GroupedElGamalCiphertext3Handles{}
	copy(value[:], dec.ReadBytes(len(value)))
	return value
}
