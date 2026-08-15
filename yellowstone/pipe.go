package yellowstone

import (
	"errors"
	"io"
	"math/big"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

// AccountSink receives realtime account updates, keyed and slot-ordered.
// *rpc.Client satisfies it once its cache is enabled (see rpc.Client's
// EnableCache and CacheStoreStreamed).
type AccountSink interface {
	CacheStoreStreamed(account solana.PublicKey, data *rpc.Account, slot uint64)
}

// CacheSink additionally receives chain-head updates (slots, block heights,
// latest blockhashes). *rpc.Client satisfies it once its cache is enabled.
type CacheSink interface {
	AccountSink
	CacheStoreSlot(commitment rpc.CommitmentType, slot uint64)
	CacheStoreBlockHeight(commitment rpc.CommitmentType, height uint64)
	CacheStoreLatestBlockhash(commitment rpc.CommitmentType, hash solana.Hash, lastValidBlockHeight, slot uint64)
}

// maxBlockhashAge is how many blocks a blockhash stays valid for; a fresh
// block's hash has lastValidBlockHeight = blockHeight + maxBlockhashAge.
const maxBlockhashAge = 150

// Pipe forwards every account, slot and block-metadata update from the
// stream into sink until the stream ends, returning the stream's terminal
// error (nil on clean shutdown). It powers the rpc.Client cache end to
// end: account updates keep GetAccountInfo/GetMultipleAccounts local, slot
// updates keep GetSlot local, and block metadata keeps
// GetBlockHeight/GetLatestBlockhash/IsBlockhashValid local.
//
// commitment must be the commitment level of the subscription: slot
// updates carry their own status, but account and block-meta updates
// inherit the stream's level. Subscribe with the filters matching what you
// want mirrored — Slots() for slots, BlocksMeta() for blockhashes and
// heights, account filters for accounts:
//
//	req := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED)
//	yellowstone.AddAccounts(req, "watched", yellowstone.AccountsByOwner(owner))
//	yellowstone.AddSlots(req, "slots", yellowstone.Slots())
//	yellowstone.AddBlocksMeta(req, "blocks", yellowstone.BlocksMeta())
//	stream, err := client.Subscribe(ctx, req)
//	go yellowstone.Pipe(stream, rpc.CommitmentConfirmed, rpcClient)
func Pipe(stream *Stream, commitment rpc.CommitmentType, sink CacheSink) error {
	for {
		update, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil // clean shutdown
			}
			return err
		}

		switch {
		case update.GetAccount() != nil:
			storeAccount(update.GetAccount(), sink)

		case update.GetSlot() != nil:
			slot := update.GetSlot()
			if c, ok := slotStatusCommitment(slot.Status); ok {
				sink.CacheStoreSlot(c, slot.Slot)
			}

		case update.GetBlockMeta() != nil:
			meta := update.GetBlockMeta()
			if meta.BlockHeight != nil {
				height := meta.BlockHeight.BlockHeight
				sink.CacheStoreBlockHeight(commitment, height)
				if hash, err := solana.HashFromBase58(meta.Blockhash); err == nil {
					sink.CacheStoreLatestBlockhash(commitment, hash, height+maxBlockhashAge, meta.Slot)
				}
			}
		}
	}
}

// slotStatusCommitment maps geyser slot statuses onto RPC commitment
// levels; lifecycle-only statuses (created bank, dead, ...) report false.
func slotStatusCommitment(status pb.SlotStatus) (rpc.CommitmentType, bool) {
	switch status {
	case pb.SlotStatus_SLOT_PROCESSED:
		return rpc.CommitmentProcessed, true
	case pb.SlotStatus_SLOT_CONFIRMED:
		return rpc.CommitmentConfirmed, true
	case pb.SlotStatus_SLOT_FINALIZED:
		return rpc.CommitmentFinalized, true
	default:
		return "", false
	}
}

// PipeAccounts forwards every account update from the stream into sink
// until the stream ends, returning the stream's terminal error (nil on
// clean shutdown). Account data is copied, so sink retention is safe.
// Non-account updates (slots, pings, ...) are ignored.
//
// Subscribe with the account filters you want mirrored locally and reads
// for those accounts served through the sink never hit the RPC endpoint:
//
//	req := yellowstone.NewRequest(pb.CommitmentLevel_PROCESSED)
//	yellowstone.AddAccounts(req, "watched", yellowstone.AccountsByOwner(owner))
//	stream, err := client.Subscribe(ctx, req)
//	go yellowstone.PipeAccounts(stream, rpcClient)
func PipeAccounts(stream *Stream, sink AccountSink) error {
	for {
		update, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil // clean shutdown
			}
			return err
		}
		if acct := update.GetAccount(); acct != nil {
			storeAccount(acct, sink)
		}
	}
}

// storeAccount converts one geyser account update (copying its data out of
// the protobuf buffer, since cached entries outlive the recv) and hands it
// to the sink.
func storeAccount(acct *pb.SubscribeUpdateAccount, sink AccountSink) {
	converted := ConvertAccount(acct)
	if converted == nil {
		return
	}

	data := make([]byte, len(converted.Data))
	copy(data, converted.Data)

	sink.CacheStoreStreamed(converted.Pubkey, &rpc.Account{
		Lamports:   converted.Lamports,
		Owner:      converted.Owner,
		Data:       rpc.DataBytesOrJSONFromBytes(data),
		Executable: converted.Executable,
		RentEpoch:  new(big.Int).SetUint64(converted.RentEpoch),
		Space:      uint64(len(data)),
	}, converted.Slot)
}
