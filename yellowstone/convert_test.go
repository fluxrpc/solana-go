package yellowstone

import (
	"bytes"
	"testing"

	solana "github.com/fluxrpc/solana-go"
	rpc "github.com/fluxrpc/solana-go/rpc"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

const (
	usdcMint     = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	tokenProgram = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
)

// testTransaction builds and signs a transaction with the root SDK; v0 adds
// address table lookups.
func testTransaction(t testing.TB, v0 bool) *solana.Transaction {
	t.Helper()
	key, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	payer := key.PublicKey()
	tx := &solana.Transaction{
		Message: solana.Message{
			AccountKeys: []solana.PublicKey{payer, solana.MustPublicKeyFromBase58(tokenProgram)},
			Header:      solana.MessageHeader{NumRequiredSignatures: 1, NumReadonlyUnsignedAccounts: 1},
			Instructions: []solana.CompiledInstruction{{
				ProgramIDIndex: 1,
				Accounts:       []uint16{0, 2, 3},
				Data:           solana.Base58{2, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
			}},
			RecentBlockhash: solana.Hash{7, 7, 7},
		},
	}
	if v0 {
		tx.Message.SetAddressTableLookups([]solana.MessageAddressTableLookup{{
			AccountKey:      solana.MustPublicKeyFromBase58(usdcMint),
			WritableIndexes: solana.Uint8SliceAsNum{4, 9},
			ReadonlyIndexes: solana.Uint8SliceAsNum{1},
		}})
	}
	if _, err := tx.Sign(func(pub solana.PublicKey) *solana.PrivateKey {
		return &key
	}); err != nil {
		t.Fatal(err)
	}
	return tx
}

// txToProto decomposes an SDK transaction into the solana-storage pb form,
// the way geyser presents it on the wire.
func txToProto(tx *solana.Transaction) *pb.Transaction {
	msg := tx.Message
	out := &pb.Transaction{
		Message: &pb.Message{
			Header: &pb.MessageHeader{
				NumRequiredSignatures:       uint32(msg.Header.NumRequiredSignatures),
				NumReadonlySignedAccounts:   uint32(msg.Header.NumReadonlySignedAccounts),
				NumReadonlyUnsignedAccounts: uint32(msg.Header.NumReadonlyUnsignedAccounts),
			},
			RecentBlockhash: msg.RecentBlockhash.Bytes(),
			Versioned:       msg.GetVersion() == solana.MessageVersionV0,
		},
	}
	for _, sig := range tx.Signatures {
		out.Signatures = append(out.Signatures, sig.Bytes())
	}
	for _, key := range msg.AccountKeys {
		out.Message.AccountKeys = append(out.Message.AccountKeys, key.Bytes())
	}
	for _, ins := range msg.Instructions {
		accounts := make([]byte, len(ins.Accounts))
		for i, idx := range ins.Accounts {
			accounts[i] = byte(idx)
		}
		out.Message.Instructions = append(out.Message.Instructions, &pb.CompiledInstruction{
			ProgramIdIndex: uint32(ins.ProgramIDIndex),
			Accounts:       accounts,
			Data:           ins.Data,
		})
	}
	for _, lookup := range msg.AddressTableLookups {
		out.Message.AddressTableLookups = append(out.Message.AddressTableLookups, &pb.MessageAddressTableLookup{
			AccountKey:      lookup.AccountKey.Bytes(),
			WritableIndexes: lookup.WritableIndexes,
			ReadonlyIndexes: lookup.ReadonlyIndexes,
		})
	}
	return out
}

func TestConvertTransactionRoundTrip(t *testing.T) {
	for _, v0 := range []bool{false, true} {
		original := testTransaction(t, v0)
		wire, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		converted, meta, err := ConvertTransaction(&pb.SubscribeUpdateTransaction{
			Slot: 100,
			Transaction: &pb.SubscribeUpdateTransactionInfo{
				Signature:   original.Signatures[0].Bytes(),
				Transaction: txToProto(original),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if meta != nil {
			t.Fatalf("meta = %+v, want nil", meta)
		}

		rewire, err := converted.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(wire, rewire) {
			t.Fatalf("v0=%v: converted transaction is not byte-identical\n got %x\nwant %x", v0, rewire, wire)
		}
		if err := converted.VerifySignatures(); err != nil {
			t.Fatalf("v0=%v: %v", v0, err)
		}
	}
}

func TestConvertTransactionMeta(t *testing.T) {
	stackHeight := uint32(2)
	computeUnits := uint64(4200)
	costUnits := uint64(999)
	pbMeta := &pb.TransactionStatusMeta{
		Err:          &pb.TransactionError{Err: []byte{8, 0, 0, 0, 1}},
		Fee:          5000,
		PreBalances:  []uint64{100, 200},
		PostBalances: []uint64{90, 210},
		InnerInstructions: []*pb.InnerInstructions{{
			Index: 1,
			Instructions: []*pb.InnerInstruction{{
				ProgramIdIndex: 3,
				Accounts:       []byte{0, 2},
				Data:           []byte{9, 9},
				StackHeight:    &stackHeight,
			}},
		}},
		LogMessages: []string{"Program log: hi"},
		PreTokenBalances: []*pb.TokenBalance{{
			AccountIndex: 2,
			Mint:         usdcMint,
			Owner:        tokenProgram,
			ProgramId:    tokenProgram,
			UiTokenAmount: &pb.UiTokenAmount{
				UiAmount:       1.5,
				Decimals:       6,
				Amount:         "1500000",
				UiAmountString: "1.5",
			},
		}},
		PostTokenBalances: []*pb.TokenBalance{{AccountIndex: 2, Mint: usdcMint}},
		Rewards: []*pb.Reward{{
			Pubkey:      tokenProgram,
			Lamports:    -12,
			PostBalance: 88,
			RewardType:  pb.RewardType_Rent,
			Commission:  "5",
		}},
		LoadedWritableAddresses: [][]byte{bytes.Repeat([]byte{1}, 32)},
		LoadedReadonlyAddresses: [][]byte{bytes.Repeat([]byte{2}, 32)},
		ReturnData: &pb.ReturnData{
			ProgramId: solana.MustPublicKeyFromBase58(tokenProgram).Bytes(),
			Data:      []byte{4, 5, 6},
		},
		ComputeUnitsConsumed: &computeUnits,
		CostUnits:            &costUnits,
	}

	tx := testTransaction(t, false)
	_, meta, err := ConvertTransaction(&pb.SubscribeUpdateTransaction{
		Transaction: &pb.SubscribeUpdateTransactionInfo{
			Transaction: txToProto(tx),
			Meta:        pbMeta,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	txErr, ok := meta.Err.(*TransactionError)
	if !ok || txErr.Variant() != 8 || !bytes.Equal(txErr.Raw, pbMeta.Err.Err) {
		t.Fatalf("Err = %#v", meta.Err)
	}
	if txErr.Error() == "" {
		t.Fatal("empty error string")
	}
	if meta.Fee != 5000 || meta.PreBalances[1] != 200 || meta.PostBalances[1] != 210 {
		t.Fatalf("balances = %+v", meta)
	}
	if len(meta.LogMessages) != 1 || meta.LogMessages[0] != "Program log: hi" {
		t.Fatalf("logs = %+v", meta.LogMessages)
	}

	inner := meta.InnerInstructions[0]
	ins := inner.Instructions[0]
	if inner.Index != 1 || ins.ProgramIDIndex != 3 || ins.StackHeight != 2 ||
		len(ins.Accounts) != 2 || ins.Accounts[1] != 2 || !bytes.Equal(ins.Data, []byte{9, 9}) {
		t.Fatalf("inner = %+v", inner)
	}

	pre := meta.PreTokenBalances[0]
	if pre.AccountIndex != 2 || pre.Mint.String() != usdcMint ||
		pre.Owner.String() != tokenProgram || pre.ProgramId.String() != tokenProgram {
		t.Fatalf("pre token balance = %+v", pre)
	}
	amount := pre.UiTokenAmount
	if amount.Amount != "1500000" || amount.Decimals != 6 || *amount.UiAmount != 1.5 || amount.UiAmountString != "1.5" {
		t.Fatalf("ui token amount = %+v", amount)
	}
	post := meta.PostTokenBalances[0]
	if post.Owner != nil || post.ProgramId != nil || post.Mint.String() != usdcMint {
		t.Fatalf("post token balance = %+v", post)
	}

	reward := meta.Rewards[0]
	if reward.Pubkey.String() != tokenProgram || reward.Lamports != -12 || reward.PostBalance != 88 ||
		reward.RewardType != rpc.RewardTypeRent || reward.Commission == nil || *reward.Commission != 5 {
		t.Fatalf("reward = %+v", reward)
	}

	if meta.LoadedAddresses.Writable[0] != (solana.PublicKey{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}) {
		t.Fatalf("writable = %+v", meta.LoadedAddresses.Writable)
	}
	if meta.LoadedAddresses.ReadOnly[0][0] != 2 {
		t.Fatalf("readonly = %+v", meta.LoadedAddresses.ReadOnly)
	}
	if meta.ReturnData.ProgramId.String() != tokenProgram ||
		!bytes.Equal(meta.ReturnData.Data.Content, []byte{4, 5, 6}) ||
		meta.ReturnData.Data.Encoding != solana.EncodingBase64 {
		t.Fatalf("return data = %+v", meta.ReturnData)
	}
	if *meta.ComputeUnitsConsumed != 4200 || *meta.CostUnits != 999 {
		t.Fatalf("units = %+v", meta)
	}
}

func TestConvertTransactionMetaSuccessHasNilErr(t *testing.T) {
	tx := testTransaction(t, false)
	_, meta, err := ConvertTransaction(&pb.SubscribeUpdateTransaction{
		Transaction: &pb.SubscribeUpdateTransactionInfo{
			Transaction: txToProto(tx),
			Meta:        &pb.TransactionStatusMeta{Fee: 5000},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if meta.Err != nil {
		t.Fatalf("Err = %#v, want nil for success", meta.Err)
	}
}

func TestConvertTransactionMissing(t *testing.T) {
	if _, _, err := ConvertTransaction(&pb.SubscribeUpdateTransaction{}); err == nil {
		t.Fatal("expected error for empty update")
	}
}

func TestConvertAccount(t *testing.T) {
	key, err := solana.NewRandomPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pubkey := key.PublicKey()
	sig := solana.Signature{9, 9}
	data := []byte{1, 2, 3, 4}

	update := ConvertAccount(&pb.SubscribeUpdateAccount{
		Slot:      341197053,
		IsStartup: true,
		Account: &pb.SubscribeUpdateAccountInfo{
			Pubkey:       pubkey.Bytes(),
			Lamports:     500,
			Owner:        solana.MustPublicKeyFromBase58(tokenProgram).Bytes(),
			Executable:   true,
			RentEpoch:    361,
			Data:         data,
			WriteVersion: 42,
			TxnSignature: sig.Bytes(),
		},
	})
	if update.Pubkey != pubkey || update.Lamports != 500 || update.Owner.String() != tokenProgram {
		t.Fatalf("update = %+v", update)
	}
	if !update.Executable || update.RentEpoch != 361 || update.WriteVersion != 42 {
		t.Fatalf("update = %+v", update)
	}
	if update.Slot != 341197053 || !update.IsStartup {
		t.Fatalf("update = %+v", update)
	}
	if *update.TxnSignature != sig {
		t.Fatalf("signature = %s", update.TxnSignature)
	}
	// Data aliases the pb buffer by contract.
	if &update.Data[0] != &data[0] {
		t.Fatal("Data should alias the pb buffer")
	}

	if ConvertAccount(&pb.SubscribeUpdateAccount{Slot: 1}) != nil {
		t.Fatal("expected nil for update without account info")
	}
	noSig := ConvertAccount(&pb.SubscribeUpdateAccount{Account: &pb.SubscribeUpdateAccountInfo{}})
	if noSig.TxnSignature != nil {
		t.Fatal("expected nil TxnSignature")
	}
}

func TestConvertBlockhash(t *testing.T) {
	hash, err := ConvertBlockhash("EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N")
	if err != nil {
		t.Fatal(err)
	}
	if hash.IsZero() || hash.String() != "EkSnNWid2cvwEVnVx9aBqawnmiCNiDgp3gUdkDPTKN1N" {
		t.Fatalf("hash = %s", hash)
	}
	if _, err := ConvertBlockhash("not-base58!"); err == nil {
		t.Fatal("expected error for invalid blockhash")
	}
}
