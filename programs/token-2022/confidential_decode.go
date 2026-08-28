package token2022

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type ConfidentialInstructionNoData struct{}

type InitializeConfidentialTransferMintData struct {
	Authority              *solana.PublicKey
	AutoApproveNewAccounts bool
	AuditorElGamalPubkey   *ElGamalPubkey
}

type UpdateConfidentialTransferMintData struct {
	AutoApproveNewAccounts bool
	AuditorElGamalPubkey   *ElGamalPubkey
}

type ConfigureConfidentialTransferAccountData struct {
	DecryptableZeroBalance             AeCiphertext
	MaximumPendingBalanceCreditCounter uint64
	ProofInstructionOffset             int8
}

type EmptyConfidentialTransferAccountData struct {
	ProofInstructionOffset int8
}

type ConfidentialDepositData struct {
	Amount   uint64
	Decimals uint8
}

type ConfidentialWithdrawData struct {
	Amount                         uint64
	Decimals                       uint8
	NewDecryptableAvailableBalance AeCiphertext
	EqualityProofInstructionOffset int8
	RangeProofInstructionOffset    int8
}

type ConfidentialTransferData struct {
	NewSourceDecryptableAvailableBalance     AeCiphertext
	TransferAmountAuditorCiphertextLo        ElGamalCiphertext
	TransferAmountAuditorCiphertextHi        ElGamalCiphertext
	EqualityProofInstructionOffset           int8
	CiphertextValidityProofInstructionOffset int8
	RangeProofInstructionOffset              int8
}

type ApplyConfidentialPendingBalanceData struct {
	ExpectedPendingBalanceCreditCounter uint64
	NewDecryptableAvailableBalance      AeCiphertext
}

type ConfidentialTransferWithFeeData struct {
	NewSourceDecryptableAvailableBalance              AeCiphertext
	TransferAmountAuditorCiphertextLo                 ElGamalCiphertext
	TransferAmountAuditorCiphertextHi                 ElGamalCiphertext
	EqualityProofInstructionOffset                    int8
	TransferAmountCiphertextValidityInstructionOffset int8
	FeeSigmaProofInstructionOffset                    int8
	FeeCiphertextValidityProofInstructionOffset       int8
	RangeProofInstructionOffset                       int8
}

type ConfidentialTransferInstructionData struct {
	InitializeMint                *InitializeConfidentialTransferMintData
	UpdateMint                    *UpdateConfidentialTransferMintData
	ConfigureAccount              *ConfigureConfidentialTransferAccountData
	ApproveAccount                *ConfidentialInstructionNoData
	EmptyAccount                  *EmptyConfidentialTransferAccountData
	Deposit                       *ConfidentialDepositData
	Withdraw                      *ConfidentialWithdrawData
	Transfer                      *ConfidentialTransferData
	ApplyPendingBalance           *ApplyConfidentialPendingBalanceData
	EnableConfidentialCredits     *ConfidentialInstructionNoData
	DisableConfidentialCredits    *ConfidentialInstructionNoData
	EnableNonConfidentialCredits  *ConfidentialInstructionNoData
	DisableNonConfidentialCredits *ConfidentialInstructionNoData
	TransferWithFee               *ConfidentialTransferWithFeeData
	ConfigureAccountWithRegistry  *ConfidentialInstructionNoData
}

type InitializeConfidentialTransferFeeData struct {
	Authority                              *solana.PublicKey
	WithdrawWithheldAuthorityElGamalPubkey ElGamalPubkey
}

type WithdrawConfidentialWithheldTokensFromMintData struct {
	ProofInstructionOffset         int8
	NewDecryptableAvailableBalance AeCiphertext
}

type WithdrawConfidentialWithheldTokensFromAccountsData struct {
	NumTokenAccounts               uint8
	ProofInstructionOffset         int8
	NewDecryptableAvailableBalance AeCiphertext
}

type ConfidentialTransferFeeInstructionData struct {
	InitializeConfig                   *InitializeConfidentialTransferFeeData
	WithdrawWithheldTokensFromMint     *WithdrawConfidentialWithheldTokensFromMintData
	WithdrawWithheldTokensFromAccounts *WithdrawConfidentialWithheldTokensFromAccountsData
	HarvestWithheldTokensToMint        *ConfidentialInstructionNoData
	EnableHarvestToMint                *ConfidentialInstructionNoData
	DisableHarvestToMint               *ConfidentialInstructionNoData
}

type InitializeConfidentialMintBurnData struct {
	SupplyElGamalPubkey ElGamalPubkey
	DecryptableSupply   AeCiphertext
}

type RotateConfidentialSupplyKeyData struct {
	NewSupplyElGamalPubkey ElGamalPubkey
	ProofInstructionOffset int8
}

type UpdateConfidentialSupplyData struct {
	NewDecryptableSupply AeCiphertext
}

type ConfidentialMintData struct {
	NewDecryptableSupply                     AeCiphertext
	MintAmountAuditorCiphertextLo            ElGamalCiphertext
	MintAmountAuditorCiphertextHi            ElGamalCiphertext
	EqualityProofInstructionOffset           int8
	CiphertextValidityProofInstructionOffset int8
	RangeProofInstructionOffset              int8
}

type ConfidentialBurnData struct {
	NewDecryptableAvailableBalance           AeCiphertext
	BurnAmountAuditorCiphertextLo            ElGamalCiphertext
	BurnAmountAuditorCiphertextHi            ElGamalCiphertext
	EqualityProofInstructionOffset           int8
	CiphertextValidityProofInstructionOffset int8
	RangeProofInstructionOffset              int8
}

type ConfidentialMintBurnInstructionData struct {
	InitializeMint          *InitializeConfidentialMintBurnData
	RotateSupplyKey         *RotateConfidentialSupplyKeyData
	UpdateDecryptableSupply *UpdateConfidentialSupplyData
	Mint                    *ConfidentialMintData
	Burn                    *ConfidentialBurnData
	ApplyPendingBurn        *ConfidentialInstructionNoData
}

func (service ConfidentialTransferService) DecodeConfidentialTransferInstruction(subInstruction uint8, data []byte) (*ConfidentialTransferInstructionData, error) {
	decoded := &ConfidentialTransferInstructionData{}
	var dec *bin.Decoder
	var err error
	switch subInstruction {
	case 0:
		dec, err = service.confidentialDecoder(data, 65, "initialize mint")
		if err == nil {
			decoded.InitializeMint = &InitializeConfidentialTransferMintData{
				Authority:              service.readOptionalPublicKey(dec),
				AutoApproveNewAccounts: dec.ReadBool(),
				AuditorElGamalPubkey:   service.readOptionalElGamalPubkey(dec),
			}
		}
	case 1:
		dec, err = service.confidentialDecoder(data, 33, "update mint")
		if err == nil {
			decoded.UpdateMint = &UpdateConfidentialTransferMintData{
				AutoApproveNewAccounts: dec.ReadBool(),
				AuditorElGamalPubkey:   service.readOptionalElGamalPubkey(dec),
			}
		}
	case 2:
		dec, err = service.confidentialDecoder(data, 45, "configure account")
		if err == nil {
			payload := &ConfigureConfidentialTransferAccountData{}
			payload.DecryptableZeroBalance = service.readAeCiphertext(dec)
			payload.MaximumPendingBalanceCreditCounter = dec.ReadUint64()
			payload.ProofInstructionOffset = int8(dec.ReadUint8())
			decoded.ConfigureAccount = payload
		}
	case 3:
		err = service.validateStateSize(data, 0, "confidential transfer approve account")
		decoded.ApproveAccount = &ConfidentialInstructionNoData{}
	case 4:
		dec, err = service.confidentialDecoder(data, 1, "empty account")
		if err == nil {
			decoded.EmptyAccount = &EmptyConfidentialTransferAccountData{ProofInstructionOffset: int8(dec.ReadUint8())}
		}
	case 5:
		dec, err = service.confidentialDecoder(data, 9, "deposit")
		if err == nil {
			decoded.Deposit = &ConfidentialDepositData{Amount: dec.ReadUint64(), Decimals: dec.ReadUint8()}
		}
	case 6:
		dec, err = service.confidentialDecoder(data, 47, "withdraw")
		if err == nil {
			payload := &ConfidentialWithdrawData{Amount: dec.ReadUint64(), Decimals: dec.ReadUint8()}
			payload.NewDecryptableAvailableBalance = service.readAeCiphertext(dec)
			payload.EqualityProofInstructionOffset = int8(dec.ReadUint8())
			payload.RangeProofInstructionOffset = int8(dec.ReadUint8())
			decoded.Withdraw = payload
		}
	case 7:
		dec, err = service.confidentialDecoder(data, 167, "transfer")
		if err == nil {
			decoded.Transfer = service.decodeConfidentialTransferData(dec)
		}
	case 8:
		dec, err = service.confidentialDecoder(data, 44, "apply pending balance")
		if err == nil {
			payload := &ApplyConfidentialPendingBalanceData{ExpectedPendingBalanceCreditCounter: dec.ReadUint64()}
			payload.NewDecryptableAvailableBalance = service.readAeCiphertext(dec)
			decoded.ApplyPendingBalance = payload
		}
	case 9, 10, 11, 12, 14:
		err = service.decodeConfidentialNoData(subInstruction, data, decoded)
	case 13:
		dec, err = service.confidentialDecoder(data, 169, "transfer with fee")
		if err == nil {
			payload := &ConfidentialTransferWithFeeData{}
			payload.NewSourceDecryptableAvailableBalance = service.readAeCiphertext(dec)
			payload.TransferAmountAuditorCiphertextLo = service.readElGamalCiphertext(dec)
			payload.TransferAmountAuditorCiphertextHi = service.readElGamalCiphertext(dec)
			payload.EqualityProofInstructionOffset = int8(dec.ReadUint8())
			payload.TransferAmountCiphertextValidityInstructionOffset = int8(dec.ReadUint8())
			payload.FeeSigmaProofInstructionOffset = int8(dec.ReadUint8())
			payload.FeeCiphertextValidityProofInstructionOffset = int8(dec.ReadUint8())
			payload.RangeProofInstructionOffset = int8(dec.ReadUint8())
			decoded.TransferWithFee = payload
		}
	default:
		return nil, fmt.Errorf("token-2022: unknown confidential-transfer instruction: %d", subInstruction)
	}
	if err != nil {
		return nil, err
	}
	if dec != nil && dec.Err() != nil {
		return nil, fmt.Errorf("decode token-2022 confidential transfer: %w", dec.Err())
	}
	return decoded, nil
}

func (service ConfidentialTransferService) DecodeConfidentialTransferFeeInstruction(subInstruction uint8, data []byte) (*ConfidentialTransferFeeInstructionData, error) {
	decoded := &ConfidentialTransferFeeInstructionData{}
	var dec *bin.Decoder
	var err error
	switch subInstruction {
	case 0:
		dec, err = service.confidentialDecoder(data, 64, "fee initialize config")
		if err == nil {
			payload := &InitializeConfidentialTransferFeeData{Authority: service.readOptionalPublicKey(dec)}
			payload.WithdrawWithheldAuthorityElGamalPubkey = service.readElGamalPubkey(dec)
			decoded.InitializeConfig = payload
		}
	case 1:
		dec, err = service.confidentialDecoder(data, 37, "fee withdraw from mint")
		if err == nil {
			payload := &WithdrawConfidentialWithheldTokensFromMintData{ProofInstructionOffset: int8(dec.ReadUint8())}
			payload.NewDecryptableAvailableBalance = service.readAeCiphertext(dec)
			decoded.WithdrawWithheldTokensFromMint = payload
		}
	case 2:
		dec, err = service.confidentialDecoder(data, 38, "fee withdraw from accounts")
		if err == nil {
			payload := &WithdrawConfidentialWithheldTokensFromAccountsData{
				NumTokenAccounts:       dec.ReadUint8(),
				ProofInstructionOffset: int8(dec.ReadUint8()),
			}
			payload.NewDecryptableAvailableBalance = service.readAeCiphertext(dec)
			decoded.WithdrawWithheldTokensFromAccounts = payload
		}
	case 3, 4, 5:
		err = service.decodeConfidentialFeeNoData(subInstruction, data, decoded)
	default:
		return nil, fmt.Errorf("token-2022: unknown confidential-transfer-fee instruction: %d", subInstruction)
	}
	if err != nil {
		return nil, err
	}
	if dec != nil && dec.Err() != nil {
		return nil, fmt.Errorf("decode token-2022 confidential transfer fee: %w", dec.Err())
	}
	return decoded, nil
}

func (service ConfidentialTransferService) DecodeConfidentialMintBurnInstruction(subInstruction uint8, data []byte) (*ConfidentialMintBurnInstructionData, error) {
	decoded := &ConfidentialMintBurnInstructionData{}
	switch subInstruction {
	case 0:
		dec, err := service.confidentialDecoder(data, 68, "mint burn initialize mint")
		if err != nil {
			return nil, err
		}
		decoded.InitializeMint = &InitializeConfidentialMintBurnData{
			SupplyElGamalPubkey: service.readElGamalPubkey(dec),
			DecryptableSupply:   service.readAeCiphertext(dec),
		}
		return decoded, service.confidentialDecodeError(dec, "mint burn")
	case 1:
		dec, err := service.confidentialDecoder(data, 33, "mint burn rotate supply key")
		if err != nil {
			return nil, err
		}
		decoded.RotateSupplyKey = &RotateConfidentialSupplyKeyData{
			NewSupplyElGamalPubkey: service.readElGamalPubkey(dec),
			ProofInstructionOffset: int8(dec.ReadUint8()),
		}
		return decoded, service.confidentialDecodeError(dec, "mint burn")
	case 2:
		dec, err := service.confidentialDecoder(data, 36, "mint burn update decryptable supply")
		if err != nil {
			return nil, err
		}
		decoded.UpdateDecryptableSupply = &UpdateConfidentialSupplyData{NewDecryptableSupply: service.readAeCiphertext(dec)}
		return decoded, service.confidentialDecodeError(dec, "mint burn")
	case 3:
		dec, err := service.confidentialDecoder(data, 167, "mint burn mint")
		if err != nil {
			return nil, err
		}
		payload := service.decodeConfidentialMintData(dec)
		decoded.Mint = &payload
		return decoded, service.confidentialDecodeError(dec, "mint burn")
	case 4:
		dec, err := service.confidentialDecoder(data, 167, "mint burn burn")
		if err != nil {
			return nil, err
		}
		payload := service.decodeConfidentialBurnData(dec)
		decoded.Burn = &payload
		return decoded, service.confidentialDecodeError(dec, "mint burn")
	case 5:
		if err := service.validateStateSize(data, 0, "confidential mint burn apply pending burn"); err != nil {
			return nil, err
		}
		decoded.ApplyPendingBurn = &ConfidentialInstructionNoData{}
		return decoded, nil
	default:
		return nil, fmt.Errorf("token-2022: unknown confidential-mint-burn instruction: %d", subInstruction)
	}
}

func (service ConfidentialTransferService) decodeConfidentialMintData(dec *bin.Decoder) ConfidentialMintData {
	payload := ConfidentialMintData{}
	payload.NewDecryptableSupply = service.readAeCiphertext(dec)
	payload.MintAmountAuditorCiphertextLo = service.readElGamalCiphertext(dec)
	payload.MintAmountAuditorCiphertextHi = service.readElGamalCiphertext(dec)
	payload.EqualityProofInstructionOffset = int8(dec.ReadUint8())
	payload.CiphertextValidityProofInstructionOffset = int8(dec.ReadUint8())
	payload.RangeProofInstructionOffset = int8(dec.ReadUint8())
	return payload
}

func (service ConfidentialTransferService) decodeConfidentialBurnData(dec *bin.Decoder) ConfidentialBurnData {
	payload := ConfidentialBurnData{}
	payload.NewDecryptableAvailableBalance = service.readAeCiphertext(dec)
	payload.BurnAmountAuditorCiphertextLo = service.readElGamalCiphertext(dec)
	payload.BurnAmountAuditorCiphertextHi = service.readElGamalCiphertext(dec)
	payload.EqualityProofInstructionOffset = int8(dec.ReadUint8())
	payload.CiphertextValidityProofInstructionOffset = int8(dec.ReadUint8())
	payload.RangeProofInstructionOffset = int8(dec.ReadUint8())
	return payload
}

func (ConfidentialTransferService) confidentialDecodeError(dec *bin.Decoder, family string) error {
	if err := dec.Err(); err != nil {
		return fmt.Errorf("decode token-2022 confidential %s: %w", family, err)
	}
	return nil
}

func (service ConfidentialTransferService) confidentialDecoder(data []byte, size int, name string) (*bin.Decoder, error) {
	if err := service.validateStateSize(data, size, "confidential transfer "+name); err != nil {
		return nil, err
	}
	return bin.NewDecoder(data), nil
}

func (ConfidentialTransferService) decodeConfidentialNoData(subInstruction uint8, data []byte, decoded *ConfidentialTransferInstructionData) error {
	if len(data) != 0 {
		return fmt.Errorf("decode confidential transfer instruction %d: expected 0 bytes, got %d", subInstruction, len(data))
	}
	payload := &ConfidentialInstructionNoData{}
	switch subInstruction {
	case 9:
		decoded.EnableConfidentialCredits = payload
	case 10:
		decoded.DisableConfidentialCredits = payload
	case 11:
		decoded.EnableNonConfidentialCredits = payload
	case 12:
		decoded.DisableNonConfidentialCredits = payload
	case 14:
		decoded.ConfigureAccountWithRegistry = payload
	}
	return nil
}

func (ConfidentialTransferService) decodeConfidentialFeeNoData(subInstruction uint8, data []byte, decoded *ConfidentialTransferFeeInstructionData) error {
	if len(data) != 0 {
		return fmt.Errorf("decode confidential transfer fee instruction %d: expected 0 bytes, got %d", subInstruction, len(data))
	}
	payload := &ConfidentialInstructionNoData{}
	switch subInstruction {
	case 3:
		decoded.HarvestWithheldTokensToMint = payload
	case 4:
		decoded.EnableHarvestToMint = payload
	case 5:
		decoded.DisableHarvestToMint = payload
	}
	return nil
}

func (service ConfidentialTransferService) decodeConfidentialTransferData(dec *bin.Decoder) *ConfidentialTransferData {
	payload := &ConfidentialTransferData{}
	payload.NewSourceDecryptableAvailableBalance = service.readAeCiphertext(dec)
	payload.TransferAmountAuditorCiphertextLo = service.readElGamalCiphertext(dec)
	payload.TransferAmountAuditorCiphertextHi = service.readElGamalCiphertext(dec)
	payload.EqualityProofInstructionOffset = int8(dec.ReadUint8())
	payload.CiphertextValidityProofInstructionOffset = int8(dec.ReadUint8())
	payload.RangeProofInstructionOffset = int8(dec.ReadUint8())
	return payload
}

func (ConfidentialTransferService) readOptionalPublicKey(dec *bin.Decoder) *solana.PublicKey {
	key := dec.ReadPublicKey()
	if key.IsZero() {
		return nil
	}
	return &key
}

func (service ConfidentialTransferService) readOptionalElGamalPubkey(dec *bin.Decoder) *ElGamalPubkey {
	key := service.readElGamalPubkey(dec)
	if key == (ElGamalPubkey{}) {
		return nil
	}
	return &key
}

func (ConfidentialTransferService) readElGamalPubkey(dec *bin.Decoder) ElGamalPubkey {
	value := ElGamalPubkey{}
	copy(value[:], dec.ReadBytes(len(value)))
	return value
}

func (ConfidentialTransferService) readElGamalCiphertext(dec *bin.Decoder) ElGamalCiphertext {
	value := ElGamalCiphertext{}
	copy(value[:], dec.ReadBytes(len(value)))
	return value
}

func (ConfidentialTransferService) readAeCiphertext(dec *bin.Decoder) AeCiphertext {
	value := AeCiphertext{}
	copy(value[:], dec.ReadBytes(len(value)))
	return value
}
