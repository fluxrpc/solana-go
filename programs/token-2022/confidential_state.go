package token2022

import (
	"fmt"

	solana "github.com/fluxrpc/solana-go"
	bin "github.com/fluxrpc/solana-go/binary"
)

type ElGamalPubkey [32]byte

type ElGamalCiphertext [64]byte

type AeCiphertext [36]byte

type ConfidentialTransferMint struct {
	Authority              solana.PublicKey
	AutoApproveNewAccounts bool
	AuditorElGamalPubkey   ElGamalPubkey
}

type ConfidentialTransferAccount struct {
	Approved                            bool
	ElGamalPubkey                       ElGamalPubkey
	PendingBalanceLo                    ElGamalCiphertext
	PendingBalanceHi                    ElGamalCiphertext
	AvailableBalance                    ElGamalCiphertext
	DecryptableAvailableBalance         AeCiphertext
	AllowConfidentialCredits            bool
	AllowNonConfidentialCredits         bool
	PendingBalanceCreditCounter         uint64
	MaximumPendingBalanceCreditCounter  uint64
	ExpectedPendingBalanceCreditCounter uint64
	ActualPendingBalanceCreditCounter   uint64
}

type ConfidentialTransferFeeConfig struct {
	Authority                              solana.PublicKey
	WithdrawWithheldAuthorityElGamalPubkey ElGamalPubkey
	HarvestToMintEnabled                   bool
	WithheldAmount                         ElGamalCiphertext
}

type ConfidentialTransferFeeAmount struct {
	WithheldAmount ElGamalCiphertext
}

type ConfidentialMintBurn struct {
	ConfidentialSupply  ElGamalCiphertext
	DecryptableSupply   AeCiphertext
	SupplyElGamalPubkey ElGamalPubkey
	PendingBurn         ElGamalCiphertext
}

func (service ConfidentialTransferService) DecodeConfidentialTransferMint(data []byte) (ConfidentialTransferMint, error) {
	if err := service.validateStateSize(data, 65, "confidential transfer mint"); err != nil {
		return ConfidentialTransferMint{}, err
	}
	dec := bin.NewDecoder(data)
	state := ConfidentialTransferMint{
		Authority:              dec.ReadPublicKey(),
		AutoApproveNewAccounts: dec.ReadBool(),
	}
	copy(state.AuditorElGamalPubkey[:], dec.ReadBytes(len(state.AuditorElGamalPubkey)))
	if err := dec.Err(); err != nil {
		return ConfidentialTransferMint{}, fmt.Errorf("decode confidential transfer mint: %w", err)
	}
	return state, nil
}

func (service ConfidentialTransferService) DecodeConfidentialTransferAccount(data []byte) (ConfidentialTransferAccount, error) {
	if err := service.validateStateSize(data, 295, "confidential transfer account"); err != nil {
		return ConfidentialTransferAccount{}, err
	}
	dec := bin.NewDecoder(data)
	state := ConfidentialTransferAccount{Approved: dec.ReadBool()}
	copy(state.ElGamalPubkey[:], dec.ReadBytes(len(state.ElGamalPubkey)))
	copy(state.PendingBalanceLo[:], dec.ReadBytes(len(state.PendingBalanceLo)))
	copy(state.PendingBalanceHi[:], dec.ReadBytes(len(state.PendingBalanceHi)))
	copy(state.AvailableBalance[:], dec.ReadBytes(len(state.AvailableBalance)))
	copy(state.DecryptableAvailableBalance[:], dec.ReadBytes(len(state.DecryptableAvailableBalance)))
	state.AllowConfidentialCredits = dec.ReadBool()
	state.AllowNonConfidentialCredits = dec.ReadBool()
	state.PendingBalanceCreditCounter = dec.ReadUint64()
	state.MaximumPendingBalanceCreditCounter = dec.ReadUint64()
	state.ExpectedPendingBalanceCreditCounter = dec.ReadUint64()
	state.ActualPendingBalanceCreditCounter = dec.ReadUint64()
	if err := dec.Err(); err != nil {
		return ConfidentialTransferAccount{}, fmt.Errorf("decode confidential transfer account: %w", err)
	}
	return state, nil
}

func (service ConfidentialTransferService) DecodeConfidentialTransferFeeConfig(data []byte) (ConfidentialTransferFeeConfig, error) {
	if err := service.validateStateSize(data, 129, "confidential transfer fee config"); err != nil {
		return ConfidentialTransferFeeConfig{}, err
	}
	dec := bin.NewDecoder(data)
	state := ConfidentialTransferFeeConfig{Authority: dec.ReadPublicKey()}
	copy(state.WithdrawWithheldAuthorityElGamalPubkey[:], dec.ReadBytes(len(state.WithdrawWithheldAuthorityElGamalPubkey)))
	state.HarvestToMintEnabled = dec.ReadBool()
	copy(state.WithheldAmount[:], dec.ReadBytes(len(state.WithheldAmount)))
	if err := dec.Err(); err != nil {
		return ConfidentialTransferFeeConfig{}, fmt.Errorf("decode confidential transfer fee config: %w", err)
	}
	return state, nil
}

func (service ConfidentialTransferService) DecodeConfidentialTransferFeeAmount(data []byte) (ConfidentialTransferFeeAmount, error) {
	if err := service.validateStateSize(data, 64, "confidential transfer fee amount"); err != nil {
		return ConfidentialTransferFeeAmount{}, err
	}
	state := ConfidentialTransferFeeAmount{}
	copy(state.WithheldAmount[:], data)
	return state, nil
}

func (service ConfidentialTransferService) DecodeConfidentialMintBurn(data []byte) (ConfidentialMintBurn, error) {
	if err := service.validateStateSize(data, 196, "confidential mint burn"); err != nil {
		return ConfidentialMintBurn{}, err
	}
	dec := bin.NewDecoder(data)
	state := ConfidentialMintBurn{}
	copy(state.ConfidentialSupply[:], dec.ReadBytes(len(state.ConfidentialSupply)))
	copy(state.DecryptableSupply[:], dec.ReadBytes(len(state.DecryptableSupply)))
	copy(state.SupplyElGamalPubkey[:], dec.ReadBytes(len(state.SupplyElGamalPubkey)))
	copy(state.PendingBurn[:], dec.ReadBytes(len(state.PendingBurn)))
	if err := dec.Err(); err != nil {
		return ConfidentialMintBurn{}, fmt.Errorf("decode confidential mint burn: %w", err)
	}
	return state, nil
}

func (ConfidentialTransferService) validateStateSize(data []byte, size int, name string) error {
	if len(data) != size {
		return fmt.Errorf("decode %s: expected %d bytes, got %d", name, size, len(data))
	}
	return nil
}
