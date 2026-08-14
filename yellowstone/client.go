// Package yellowstone is a throughput-optimized client for the Yellowstone
// gRPC Geyser plugin (Dragon's Mouth). It wraps the generated protobuf
// bindings with a small connection/option layer, a thin Subscribe stream and
// allocation-light converters into the core solana-go types.
package yellowstone

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// options collects the dialing configuration; the defaults favor throughput
// against a real geyser endpoint (huge receive window, keepalive pings).
type options struct {
	token       string
	insecure    bool
	maxRecvSize int
	keepalive   time.Duration
	grpcOpts    []grpc.DialOption
}

// Option configures Connect.
type Option func(*options)

// WithToken sets the access token, sent as `x-token` metadata on every
// unary and stream call.
func WithToken(token string) Option {
	return func(o *options) { o.token = token }
}

// WithInsecure disables TLS even for https:// or :443 endpoints.
func WithInsecure() Option {
	return func(o *options) { o.insecure = true }
}

// WithMaxRecvSize sets the maximum gRPC message size the client accepts.
// The default is 1<<30: geyser block updates routinely exceed gRPC's 4MB
// default, which makes streams die mid-subscription.
func WithMaxRecvSize(size int) Option {
	return func(o *options) { o.maxRecvSize = size }
}

// WithKeepalive sets the HTTP/2 keepalive ping interval (default 30s, with
// pings permitted while no stream is active).
func WithKeepalive(interval time.Duration) Option {
	return func(o *options) { o.keepalive = interval }
}

// WithGRPCOptions appends raw grpc.DialOptions, applied after the ones this
// package derives — an escape hatch for anything not covered above.
func WithGRPCOptions(opts ...grpc.DialOption) Option {
	return func(o *options) { o.grpcOpts = append(o.grpcOpts, opts...) }
}

// Client is a Yellowstone gRPC Geyser client. It is safe for concurrent use;
// all methods share one HTTP/2 connection.
type Client struct {
	conn   *grpc.ClientConn
	geyser pb.GeyserClient
	token  string
}

// Connect creates a client for the given endpoint. Accepted endpoint forms
// are "host:port", "http://host[:port]" and "https://host[:port]". TLS is
// used for https:// endpoints and for port 443; anything else dials
// plaintext (override with WithInsecure to force plaintext everywhere).
//
// The connection is created lazily (grpc.NewClient); the first RPC performs
// the actual dial.
func Connect(ctx context.Context, endpoint string, opts ...Option) (*Client, error) {
	o := options{
		maxRecvSize: 1 << 30,
		keepalive:   30 * time.Second,
	}
	for _, opt := range opts {
		opt(&o)
	}

	target := endpoint
	useTLS := false
	switch {
	case strings.HasPrefix(endpoint, "https://"):
		target = strings.TrimPrefix(endpoint, "https://")
		useTLS = true
	case strings.HasPrefix(endpoint, "http://"):
		target = strings.TrimPrefix(endpoint, "http://")
	}
	target = strings.TrimSuffix(target, "/")
	if strings.HasSuffix(target, ":443") {
		useTLS = true
	}

	creds := insecure.NewCredentials()
	if useTLS && !o.insecure {
		creds = credentials.NewTLS(&tls.Config{})
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(o.maxRecvSize)),
	}
	if o.keepalive > 0 {
		dialOpts = append(dialOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                o.keepalive,
			Timeout:             o.keepalive,
			PermitWithoutStream: true,
		}))
	}
	dialOpts = append(dialOpts, o.grpcOpts...)

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, geyser: pb.NewGeyserClient(conn), token: o.token}, nil
}

// Close tears down the underlying connection; all active streams die.
func (c *Client) Close() error {
	return c.conn.Close()
}

// withToken attaches the x-token metadata when a token is configured.
func (c *Client) withToken(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "x-token", c.token)
}

// Ping round-trips count through the endpoint.
func (c *Client) Ping(ctx context.Context, count int32) (*pb.PongResponse, error) {
	return c.geyser.Ping(c.withToken(ctx), &pb.PingRequest{Count: count})
}

// GetVersion returns the endpoint's version info (a JSON string).
func (c *Client) GetVersion(ctx context.Context) (*pb.GetVersionResponse, error) {
	return c.geyser.GetVersion(c.withToken(ctx), &pb.GetVersionRequest{})
}

// GetSlot returns the current slot at the given commitment.
func (c *Client) GetSlot(ctx context.Context, commitment pb.CommitmentLevel) (*pb.GetSlotResponse, error) {
	return c.geyser.GetSlot(c.withToken(ctx), &pb.GetSlotRequest{Commitment: &commitment})
}

// GetLatestBlockhash returns the latest blockhash at the given commitment.
func (c *Client) GetLatestBlockhash(ctx context.Context, commitment pb.CommitmentLevel) (*pb.GetLatestBlockhashResponse, error) {
	return c.geyser.GetLatestBlockhash(c.withToken(ctx), &pb.GetLatestBlockhashRequest{Commitment: &commitment})
}

// GetBlockHeight returns the current block height at the given commitment.
func (c *Client) GetBlockHeight(ctx context.Context, commitment pb.CommitmentLevel) (*pb.GetBlockHeightResponse, error) {
	return c.geyser.GetBlockHeight(c.withToken(ctx), &pb.GetBlockHeightRequest{Commitment: &commitment})
}

// IsBlockhashValid reports whether the base58 blockhash is still valid at
// the given commitment.
func (c *Client) IsBlockhashValid(ctx context.Context, blockhash string, commitment pb.CommitmentLevel) (*pb.IsBlockhashValidResponse, error) {
	return c.geyser.IsBlockhashValid(c.withToken(ctx), &pb.IsBlockhashValidRequest{
		Blockhash:  blockhash,
		Commitment: &commitment,
	})
}
