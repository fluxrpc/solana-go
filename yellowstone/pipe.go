package yellowstone

import (
	"errors"
	"io"
	"math/big"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
	pb "github.com/fluxrpc/solana-go/yellowstone/proto"
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

// Pipe forwards every account, slot and block-metadata update into sink until
// the stream ends, returning the stream's terminal
// error (nil on clean shutdown). It powers the rpc.Client cache end to
// end: account updates keep GetAccountInfo/GetMultipleAccounts local, slot
// updates keep GetSlot local, and block metadata keeps
// GetBlockHeight/GetLatestBlockhash/IsBlockhashValid local.
//
// commitment must be the commitment level of the subscription: slot
// updates carry their own status, but account and block-meta updates
// inherit the stream's level. Subscribe with the filters matching what you
// want mirrored — AllSlots for slots, AllBlocksMeta for blockhashes and
// heights, and account-filter methods for accounts:
//
//	req := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED).
//		AccountsByOwner("watched", owner).
//		AllSlots("slots").
//		AllBlocksMeta("blocks")
//	stream, err := client.Subscribe(ctx, req)
//	go stream.Pipe(rpc.CommitmentConfirmed, rpcClient)
func (s *Stream) Pipe(commitment rpc.CommitmentType, sink CacheSink) error {
	for {
		update, err := s.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil // clean shutdown
			}
			return err
		}

		switch {
		case update.GetAccount() != nil:
			update.storeAccount(sink)

		case update.GetSlot() != nil:
			if c, ok := update.slotCommitment(); ok {
				sink.CacheStoreSlot(c, update.GetSlot().Slot)
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

// slotCommitment maps this update's geyser slot status onto an RPC commitment
// level; lifecycle-only statuses (created bank, dead, ...) report false.
func (u *Update) slotCommitment() (rpc.CommitmentType, bool) {
	switch u.GetSlot().Status {
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

// PipeAccounts forwards every account update into sink
// until the stream ends, returning the stream's terminal error (nil on
// clean shutdown). Account data is copied, so sink retention is safe.
// Non-account updates (slots, pings, ...) are ignored.
//
// Subscribe with the account filters you want mirrored locally and reads
// for those accounts served through the sink never hit the RPC endpoint:
//
//	req := yellowstone.NewRequest(pb.CommitmentLevel_PROCESSED).
//		AccountsByOwner("watched", owner)
//	stream, err := client.Subscribe(ctx, req)
//	go stream.PipeAccounts(rpcClient)
func (s *Stream) PipeAccounts(sink AccountSink) error {
	for {
		update, err := s.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil // clean shutdown
			}
			return err
		}
		if acct := update.GetAccount(); acct != nil {
			update.storeAccount(sink)
		}
	}
}

// storeAccount converts this geyser account update (copying its data out of
// the protobuf buffer, since cached entries outlive the recv) and hands it
// to the sink.
func (u *Update) storeAccount(sink AccountSink) {
	converted := u.Account()
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
