package token2022

import solana "github.com/fluxrpc/solana-go"

const (
	ConfidentialTransfer_InitializeMint uint8 = iota
	ConfidentialTransfer_UpdateMint
	ConfidentialTransfer_ConfigureAccount
	ConfidentialTransfer_ApproveAccount
	ConfidentialTransfer_EmptyAccount
	ConfidentialTransfer_Deposit
	ConfidentialTransfer_Withdraw
	ConfidentialTransfer_Transfer
	ConfidentialTransfer_ApplyPendingBalance
	ConfidentialTransfer_EnableConfidentialCredits
	ConfidentialTransfer_DisableConfidentialCredits
	ConfidentialTransfer_EnableNonConfidentialCredits
	ConfidentialTransfer_DisableNonConfidentialCredits
	ConfidentialTransfer_TransferWithSplitProofs
	ConfidentialTransfer_TransferWithSplitProofsInParallel
)

const (
	ConfidentialTransferFee_InitializeConfidentialTransferFeeConfig uint8 = iota
	ConfidentialTransferFee_WithdrawWithheldTokensFromMint
	ConfidentialTransferFee_WithdrawWithheldTokensFromAccounts
	ConfidentialTransferFee_HarvestWithheldTokensToMint
	ConfidentialTransferFee_EnableHarvestToMint
	ConfidentialTransferFee_DisableHarvestToMint
)

const (
	ConfidentialMintBurn_InitializeMint uint8 = iota
	ConfidentialMintBurn_UpdateDecryptableSupply
	ConfidentialMintBurn_RotateSupplyElGamalPubkey
	ConfidentialMintBurn_Mint
	ConfidentialMintBurn_Burn
)

type ConfidentialTransferExtension struct {
	extensionInstruction
	SubInstruction uint8
	RawData        []byte
}

func NewConfidentialTransferInstruction(subInstruction uint8, rawData []byte, accounts ...solana.AccountMeta) *ConfidentialTransferExtension {
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
}

func NewConfidentialTransferFeeInstruction(subInstruction uint8, rawData []byte, accounts ...solana.AccountMeta) *ConfidentialTransferFeeExtension {
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
}

func NewConfidentialMintBurnInstruction(subInstruction uint8, rawData []byte, accounts ...solana.AccountMeta) *ConfidentialMintBurnExtension {
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
