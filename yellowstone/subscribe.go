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

// Request owns a Yellowstone subscription request and its named filters.
// The embedded protobuf request remains accessible for advanced filters not
// covered by the convenience methods.
type Request struct {
	*pb.SubscribeRequest
}

// Update owns one Yellowstone stream update and provides conversion methods
// for account and transaction payloads.
type Update struct {
	*pb.SubscribeUpdate
}

// NewUpdate wraps a raw Yellowstone protobuf update so its typed conversion
// methods can be used outside a Stream.
func NewUpdate(update *pb.SubscribeUpdate) *Update {
	return &Update{SubscribeUpdate: update}
}

// Subscribe opens the bidi stream and sends req before returning, so the
// first Recv already delivers filtered updates.
func (c *Client) Subscribe(ctx context.Context, req *Request) (*Stream, error) {
	ctx, cancel := context.WithCancel(c.withToken(ctx))
	stream, err := c.geyser.Subscribe(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	if err := stream.Send(req.SubscribeRequest); err != nil {
		cancel()
		return nil, err
	}
	return &Stream{stream: stream, cancel: cancel}, nil
}

// Recv blocks for the next update. The returned update (including any
// account data byte slices) is owned by the caller, but conversion methods
// alias into it — see Update.Account.
func (s *Stream) Recv() (*Update, error) {
	update, err := s.stream.Recv()
	if err != nil {
		return nil, err
	}
	return NewUpdate(update), nil
}

// Update replaces the subscription's filters without reconnecting.
func (s *Stream) Update(req *Request) error {
	return s.stream.Send(req.SubscribeRequest)
}

// Close ends the subscription; a pending Recv unblocks with an error.
func (s *Stream) Close() error {
	err := s.stream.CloseSend()
	s.cancel()
	return err
}

// NewRequest returns a SubscribeRequest at the given commitment with every
// filter map initialized, ready for the Add* helpers below.
func NewRequest(commitment pb.CommitmentLevel) *Request {
	return &Request{SubscribeRequest: &pb.SubscribeRequest{
		Accounts:           map[string]*pb.SubscribeRequestFilterAccounts{},
		Slots:              map[string]*pb.SubscribeRequestFilterSlots{},
		Transactions:       map[string]*pb.SubscribeRequestFilterTransactions{},
		TransactionsStatus: map[string]*pb.SubscribeRequestFilterTransactions{},
		Blocks:             map[string]*pb.SubscribeRequestFilterBlocks{},
		BlocksMeta:         map[string]*pb.SubscribeRequestFilterBlocksMeta{},
		Entry:              map[string]*pb.SubscribeRequestFilterEntry{},
		Commitment:         &commitment,
	}}
}

// AddAccounts registers an account filter under name. Matching updates carry
// name in Update.Filters.
func (r *Request) AddAccounts(name string, filter *pb.SubscribeRequestFilterAccounts) *Request {
	r.Accounts[name] = filter
	return r
}

// AccountsByOwner registers an account filter matching any owner key.
func (r *Request) AccountsByOwner(name string, owners ...string) *Request {
	return r.AddAccounts(name, &pb.SubscribeRequestFilterAccounts{Owner: owners})
}

// AccountsByKey registers an account filter matching any account key.
func (r *Request) AccountsByKey(name string, keys ...string) *Request {
	return r.AddAccounts(name, &pb.SubscribeRequestFilterAccounts{Account: keys})
}

// AddSlots registers a slot filter under name.
func (r *Request) AddSlots(name string, filter *pb.SubscribeRequestFilterSlots) *Request {
	r.Slots[name] = filter
	return r
}

// AllSlots registers a filter matching every slot-status update.
func (r *Request) AllSlots(name string) *Request {
	return r.AddSlots(name, &pb.SubscribeRequestFilterSlots{})
}

// AddTransactions registers a transaction filter under name.
func (r *Request) AddTransactions(name string, filter *pb.SubscribeRequestFilterTransactions) *Request {
	r.Transactions[name] = filter
	return r
}

// TransactionsByAccount registers a transaction filter matching any account.
func (r *Request) TransactionsByAccount(name string, accounts ...string) *Request {
	return r.AddTransactions(name, &pb.SubscribeRequestFilterTransactions{AccountInclude: accounts})
}

// TransactionsByAccountRequired registers a transaction filter that requires
// every supplied account.
func (r *Request) TransactionsByAccountRequired(name string, accounts ...string) *Request {
	return r.AddTransactions(name, &pb.SubscribeRequestFilterTransactions{AccountRequired: accounts})
}

// AddTransactionStatuses registers a transaction-status filter under name.
func (r *Request) AddTransactionStatuses(name string, filter *pb.SubscribeRequestFilterTransactions) *Request {
	r.TransactionsStatus[name] = filter
	return r
}

// TransactionStatusesByAccount registers a status-only filter matching any
// supplied account.
func (r *Request) TransactionStatusesByAccount(name string, accounts ...string) *Request {
	return r.AddTransactionStatuses(name, &pb.SubscribeRequestFilterTransactions{AccountInclude: accounts})
}

// TransactionStatusesByAccountRequired registers a status-only filter that
// requires every supplied account.
func (r *Request) TransactionStatusesByAccountRequired(name string, accounts ...string) *Request {
	return r.AddTransactionStatuses(name, &pb.SubscribeRequestFilterTransactions{AccountRequired: accounts})
}

// AddBlocks registers a full-block filter under name.
func (r *Request) AddBlocks(name string, filter *pb.SubscribeRequestFilterBlocks) *Request {
	r.Blocks[name] = filter
	return r
}

// BlocksIncluding registers a block filter whose transactions mention any of
// the supplied accounts. With no accounts it matches full blocks.
func (r *Request) BlocksIncluding(name string, accounts ...string) *Request {
	return r.AddBlocks(name, &pb.SubscribeRequestFilterBlocks{AccountInclude: accounts})
}

// AddBlocksMeta registers a block-metadata filter under name.
func (r *Request) AddBlocksMeta(name string, filter *pb.SubscribeRequestFilterBlocksMeta) *Request {
	r.BlocksMeta[name] = filter
	return r
}

// AllBlocksMeta registers a filter matching all block metadata.
func (r *Request) AllBlocksMeta(name string) *Request {
	return r.AddBlocksMeta(name, &pb.SubscribeRequestFilterBlocksMeta{})
}

// AddEntries registers a ledger-entry filter under name.
func (r *Request) AddEntries(name string, filter *pb.SubscribeRequestFilterEntry) *Request {
	r.Entry[name] = filter
	return r
}

// AllEntries registers a filter matching every ledger entry.
func (r *Request) AllEntries(name string) *Request {
	return r.AddEntries(name, &pb.SubscribeRequestFilterEntry{})
}
