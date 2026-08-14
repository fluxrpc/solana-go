package yellowstone

import (
	"context"

	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
)

// Stream is a live geyser subscription. Recv is safe from one goroutine
// while another calls Update/Close (the usual gRPC stream contract).
type Stream struct {
	stream grpc.BidiStreamingClient[pb.SubscribeRequest, pb.SubscribeUpdate]
	cancel context.CancelFunc
}

// Subscribe opens the bidi stream and sends req before returning, so the
// first Recv already delivers filtered updates.
func (c *Client) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*Stream, error) {
	ctx, cancel := context.WithCancel(c.withToken(ctx))
	stream, err := c.geyser.Subscribe(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	if err := stream.Send(req); err != nil {
		cancel()
		return nil, err
	}
	return &Stream{stream: stream, cancel: cancel}, nil
}

// Recv blocks for the next update. The returned update (including any
// account data byte slices) is owned by the caller, but converters in this
// package alias into it — see ConvertAccount.
func (s *Stream) Recv() (*pb.SubscribeUpdate, error) {
	return s.stream.Recv()
}

// Update replaces the subscription's filters without reconnecting.
func (s *Stream) Update(req *pb.SubscribeRequest) error {
	return s.stream.Send(req)
}

// Close ends the subscription; a pending Recv unblocks with an error.
func (s *Stream) Close() error {
	err := s.stream.CloseSend()
	s.cancel()
	return err
}

// NewRequest returns a SubscribeRequest at the given commitment with every
// filter map initialized, ready for the Add* helpers below.
func NewRequest(commitment pb.CommitmentLevel) *pb.SubscribeRequest {
	return &pb.SubscribeRequest{
		Accounts:           map[string]*pb.SubscribeRequestFilterAccounts{},
		Slots:              map[string]*pb.SubscribeRequestFilterSlots{},
		Transactions:       map[string]*pb.SubscribeRequestFilterTransactions{},
		TransactionsStatus: map[string]*pb.SubscribeRequestFilterTransactions{},
		Blocks:             map[string]*pb.SubscribeRequestFilterBlocks{},
		BlocksMeta:         map[string]*pb.SubscribeRequestFilterBlocksMeta{},
		Entry:              map[string]*pb.SubscribeRequestFilterEntry{},
		Commitment:         &commitment,
	}
}

// AddAccounts registers an accounts filter under name; updates matching it
// carry name in SubscribeUpdate.Filters.
func AddAccounts(req *pb.SubscribeRequest, name string, filter *pb.SubscribeRequestFilterAccounts) {
	req.Accounts[name] = filter
}

// AddSlots registers a slots filter under name.
func AddSlots(req *pb.SubscribeRequest, name string, filter *pb.SubscribeRequestFilterSlots) {
	req.Slots[name] = filter
}

// AddTransactions registers a transactions filter under name.
func AddTransactions(req *pb.SubscribeRequest, name string, filter *pb.SubscribeRequestFilterTransactions) {
	req.Transactions[name] = filter
}

// AddBlocks registers a blocks filter under name.
func AddBlocks(req *pb.SubscribeRequest, name string, filter *pb.SubscribeRequestFilterBlocks) {
	req.Blocks[name] = filter
}

// AddBlocksMeta registers a block-meta filter under name.
func AddBlocksMeta(req *pb.SubscribeRequest, name string, filter *pb.SubscribeRequestFilterBlocksMeta) {
	req.BlocksMeta[name] = filter
}

// AddEntries registers an entries filter under name.
func AddEntries(req *pb.SubscribeRequest, name string, filter *pb.SubscribeRequestFilterEntry) {
	req.Entry[name] = filter
}

// AccountsByOwner matches accounts owned by any of the base58 program keys.
func AccountsByOwner(owners ...string) *pb.SubscribeRequestFilterAccounts {
	return &pb.SubscribeRequestFilterAccounts{Owner: owners}
}

// AccountsByKey matches the given base58 account keys.
func AccountsByKey(keys ...string) *pb.SubscribeRequestFilterAccounts {
	return &pb.SubscribeRequestFilterAccounts{Account: keys}
}

// TransactionsByAccount matches transactions that mention ANY of the given
// base58 accounts (pb account_include semantics).
func TransactionsByAccount(accounts ...string) *pb.SubscribeRequestFilterTransactions {
	return &pb.SubscribeRequestFilterTransactions{AccountInclude: accounts}
}

// TransactionsByAccountRequired matches transactions that mention ALL of the
// given base58 accounts (pb account_required semantics).
func TransactionsByAccountRequired(accounts ...string) *pb.SubscribeRequestFilterTransactions {
	return &pb.SubscribeRequestFilterTransactions{AccountRequired: accounts}
}

// Slots matches every slot status update.
func Slots() *pb.SubscribeRequestFilterSlots {
	return &pb.SubscribeRequestFilterSlots{}
}

// Blocks matches full blocks; with accounts given, only transactions
// mentioning them are included (pb account_include semantics).
func Blocks(accountInclude ...string) *pb.SubscribeRequestFilterBlocks {
	return &pb.SubscribeRequestFilterBlocks{AccountInclude: accountInclude}
}

// BlocksMeta matches block metadata (no transactions or accounts).
func BlocksMeta() *pb.SubscribeRequestFilterBlocksMeta {
	return &pb.SubscribeRequestFilterBlocksMeta{}
}

// Entries matches ledger entries.
func Entries() *pb.SubscribeRequestFilterEntry {
	return &pb.SubscribeRequestFilterEntry{}
}
