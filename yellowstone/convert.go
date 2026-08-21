package yellowstone

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	solana "github.com/fluxrpc/solana-go"
	rpc "github.com/fluxrpc/solana-go/rpc"
	pb "github.com/fluxrpc/solana-go/yellowstone/proto"
)

// TransactionError is the failure of a geyser-streamed transaction. Geyser
// carries the error as bincode-serialized bytes of the Rust
// TransactionError enum; decoding the full enum is out of scope, so the raw
// bytes are exposed as-is with the leading variant index split out
// best-effort. A non-nil *TransactionError in TransactionMeta.Err means the
// transaction failed.
type TransactionError struct {
	// Raw is the bincode encoding of the Rust TransactionError enum,
	// aliasing the pb buffer.
	Raw []byte
}

// Variant returns the bincode enum variant index (the first 4 bytes,
// little-endian), or -1 if the payload is too short.
func (e *TransactionError) Variant() int {
	if len(e.Raw) < 4 {
		return -1
	}
	return int(binary.LittleEndian.Uint32(e.Raw))
}

// Error implements the error interface.
func (e *TransactionError) Error() string {
	return fmt.Sprintf("transaction error (bincode variant %d, %d bytes)", e.Variant(), len(e.Raw))
}

// pubkeyFromBytes copies up to 32 bytes into a PublicKey; unlike
// solana.PublicKeyFromBytes it does not panic on short input, so malformed
// stream data cannot take the consumer down.
func pubkeyFromBytes(in []byte) (out solana.PublicKey) {
	copy(out[:], in)
	return
}

func hashFromBytes(in []byte) (out solana.Hash) {
	copy(out[:], in)
	return
}

// Transaction converts this update's transaction payload into the core SDK's
// transaction and metadata types. For throughput, byte-slice fields alias the
// protobuf buffers; copy them before retaining them beyond subsequent receives.
func (u *Update) Transaction() (*solana.Transaction, *rpc.TransactionMeta, error) {
	if u == nil || u.SubscribeUpdate == nil {
		return nil, nil, errors.New("yellowstone update has no transaction")
	}
	info := u.GetTransaction().GetTransaction()
	if info == nil || info.Transaction == nil {
		return nil, nil, errors.New("transaction update has no transaction")
	}
	msg := info.Transaction.Message
	if msg == nil {
		return nil, nil, errors.New("transaction update has no message")
	}
	if msg.Config != nil {
		return nil, nil, errors.New("transaction update is version 1; use TransactionV1")
	}

	tx := &solana.Transaction{
		Signatures: make([]solana.Signature, len(info.Transaction.Signatures)),
	}
	for i, sig := range info.Transaction.Signatures {
		tx.Signatures[i] = solana.SignatureFromBytes(sig)
	}

	if h := msg.Header; h != nil {
		tx.Message.Header = solana.MessageHeader{
			NumRequiredSignatures:       uint8(h.NumRequiredSignatures),
			NumReadonlySignedAccounts:   uint8(h.NumReadonlySignedAccounts),
			NumReadonlyUnsignedAccounts: uint8(h.NumReadonlyUnsignedAccounts),
		}
	}
	tx.Message.AccountKeys = make([]solana.PublicKey, len(msg.AccountKeys))
	for i, key := range msg.AccountKeys {
		tx.Message.AccountKeys[i] = pubkeyFromBytes(key)
	}
	tx.Message.RecentBlockhash = hashFromBytes(msg.RecentBlockhash)
	tx.Message.Instructions = convertInstructions(msg.Instructions)

	if msg.Versioned {
		if _, err := tx.Message.SetVersion(solana.MessageVersionV0); err != nil {
			return nil, nil, err
		}
		if len(msg.AddressTableLookups) > 0 {
			lookups := make([]solana.MessageAddressTableLookup, len(msg.AddressTableLookups))
			for i, lookup := range msg.AddressTableLookups {
				lookups[i] = solana.MessageAddressTableLookup{
					AccountKey:      pubkeyFromBytes(lookup.AccountKey),
					WritableIndexes: solana.Uint8SliceAsNum(lookup.WritableIndexes),
					ReadonlyIndexes: solana.Uint8SliceAsNum(lookup.ReadonlyIndexes),
				}
			}
			tx.Message.SetAddressTableLookups(lookups)
		}
	}

	meta, err := convertMeta(info.Meta)
	if err != nil {
		return nil, nil, err
	}
	return tx, meta, nil
}

// TransactionV1 converts a streamed v1 transaction and its metadata.
func (u *Update) TransactionV1() (*solana.TransactionV1, *rpc.TransactionMeta, error) {
	if u == nil || u.SubscribeUpdate == nil {
		return nil, nil, errors.New("yellowstone update has no transaction")
	}
	info := u.GetTransaction().GetTransaction()
	if info == nil || info.Transaction == nil || info.Transaction.Message == nil {
		return nil, nil, errors.New("transaction update has no transaction message")
	}
	msg := info.Transaction.Message
	if msg.Config == nil {
		return nil, nil, errors.New("transaction update is not version 1")
	}
	if len(msg.AddressTableLookups) != 0 {
		return nil, nil, errors.New("v1 transaction contains address table lookups")
	}

	tx := &solana.TransactionV1{
		Config: solana.TransactionConfig{
			PriorityFeeLamports:         msg.Config.PriorityFee,
			ComputeUnitLimit:            msg.Config.ComputeUnitLimit,
			LoadedAccountsDataSizeLimit: msg.Config.LoadedAccountsDataSizeLimit,
			HeapSize:                    msg.Config.HeapSize,
		},
		LifetimeSpecifier: hashFromBytes(msg.RecentBlockhash),
		AccountKeys:       convertPubkeys(msg.AccountKeys),
		Instructions:      convertInstructions(msg.Instructions),
		Signatures:        make([]solana.Signature, len(info.Transaction.Signatures)),
	}
	if h := msg.Header; h != nil {
		tx.Header = solana.MessageHeader{
			NumRequiredSignatures:       uint8(h.NumRequiredSignatures),
			NumReadonlySignedAccounts:   uint8(h.NumReadonlySignedAccounts),
			NumReadonlyUnsignedAccounts: uint8(h.NumReadonlyUnsignedAccounts),
		}
	}
	for i, sig := range info.Transaction.Signatures {
		tx.Signatures[i] = solana.SignatureFromBytes(sig)
	}
	meta, err := convertMeta(info.Meta)
	return tx, meta, err
}

func convertInstructions(in []*pb.CompiledInstruction) []solana.CompiledInstruction {
	out := make([]solana.CompiledInstruction, len(in))
	for i, ins := range in {
		out[i] = solana.CompiledInstruction{
			ProgramIDIndex: uint16(ins.ProgramIdIndex),
			Accounts:       convertAccountIndexes(ins.Accounts),
			Data:           solana.Base58(ins.Data),
		}
	}
	return out
}

func convertAccountIndexes(in []byte) []uint16 {
	out := make([]uint16, len(in))
	for i, idx := range in {
		out[i] = uint16(idx)
	}
	return out
}

func convertMeta(m *pb.TransactionStatusMeta) (*rpc.TransactionMeta, error) {
	if m == nil {
		return nil, nil
	}
	meta := &rpc.TransactionMeta{
		Fee:                  m.Fee,
		PreBalances:          m.PreBalances,
		PostBalances:         m.PostBalances,
		LogMessages:          m.LogMessages,
		ComputeUnitsConsumed: m.ComputeUnitsConsumed,
		CostUnits:            m.CostUnits,
	}
	if m.Err != nil && len(m.Err.Err) > 0 {
		meta.Err = &TransactionError{Raw: m.Err.Err}
	}

	if len(m.InnerInstructions) > 0 {
		meta.InnerInstructions = make([]rpc.InnerInstruction, len(m.InnerInstructions))
		for i, inner := range m.InnerInstructions {
			instructions := make([]rpc.CompiledInstruction, len(inner.Instructions))
			for j, ins := range inner.Instructions {
				instructions[j] = rpc.CompiledInstruction{
					ProgramIDIndex: uint16(ins.ProgramIdIndex),
					Accounts:       convertAccountIndexes(ins.Accounts),
					Data:           solana.Base58(ins.Data),
					StackHeight:    uint16(ins.GetStackHeight()),
				}
			}
			meta.InnerInstructions[i] = rpc.InnerInstruction{
				Index:        uint16(inner.Index),
				Instructions: instructions,
			}
		}
	}

	var err error
	if meta.PreTokenBalances, err = convertTokenBalances(m.PreTokenBalances); err != nil {
		return nil, fmt.Errorf("pre token balances: %w", err)
	}
	if meta.PostTokenBalances, err = convertTokenBalances(m.PostTokenBalances); err != nil {
		return nil, fmt.Errorf("post token balances: %w", err)
	}

	if len(m.Rewards) > 0 {
		meta.Rewards = make([]rpc.BlockReward, len(m.Rewards))
		for i, reward := range m.Rewards {
			pubkey, err := solana.PublicKeyFromBase58(reward.Pubkey)
			if err != nil {
				return nil, fmt.Errorf("reward pubkey: %w", err)
			}
			meta.Rewards[i] = rpc.BlockReward{
				Pubkey:      pubkey,
				Lamports:    reward.Lamports,
				PostBalance: reward.PostBalance,
				RewardType:  rpc.RewardType(reward.RewardType.String()),
			}
			if reward.Commission != "" {
				commission, err := strconv.ParseUint(reward.Commission, 10, 8)
				if err != nil {
					return nil, fmt.Errorf("reward commission: %w", err)
				}
				c := uint8(commission)
				meta.Rewards[i].Commission = &c
			}
			if reward.CommissionBps != "" {
				commission, err := strconv.ParseUint(reward.CommissionBps, 10, 16)
				if err != nil {
					return nil, fmt.Errorf("reward commission bps: %w", err)
				}
				c := uint16(commission)
				meta.Rewards[i].CommissionBps = &c
			}
		}
	}

	meta.LoadedAddresses = rpc.LoadedAddresses{
		Writable: convertPubkeys(m.LoadedWritableAddresses),
		ReadOnly: convertPubkeys(m.LoadedReadonlyAddresses),
	}
	if m.ReturnData != nil {
		meta.ReturnData = rpc.ReturnData{
			ProgramId: pubkeyFromBytes(m.ReturnData.ProgramId),
			Data:      solana.Data{Content: m.ReturnData.Data, Encoding: solana.EncodingBase64},
		}
	}
	return meta, nil
}

func convertTokenBalances(in []*pb.TokenBalance) ([]rpc.TokenBalance, error) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make([]rpc.TokenBalance, len(in))
	for i, balance := range in {
		mint, err := solana.PublicKeyFromBase58(balance.Mint)
		if err != nil {
			return nil, fmt.Errorf("mint: %w", err)
		}
		out[i] = rpc.TokenBalance{
			AccountIndex: uint16(balance.AccountIndex),
			Mint:         mint,
		}
		if balance.Owner != "" {
			owner, err := solana.PublicKeyFromBase58(balance.Owner)
			if err != nil {
				return nil, fmt.Errorf("owner: %w", err)
			}
			out[i].Owner = &owner
		}
		if balance.ProgramId != "" {
			program, err := solana.PublicKeyFromBase58(balance.ProgramId)
			if err != nil {
				return nil, fmt.Errorf("program id: %w", err)
			}
			out[i].ProgramId = &program
		}
		if amount := balance.UiTokenAmount; amount != nil {
			uiAmount := amount.UiAmount
			out[i].UiTokenAmount = &rpc.UiTokenAmount{
				Amount:         amount.Amount,
				Decimals:       uint8(amount.Decimals),
				UiAmount:       &uiAmount,
				UiAmountString: amount.UiAmountString,
			}
		}
	}
	return out, nil
}

func convertPubkeys(in [][]byte) []solana.PublicKey {
	if len(in) == 0 {
		return nil
	}
	out := make([]solana.PublicKey, len(in))
	for i, key := range in {
		out[i] = pubkeyFromBytes(key)
	}
	return out
}

// AccountUpdate is a geyser account update flattened into SDK types.
type AccountUpdate struct {
	Pubkey       solana.PublicKey
	Lamports     uint64
	Owner        solana.PublicKey
	Executable   bool
	RentEpoch    uint64
	Data         []byte
	WriteVersion uint64
	TxnSignature *solana.Signature
	Slot         uint64
	IsStartup    bool
}

// Account converts this update's account payload. The returned Data aliases
// the protobuf buffer; copy it before retaining it beyond subsequent receives.
func (u *Update) Account() *AccountUpdate {
	if u == nil || u.SubscribeUpdate == nil {
		return nil
	}
	account := u.GetAccount()
	info := account.GetAccount()
	if info == nil {
		return nil
	}
	update := &AccountUpdate{
		Pubkey:       pubkeyFromBytes(info.Pubkey),
		Lamports:     info.Lamports,
		Owner:        pubkeyFromBytes(info.Owner),
		Executable:   info.Executable,
		RentEpoch:    info.RentEpoch,
		Data:         info.Data,
		WriteVersion: info.WriteVersion,
		Slot:         account.Slot,
		IsStartup:    account.IsStartup,
	}
	if len(info.TxnSignature) > 0 {
		sig := solana.SignatureFromBytes(info.TxnSignature)
		update.TxnSignature = &sig
	}
	return update
}
