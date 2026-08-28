package token2022

import "fmt"

func (service ConfidentialTransferService) VerifyZKProofData(data ZKProofData) error {
	if err := service.validateZKProofData(data); err != nil {
		return service.invalidProofError(err)
	}
	var err error
	switch data.Discriminator {
	case 1:
		err = service.VerifyZeroCiphertextProof(data)
	case 2:
		err = service.VerifyCiphertextCiphertextEqualityProof(data)
	case 3:
		err = service.VerifyCiphertextCommitmentEqualityProof(data)
	case 4:
		err = service.VerifyPubkeyValidityProof(data)
	case 5:
		err = service.VerifyPercentageWithCapProof(data)
	case 6, 7, 8:
		err = service.VerifyBatchedRangeProof(data)
	case 9:
		err = service.VerifyGroupedCiphertext2HandlesValidityProof(data)
	case 10:
		err = service.VerifyBatchedGroupedCiphertext2HandlesValidityProof(data)
	case 11:
		err = service.VerifyGroupedCiphertext3HandlesValidityProof(data)
	case 12:
		err = service.VerifyBatchedGroupedCiphertext3HandlesValidityProof(data)
	}
	if err != nil {
		return service.invalidProofError(err)
	}
	return nil
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
