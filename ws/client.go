package ws

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/fluxrpc/solana-go/rpc"
)

// ErrSubscriptionClosed is returned by Recv after Unsubscribe, or when the
// connection is lost (the connection error is wrapped alongside it).
var ErrSubscriptionClosed = errors.New("subscription closed")

// DefaultSubscriptionBuffer is the per-subscription notification buffer.
const DefaultSubscriptionBuffer = 256

// Options configures a Client beyond the defaults.
type Options struct {
	// HTTPHeader is attached to the WebSocket handshake (e.g. auth).
	HTTPHeader http.Header

	// SubscriptionBuffer overrides DefaultSubscriptionBuffer.
	SubscriptionBuffer int

	// HandshakeTimeout bounds the dial+handshake when the Connect context
	// has no deadline of its own. Default 30s.
	HandshakeTimeout time.Duration
}

// Client is a Solana WebSocket subscription client. It is safe for
// concurrent use.
type Client struct {
	conn   net.Conn
	reader *wsutil.Reader
	buffer int

	reqID atomic.Uint64

	mu      sync.Mutex
	pending map[uint64]*pendingReq
	subs    map[uint64]*subEntry
	err     error
	closed  chan struct{}

	writeMu sync.Mutex
}

type subEntry struct {
	ch      chan []byte
	dropped atomic.Uint64
}

type subResponse struct {
	result json.RawMessage
	err    *rpc.RPCError
}

// pendingReq is an in-flight request. For subscribe requests, entry is the
// notification channel to register the moment the ack arrives — inside the
// read loop, before any subsequent frame is processed — so notifications
// sent immediately behind the ack can never be lost.
type pendingReq struct {
	ch    chan subResponse
	entry *subEntry
}

// Connect dials a WebSocket endpoint (ws:// or wss://).
func Connect(ctx context.Context, url string) (*Client, error) {
	return ConnectWithOptions(ctx, url, nil)
}

// ConnectWithOptions dials a WebSocket endpoint with explicit options.
func ConnectWithOptions(ctx context.Context, url string, opts *Options) (*Client, error) {
	dialer := ws.Dialer{Timeout: 30 * time.Second}
	buffer := DefaultSubscriptionBuffer
	if opts != nil {
		if opts.HTTPHeader != nil {
			dialer.Header = ws.HandshakeHeaderHTTP(opts.HTTPHeader)
		}
		if opts.SubscriptionBuffer > 0 {
			buffer = opts.SubscriptionBuffer
		}
		if opts.HandshakeTimeout > 0 {
			dialer.Timeout = opts.HandshakeTimeout
		}
	}

	conn, br, _, err := dialer.Dial(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("ws dial: %w", err)
	}

	// br carries any bytes the server sent right behind the handshake.
	var source io.Reader = conn
	if br != nil {
		source = br
	}

	c := &Client{
		conn:    conn,
		buffer:  buffer,
		pending: make(map[uint64]*pendingReq),
		subs:    make(map[uint64]*subEntry),
		closed:  make(chan struct{}),
	}
	c.reader = &wsutil.Reader{
		Source: source,
		State:  ws.StateClientSide,
		// Control frames can arrive in the middle of a fragmented message;
		// answer them inline so ReadFrom keeps flowing.
		OnIntermediate: func(hdr ws.Header, rd io.Reader) error {
			var buf [ws.MaxControlFramePayloadSize]byte
			return c.handleControlFrom(hdr, rd, buf[:])
		},
	}
	go c.readLoop()
	return c, nil
}

// Close tears down the connection; all subscriptions' Recv calls return.
func (c *Client) Close() error {
	c.writeMu.Lock()
	ws.WriteFrame(c.conn, ws.MaskFrameInPlace(ws.NewCloseFrame(ws.NewCloseFrameBody(ws.StatusNormalClosure, ""))))
	c.writeMu.Unlock()
	err := c.conn.Close()
	c.fail(ErrSubscriptionClosed)
	return err
}

// Err returns the terminal connection error, if any.
func (c *Client) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *Client) fail(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.err != nil {
		return
	}
	c.err = err
	close(c.closed)
	for id, entry := range c.subs {
		close(entry.ch)
		delete(c.subs, id)
	}
	for id, p := range c.pending {
		close(p.ch)
		delete(c.pending, id)
	}
}

// readLoop is the only reader. Message bytes accumulate into one reused
// buffer; the routed payload is copied out at its exact size.
func (c *Client) readLoop() {
	var msgBuf bytes.Buffer
	var ctrlBuf [ws.MaxControlFramePayloadSize]byte

	for {
		hdr, err := c.reader.NextFrame()
		if err != nil {
			c.fail(fmt.Errorf("%w: %w", ErrSubscriptionClosed, err))
			c.conn.Close()
			return
		}

		if hdr.OpCode.IsControl() {
			if err := c.handleControlFrom(hdr, c.reader, ctrlBuf[:]); err != nil {
				c.fail(fmt.Errorf("%w: %w", ErrSubscriptionClosed, err))
				c.conn.Close()
				return
			}
			continue
		}

		msgBuf.Reset()
		if hdr.Length > 0 {
			msgBuf.Grow(int(hdr.Length))
		}
		if _, err := msgBuf.ReadFrom(c.reader); err != nil {
			c.fail(fmt.Errorf("%w: %w", ErrSubscriptionClosed, err))
			c.conn.Close()
			return
		}
		c.route(msgBuf.Bytes())
	}
}

// handleControlFrom answers pings and surfaces close frames. Pong writes
// share the request write lock, so they can never interleave with a request
// frame.
func (c *Client) handleControlFrom(hdr ws.Header, rd io.Reader, buf []byte) error {
	payload := buf[:hdr.Length]
	if _, err := io.ReadFull(rd, payload); err != nil && err != io.EOF {
		return err
	}
	switch hdr.OpCode {
	case ws.OpPing:
		frame := ws.MaskFrameInPlace(ws.NewPongFrame(payload))
		c.writeMu.Lock()
		err := ws.WriteFrame(c.conn, frame)
		c.writeMu.Unlock()
		return err
	case ws.OpClose:
		status, reason := ws.ParseCloseFrameData(payload)
		return fmt.Errorf("connection closed by server: %d %s", status, reason)
	}
	return nil // pong: ignore
}

// route dispatches one message. data is only valid during the call.
func (c *Client) route(data []byte) {
	var env struct {
		ID     uint64          `json:"id"`
		Result json.RawMessage `json:"result"`
		Error  *rpc.RPCError   `json:"error"`
		Method string          `json:"method"`
		Params struct {
			Subscription uint64          `json:"subscription"`
			Result       json.RawMessage `json:"result"`
		} `json:"params"`
	}
	if err := sonic.Unmarshal(data, &env); err != nil {
		return // tolerate unknown frames
	}

	if env.Method != "" {
		// The non-blocking send happens under mu so a concurrent
		// Unsubscribe/fail can never close the channel mid-send.
		c.mu.Lock()
		if entry := c.subs[env.Params.Subscription]; entry != nil {
			select {
			case entry.ch <- bytes.Clone(env.Params.Result):
			default:
				entry.dropped.Add(1)
			}
		}
		c.mu.Unlock()
		return
	}

	c.mu.Lock()
	p := c.pending[env.ID]
	delete(c.pending, env.ID)
	if p != nil && p.entry != nil && env.Error == nil {
		var subID uint64
		if sonic.Unmarshal(env.Result, &subID) == nil {
			c.subs[subID] = p.entry
		}
	}
	c.mu.Unlock()
	if p != nil {
		p.ch <- subResponse{result: bytes.Clone(env.Result), err: env.Error}
	}
}

// request sends a JSON-RPC request over the socket and waits for its reply.
// A non-nil entry is registered as a subscription by the read loop as soon
// as the ack arrives.
func (c *Client) request(ctx context.Context, method string, params []any, entry *subEntry) (json.RawMessage, error) {
	id := c.reqID.Add(1)
	payload, err := sonic.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan subResponse, 1)
	c.mu.Lock()
	if c.err != nil {
		err := c.err
		c.mu.Unlock()
		return nil, err
	}
	c.pending[id] = &pendingReq{ch: ch, entry: entry}
	c.mu.Unlock()

	c.writeMu.Lock()
	err = wsutil.WriteClientText(c.conn, payload)
	c.writeMu.Unlock()
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("%s: %w", method, err)
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			return nil, c.Err()
		}
		if resp.err != nil {
			return nil, fmt.Errorf("%s: %w", method, resp.err)
		}
		return resp.result, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case <-c.closed:
		return nil, c.Err()
	}
}

// Subscription is a live subscription delivering notifications of type T.
type Subscription[T any] struct {
	client      *Client
	id          uint64
	unsubMethod string
	entry       *subEntry
}

// subscribe issues the subscribe request; the read loop registers the
// notification channel the instant the ack arrives, so no notification
// delivered after the ack is ever missed.
func subscribe[T any](ctx context.Context, c *Client, method, unsubMethod string, params []any) (*Subscription[T], error) {
	entry := &subEntry{ch: make(chan []byte, c.buffer)}
	result, err := c.request(ctx, method, params, entry)
	if err != nil {
		return nil, err
	}
	var subID uint64
	if err := sonic.Unmarshal(result, &subID); err != nil {
		return nil, fmt.Errorf("%s: unexpected subscription id %s", method, result)
	}
	return &Subscription[T]{client: c, id: subID, unsubMethod: unsubMethod, entry: entry}, nil
}

// Recv returns the next notification, decoded on the caller's goroutine.
func (s *Subscription[T]) Recv(ctx context.Context) (*T, error) {
	select {
	case raw, ok := <-s.entry.ch:
		if !ok {
			if err := s.client.Err(); err != nil {
				return nil, err
			}
			return nil, ErrSubscriptionClosed
		}
		out := new(T)
		if err := sonic.Unmarshal(raw, out); err != nil {
			return nil, fmt.Errorf("decoding notification: %w", err)
		}
		return out, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Dropped reports how many notifications were discarded because the
// subscription buffer was full.
func (s *Subscription[T]) Dropped() uint64 {
	return s.entry.dropped.Load()
}

// Unsubscribe cancels the subscription server-side and closes its channel.
func (s *Subscription[T]) Unsubscribe(ctx context.Context) error {
	s.Release()
	_, err := s.client.request(ctx, s.unsubMethod, []any{s.id}, nil)
	return err
}

// Release drops the subscription's client-side resources without
// contacting the server. Use it instead of Unsubscribe when the server has
// already removed the subscription — signatureSubscribe is single-shot and
// auto-cancels after its final notification — so the unsubscribe round
// trip would be wasted. Safe to call more than once.
func (s *Subscription[T]) Release() {
	s.client.mu.Lock()
	// Identity check: the server may have reassigned this subscription id
	// to a newer subscription; never tear that one down.
	if entry := s.client.subs[s.id]; entry == s.entry {
		delete(s.client.subs, s.id)
		close(entry.ch)
	}
	s.client.mu.Unlock()
}
