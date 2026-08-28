package token2022

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestConfidentialTransferServiceDecodeTransferMint(t *testing.T) {
	data := make([]byte, 65)
	for i := range data {
		data[i] = byte(i + 1)
	}
	data[32] = 1
	state, err := (ConfidentialTransferService{}).DecodeConfidentialTransferMint(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(state.Authority[:], data[:32]) || !state.AutoApproveNewAccounts || !bytes.Equal(state.AuditorElGamalPubkey[:], data[33:]) {
		t.Fatalf("state = %+v", state)
	}
}

func TestConfidentialTransferServiceDecodeTransferAccount(t *testing.T) {
	data := make([]byte, 295)
	for i := range data {
		data[i] = byte(i)
	}
	data[0], data[261], data[262] = 1, 1, 0
	state, err := (ConfidentialTransferService{}).DecodeConfidentialTransferAccount(data)
	if err != nil {
		t.Fatal(err)
	}
	if !state.Approved || !bytes.Equal(state.ElGamalPubkey[:], data[1:33]) || !bytes.Equal(state.PendingBalanceLo[:], data[33:97]) {
		t.Fatalf("leading fields = %+v", state)
	}
	if !bytes.Equal(state.PendingBalanceHi[:], data[97:161]) || !bytes.Equal(state.AvailableBalance[:], data[161:225]) || !bytes.Equal(state.DecryptableAvailableBalance[:], data[225:261]) {
		t.Fatalf("balance fields = %+v", state)
	}
	if !state.AllowConfidentialCredits || state.AllowNonConfidentialCredits {
		t.Fatalf("credit flags = %v, %v", state.AllowConfidentialCredits, state.AllowNonConfidentialCredits)
	}
	if state.PendingBalanceCreditCounter != binary.LittleEndian.Uint64(data[263:271]) || state.MaximumPendingBalanceCreditCounter != binary.LittleEndian.Uint64(data[271:279]) || state.ExpectedPendingBalanceCreditCounter != binary.LittleEndian.Uint64(data[279:287]) || state.ActualPendingBalanceCreditCounter != binary.LittleEndian.Uint64(data[287:295]) {
		t.Fatalf("counters = %d, %d, %d, %d", state.PendingBalanceCreditCounter, state.MaximumPendingBalanceCreditCounter, state.ExpectedPendingBalanceCreditCounter, state.ActualPendingBalanceCreditCounter)
	}
}

func TestConfidentialTransferServiceDecodeFeeState(t *testing.T) {
	configData := make([]byte, 129)
	for i := range configData {
		configData[i] = byte(i + 1)
	}
	configData[64] = 1
	service := ConfidentialTransferService{}
	config, err := service.DecodeConfidentialTransferFeeConfig(configData)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(config.Authority[:], configData[:32]) || !bytes.Equal(config.WithdrawWithheldAuthorityElGamalPubkey[:], configData[32:64]) || !config.HarvestToMintEnabled || !bytes.Equal(config.WithheldAmount[:], configData[65:]) {
		t.Fatalf("config = %+v", config)
	}
	amountData := make([]byte, 64)
	for i := range amountData {
		amountData[i] = byte(255 - i)
	}
	amount, err := service.DecodeConfidentialTransferFeeAmount(amountData)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(amount.WithheldAmount[:], amountData) {
		t.Fatalf("amount = %x", amount.WithheldAmount)
	}
}

func TestConfidentialTransferServiceDecodeMintBurn(t *testing.T) {
	data := make([]byte, 196)
	for i := range data {
		data[i] = byte(i + 1)
	}
	state, err := (ConfidentialTransferService{}).DecodeConfidentialMintBurn(data)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(state.ConfidentialSupply[:], data[:64]) || !bytes.Equal(state.DecryptableSupply[:], data[64:100]) || !bytes.Equal(state.SupplyElGamalPubkey[:], data[100:132]) || !bytes.Equal(state.PendingBurn[:], data[132:]) {
		t.Fatalf("state = %+v", state)
	}
}

func TestConfidentialTransferServiceRejectsInvalidStateSizes(t *testing.T) {
	service := ConfidentialTransferService{}
	if _, err := service.DecodeConfidentialTransferMint(make([]byte, 64)); err == nil {
		t.Fatal("expected transfer mint size error")
	}
	if _, err := service.DecodeConfidentialTransferAccount(make([]byte, 296)); err == nil {
		t.Fatal("expected transfer account size error")
	}
	if _, err := service.DecodeConfidentialTransferFeeConfig(make([]byte, 128)); err == nil {
		t.Fatal("expected fee config size error")
	}
	if _, err := service.DecodeConfidentialTransferFeeAmount(make([]byte, 65)); err == nil {
		t.Fatal("expected fee amount size error")
	}
	if _, err := service.DecodeConfidentialMintBurn(make([]byte, 195)); err == nil {
		t.Fatal("expected mint burn size error")
	}
}
