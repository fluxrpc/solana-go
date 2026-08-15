package yellowstone

import (
	"errors"
	"io"
	"math/big"

	solana "github.com/fluxrpc/solana-go"
	"github.com/fluxrpc/solana-go/rpc"
)

// AccountSink receives realtime account updates, keyed and slot-ordered.
// *rpccache.Client satisfies it.
type AccountSink interface {
	StoreStreamed(account solana.PublicKey, data *rpc.Account, slot uint64)
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
//	go yellowstone.PipeAccounts(stream, cachedClient)
func PipeAccounts(stream *Stream, sink AccountSink) error {
	for {
		update, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil // clean shutdown
			}
			return err
		}
		acct := update.GetAccount()
		if acct == nil || acct.Account == nil {
			continue
		}

		converted := ConvertAccount(acct)
		if converted == nil {
			continue
		}

		// Copy out of the protobuf buffer: cached data outlives this recv.
		data := make([]byte, len(converted.Data))
		copy(data, converted.Data)

		sink.StoreStreamed(converted.Pubkey, &rpc.Account{
			Lamports:   converted.Lamports,
			Owner:      converted.Owner,
			Data:       rpc.DataBytesOrJSONFromBytes(data),
			Executable: converted.Executable,
			RentEpoch:  new(big.Int).SetUint64(converted.RentEpoch),
			Space:      uint64(len(data)),
		}, converted.Slot)
	}
}
