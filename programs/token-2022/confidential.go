package token2022

import solana "github.com/fluxrpc/solana-go"

type ConfidentialTransferExtension struct {
	extensionInstruction
	SubInstruction uint8
	RawData        []byte
	Decoded        *ConfidentialTransferInstructionData
}

func (ConfidentialTransferService) RawConfidentialTransferInstruction(subInstruction uint8, rawData []byte, accounts ...solana.AccountMeta) *ConfidentialTransferExtension {
	metas := make(solana.AccountMetaSlice, len(accounts))
	for i := range accounts {
		metas[i] = &accounts[i]
	}
	return &ConfidentialTransferExtension{extensionInstruction: extensionInstruction{metas}, SubInstruction: subInstruction, RawData: rawData}
}

func (inst *ConfidentialTransferExtension) Data() ([]byte, error) {
	data := make([]byte, 2, 2+len(inst.RawData))
	data[0] = byte(ConfidentialTransferExtensionInstruction)
	data[1] = inst.SubInstruction
	return append(data, inst.RawData...), nil
}

type ConfidentialTransferFeeExtension struct {
	extensionInstruction
	SubInstruction uint8
	RawData        []byte
	Decoded        *ConfidentialTransferFeeInstructionData
}

func (ConfidentialTransferService) RawConfidentialTransferFeeInstruction(subInstruction uint8, rawData []byte, accounts ...solana.AccountMeta) *ConfidentialTransferFeeExtension {
	metas := make(solana.AccountMetaSlice, len(accounts))
	for i := range accounts {
		metas[i] = &accounts[i]
	}
	return &ConfidentialTransferFeeExtension{extensionInstruction: extensionInstruction{metas}, SubInstruction: subInstruction, RawData: rawData}
}

func (inst *ConfidentialTransferFeeExtension) Data() ([]byte, error) {
	data := make([]byte, 2, 2+len(inst.RawData))
	data[0] = byte(ConfidentialTransferFeeExtensionInstruction)
	data[1] = inst.SubInstruction
	return append(data, inst.RawData...), nil
}

type ConfidentialMintBurnExtension struct {
	extensionInstruction
	SubInstruction uint8
	RawData        []byte
	Decoded        *ConfidentialMintBurnInstructionData
}

func (ConfidentialTransferService) RawConfidentialMintBurnInstruction(subInstruction uint8, rawData []byte, accounts ...solana.AccountMeta) *ConfidentialMintBurnExtension {
	metas := make(solana.AccountMetaSlice, len(accounts))
	for i := range accounts {
		metas[i] = &accounts[i]
	}
	return &ConfidentialMintBurnExtension{extensionInstruction: extensionInstruction{metas}, SubInstruction: subInstruction, RawData: rawData}
}

func (inst *ConfidentialMintBurnExtension) Data() ([]byte, error) {
	data := make([]byte, 2, 2+len(inst.RawData))
	data[0] = byte(ConfidentialMintBurnExtensionInstruction)
	data[1] = inst.SubInstruction
	return append(data, inst.RawData...), nil
}
