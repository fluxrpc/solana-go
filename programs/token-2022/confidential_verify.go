package token2022

import "fmt"

func (service ConfidentialTransferService) VerifyZKProofData(data ZKProofData) error {
	if err := service.validateZKProofData(data); err != nil {
		return err
	}
	switch data.Discriminator {
	case 1:
		return service.VerifyZeroCiphertextProof(data)
	case 2:
		return service.VerifyCiphertextCiphertextEqualityProof(data)
	case 3:
		return service.VerifyCiphertextCommitmentEqualityProof(data)
	case 4:
		return service.VerifyPubkeyValidityProof(data)
	case 5:
		return service.VerifyPercentageWithCapProof(data)
	case 6, 7, 8:
		return service.VerifyBatchedRangeProof(data)
	case 9:
		return service.VerifyGroupedCiphertext2HandlesValidityProof(data)
	case 10:
		return service.VerifyBatchedGroupedCiphertext2HandlesValidityProof(data)
	case 11:
		return service.VerifyGroupedCiphertext3HandlesValidityProof(data)
	case 12:
		return service.VerifyBatchedGroupedCiphertext3HandlesValidityProof(data)
	default:
		return fmt.Errorf("verify ZK proof data: unsupported discriminator %d", data.Discriminator)
	}
}

func (service ConfidentialTransferService) validateZKProofData(data ZKProofData) error {
	contextSize, proofSize, ok := service.zkProofSizes(data.Discriminator)
	if !ok {
		return fmt.Errorf("verify ZK proof data: unsupported discriminator %d", data.Discriminator)
	}
	if len(data.Context) != contextSize || len(data.Proof) != proofSize {
		return fmt.Errorf("verify ZK proof data %d: expected %d context and %d proof bytes, got %d and %d", data.Discriminator, contextSize, proofSize, len(data.Context), len(data.Proof))
	}
	return nil
}
